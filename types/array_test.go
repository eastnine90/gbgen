package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestArrayFeature_Evaluate_OK(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"json-arr": {"defaultValue": [1,2,3]}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := ArrayFeature("json-arr").Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true, got false")
	}
	if res.Raw == nil {
		t.Fatalf("expected Raw to be set, got nil")
	}
	if got := len(res.Value); got != 3 {
		t.Fatalf("expected len(Value)=3, got %d", got)
	}
	if got := res.Value[0]; got != float64(1) {
		t.Fatalf("expected Value[0]=1, got %#v", got)
	}
}

func TestArrayFeature_Evaluate_TypeMismatch(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"json-arr": {"defaultValue": {"k":"v"}}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = ArrayFeature("json-arr").Evaluate(ctx, client)
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
	if tm.FeatureKey != "json-arr" {
		t.Fatalf("expected FeatureKey=json-arr, got %q", tm.FeatureKey)
	}
	if tm.Expected != "[]any" {
		t.Fatalf("expected Expected=[]any, got %q", tm.Expected)
	}
}

func TestArrayFeature_Get_Sugar(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"json-arr": {"defaultValue": [1,2,3]},
		"bad-arr": {"defaultValue": {"k":"v"}}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	v, ok := ArrayFeature("json-arr").Get(ctx, client)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(v) != 3 {
		t.Fatalf("expected len=3, got %d", len(v))
	}

	if v2, ok := ArrayFeature("bad-arr").Get(ctx, client); ok || v2 != nil {
		t.Fatalf("expected (nil, false) on type mismatch, got (%#v, %v)", v2, ok)
	}
	if v3 := ArrayFeature("bad-arr").GetOr(ctx, client, []any{"fallback"}); len(v3) != 1 || v3[0] != "fallback" {
		t.Fatalf("expected default slice, got %#v", v3)
	}
}

func TestArrayFeature_Key_And_GetOr_OK(t *testing.T) {
	ctx := context.Background()

	if got := ArrayFeature("json-arr").Key(); got != "json-arr" {
		t.Fatalf("expected Key=json-arr, got %q", got)
	}

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"json-arr": {"defaultValue": [1,2,3]}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Exercise the non-default branch of GetOr.
	if v := ArrayFeature("json-arr").GetOr(ctx, client, []any{"fallback"}); len(v) != 3 {
		t.Fatalf("expected len=3, got %d", len(v))
	}
}
