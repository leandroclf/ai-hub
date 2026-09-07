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
