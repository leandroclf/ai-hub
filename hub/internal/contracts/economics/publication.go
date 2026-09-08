package economics

import (
	"encoding/json"
	"errors"
)

type PublishedContract struct {
	Kind              string        `json:"contract_kind"`
	Currency          string        `json:"currency"`
	UnitPrice         Decimal       `json:"unit_price"`
	SettlementParty   string        `json:"settlement_party"`
	Meter             string        `json:"meter"`
	UnitScope         string        `json:"unit_scope"`
	Rules             []PricingRule `json:"pricing_rules,omitempty"`
	StrictBalance     bool          `json:"strict_balance"`
	ReservationAmount Decimal       `json:"reservation_amount,omitempty"`
}

func DecodePublished(id string, version int, raw json.RawMessage) (PublishedContract, []PricingRule, error) {
	var c PublishedContract
	if json.Unmarshal(raw, &c) != nil || id == "" || version < 1 || (c.Kind != "PURCHASE" && c.Kind != "SALE") {
		return c, nil, errors.New("invalid published economic contract")
	}
	rules := append([]PricingRule(nil), c.Rules...)
	if len(rules) == 0 {
		incidence := map[string]string{"SUCCESS": "SUCCEEDED", "PARTIAL_SUCCESS": "PARTIALLY_SUCCEEDED", "SUBMISSION": "SUBMITTED", "STATUS": "STATUS", "FETCH": "FETCH"}[c.Meter]
		if incidence == "" {
			return c, nil, errors.New("explicit economic incidence required")
		}
		rules = []PricingRule{{Meter: c.Meter, Amount: c.UnitPrice, Incidence: []string{incidence}, UnitScope: c.UnitScope}}
	}
	for i := range rules {
		rules[i].ContractID = id
		rules[i].Version = version
	}
	snapshot := Snapshot{ContractID: id, Version: version, Currency: c.Currency, SettlementParty: c.SettlementParty, Sell: rules}
	if err := snapshot.Validate(); err != nil {
		return c, nil, err
	}
	if c.StrictBalance {
		if c.Kind != "SALE" || c.ReservationAmount.Sign() <= 0 {
			return c, nil, errors.New("strict sale requires explicit positive reservation amount")
		}
	}
	return c, rules, nil
}

func Freeze(purchaseID string, purchaseVersion int, purchase json.RawMessage, saleID string, saleVersion int, sale json.RawMessage) (Snapshot, PublishedContract, error) {
	buy, buyRules, err := DecodePublished(purchaseID, purchaseVersion, purchase)
	if err != nil {
		return Snapshot{}, PublishedContract{}, err
	}
	sell, sellRules, err := DecodePublished(saleID, saleVersion, sale)
	if err != nil {
		return Snapshot{}, PublishedContract{}, err
	}
	if buy.Kind != "PURCHASE" || sell.Kind != "SALE" || buy.Currency != sell.Currency {
		return Snapshot{}, PublishedContract{}, errors.New("incompatible purchase and sale contracts")
	}
	snapshot := Snapshot{ContractID: saleID, Version: saleVersion, Currency: sell.Currency, SettlementParty: buy.SettlementParty, Buy: buyRules, Sell: sellRules}
	return snapshot, sell, snapshot.Validate()
}
