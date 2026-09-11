package orbita

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
)

type productStep struct {
	StepID                string
	CommandID             string
	ServiceID             string
	ServiceVersion        int
	Required              bool
	DependsOn             []string
	InputMapping          map[string]string
	State                 string
	Command               []byte
	Result                []byte
	ErrorCode             string
	ErrorMessage          string
	CompensationServiceID sql.NullString
	CompensationCommand   []byte
	CompensatesStepID     sql.NullString
}

type ProductFactOutcome struct {
	Finalize        bool
	Status          Status
	ExpectedVersion int
	Body            FinalBody
}

func (s *Store) ProductStepByCommand(ctx context.Context, commandID string) (string, bool, error) {
	var stepID string
	err := s.db.QueryRowContext(ctx, `SELECT step_id FROM operation_steps WHERE command_id=$1`, commandID).Scan(&stepID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return stepID, true, nil
}

// ClaimProductStep marks a durable step RUNNING while the intent lease is
// held. The transition is deliberately conditional: a stale publisher cannot
// resurrect a step already completed by a newer observation.
func (s *Store) ClaimProductStep(ctx context.Context, commandID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE operation_steps SET state='RUNNING',version=version+1,updated_at=clock_timestamp() WHERE command_id=$1 AND state IN ('READY','PENDING')`, commandID)
	return err
}

// ApplyProductFact records one external observation and unlocks only the
// dependent stages whose inputs are now durable. It never calls a provider or
// finalizes the protocol inside the transaction.
func (s *Store) ApplyProductFact(ctx context.Context, protocolID, commandID, kind string, response any, errorMessage string) (ProductFactOutcome, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ProductFactOutcome{}, err
	}
	defer tx.Rollback()
	var step productStep
	var allowPartial bool
	var failurePolicy string
	var protocolVersion int
	var dependsRaw, mappingRaw []byte
	if err = tx.QueryRowContext(ctx, `SELECT s.step_id,s.command_id,s.service_id,s.service_version,s.required,s.depends_on,s.input_mapping,s.state,s.command,COALESCE(s.result,'null'::jsonb),COALESCE(s.error_code,''),COALESCE(s.error_message,''),s.compensation_service_id,COALESCE(s.compensation_command,'null'::jsonb),s.compensates_step_id,p.allow_partial,p.failure_policy,pr.version
		FROM operation_steps s JOIN operation_plans p ON p.protocol_id=s.protocol_id JOIN protocols pr ON pr.protocol_id=s.protocol_id
		WHERE s.protocol_id=$1 AND s.command_id=$2 FOR UPDATE`, protocolID, commandID).Scan(
		&step.StepID, &step.CommandID, &step.ServiceID, &step.ServiceVersion, &step.Required, &dependsRaw,
		&mappingRaw, &step.State, &step.Command, &step.Result, &step.ErrorCode, &step.ErrorMessage,
		&step.CompensationServiceID, &step.CompensationCommand, &step.CompensatesStepID, &allowPartial, &failurePolicy, &protocolVersion); err != nil {
		return ProductFactOutcome{}, err
	}
	if err = json.Unmarshal(dependsRaw, &step.DependsOn); err != nil {
		return ProductFactOutcome{}, err
	}
	if err = json.Unmarshal(mappingRaw, &step.InputMapping); err != nil {
		return ProductFactOutcome{}, err
	}
	if kind != "UNKNOWN" && kind != "SUCCEEDED" && kind != "FAILED" {
		return ProductFactOutcome{}, errors.New("invalid product observation kind")
	}
	compensation := step.CompensatesStepID.Valid
	if kind == "UNKNOWN" {
		if step.State == "SUCCEEDED" || step.State == "FAILED" || step.State == "SKIPPED" || step.State == "COMPENSATED" || step.State == "COMPENSATION_FAILED" {
			if err = tx.Commit(); err != nil {
				return ProductFactOutcome{}, err
			}
			return ProductFactOutcome{}, nil
		}
		_, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='WAITING_PROVIDER',version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND command_id=$2 AND state NOT IN ('SUCCEEDED','FAILED','SKIPPED','COMPENSATED','COMPENSATION_FAILED')`, protocolID, commandID)
		if err != nil {
			return ProductFactOutcome{}, err
		}
		if err = tx.Commit(); err != nil {
			return ProductFactOutcome{}, err
		}
		return ProductFactOutcome{}, nil
	}
	if step.State != "SUCCEEDED" && step.State != "FAILED" && step.State != "SKIPPED" && step.State != "COMPENSATED" && step.State != "COMPENSATION_FAILED" {
		resultRaw, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			return ProductFactOutcome{}, marshalErr
		}
		next := "FAILED"
		if kind == "SUCCEEDED" {
			next = "SUCCEEDED"
		}
		if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state=$3,result=$4,error_message=NULLIF($5,''),version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND command_id=$2`, protocolID, commandID, next, resultRaw, errorMessage); err != nil {
			return ProductFactOutcome{}, err
		}
		step.State, step.Result = next, resultRaw
		step.ErrorMessage = errorMessage
		if compensation {
			originalState := "COMPENSATION_FAILED"
			if next == "SUCCEEDED" {
				originalState = "COMPENSATED"
			}
			if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state=$3,version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2 AND state='COMPENSATING'`, protocolID, step.CompensatesStepID.String, originalState); err != nil {
				return ProductFactOutcome{}, err
			}
		}
	}

	var steps []productStep
	rows, err := tx.QueryContext(ctx, `SELECT step_id,command_id,service_id,service_version,required,depends_on,input_mapping,state,command,COALESCE(result,'null'::jsonb),COALESCE(error_code,''),COALESCE(error_message,''),compensation_service_id,COALESCE(compensation_command,'null'::jsonb),compensates_step_id FROM operation_steps WHERE protocol_id=$1 ORDER BY step_id FOR UPDATE`, protocolID)
	if err != nil {
		return ProductFactOutcome{}, err
	}
	for rows.Next() {
		var current productStep
		var currentDepends, currentMapping []byte
		if err = rows.Scan(&current.StepID, &current.CommandID, &current.ServiceID, &current.ServiceVersion, &current.Required, &currentDepends, &currentMapping, &current.State, &current.Command, &current.Result, &current.ErrorCode, &current.ErrorMessage, &current.CompensationServiceID, &current.CompensationCommand, &current.CompensatesStepID); err != nil {
			rows.Close()
			return ProductFactOutcome{}, err
		}
		if err = json.Unmarshal(currentDepends, &current.DependsOn); err != nil {
			rows.Close()
			return ProductFactOutcome{}, err
		}
		if err = json.Unmarshal(currentMapping, &current.InputMapping); err != nil {
			rows.Close()
			return ProductFactOutcome{}, err
		}
		steps = append(steps, current)
	}
	if err = rows.Close(); err != nil {
		return ProductFactOutcome{}, err
	}
	byID := make(map[string]*productStep, len(steps))
	for i := range steps {
		byID[steps[i].StepID] = &steps[i]
	}

	// STOP prevents new productive effects after the first failure. A
	// dependent step is SKIPPED regardless of policy because its precondition
	// is not satisfied; this is persisted, not inferred by a goroutine.
	failed := false
	for _, current := range steps {
		if current.State == "FAILED" {
			failed = true
		}
	}
	for _, current := range steps {
		if current.State != "PENDING" && current.State != "READY" {
			continue
		}
		if failed && (failurePolicy == "STOP" || failurePolicy == "COMPENSATE") {
			if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='SKIPPED',version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2`, protocolID, current.StepID); err != nil {
				return ProductFactOutcome{}, err
			}
			byID[current.StepID].State = "SKIPPED"
			continue
		}
		depFailed := false
		ready := true
		for _, dep := range current.DependsOn {
			parent := byID[dep]
			if parent == nil || parent.State == "FAILED" || parent.State == "SKIPPED" || parent.State == "COMPENSATION_FAILED" {
				depFailed = true
			}
			if parent == nil || parent.State != "SUCCEEDED" {
				ready = false
			}
		}
		if depFailed {
			if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='SKIPPED',version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2`, protocolID, current.StepID); err != nil {
				return ProductFactOutcome{}, err
			}
			byID[current.StepID].State = "SKIPPED"
			continue
		}
		if !ready || current.State != "PENDING" {
			continue
		}
		command, request, buildErr := productStepCommand(current, byID)
		if buildErr != nil {
			return ProductFactOutcome{}, buildErr
		}
		command.RequestBody = request
		rawCommand, marshalErr := json.Marshal(command)
		if marshalErr != nil {
			return ProductFactOutcome{}, marshalErr
		}
		if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='READY',command=$3,version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2`, protocolID, current.StepID, rawCommand); err != nil {
			return ProductFactOutcome{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE command_intents SET state='READY',command=$2,next_attempt_at=clock_timestamp(),last_error=NULL WHERE command_id=$1 AND state='PENDING'`, current.CommandID, rawCommand); err != nil {
			return ProductFactOutcome{}, err
		}
		byID[current.StepID].State = "READY"
	}
	if failed && failurePolicy == "COMPENSATE" {
		compensationRecords := make([]productStep, 0)
		for _, current := range append([]productStep(nil), steps...) {
			if current.State != "SUCCEEDED" || current.CompensationServiceID.String == "" {
				continue
			}
			var compensationExists bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM operation_steps WHERE protocol_id=$1 AND compensates_step_id=$2)`, protocolID, current.StepID).Scan(&compensationExists); err != nil {
				return ProductFactOutcome{}, err
			}
			if compensationExists {
				continue
			}
			if len(current.CompensationCommand) == 0 || string(current.CompensationCommand) == "null" {
				if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='COMPENSATION_FAILED',version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2 AND state='SUCCEEDED'`, protocolID, current.StepID); err != nil {
					return ProductFactOutcome{}, err
				}
				byID[current.StepID].State = "COMPENSATION_FAILED"
				continue
			}
			var compensation dispatch.Command
			if err = json.Unmarshal(current.CompensationCommand, &compensation); err != nil {
				return ProductFactOutcome{}, err
			}
			if len(current.Result) > 0 && string(current.Result) != "null" {
				compensation.RequestBody = json.RawMessage(current.Result)
			}
			compensationRaw, marshalErr := json.Marshal(compensation)
			if marshalErr != nil {
				return ProductFactOutcome{}, marshalErr
			}
			compensationStepID := "compensate_" + current.StepID
			if _, err = tx.ExecContext(ctx, `UPDATE operation_steps SET state='COMPENSATING',version=version+1,updated_at=clock_timestamp() WHERE protocol_id=$1 AND step_id=$2 AND state='SUCCEEDED'`, protocolID, current.StepID); err != nil {
				return ProductFactOutcome{}, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO operation_steps(protocol_id,tenant_id,cell_id,step_id,command_id,service_id,service_version,required,depends_on,input_mapping,state,command,compensates_step_id) SELECT $1,p.tenant_id,p.cell_id,$2,$3,$4,$5,FALSE,'[]'::jsonb,'{}'::jsonb,'READY',$6,$7 FROM operation_plans p WHERE p.protocol_id=$1`, protocolID, compensationStepID, compensation.CommandID, compensation.ServiceCode, compensation.ServiceVersion, compensationRaw, current.StepID); err != nil {
				return ProductFactOutcome{}, err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO command_intents(command_id,protocol_id,tenant_id,application_id,cell_id,dispatch_mode,command,state) SELECT $1,p.protocol_id,p.tenant_id,p.application_id,p.cell_id,$2,$3,'READY' FROM protocols p WHERE p.protocol_id=$4`, compensation.CommandID, compensation.DispatchMode, compensationRaw, protocolID); err != nil {
				return ProductFactOutcome{}, err
			}
			compensationRecord := productStep{StepID: compensationStepID, CommandID: compensation.CommandID, ServiceID: compensation.ServiceCode, ServiceVersion: compensation.ServiceVersion, State: "READY", Command: compensationRaw, CompensatesStepID: sql.NullString{String: current.StepID, Valid: true}}
			compensationRecords = append(compensationRecords, compensationRecord)
			byID[current.StepID].State = "COMPENSATING"
		}
		steps = append(steps, compensationRecords...)
		byID = make(map[string]*productStep, len(steps))
		for i := range steps {
			byID[steps[i].StepID] = &steps[i]
		}
	}

	allTerminal := true
	anyRequiredFailure := false
	anyFailure := false
	anySuccess := false
	resultSteps := make(map[string]any, len(steps))
	failedSteps := make([]string, 0)
	skippedSteps := make([]string, 0)
	for _, current := range steps {
		latest := byID[current.StepID]
		if latest == nil {
			continue
		}
		switch latest.State {
		case "SUCCEEDED":
			anySuccess = true
			if !latest.CompensatesStepID.Valid {
				var value any
				if json.Unmarshal(latest.Result, &value) == nil {
					resultSteps[latest.StepID] = value
				}
			}
		case "FAILED":
			anyFailure = true
			if !latest.CompensatesStepID.Valid {
				failedSteps = append(failedSteps, latest.StepID)
			}
			if latest.Required {
				anyRequiredFailure = true
			}
		case "SKIPPED":
			if !latest.CompensatesStepID.Valid {
				skippedSteps = append(skippedSteps, latest.StepID)
			}
			if latest.Required {
				anyRequiredFailure = true
			}
		case "COMPENSATED":
			// The original effect was compensated; the product still fails
			// because another required step failed, but this is not a second
			// failure or a partial-success signal.
		case "COMPENSATION_FAILED":
			anyFailure = true
			if latest.CompensatesStepID.Valid {
				failedSteps = append(failedSteps, latest.CompensatesStepID.String+"/compensation")
			} else {
				failedSteps = append(failedSteps, latest.StepID)
			}
			if latest.Required {
				anyRequiredFailure = true
			}
		default:
			allTerminal = false
		}
	}
	if err = tx.Commit(); err != nil {
		return ProductFactOutcome{}, err
	}
	if !allTerminal {
		return ProductFactOutcome{}, nil
	}
	status := StatusSucceeded
	if anyRequiredFailure || (!allowPartial && anyFailure) {
		status = StatusFailed
	} else if anyFailure || len(skippedSteps) > 0 {
		if allowPartial && anySuccess {
			status = StatusPartiallySucceeded
		} else {
			status = StatusFailed
		}
	}
	body := FinalBody{ProtocolID: protocolID, Result: map[string]any{"steps": resultSteps, "failed_steps": failedSteps, "skipped_steps": skippedSteps}, ErrorMessage: strings.TrimSpace(errorMessage)}
	return ProductFactOutcome{Finalize: true, Status: status, ExpectedVersion: protocolVersion, Body: body}, nil
}

func productStepCommand(step productStep, byID map[string]*productStep) (dispatch.Command, json.RawMessage, error) {
	var command dispatch.Command
	if err := json.Unmarshal(step.Command, &command); err != nil {
		return dispatch.Command{}, nil, err
	}
	input := map[string]json.RawMessage{}
	for dst, source := range step.InputMapping {
		parts := strings.Split(source, ".")
		if len(parts) != 2 {
			return dispatch.Command{}, nil, fmt.Errorf("mapeamento inválido na etapa %s", step.StepID)
		}
		dep := byID[parts[0]]
		if dep == nil || len(dep.Result) == 0 || string(dep.Result) == "null" {
			return dispatch.Command{}, nil, fmt.Errorf("saída ausente para %s", source)
		}
		var object map[string]json.RawMessage
		if json.Unmarshal(dep.Result, &object) != nil {
			return dispatch.Command{}, nil, fmt.Errorf("saída da etapa %s não é objeto", parts[0])
		}
		value, ok := object[parts[1]]
		if !ok {
			return dispatch.Command{}, nil, fmt.Errorf("campo ausente para %s", source)
		}
		input[dst] = value
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return dispatch.Command{}, nil, err
	}
	var snapshot atlas.OfferSnapshot
	if err = json.Unmarshal(command.ConfigSnapshot, &snapshot); err != nil {
		return dispatch.Command{}, nil, err
	}
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err != nil {
		return dispatch.Command{}, nil, err
	}
	validated, err := atlas.TransformJSON(raw, nil, target.InputSchema)
	if err != nil {
		return dispatch.Command{}, nil, fmt.Errorf("entrada da etapa %s inválida: %w", step.StepID, err)
	}
	return command, validated, nil
}
