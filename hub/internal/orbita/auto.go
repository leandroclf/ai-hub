package orbita

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ai-hub/hub/internal/atlas"
)

// AwaitAuto observes only the authoritative local protocol. Waiting does not
// dispatch, poll a provider, renew TTL or cancel the durable execution intent.
func (s *Store) AwaitAuto(ctx context.Context, p Protocol, seconds int) (Protocol, error) {
	if IsTerminal(p.Status) {
		return p, nil
	}
	if seconds <= 0 {
		return Protocol{}, errors.New("invalid frozen AUTO wait")
	}
	until := p.AcceptedAt.Add(time.Duration(seconds) * time.Second)
	if p.ClientDeadlineAt.Before(until) {
		until = p.ClientDeadlineAt
	}
	for {
		current, err := s.Get(ctx, p.TenantID, p.ProtocolID)
		if err != nil {
			return Protocol{}, err
		}
		if IsTerminal(current.Status) || !time.Now().Before(until) {
			return current, nil
		}
		delay := min(50*time.Millisecond, time.Until(until))
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Protocol{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func (h *Handlers) respondAuto(w http.ResponseWriter, r *http.Request, p Protocol, seconds int) {
	current, err := h.store.AwaitAuto(r.Context(), p, seconds)
	if err != nil {
		acceptedError(w, p.ProtocolID, 503, "auto_observation_unavailable")
		return
	}
	status := http.StatusAccepted
	if IsTerminal(current.Status) {
		status = statusForTerminal(current.Status)
	}
	h.respondWithProtocol(w, current, status)
}

func (h *Handlers) respondExistingAdmission(w http.ResponseWriter, r *http.Request, p Protocol) {
	if p.Mode != "AUTO" || IsTerminal(p.Status) {
		h.respondWithProtocol(w, p, http.StatusOK)
		return
	}
	var raw []byte
	err := h.store.db.QueryRowContext(r.Context(), `SELECT config_snapshot FROM protocols WHERE protocol_id=$1 AND tenant_id=$2 AND application_id=$3`, p.ProtocolID, p.TenantID, p.ApplicationID).Scan(&raw)
	var snapshot atlas.OfferSnapshot
	if err != nil || json.Unmarshal(raw, &snapshot) != nil {
		acceptedError(w, p.ProtocolID, 503, "snapshot_unavailable")
		return
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil {
		acceptedError(w, p.ProtocolID, 503, "snapshot_unavailable")
		return
	}
	h.respondAuto(w, r, p, target.AutoWaitSeconds)
}
