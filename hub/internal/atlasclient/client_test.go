package atlasclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
)

func TestOfferUsesValidProjectionWhenAtlasIsUnavailableAndRejectsExpiredProjection(t *testing.T) {
	snapshot := atlas.OfferSnapshot{
		Hash:       "projection-hash",
		ValidUntil: time.Now().Add(150 * time.Millisecond),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snapshot)
	}))
	client := New(server.URL, time.Hour)
	client.http = server.Client()

	first, err := client.Offer(t.Context(), "tenant", "application", "service", 1, "account")
	if err != nil || first.Hash != snapshot.Hash {
		t.Fatalf("projeção inicial recusada: snapshot=%+v err=%v", first, err)
	}
	server.Close()

	fromCache, err := client.Offer(t.Context(), "tenant", "application", "service", 1, "account")
	if err != nil || fromCache.Hash != snapshot.Hash {
		t.Fatalf("projeção válida não foi reutilizada com Atlas indisponível: snapshot=%+v err=%v", fromCache, err)
	}

	time.Sleep(200 * time.Millisecond)
	if _, err = client.Offer(t.Context(), "tenant", "application", "service", 1, "account"); err == nil {
		t.Fatal("projeção vencida foi aceita após indisponibilidade do Atlas")
	}
}
