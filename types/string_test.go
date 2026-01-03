package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestStringFeature_Evaluate_OK(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"button-color": {"defaultValue": "blue"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := StringFeature("button-color").Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true, got false")
	}
	if res.Value != "blue" {
		t.Fatalf("expected Value=blue, got %q", res.Value)
	}
}

func TestStringFeature_Evaluate_TypeMismatch(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"button-color": {"defaultValue": true}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = StringFeature("button-color").Evaluate(ctx, client)
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
	if tm.FeatureKey != "button-color" {
		t.Fatalf("expected FeatureKey=button-color, got %q", tm.FeatureKey)
	}
	if tm.Expected != "string" {
		t.Fatalf("expected Expected=string, got %q", tm.Expected)
	}
}

func TestStringFeature_Get_Sugar(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"button-color": {"defaultValue": "blue"},
		"bad-string": {"defaultValue": true}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if v, ok := StringFeature("button-color").Get(ctx, client); !ok || v != "blue" {
		t.Fatalf("expected (blue, true), got (%q, %v)", v, ok)
	}
	if v, ok := StringFeature("bad-string").Get(ctx, client); ok || v != "" {
		t.Fatalf("expected (\"\", false) on type mismatch, got (%q, %v)", v, ok)
	}
	if v := StringFeature("bad-string").GetOr(ctx, client, "fallback"); v != "fallback" {
		t.Fatalf("expected default=fallback, got %q", v)
	}
}

func TestStringFeature_Key_And_GetOr_OK(t *testing.T) {
	ctx := context.Background()

	if got := StringFeature("button-color").Key(); got != "button-color" {
		t.Fatalf("expected Key=button-color, got %q", got)
	}

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"button-color": {"defaultValue": "blue"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Exercise the non-default branch of GetOr.
	if v := StringFeature("button-color").GetOr(ctx, client, "fallback"); v != "blue" {
		t.Fatalf("expected blue, got %q", v)
	}
}
