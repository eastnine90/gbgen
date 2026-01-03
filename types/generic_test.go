package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestTypedFeature_Evaluate_OK(t *testing.T) {
	type Config struct {
		Currency string `json:"currency"`
		MaxItems int    `json:"maxItems"`
	}

	ctx := context.Background()
	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": {"currency":"USD","maxItems":3}}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := WithType[Config](JSONFeature("checkout-config")).Evaluate(ctx, client)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true")
	}
	if res.Raw == nil {
		t.Fatalf("expected Raw to be set")
	}
	if res.Value.Currency != "USD" || res.Value.MaxItems != 3 {
		t.Fatalf("unexpected decoded value: %#v", res.Value)
	}

	// Exercise the non-default branch of GetOr.
	def := Config{Currency: "DEF", MaxItems: 9}
	got := WithType[Config](JSONFeature("checkout-config")).GetOr(ctx, client, def)
	if got.Currency != "USD" || got.MaxItems != 3 {
		t.Fatalf("expected decoded value from GetOr, got %#v", got)
	}
}

func TestTypedFeature_Get_GetOr_TypeMismatch(t *testing.T) {
	type Config struct {
		Currency string `json:"currency"`
	}

	ctx := context.Background()
	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": [1,2,3]}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// Evaluate should return a TypeMismatchError.
	_, err = WithType[Config](JSONFeature("checkout-config")).Evaluate(ctx, client)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("expected errors.Is(err, ErrTypeMismatch)=true, got false (err=%T %v)", err, err)
	}

	// Get should return ok=false on mismatch.
	if v, ok := WithType[Config](JSONFeature("checkout-config")).Get(ctx, client); ok {
		t.Fatalf("expected ok=false, got true (v=%#v)", v)
	}

	// GetOr should return the default on mismatch.
	def := Config{Currency: "DEF"}
	got := WithType[Config](JSONFeature("checkout-config")).GetOr(ctx, client, def)
	if got != def {
		t.Fatalf("expected default %#v, got %#v", def, got)
	}
}

func TestTypedFeature_GetOr_MissingKey_ReturnsDefault(t *testing.T) {
	type Config struct {
		A string `json:"a"`
	}

	ctx := context.Background()
	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	def := Config{A: "default"}
	got := WithType[Config](JSONFeature("missing")).GetOr(ctx, client, def)
	if got != def {
		t.Fatalf("expected default %#v, got %#v", def, got)
	}

	if v, ok := WithType[Config](JSONFeature("missing")).Get(ctx, client); ok {
		t.Fatalf("expected ok=false for missing key, got true (v=%#v)", v)
	}

	_, err = WithType[Config](JSONFeature("missing")).Evaluate(ctx, client)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingKey) {
		t.Fatalf("expected errors.Is(err, ErrMissingKey)=true, got false (err=%T %v)", err, err)
	}
}
