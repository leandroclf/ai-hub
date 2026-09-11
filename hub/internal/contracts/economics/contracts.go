package economics

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// Decimal preserves the lexical decimal value across JSON, Go and PostgreSQL.
// No monetary calculation or serialization passes through binary floating point.
var ErrLimitExceeded = errors.New("libra: strict limit exceeded")

type Decimal string

var decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,21})(\.[0-9]{1,8})?$`)

func ParseDecimal(v string) (Decimal, error) {
	if !decimalPattern.MatchString(v) {
		return "", errors.New("libra: decimal must have at most 22 integer and 8 fractional digits")
	}
	return Decimal(v), nil
}
func (d *Decimal) UnmarshalJSON(b []byte) error {
	var v string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
	} else {
		v = string(b)
	}
	p, err := ParseDecimal(v)
	if err != nil {
		return err
	}
	*d = p
	return nil
}
func (d Decimal) Rat() (*big.Rat, error) {
	if _, err := ParseDecimal(string(d)); err != nil {
		return nil, err
	}
	r, ok := new(big.Rat).SetString(string(d))
	if !ok {
		return nil, errors.New("libra: invalid decimal")
	}
	return r, nil
}
func (d Decimal) Canonical() (Decimal, error) {
	r, e := d.Rat()
	if e != nil {
		return "", e
	}
	return Decimal(r.FloatString(8)), nil
}
func (d Decimal) Sign() int {
	r, e := d.Rat()
	if e != nil {
		return -1
	}
	return r.Sign()
}

type Tier struct {
	UpTo   int64   `json:"up_to"`
	Amount Decimal `json:"amount"`
}
type Plan struct {
	Kind          string `json:"kind"`
	Period        string `json:"period"`
	IncludedUnits int64  `json:"included_units,omitempty"`
	Excess        string `json:"excess,omitempty"`
	Tiers         []Tier `json:"tiers,omitempty"`
}
type PricingRule struct {
	ContractID string   `json:"contract_id,omitempty"`
	Version    int      `json:"version,omitempty"`
	Meter      string   `json:"meter"`
	Amount     Decimal  `json:"amount"`
	Incidence  []string `json:"incidence"`
	UnitScope  string   `json:"unit_scope"`
	Plan       *Plan    `json:"plan,omitempty"`
}
type Snapshot struct {
	ContractID      string        `json:"contract_id"`
	Version         int           `json:"version"`
	Currency        string        `json:"currency"`
	SettlementParty string        `json:"settlement_party"`
	Buy             []PricingRule `json:"buy"`
	Sell            []PricingRule `json:"sell"`
}
type EconomicEvent struct {
	ProtocolID  string `json:"protocol_id"`
	TenantID    string `json:"tenant_id"`
	OperationID string `json:"operation_id,omitempty"`
	StepID      string `json:"step_id,omitempty"`
	AttemptID   string `json:"attempt_id,omitempty"`
	// EconomicKind separates the financial incidence from the operational
	// observation. For example, an externally accepted SUBMIT remains
	// UNKNOWN to Orbita while Libra must record the SUBMITTED incidence.
	// Empty means that Kind itself is the contracted incidence.
	EconomicKind        string    `json:"economic_kind,omitempty"`
	ProviderAccountID   string    `json:"provider_account_id,omitempty"`
	CredentialBindingID string    `json:"credential_binding_id,omitempty"`
	Status              string    `json:"status,omitempty"`
	Kind                string    `json:"kind,omitempty"`
	EvidenceID          string    `json:"evidence_id"`
	OccurredAt          time.Time `json:"occurred_at"`
	ExternalState       string    `json:"external_state,omitempty"`
	SafeToRelease       bool      `json:"safe_to_release,omitempty"`
	EconomicSnapshot    Snapshot  `json:"economic_snapshot"`
}

func (s Snapshot) Validate() error {
	if s.ContractID == "" || s.Version < 1 || len(s.Currency) != 3 || s.Currency != strings.ToUpper(s.Currency) || (s.SettlementParty != "HUB" && s.SettlementParty != "CLIENT_DIRECT") {
		return errors.New("libra: invalid economic snapshot")
	}
	if len(s.Buy) == 0 && len(s.Sell) == 0 {
		return errors.New("libra: economic incidence is required")
	}
	for _, rules := range [][]PricingRule{s.Buy, s.Sell} {
		seen := map[string]bool{}
		for _, r := range rules {
			if r.Meter == "" || seen[r.Meter] || len(r.Incidence) == 0 {
				return errors.New("libra: invalid or duplicate meter")
			}
			seen[r.Meter] = true
			if _, e := r.Amount.Rat(); e != nil || r.Amount.Sign() < 0 {
				return errors.New("libra: invalid nonnegative tariff")
			}
			switch r.UnitScope {
			case "PROTOCOL", "STEP", "OPERATION", "ATTEMPT":
			default:
				return errors.New("libra: invalid economic unit scope")
			}
			if _, e := Price(r, 0, 1); e != nil {
				return e
			}
		}
	}
	return nil
}

// Price returns the exact delta of a published plan. Volume tiers may generate
// a negative delta at a threshold; the journal then records compensating sides.
func Price(r PricingRule, previous, quantity int64) (Decimal, error) {
	if previous < 0 || quantity < 1 || quantity > 1_000_000_000 || previous > 1_000_000_000-quantity {
		return "", errors.New("libra: quantity outside qualified envelope")
	}
	a, e := r.Amount.Rat()
	if e != nil {
		return "", e
	}
	if r.Plan == nil || r.Plan.Kind == "UNIT" || r.Plan.Kind == "PACKAGE" {
		return Decimal(new(big.Rat).Mul(a, new(big.Rat).SetInt64(quantity)).FloatString(8)), nil
	}
	p := r.Plan
	if p.Period == "" {
		return "", errors.New("libra: plan requires explicit period")
	}
	if p.Kind == "ALLOWANCE" {
		if p.IncludedUnits < 0 || (p.Excess != "DENY" && p.Excess != "CHARGE") {
			return "", errors.New("libra: invalid allowance policy")
		}
		bill := max(int64(0), previous+quantity-p.IncludedUnits) - max(int64(0), previous-p.IncludedUnits)
		if bill > 0 && p.Excess == "DENY" {
			return "", ErrLimitExceeded
		}
		return Decimal(new(big.Rat).Mul(a, new(big.Rat).SetInt64(bill)).FloatString(8)), nil
	}
	if p.Kind != "MARGINAL" && p.Kind != "VOLUME" {
		return "", errors.New("libra: unsupported explicit pricing policy")
	}
	if len(p.Tiers) == 0 {
		return "", errors.New("libra: tiers required")
	}
	last := int64(0)
	for i, t := range p.Tiers {
		if t.Amount.Sign() < 0 || (t.UpTo <= last && t.UpTo != 0) || (t.UpTo == 0 && i != len(p.Tiers)-1) {
			return "", errors.New("libra: invalid ordered tiers")
		}
		last = t.UpTo
	}
	total := func(n int64) (*big.Rat, error) {
		v := new(big.Rat)
		start := int64(0)
		for _, t := range p.Tiers {
			end := t.UpTo
			if end == 0 {
				end = n
			}
			rate, _ := t.Amount.Rat()
			if p.Kind == "VOLUME" {
				if n <= end {
					return new(big.Rat).Mul(rate, new(big.Rat).SetInt64(n)), nil
				}
			} else {
				count := max(int64(0), min(n, end)-start)
				v.Add(v, new(big.Rat).Mul(rate, new(big.Rat).SetInt64(count)))
				if n <= end {
					return v, nil
				}
			}
			start = end
		}
		return nil, errors.New("libra: quantity exceeds published tiers")
	}
	x, e := total(previous + quantity)
	if e != nil {
		return "", e
	}
	y, e := total(previous)
	if e != nil {
		return "", e
	}
	out := Decimal(new(big.Rat).Sub(x, y).FloatString(8))
	if _, e := ParseDecimal(string(out)); e != nil {
		return "", fmt.Errorf("libra: price overflow: %w", e)
	}
	return out, nil
}
