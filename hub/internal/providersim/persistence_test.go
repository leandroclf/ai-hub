package providersim

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerStatePersistsDeduplicationAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "provider.json")
	first := NewServerWithState(path)
	first.mu.Lock()
	first.seq = 7
	first.effects = 1
	first.protocols["protocol-1"] = "prov-req-000007"
	first.ops["prov-req-000007"] = &operation{result: OperationResult{ProviderRequestID: "prov-req-000007", Status: "SUCCEEDED"}, ready: true}
	first.saveStateLocked()
	first.mu.Unlock()

	second := NewServerWithState(path)
	second.mu.Lock()
	defer second.mu.Unlock()
	if second.protocols["protocol-1"] != "prov-req-000007" || second.effects != 1 || second.seq != 7 {
		t.Fatalf("state lost after restart: protocols=%v effects=%d seq=%d", second.protocols, second.effects, second.seq)
	}
	if second.ops["prov-req-000007"].result.Status != "SUCCEEDED" {
		t.Fatalf("operation result lost after restart")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state file not created: %v", err)
	}
}
