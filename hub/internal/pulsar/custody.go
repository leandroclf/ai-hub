package pulsar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"ai-hub/hub/internal/platform/idgen"
)

type ClaimedDelivery struct {
	ID, ProtocolID, EventID, TenantID, CellID, URL, DestinationID, SecretRef, SecretVersion, Hash, Owner, AttemptID string
	Body                                                                                                            []byte
	Epoch                                                                                                           int64
	Attempt, DestinationVersion, MaxAttempts, TimeoutSeconds                                                        int
}

func (s *Store) ConserveFinal(ctx context.Context, envelopeID string, f protocolFact) error {
	if len(f.Representation) == 0 || len(f.Representation) > 512*1024 || f.CellID == "" || f.TenantID == "" || f.ApplicationID == "" || f.EventID == "" || len(f.WebhookDestinations) > 100 {
		return errors.New("invalid final representation")
	}
	sum := sha256.Sum256(f.Representation)
	hash := hex.EncodeToString(sum[:])
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "INSERT INTO inbox(consumer,event_id) VALUES('pulsar',$1) ON CONFLICT DO NOTHING", envelopeID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return tx.Commit()
	}
	seen := make(map[string]struct{}, len(f.WebhookDestinations))
	for _, d := range f.WebhookDestinations {
		if d.ID == "" || d.Version < 1 || d.URL == "" || d.SecretRef == "" || d.SecretVersion == "" || d.MaxAttempts < 1 || d.MaxAttempts > 20 || d.TimeoutSecond < 1 || d.TimeoutSecond > 15 || (d.ApplicationID != "" && d.ApplicationID != f.ApplicationID) {
			return errors.New("invalid webhook destination snapshot")
		}
		key := d.ID + ":" + fmt.Sprint(d.Version)
		if _, ok := seen[key]; ok {
			return errors.New("duplicate webhook destination snapshot")
		}
		seen[key] = struct{}{}
		_, err = tx.ExecContext(ctx, `INSERT INTO deliveries(delivery_id,protocol_id,event_id,destination_url,state,next_attempt_at,tenant_id,cell_id,destination_id,destination_version,representation,body_sha256) VALUES($1,$2,$3,$4,'PENDING',clock_timestamp(),$5,$6,$7,$8,$9,$10) ON CONFLICT(event_id,destination_id,destination_version) DO NOTHING`, idgen.New(), f.ProtocolID, f.EventID, d.URL, f.TenantID, f.CellID, d.ID, d.Version, f.Representation, hash)
		if err != nil {
			return err
		}
	}
	// Even when there is no configured destination, the local disposition is
	// durable. Retrying that same event cannot create a new historical delivery.
	return tx.Commit()
}

func (s *Store) ClaimDelivery(ctx context.Context, cell, owner string) (ClaimedDelivery, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ClaimedDelivery{}, err
	}
	defer tx.Rollback()
	var d ClaimedDelivery
	d.Owner = owner
	d.AttemptID = idgen.New()
	err = tx.QueryRowContext(ctx, `SELECT d.delivery_id,d.protocol_id,d.event_id,d.tenant_id,d.cell_id,d.destination_url,d.destination_id,d.destination_version,d.representation,d.body_sha256,d.epoch,d.attempts_count,v.secret_ref,v.secret_version,v.max_attempts,v.timeout_seconds
 FROM deliveries d JOIN webhook_destination_versions v ON v.id=d.destination_id AND v.version=d.destination_version
 WHERE d.cell_id=$1 AND d.state IN ('PENDING','RETRY_SCHEDULED') AND d.next_attempt_at<=clock_timestamp() AND (d.lease_until IS NULL OR d.lease_until<=clock_timestamp())
	ORDER BY d.next_attempt_at,d.delivery_id FOR UPDATE OF d SKIP LOCKED LIMIT 1`, cell).Scan(&d.ID, &d.ProtocolID, &d.EventID, &d.TenantID, &d.CellID, &d.URL, &d.DestinationID, &d.DestinationVersion, &d.Body, &d.Hash, &d.Epoch, &d.Attempt, &d.SecretRef, &d.SecretVersion, &d.MaxAttempts, &d.TimeoutSeconds)
	if err != nil {
		return ClaimedDelivery{}, err
	}
	d.Epoch++
	d.Attempt++
	_, err = tx.ExecContext(ctx, `UPDATE deliveries SET lease_owner=$2,lease_until=clock_timestamp()+interval '30 seconds',epoch=$3,attempts_count=$4 WHERE delivery_id=$1`, d.ID, owner, d.Epoch, d.Attempt)
	if err != nil {
		return ClaimedDelivery{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO webhook_attempts(id,delivery_id,owner,epoch,state,body_sha256) VALUES($1,$2,$3,$4,'PREPARED',$5)`, d.AttemptID, d.ID, owner, d.Epoch, d.Hash)
	if err != nil {
		return ClaimedDelivery{}, err
	}
	if err = tx.Commit(); err != nil {
		return ClaimedDelivery{}, err
	}
	return d, nil
}

func (s *Store) CompleteDelivery(ctx context.Context, d ClaimedDelivery, status int, code string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	state, attemptState := "RETRY_SCHEDULED", "FAILED"
	if status >= 200 && status < 300 && code == "" {
		state = "DELIVERED"
		attemptState = "ACKNOWLEDGED"
	} else if d.Attempt >= d.MaxAttempts {
		state = "EXHAUSTED"
	}
	res, err := tx.ExecContext(ctx, `UPDATE deliveries SET state=$4,lease_owner=NULL,lease_until=NULL,next_attempt_at=clock_timestamp()+($5*interval '1 second'),updated_at=clock_timestamp() WHERE delivery_id=$1 AND lease_owner=$2 AND epoch=$3 AND lease_until>clock_timestamp()`, d.ID, d.Owner, d.Epoch, state, d.Attempt*d.Attempt)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("delivery lease lost")
	}
	_, err = tx.ExecContext(ctx, "UPDATE webhook_attempts SET state=$2,received_at=clock_timestamp(),http_status=$3,error_code=NULLIF($4,'') WHERE id=$1", d.AttemptID, attemptState, status, code)
	if err != nil {
		return err
	}
	return tx.Commit()
}
