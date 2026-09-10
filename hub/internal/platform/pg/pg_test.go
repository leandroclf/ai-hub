package pg

import "testing"

func TestRuntimeDSNFixesTenantSessionSetting(t *testing.T) {
	got, err := RuntimeDSN("postgres://runtime:secret@db.example/hub_core?sslmode=require", "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" || !contains(got, "app.tenant_id%3Dtenant-a") {
		t.Fatalf("tenant session setting missing from DSN: %s", got)
	}
	if _, err = RuntimeDSN("postgres://runtime/db", "bad\nvalue"); err == nil {
		t.Fatal("invalid tenant accepted")
	}
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
