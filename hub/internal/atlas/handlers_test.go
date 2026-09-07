package atlas

import "testing"

func TestValidateService(t *testing.T) {
	tests := []struct {
		name string
		svc  Service
		want bool
	}{
		{"valido", Service{Code: "consulta", Version: 1, Description: "API publica", Modes: []string{"SYNC", "ASYNC"}, ClientSLASeconds: 30}, true},
		{"sem codigo", Service{Version: 1, Description: "API", Modes: []string{"SYNC"}, ClientSLASeconds: 30}, false},
		{"modo invalido", Service{Code: "consulta", Version: 1, Description: "API", Modes: []string{"SOAP"}, ClientSLASeconds: 30}, false},
		{"sla invalido", Service{Code: "consulta", Version: 1, Description: "API", Modes: []string{"SYNC"}}, false},
		{"modo duplicado", Service{Code: "consulta", Version: 1, Description: "API", Modes: []string{"SYNC", "SYNC"}, ClientSLASeconds: 30}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateService(tt.svc)
			if (err == nil) != tt.want {
				t.Fatalf("validateService() error=%v, want valid=%v", err, tt.want)
			}
		})
	}
}

func TestValidateProviderAccount(t *testing.T) {
	valid := []ProviderAccount{
		{ProviderAccountID: "basic", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "BASIC", AuthUsername: "hub", AuthSecretRef: "vault://basic"},
		{ProviderAccountID: "oauth", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "OAUTH_CLIENT_CREDENTIALS", OAuthTokenURL: "http://provider:8090/oauth/token", OAuthClientID: "client", OAuthClientSecretRef: "vault://oauth"},
		{ProviderAccountID: "mtls", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "MTLS_OAUTH", OAuthTokenURL: "http://provider:8090/oauth/token", OAuthClientID: "client", OAuthClientSecretRef: "vault://oauth", MTLSCertificateRef: "vault://cert"},
	}
	for _, account := range valid {
		if err := validateProviderAccount(&account); err != nil {
			t.Errorf("conta valida rejeitada: %s: %v", account.ProviderAccountID, err)
		}
	}
	invalid := []ProviderAccount{
		{ProviderAccountID: "missing-basic", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "BASIC"},
		{ProviderAccountID: "missing-mtls", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "MTLS_OAUTH", OAuthTokenURL: "http://provider:8090/oauth/token", OAuthClientID: "client", OAuthClientSecretRef: "vault://oauth"},
		{ProviderAccountID: "unknown", ProviderID: "p", BaseURL: "http://provider:8090", AuthType: "KERBEROS"},
	}
	for _, account := range invalid {
		if err := validateProviderAccount(&account); err == nil {
			t.Errorf("conta invalida aceita: %s", account.ProviderAccountID)
		}
	}
}
