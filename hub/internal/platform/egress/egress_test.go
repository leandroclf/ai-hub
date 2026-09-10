package egress

import (
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
