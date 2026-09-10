// Package egress constrains outbound connections at DNS resolution and dial time.
package egress

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var ErrDenied = errors.New("destino externo não autorizado")

const maxPooledOrigins = 256

type Policy struct {
	// Private maps an exact host:port to approved CIDRs; never a blanket bypass.
	Private     map[string][]netip.Prefix
	HTTPOrigins map[string]bool
	Resolver    *net.Resolver
}

func FromEnv() Policy {
	p := Policy{Private: map[string][]netip.Prefix{}, HTTPOrigins: map[string]bool{}}
	for _, rule := range strings.Split(os.Getenv("EGRESS_PRIVATE_RULES"), ";") {
		parts := strings.SplitN(rule, "=", 2)
		if len(parts) != 2 {
			continue
		}
		for _, cidr := range strings.Split(parts[1], ",") {
			prefix, err := netip.ParsePrefix(cidr)
			if err == nil {
				p.Private[strings.ToLower(parts[0])] = append(p.Private[strings.ToLower(parts[0])], prefix)
			}
		}
	}
	if os.Getenv("ENVIRONMENT") == "local" {
		for _, origin := range strings.Split(os.Getenv("EGRESS_HTTP_ORIGINS"), ",") {
			if origin != "" {
				p.HTTPOrigins[origin] = true
			}
		}
	}
	return p
}

func address(u *url.URL) string {
	port := u.Port()
	if port == "" {
		port = "443"
		if u.Scheme == "http" {
			port = "80"
		}
	}
	return net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

func (p Policy) Validate(target string) error {
	u, err := url.Parse(target)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" || strings.Contains(u.Host, "%") {
		return ErrDenied
	}
	if u.Scheme != "https" && (u.Scheme != "http" || !p.HTTPOrigins[u.Scheme+"://"+u.Host]) {
		return ErrDenied
	}
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !p.allowed(address(u), ip) {
		return ErrDenied
	}
	return nil
}

func ValidateURL(target string) error { return FromEnv().Validate(target) }

var blocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("64:ff9b::/96"),
}

func (p Policy) allowed(hostport string, ip netip.Addr) bool {
	ip = ip.Unmap()
	// Link-local and metadata addresses cannot be whitelisted.
	if !ip.IsValid() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	for _, prefix := range p.Private[strings.ToLower(hostport)] {
		if prefix.Contains(ip) {
			return true
		}
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() {
		return false
	}
	for _, prefix := range blocked {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func (p Policy) Client(target string, timeout time.Duration) (*http.Client, error) {
	if err := p.Validate(target); err != nil {
		return nil, err
	}
	u, _ := url.Parse(target)
	origin := u.Scheme + "://" + u.Host
	return &http.Client{Timeout: timeout, Transport: originTransport{origin: origin, transport: p.transport(u, timeout)}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, nil
}

// Pool reutiliza transports por origem, mantendo os limites de conexões do
// próprio http.Transport. O cliente retornado é novo a cada chamada para que
// timeout e certificados mTLS sejam propriedade da operação; o transporte,
// que é seguro para uso concorrente, permanece compartilhado. Origens acima
// do limite usam um transporte efêmero e não fazem o pool crescer sem limite.
type Pool struct {
	policy     Policy
	mu         sync.Mutex
	transports map[string]*http.Transport
}

func NewPool(policy Policy) *Pool {
	return &Pool{policy: policy, transports: make(map[string]*http.Transport)}
}

func (p *Pool) Client(target string, timeout time.Duration) (*http.Client, error) {
	if p == nil {
		return nil, ErrDenied
	}
	if err := p.policy.Validate(target); err != nil {
		return nil, err
	}
	u, _ := url.Parse(target)
	origin := u.Scheme + "://" + u.Host
	p.mu.Lock()
	transport := p.transports[origin]
	if transport == nil {
		transport = p.policy.transport(u, 0)
		if len(p.transports) < maxPooledOrigins {
			p.transports[origin] = transport
		}
	}
	p.mu.Unlock()
	return &http.Client{Timeout: timeout, Transport: originTransport{origin: origin, transport: transport}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, nil
}

// CloseIdleConnections libera conexões mantidas pelos transports pooled.
func (p *Pool) CloseIdleConnections() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, transport := range p.transports {
		transport.CloseIdleConnections()
	}
}

func (p Policy) transport(u *url.URL, responseHeaderTimeout time.Duration) *http.Transport {
	resolver := p.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	dialer := net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{Proxy: nil, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, MaxIdleConns: 64, MaxIdleConnsPerHost: 8, MaxConnsPerHost: 16, IdleConnTimeout: 30 * time.Second, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: responseHeaderTimeout}
	transport.DialContext = func(ctx context.Context, network, hostport string) (net.Conn, error) {
		if !strings.EqualFold(hostport, address(u)) {
			return nil, ErrDenied
		}
		host, port, err := net.SplitHostPort(hostport)
		if err != nil {
			return nil, ErrDenied
		}
		ips, err := resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, ErrDenied
		}
		for _, ip := range ips {
			if !p.allowed(hostport, ip) {
				return nil, ErrDenied
			}
		}
		// Dial the validated address, never resolve the hostname a second time.
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}
	return transport
}

type originTransport struct {
	origin    string
	transport *http.Transport
}

func (t originTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme+"://"+r.URL.Host != t.origin {
		return nil, ErrDenied
	}
	return t.transport.RoundTrip(r)
}

// WithCertificate creates a private mTLS transport without mutating shared pools.
func WithCertificate(client *http.Client, certificate tls.Certificate, roots *tls.Config) error {
	t, ok := client.Transport.(originTransport)
	if !ok {
		return ErrDenied
	}
	clone := t.transport.Clone()
	clone.TLSClientConfig = clone.TLSClientConfig.Clone()
	clone.TLSClientConfig.Certificates = []tls.Certificate{certificate}
	if roots != nil {
		clone.TLSClientConfig.RootCAs = roots.RootCAs
	}
	client.Transport = originTransport{origin: t.origin, transport: clone}
	return nil
}

func NewClient(target string, timeout time.Duration) (*http.Client, error) {
	return FromEnv().Client(target, timeout)
}
