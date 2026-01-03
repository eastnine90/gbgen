package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestNumberFeature_Evaluate_OK(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"rollout-percent": {"defaultValue": 12.5}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := NumberFeature("rollout-percent").Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true, got false")
	}
	if res.Value != 12.5 {
		t.Fatalf("expected Value=12.5, got %v", res.Value)
	}
}

func TestNumberFeature_Evaluate_TypeMismatch(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"rollout-percent": {"defaultValue": "twelve"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = NumberFeature("rollout-percent").Evaluate(ctx, client)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("expected errors.Is(err, ErrTypeMismatch)=true, got false (err=%T %v)", err, err)
	}
	var tm *TypeMismatchError
	if !errors.As(err, &tm) {
		t.Fatalf("expected TypeMismatchError, got %T (%v)", err, err)
	}
	if tm.FeatureKey != "rollout-percent" {
		t.Fatalf("expected FeatureKey=rollout-percent, got %q", tm.FeatureKey)
	}
	if tm.Expected != "float64" {
		t.Fatalf("expected Expected=float64, got %q", tm.Expected)
	}
}

func TestNumberFeature_Get_Sugar(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"rollout-percent": {"defaultValue": 12.5},
		"bad-number": {"defaultValue": "twelve"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if v, ok := NumberFeature("rollout-percent").Get(ctx, client); !ok || v != 12.5 {
		t.Fatalf("expected (12.5, true), got (%v, %v)", v, ok)
	}
	if v, ok := NumberFeature("bad-number").Get(ctx, client); ok || v != 0 {
		t.Fatalf("expected (0, false) on type mismatch, got (%v, %v)", v, ok)
	}
	if v := NumberFeature("bad-number").GetOr(ctx, client, 99.9); v != 99.9 {
		t.Fatalf("expected default=99.9, got %v", v)
	}
}

func TestNumberFeature_Key_And_GetOr_OK(t *testing.T) {
	ctx := context.Background()

	if got := NumberFeature("rollout-percent").Key(); got != "rollout-percent" {
		t.Fatalf("expected Key=rollout-percent, got %q", got)
	}

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"rollout-percent": {"defaultValue": 12.5}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Exercise the non-default branch of GetOr.
	if v := NumberFeature("rollout-percent").GetOr(ctx, client, 99.9); v != 12.5 {
		t.Fatalf("expected 12.5, got %v", v)
	}
}


