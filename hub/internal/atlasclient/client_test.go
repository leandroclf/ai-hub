package atlasclient

import (
	"encoding/json"
	"fmt"
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

// R5-SEG-02-S01: uma negação autoritativa (403/409) nunca pode ser
// convertida em autorização pelo cache, mesmo após a janela de validade
// expirar e uma nova resolução ser exigida.
func TestOfferDenialIsNeverMaskedByCache(t *testing.T) {
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(atlas.OfferSnapshot{Hash: "projection-hash", ValidUntil: time.Now().Add(50 * time.Millisecond)})
	}))
	defer server.Close()
	client := New(server.URL, time.Hour)
	client.http = server.Client()

	if _, err := client.Offer(t.Context(), "tenant", "application", "service", 1, "account"); err != nil {
		t.Fatalf("projeção inicial recusada: %v", err)
	}

	status = http.StatusForbidden
	time.Sleep(60 * time.Millisecond)
	if _, err := client.Offer(t.Context(), "tenant", "application", "service", 1, "account"); err == nil {
		t.Fatal("negação 403 do Atlas foi mascarada pelo cache expirado")
	}

	status = http.StatusOK
	if _, err := client.OfferAuthoritative(t.Context(), "tenant", "application", "service", 1, "account"); err != nil {
		t.Fatalf("resolução autoritativa não recuperou após denial transitória sanada: %v", err)
	}

	status = http.StatusConflict
	if _, err := client.OfferAuthoritative(t.Context(), "tenant", "application", "service", 1, "account"); err == nil {
		t.Fatal("conflito 409 do Atlas foi mascarado por OfferAuthoritative")
	}
	if _, ok := client.getCached(fmt.Sprintf("offer:%s:%s:%s:%d:%s", "tenant", "application", "service", 1, "account")); ok {
		t.Fatal("cache reteve projeção após conflito 409 autoritativo")
	}
}

// R5-SEG-02: o cache local não pode crescer sem limite por combinação de
// tenant/application/service/account.
func TestOfferCacheIsBounded(t *testing.T) {
	client := New("http://unused.invalid", time.Hour)
	for i := 0; i < maxCacheEntries+50; i++ {
		client.setCached(fmt.Sprintf("offer:key:%d", i), atlas.OfferSnapshot{Hash: "h", ValidUntil: time.Now().Add(time.Hour)})
	}
	if len(client.cache) > maxCacheEntries {
		t.Fatalf("cache cresceu sem limite: %d entradas", len(client.cache))
	}
}
