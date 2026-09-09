package atlas

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestExecuteDAGBoundedMappingAndCompensation(t *testing.T) {
	steps := []Step{
		{ID: "source", ServiceID: "fixture", ServiceVersion: 1, Required: true},
		{ID: "parallel_a", ServiceID: "fixture", ServiceVersion: 1, DependsOn: []string{"source"}, Required: true, InputMapping: map[string]string{"value": "source.value"}},
		{ID: "parallel_b", ServiceID: "fixture", ServiceVersion: 1, DependsOn: []string{"source"}, Required: true, InputMapping: map[string]string{"value": "source.value"}},
		{ID: "final", ServiceID: "fixture", ServiceVersion: 1, DependsOn: []string{"parallel_a", "parallel_b"}, Required: true, InputMapping: map[string]string{"a": "parallel_a.value", "b": "parallel_b.value"}},
	}
	var running, peak int
	var order []string
	var compensations []string
	got, err := ExecuteDAG(context.Background(), steps, 1, false, func(_ context.Context, step Step, input map[string]any) (any, error) {
		running++
		if running > peak {
			peak = running
		}
		order = append(order, step.ID)
		running--
		if step.ID == "source" {
			return map[string]any{"value": "ok"}, nil
		}
		if step.ID == "final" {
			return map[string]any{"value": input["a"].(string) + "+" + input["b"].(string)}, nil
		}
		return map[string]any{"value": step.ID + ":" + input["value"].(string)}, nil
	}, func(_ context.Context, step Step, _ any) error {
		compensations = append(compensations, step.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if peak != 1 {
		t.Fatalf("parallelism escaped limit: peak=%d", peak)
	}
	if !reflect.DeepEqual(got.Completed, order) {
		t.Fatalf("completion order mismatch: got=%v run=%v", got.Completed, order)
	}
	if got.Outputs["final"].(map[string]any)["value"] != "parallel_a:ok+parallel_b:ok" {
		t.Fatal("final output not produced")
	}
	if len(compensations) != 0 {
		t.Fatalf("unexpected compensation: %v", compensations)
	}
}

func TestExecuteDAGFailureCompensatesReverseAndPartial(t *testing.T) {
	steps := []Step{
		{ID: "required", ServiceID: "fixture", ServiceVersion: 1, Required: true},
		{ID: "optional", ServiceID: "fixture", ServiceVersion: 1, DependsOn: []string{"required"}, Required: false},
		{ID: "terminal", ServiceID: "fixture", ServiceVersion: 1, DependsOn: []string{"required"}, Required: true},
	}
	var compensated []string
	got, err := ExecuteDAG(context.Background(), steps, 2, true, func(_ context.Context, step Step, _ map[string]any) (any, error) {
		if step.ID == "optional" {
			return nil, errors.New("optional fixture failure")
		}
		if step.ID == "terminal" {
			return nil, errors.New("terminal fixture failure")
		}
		return map[string]any{"value": "ok"}, nil
	}, func(_ context.Context, step Step, _ any) error {
		compensated = append(compensated, step.ID)
		return nil
	})
	if err == nil || got.Failed != "terminal" {
		t.Fatalf("expected terminal failure, got %+v err=%v", got, err)
	}
	sort.Strings(compensated)
	if !reflect.DeepEqual(compensated, []string{"required"}) {
		t.Fatalf("unexpected compensation: %v", compensated)
	}
}
