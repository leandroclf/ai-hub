package orbita

import (
	"encoding/json"
	"errors"
	"fmt"

	"ai-hub/hub/internal/atlas"
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
		request := json.RawMessage(`{}`)
		if len(step.DependsOn) == 0 {
			request, err = atlas.TransformJSON(productInput, nil, serviceData.InputSchema)
			if err != nil {
				return ProductPlan{}, nil, fmt.Errorf("entrada da etapa %s inválida: %w", step.ID, err)
			}
		}
		childSnapshot := snapshot
		childSnapshot.Target = service
		childSnapshot.Services = nil
		childSnapshot.Hash = ""
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
			compensationSnapshot := snapshot
			compensationSnapshot.Target = compensationService
			compensationSnapshot.Services = nil
			compensationSnapshot.Hash = ""
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
			plan.Steps[len(plan.Steps)-1].CompensationCommand = compensationCommand
		}
		commands = append(commands, command)
	}
	return plan, commands, nil
}

func targetID(resource atlas.Resource) string { return resource.ID }

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
