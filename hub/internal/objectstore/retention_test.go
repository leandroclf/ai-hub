package objectstore

import (
	"reflect"
	"testing"
)

func TestTenantsFromEnvRequiresExplicitUniqueTenants(t *testing.T) {
	got := TenantsFromEnv(" tenant-a,tenant-b, tenant-a ,,tenant-c ")
	want := []string{"tenant-a", "tenant-b", "tenant-c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tenants=%v, want %v", got, want)
	}
	if got := TenantsFromEnv(""); len(got) != 0 {
		t.Fatalf("empty configuration returned %v", got)
	}
}
