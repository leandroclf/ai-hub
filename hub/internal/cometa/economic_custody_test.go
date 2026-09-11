package cometa

import "testing"

func TestEconomicKindForSource(t *testing.T) {
	if got := economicKindForSource("SUBMIT_ACCEPTED"); got != "SUBMITTED" {
		t.Fatalf("accepted submit incidence = %q", got)
	}
	for _, source := range []string{"SUBMIT", "PROVIDER", "CALLBACK", "POLL", "ADMIN_RECONCILIATION"} {
		if got := economicKindForSource(source); got != "" {
			t.Fatalf("source %s unexpectedly overrode operational kind with %q", source, got)
		}
	}
}
