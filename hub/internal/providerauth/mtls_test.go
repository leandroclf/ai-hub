package providerauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/egress"
)

type mtlsFixtureResolver struct {
	certificate string
}

func (r mtlsFixtureResolver) Resolve(_ context.Context, ref, version string) (Secret, error) {
	if version != "v1" {
		return Secret{}, errors.New("secret unavailable")
	}
	switch ref {
	case "vault://mtls":
		return Secret{Value: r.certificate, Version: version}, nil
	case "vault://mtls-secret":
		return Secret{Value: "client-secret", Version: version}, nil
	default:
		return Secret{}, errors.New("secret unavailable")
	}
}

func issueMTLSCertificate(t *testing.T, commonName string, isCA bool, usages []x509.ExtKeyUsage, parent *x509.Certificate, parentKey *rsa.PrivateKey, ip net.IP) (tls.Certificate, []byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           usages,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
	}
	if ip != nil {
		template.IPAddresses = []net.IP{ip}
	}
	if parent == nil {
		parent = template
		parentKey = key
	}
	der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
	if err != nil {
		t.Fatal(err)
	}
	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, certificatePEM, privateKeyPEM
}

func TestMTLSOAuthUsesClientCertificateAndPinnedCA(t *testing.T) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "AI Hub mTLS test CA"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	_, serverCertPEM, serverKeyPEM := issueMTLSCertificate(t, "localhost", false, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, ca, caKey, net.ParseIP("127.0.0.1"))
	_, clientCertPEM, clientKeyPEM := issueMTLSCertificate(t, "hub-cometa", false, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, ca, caKey, nil)
	serverTLS, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	clientRoots := x509.NewCertPool()
	clientRoots.AddCert(ca)

	var authenticatedRequests atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) != 1 || r.TLS.PeerCertificates[0].Subject.CommonName != "hub-cometa" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticatedRequests.Add(1)
		if r.URL.Path == "/token" {
			if err := r.ParseForm(); err != nil || r.Form.Get("client_secret") != "client-secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"mtls-token","expires_in":30}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer mtls-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{serverTLS}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: clientRoots}
	server.StartTLS()
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENVIRONMENT", "local")
	t.Setenv("EGRESS_HTTP_ORIGINS", server.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", u.Host+"=127.0.0.1/32")

	certificateSecret, err := json.Marshal(map[string]string{
		"certificate": string(clientCertPEM),
		"private_key": string(clientKeyPEM),
		"root_ca":     string(caPEM),
	})
	if err != nil {
		t.Fatal(err)
	}
	cache := NewTokenCache("127.0.0.1:1")
	defer cache.Close()
	cache.Resolver = mtlsFixtureResolver{certificate: string(certificateSecret)}
	c := egress.NewPool(egress.FromEnv())
	client, err := c.Client(server.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, server.URL+"/resource", nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{AuthType: MTLSOAuth, TokenURL: server.URL + "/token", ClientID: "fixture-client", ClientSecretRef: "vault://mtls-secret", MTLSCertificateRef: "vault://mtls", SecretVersion: "v1", BindingID: "binding-mtls", TenantID: "tenant-mtls", Environment: "local"}
	if err := cache.Apply(context.Background(), client, "account-mtls", cfg, req); err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent || authenticatedRequests.Load() != 2 {
		t.Fatalf("mTLS requests status=%d authenticated_requests=%d", response.StatusCode, authenticatedRequests.Load())
	}

	cfg.MTLSCertificateRef = "vault://missing"
	invalidRequest, _ := http.NewRequest(http.MethodGet, server.URL+"/resource", nil)
	if err := cache.Apply(context.Background(), client, "account-mtls", cfg, invalidRequest); err == nil {
		t.Fatal("certificado ausente foi aceito")
	}
	t.Log("HTTPS real exigiu certificado cliente assinado pela CA do vínculo; token e recurso autenticaram; referência ausente foi recusada sem fallback")
}
