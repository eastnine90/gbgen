package types

import (
	"context"
	"errors"
	"testing"

	"github.com/growthbook/growthbook-golang"
)

func TestJSONFeature_Evaluate_OK(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": {"currency":"USD","maxItems":3}}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res, err := JSONFeature("checkout-config").Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected Valid=true, got false")
	}
	if got := res.Value["currency"]; got != "USD" {
		t.Fatalf("expected currency=USD, got %#v", got)
	}
	// JSON numbers unmarshal as float64.
	if got := res.Value["maxItems"]; got != float64(3) {
		t.Fatalf("expected maxItems=3, got %#v", got)
	}
}

func TestJSONFeature_Evaluate_TypeMismatch(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": "not-json"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = JSONFeature("checkout-config").Evaluate(ctx, client)
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
	if tm.FeatureKey != "checkout-config" {
		t.Fatalf("expected FeatureKey=checkout-config, got %q", tm.FeatureKey)
	}
	if tm.Expected != "json" {
		t.Fatalf("expected Expected=json, got %q", tm.Expected)
	}
}

func TestJSONFeature_Get_Sugar(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": {"currency":"USD","maxItems":3}},
		"bad-json": {"defaultValue": "not-json"}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	v, ok := JSONFeature("checkout-config").Get(ctx, client)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if v["currency"] != "USD" {
		t.Fatalf("expected currency=USD, got %#v", v["currency"])
	}

	if v2, ok := JSONFeature("bad-json").Get(ctx, client); ok || v2 != nil {
		t.Fatalf("expected (nil, false) on type mismatch, got (%#v, %v)", v2, ok)
	}
	if v3 := JSONFeature("bad-json").GetOr(ctx, client, map[string]any{"fallback": true}); v3["fallback"] != true {
		t.Fatalf("expected default map, got %#v", v3)
	}
}

func TestJSONFeature_EvaluateAny_AllowsNonObject(t *testing.T) {
	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"json-str": {"defaultValue": "hello"},
		"json-arr": {"defaultValue": [1,2,3]},
		"json-obj": {"defaultValue": {"k":"v"}}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	res1, err := JSONFeature("json-str").EvaluateAny(ctx, client)
	if err != nil {
		t.Fatalf("EvaluateAny: %v", err)
	}
	if !res1.Valid || res1.Raw == nil {
		t.Fatalf("expected Valid=true and Raw set")
	}
	if got, ok := res1.Value.(string); !ok || got != "hello" {
		t.Fatalf("expected string \"hello\", got (%T) %#v", res1.Value, res1.Value)
	}

	res2, err := JSONFeature("json-arr").EvaluateAny(ctx, client)
	if err != nil {
		t.Fatalf("EvaluateAny: %v", err)
	}
	if got, ok := res2.Value.([]any); !ok || len(got) != 3 {
		t.Fatalf("expected []any len=3, got (%T) %#v", res2.Value, res2.Value)
	}

	res3, err := JSONFeature("json-obj").EvaluateAny(ctx, client)
	if err != nil {
		t.Fatalf("EvaluateAny: %v", err)
	}
	if got, ok := res3.Value.(map[string]any); !ok || got["k"] != "v" {
		t.Fatalf("expected map[k]=v, got (%T) %#v", res3.Value, res3.Value)
	}

	if v, ok := JSONFeature("json-str").GetAny(ctx, client); !ok || v.(string) != "hello" {
		t.Fatalf("expected GetAny=(\"hello\", true), got (%#v, %v)", v, ok)
	}
	if v := JSONFeature("missing").GetAnyOr(ctx, client, "fallback"); v != "fallback" {
		t.Fatalf("expected GetAnyOr default, got %#v", v)
	}
}

func TestJSONFeature_Decode_Generic(t *testing.T) {
	type Config struct {
		Currency string `json:"currency"`
		MaxItems int    `json:"maxItems"`
	}

	ctx := context.Background()

	client, err := growthbook.NewClient(ctx, growthbook.WithJsonFeatures(`{
		"checkout-config": {"defaultValue": {"currency":"USD","maxItems":3}},
		"checkout-config-array": {"defaultValue": [1,2,3]}
	}`))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	cfg, ok := WithType[Config](JSONFeature("checkout-config")).Get(ctx, client)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if cfg.Currency != "USD" || cfg.MaxItems != 3 {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}

	// Decoding array into struct should fail.
	def := Config{Currency: "DEF", MaxItems: 9}
	cfg2 := WithType[Config](JSONFeature("checkout-config-array")).GetOr(ctx, client, def)
	if cfg2 != def {
		t.Fatalf("expected default on decode failure, got %#v", cfg2)
	}
}
