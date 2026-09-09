package atlas

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// StepExecutor executa uma etapa já validada do produto. O executor recebe
// somente as saídas das dependências declaradas; nenhuma etapa pode observar
// estado implícito de outra etapa.
type StepExecutor func(context.Context, Step, map[string]any) (any, error)

// StepCompensator desfaz o efeito de uma etapa concluída. A compensação é
// chamada em ordem reversa e continua mesmo quando uma compensação falha;
// falhas são devolvidas ao chamador para reconciliação explícita.
type StepCompensator func(context.Context, Step, any) error

type Execution struct {
	Outputs      map[string]any
	Completed    []string
	Failed       string
	Partial      bool
	Compensated  []string
	Compensation map[string]error
}

// ExecuteDAG executa as camadas de PlanDAG com paralelismo limitado. Etapas
// obrigatórias falhas interrompem novas camadas e compensam o que já terminou;
// etapas opcionais podem produzir execução parcial quando allowPartial está
// habilitado. O limite é aplicado por camada e nunca cria goroutines sem
// controle por etapa.
func ExecuteDAG(ctx context.Context, steps []Step, maxParallel int, allowPartial bool, run StepExecutor, compensate StepCompensator) (Execution, error) {
	if run == nil {
		return Execution{}, errors.New("executor ausente")
	}
	layers, err := PlanDAG(steps)
	if err != nil {
		return Execution{}, err
	}
	if maxParallel < 1 {
		maxParallel = 1
	}
	byID := make(map[string]Step, len(steps))
	for _, step := range steps {
		byID[step.ID] = step
	}
	result := Execution{Outputs: map[string]any{}, Compensation: map[string]error{}}
	completed := make([]string, 0, len(steps))
	var outputsMu sync.RWMutex

	rollback := func(cause error) (Execution, error) {
		if compensate == nil {
			return result, cause
		}
		for i := len(completed) - 1; i >= 0; i-- {
			id := completed[i]
			if err := compensate(ctx, byID[id], result.Outputs[id]); err != nil {
				result.Compensation[id] = err
				continue
			}
			result.Compensated = append(result.Compensated, id)
		}
		return result, cause
	}

	for _, layer := range layers {
		if err := ctx.Err(); err != nil {
			result.Failed = "context"
			return rollback(err)
		}
		sem := make(chan struct{}, maxParallel)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstErr error
		var failedID string
		for _, id := range layer {
			step := byID[id]
			inputs := make(map[string]any, len(step.InputMapping))
			var mappingErr error
			for target, source := range step.InputMapping {
				parts := splitMapping(source)
				if parts == nil {
					mappingErr = fmt.Errorf("mapeamento inválido no passo %s", id)
					failedID = id
					break
				}
				outputsMu.RLock()
				value, ok := result.Outputs[parts[0]].(map[string]any)
				outputsMu.RUnlock()
				if !ok {
					mappingErr = fmt.Errorf("saída da dependência %s não é objeto", parts[0])
					failedID = id
					break
				}
				v, ok := value[parts[1]]
				if !ok {
					mappingErr = fmt.Errorf("campo %s ausente na saída de %s", parts[1], parts[0])
					failedID = id
					break
				}
				inputs[target] = v
			}
			if mappingErr != nil {
				firstErr = mappingErr
				break
			}
			wg.Add(1)
			go func(step Step, input map[string]any) {
				defer wg.Done()
				sem <- struct{}{}
				out, runErr := run(ctx, step, input)
				<-sem
				mu.Lock()
				defer mu.Unlock()
				if runErr != nil {
					if firstErr == nil || step.Required {
						firstErr, failedID = runErr, step.ID
					}
					if !step.Required && allowPartial {
						result.Partial = true
					}
					return
				}
				outputsMu.Lock()
				result.Outputs[step.ID] = out
				outputsMu.Unlock()
				completed = append(completed, step.ID)
			}(step, inputs)
		}
		wg.Wait()
		if firstErr != nil {
			result.Failed = failedID
			if result.Partial && !byID[failedID].Required {
				continue
			}
			return rollback(firstErr)
		}
	}
	result.Completed = append(result.Completed, completed...)
	return result, nil
}

func splitMapping(source string) []string {
	parts := make([]string, 0, 2)
	for i := 0; i < len(source); {
		j := i
		for j < len(source) && source[j] != '.' {
			j++
		}
		parts = append(parts, source[i:j])
		if j == len(source) {
			break
		}
		i = j + 1
	}
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}
	return parts
}
