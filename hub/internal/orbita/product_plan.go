package orbita

import (
	"encoding/json"
	"errors"
	"fmt"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/contracts/economics"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
)

// ProductPlan é o DTO congelado pela admissão. O estado de execução fica no
// PostgreSQL; este tipo apenas liga cada registro do catálogo a seu comando
// idempotente inicial.
type ProductPlan struct {
	TargetID      string
	TargetVersion int
	MaxParallel   int
	AllowPartial  bool
	Consolidation string
	FailurePolicy string
	Steps         []ProductPlanStep
}

type ProductPlanStep struct {
	StepID                string
	ServiceID             string
	ServiceVersion        int
	DependsOn             []string
	Required              bool
	InputMapping          map[string]string
	CompensationServiceID string
	CompensationCommand   dispatch.Command
	CommandID             string
}

// BuildProductPlan materializa comandos por etapa a partir do snapshot de
// aceite. Cada comando aponta para o serviço versionado da etapa; assim o
// Cometa continua executando uma operação externa por vez e não recebe um
// produto sem adapter executável.
func BuildProductPlan(snapshot atlas.OfferSnapshot, base dispatch.Command, productInput json.RawMessage) (ProductPlan, []dispatch.Command, error) {
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil || snapshot.Target.Kind != "products" {
		return ProductPlan{}, nil, errors.New("snapshot não é produto")
	}
	if len(target.Steps) == 0 || len(snapshot.Services) < len(target.Steps) {
		return ProductPlan{}, nil, errors.New("snapshot de produto sem serviços completos")
	}
	services := make(map[string]atlas.Resource, len(snapshot.Services))
	for _, service := range snapshot.Services {
		services[service.ID+fmt.Sprintf("@%d", service.Version)] = service
	}
	plan := ProductPlan{
		TargetID: targetID(snapshot.Target), TargetVersion: snapshot.Target.Version,
		MaxParallel: target.MaxParallel, AllowPartial: target.AllowPartial,
		Consolidation: target.Consolidation, FailurePolicy: target.FailurePolicy,
		Steps: make([]ProductPlanStep, 0, len(target.Steps)),
	}
	commands := make([]dispatch.Command, 0, len(target.Steps))
	for _, step := range target.Steps {
		service, ok := services[step.ServiceID+fmt.Sprintf("@%d", step.ServiceVersion)]
		if !ok {
			return ProductPlan{}, nil, fmt.Errorf("serviço da etapa %s ausente no snapshot", step.ID)
		}
		serviceData, err := atlas.DecodeCatalogData(service)
		if err != nil || serviceData.AdapterID == "" {
			return ProductPlan{}, nil, fmt.Errorf("serviço da etapa %s sem adapter qualificado", step.ID)
		}
		// R6-EXE-01: a etapa SHALL ter rota/conta/vínculo/contrato próprios já
		// resolvidos por Atlas (nunca herdados do produto por omissão). Sem
		// isso, a admissão é recusada antes de qualquer efeito.
		stepOffer, ok := snapshot.StepOffers[step.ID]
		if !ok || stepOffer.Account.ID == "" || stepOffer.Binding.ID == "" {
			return ProductPlan{}, nil, fmt.Errorf("etapa %s sem rota/conta/vínculo próprios resolvidos", step.ID)
		}
		request := json.RawMessage(`{}`)
		if len(step.DependsOn) == 0 {
			request, err = atlas.TransformJSON(productInput, nil, serviceData.InputSchema)
			if err != nil {
				return ProductPlan{}, nil, fmt.Errorf("entrada da etapa %s inválida: %w", step.ID, err)
			}
		}
		childSnapshot := stepSnapshot(snapshot, service, stepOffer)
		config, err := json.Marshal(childSnapshot)
		if err != nil {
			return ProductPlan{}, nil, err
		}
		commandID := idgen.New()
		command := base
		command.StepID = step.ID
		command.CommandID = commandID
		command.ServiceCode = step.ServiceID
		command.ServiceVersion = step.ServiceVersion
		command.RequestBody = request
		command.ConfigSnapshot = config
		// A conta que efetivamente executa a etapa é a resolvida para ELA,
		// nunca a do produto (R6-EXE-01: probe reproduziu rota B com conta A).
		command.ProviderAccountID = stepOffer.Account.ID
		if stepEconomic, econErr := stepEconomicSnapshot(stepOffer); econErr == nil {
			command.EconomicSnapshot = stepEconomic
		} else {
			return ProductPlan{}, nil, fmt.Errorf("política econômica da etapa %s inválida: %w", step.ID, econErr)
		}
		plan.Steps = append(plan.Steps, ProductPlanStep{
			StepID: step.ID, ServiceID: step.ServiceID, ServiceVersion: step.ServiceVersion,
			DependsOn: append([]string(nil), step.DependsOn...), Required: step.Required,
			InputMapping: cloneStringMap(step.InputMapping), CompensationServiceID: step.CompensationServiceID,
			CommandID: commandID,
		})
		if step.CompensationServiceID != "" {
			compensationService, ok := services[step.CompensationServiceID+fmt.Sprintf("@%d", step.CompensationServiceVersion)]
			if !ok {
				return ProductPlan{}, nil, fmt.Errorf("serviço de compensação da etapa %s ausente no snapshot", step.ID)
			}
			compensationData, compErr := atlas.DecodeCatalogData(compensationService)
			if compErr != nil || compensationData.AdapterID == "" {
				return ProductPlan{}, nil, fmt.Errorf("serviço de compensação da etapa %s sem adapter qualificado", step.ID)
			}
			compensationOffer, ok := snapshot.StepOffers["compensate_"+step.ID]
			if !ok || compensationOffer.Account.ID == "" || compensationOffer.Binding.ID == "" {
				return ProductPlan{}, nil, fmt.Errorf("compensação da etapa %s sem rota/conta/vínculo próprios resolvidos", step.ID)
			}
			compensationSnapshot := stepSnapshot(snapshot, compensationService, compensationOffer)
			compensationConfig, compErr := json.Marshal(compensationSnapshot)
			if compErr != nil {
				return ProductPlan{}, nil, compErr
			}
			compensationCommand := command
			compensationCommand.StepID = "compensate_" + step.ID
			compensationCommand.CommandID = idgen.New()
			compensationCommand.ServiceCode = step.CompensationServiceID
			compensationCommand.ServiceVersion = step.CompensationServiceVersion
			compensationCommand.ConfigSnapshot = compensationConfig
			compensationCommand.RequestBody = request
			compensationCommand.ProviderAccountID = compensationOffer.Account.ID
			if compensationEconomic, econErr := stepEconomicSnapshot(compensationOffer); econErr == nil {
				compensationCommand.EconomicSnapshot = compensationEconomic
			} else {
				return ProductPlan{}, nil, fmt.Errorf("política econômica da compensação da etapa %s inválida: %w", step.ID, econErr)
			}
			plan.Steps[len(plan.Steps)-1].CompensationCommand = compensationCommand
		}
		commands = append(commands, command)
	}
	return plan, commands, nil
}

func targetID(resource atlas.Resource) string { return resource.ID }

// stepSnapshot freezes the per-command snapshot for one product step from
// its OWN independently resolved offer (R6-EXE-01): account, binding and
// purchase contract all come from stepOffer, never from the parent
// product's base snapshot. Only Offer/TechnicalProfile/SaleContract remain
// at the product level — a step has no sale of its own to the customer.
func stepSnapshot(base atlas.OfferSnapshot, service atlas.Resource, stepOffer atlas.StepOffer) atlas.OfferSnapshot {
	derived := base
	derived.Target = service
	derived.Services = []atlas.Resource{service}
	derived.StepOffers = nil
	derived.SelectedRoute = stepOffer.SelectedRoute
	derived.Account = stepOffer.Account
	derived.Binding = stepOffer.Binding
	derived.PurchaseContract = stepOffer.PurchaseContract
	derived.Hash = atlas.SnapshotHash(derived)
	return derived
}

// stepEconomicSnapshot builds the command-level economic snapshot for one
// step from its OWN purchase contract (cost), never the product's shared
// snapshot (R6-EXE-01: "venda do produto e custos das etapas usam chaves
// econômicas distintas"). Product-level revenue (Sell) is charged once, at
// protocol finalization, from the product's own SaleContract — never
// duplicated per step.
func stepEconomicSnapshot(stepOffer atlas.StepOffer) (json.RawMessage, error) {
	if stepOffer.PurchaseContract.ID == "" {
		return json.RawMessage(`{}`), nil
	}
	buy, buyRules, err := economics.DecodePublished(stepOffer.PurchaseContract.ID, stepOffer.PurchaseContract.Version, stepOffer.PurchaseContract.Data)
	if err != nil {
		return nil, fmt.Errorf("etapa: contrato de compra inválido: %w", err)
	}
	if buy.Kind != "PURCHASE" {
		return nil, errors.New("etapa: contrato de compra com tipo inesperado")
	}
	snapshot := economics.Snapshot{ContractID: stepOffer.PurchaseContract.ID, Version: stepOffer.PurchaseContract.Version, Currency: buy.Currency, SettlementParty: buy.SettlementParty, Buy: buyRules}
	return json.Marshal(snapshot)
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
