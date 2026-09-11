package egress

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"testing"
	"time"
)

func TestUnsafeAddressesAndSchemes(t *testing.T) {
	p := Policy{}
	for _, target := range []string{"http://example.com", "https://user:pass@example.com", "https://169.254.169.254/latest/meta-data", "https://127.0.0.1", "https://[::ffff:127.0.0.1]", "https://10.1.1.1", "https://100.100.100.200", "file:///etc/passwd"} {
		if p.Validate(target) == nil {
			t.Errorf("unsafe destination allowed: %s", target)
		}
	}
	if p.Validate("https://example.com/api") != nil {
		t.Fatal("public TLS origin rejected")
	}
}

func TestPrivateScopeAndRedirectDoNotLeak(t *testing.T) {
	leaked := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked = true }))
	defer target.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	p := Policy{HTTPOrigins: map[string]bool{srv.URL: true}, Private: map[string][]netip.Prefix{u.Host: {netip.MustParsePrefix("127.0.0.1/32")}}}
	c, err := p.Client(srv.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.Header.Set("Authorization", "Bearer fixture")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 307 || leaked {
		t.Fatal("redirect followed or credential leaked")
	}
	if _, err := c.Get(target.URL); err == nil {
		t.Fatal("different origin accepted")
	}
	if p.allowed("other-host:80", netip.MustParseAddr("127.0.0.1")) {
		t.Fatal("private profile escaped host/port")
	}
}

func TestPoolReusesTransportPerOriginAndKeepsClientsIndependent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	p := Policy{HTTPOrigins: map[string]bool{srv.URL: true}, Private: map[string][]netip.Prefix{u.Host: {netip.MustParsePrefix("127.0.0.1/32")}}}
	pool := NewPool(p)
	first, err := pool.Client(srv.URL+"/one", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	second, err := pool.Client(srv.URL+"/two", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	a, ok := first.Transport.(originTransport)
	if !ok {
		t.Fatal("first client is not origin constrained")
	}
	b, ok := second.Transport.(originTransport)
	if !ok {
		t.Fatal("second client is not origin constrained")
	}
	if a.transport != b.transport {
		t.Fatal("clients from one origin did not share the transport pool")
	}
	if first == second || first.Timeout == second.Timeout {
		t.Fatal("per-call clients lost independent timeout ownership")
	}
	pool.CloseIdleConnections()
}

func TestR2Seg04Scenarios(t *testing.T) {
	t.Run("R2-SEG-04-S01_destino_homologado", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/health" {
				t.Fatalf("path inesperado: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		policy := Policy{Private: map[string][]netip.Prefix{u.Host: {netip.MustParsePrefix("127.0.0.1/32")}}}
		if err := policy.Validate(server.URL + "/health"); err != nil {
			t.Fatalf("HTTPS homologado recusado: %v", err)
		}
		client, err := policy.Client(server.URL, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		certificate := x509.NewCertPool()
		certificate.AddCert(server.Certificate())
		if err := WithCertificate(client, tls.Certificate{}, &tls.Config{RootCAs: certificate}); err != nil {
			t.Fatal(err)
		}
		response, err := client.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("HTTPS homologado não conectou: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("status HTTPS inesperado: %d", response.StatusCode)
		}
		transport, ok := client.Transport.(originTransport)
		if !ok || transport.origin != "https://"+u.Host {
			t.Fatalf("transporte não preservou origem homologada: %#v", client.Transport)
		}
	})

	t.Run("R2-SEG-04-S02_ssrf_e_rebinding", func(t *testing.T) {
		for _, target := range []string{
			"https://169.254.169.254/latest/meta-data",
			"https://127.0.0.1/internal",
			"https://user:password@example.com",
			"file:///etc/passwd",
		} {
			if (Policy{}).Validate(target) == nil {
				t.Fatalf("destino inseguro aceito: %s", target)
			}
		}
		leaked := false
		destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked = true }))
		defer destination.Close()
		redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
		}))
		defer redirector.Close()
		u, _ := url.Parse(redirector.URL)
		policy := Policy{HTTPOrigins: map[string]bool{redirector.URL: true}, Private: map[string][]netip.Prefix{u.Host: {netip.MustParsePrefix("127.0.0.1/32")}}}
		client, err := policy.Client(redirector.URL, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		request, _ := http.NewRequest(http.MethodGet, redirector.URL, nil)
		request.Header.Set("Authorization", "Bearer fixture")
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusTemporaryRedirect || leaked {
			t.Fatal("redirect foi seguido ou credencial vazou para outra origem")
		}
	})

	t.Run("R2-SEG-04-S03_rede_privada_legitima", func(t *testing.T) {
		allowedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
		defer allowedServer.Close()
		otherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
		defer otherServer.Close()
		u, err := url.Parse(allowedServer.URL)
		if err != nil {
			t.Fatal(err)
		}
		policy := Policy{HTTPOrigins: map[string]bool{allowedServer.URL: true}, Private: map[string][]netip.Prefix{u.Host: {netip.MustParsePrefix("127.0.0.1/32")}}}
		client, err := policy.Client(allowedServer.URL, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Get(allowedServer.URL)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("destino privado homologado falhou: %d", response.StatusCode)
		}
		if _, err := policy.Client(otherServer.URL, time.Second); err == nil {
			t.Fatal("whitelist privada escapou para outro host/porta")
		}
	})
}
