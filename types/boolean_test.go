package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestBooleanFeature_Evaluate_OK(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"dark-mode": {"defaultValue": true}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := BooleanFeature("dark-mode").Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true, got false")
	}
	if res.Value != true {
		t.Fatalf("expected Value=true, got %v", res.Value)
	}
	if res.Raw == nil {
		t.Fatalf("expected Raw to be set, got nil")
	}
	if res.Raw.Source != growthbook.DefaultValueResultSource {
		t.Fatalf("expected Raw.Source=%q, got %q", growthbook.DefaultValueResultSource, res.Raw.Source)
	}
}

func TestBooleanFeature_Evaluate_TypeMismatch(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"dark-mode": {"defaultValue": "yes"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = BooleanFeature("dark-mode").Evaluate(ctx, client)
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
	if tm.FeatureKey != "dark-mode" {
		t.Fatalf("expected FeatureKey=dark-mode, got %q", tm.FeatureKey)
	}
	if tm.Expected != "bool" {
		t.Fatalf("expected Expected=bool, got %q", tm.Expected)
	}
}

func TestBooleanFeature_Get_Sugar(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"dark-mode": {"defaultValue": true},
		"bad-bool": {"defaultValue": "yes"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if v, ok := BooleanFeature("dark-mode").Get(ctx, client); !ok || v != true {
		t.Fatalf("expected (true, true), got (%v, %v)", v, ok)
	}
	if v, ok := BooleanFeature("bad-bool").Get(ctx, client); ok || v != false {
		t.Fatalf("expected (false, false) on type mismatch, got (%v, %v)", v, ok)
	}
	if v := BooleanFeature("bad-bool").GetOr(ctx, client, true); v != true {
		t.Fatalf("expected default=true, got %v", v)
	}
}

func TestBooleanFeature_Key_And_GetOr_OK(t *testing.T) {
	ctx := context.Background()

	if got := BooleanFeature("dark-mode").Key(); got != "dark-mode" {
		t.Fatalf("expected Key=dark-mode, got %q", got)
	}

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"dark-mode": {"defaultValue": true}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Exercise the non-default branch of GetOr.
	if v := BooleanFeature("dark-mode").GetOr(ctx, client, false); v != true {
		t.Fatalf("expected true, got %v", v)
	}
}
