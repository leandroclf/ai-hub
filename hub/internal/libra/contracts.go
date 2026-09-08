package libra

import "ai-hub/hub/internal/contracts/economics"

// Shared immutable contract DTOs and exact arithmetic; persistence remains Libra-owned.
type Decimal = economics.Decimal
type Tier = economics.Tier
type Plan = economics.Plan
type PricingRule = economics.PricingRule
type Snapshot = economics.Snapshot
type EconomicEvent = economics.EconomicEvent

var ParseDecimal = economics.ParseDecimal
var Price = economics.Price
