package types

import (
	"context"

	"github.com/growthbook/growthbook-golang"
)

// JSONFeature is a typed wrapper for a JSON GrowthBook feature.
// The value is exposed as map[string]any, matching the GrowthBook Go SDK decoding.
type JSONFeature string

// Key returns the underlying GrowthBook feature key.
func (f JSONFeature) Key() string {
	return string(f)
}

// Object returns the same feature key as a JSONFeature.
//
// This is an identity helper for readability when chaining.
func (f JSONFeature) Object() JSONFeature {
	return f
}

// Array returns the same feature key as an ArrayFeature (JSON array decoded as []any).
//
// Note: this does not convert or validate values; it only reinterprets the key.
func (f JSONFeature) Array() ArrayFeature {
	return ArrayFeature(f)
}

// String returns the same feature key as a StringFeature.
//
// Note: this does not convert or validate values; it only reinterprets the key.
func (f JSONFeature) String() StringFeature {
	return StringFeature(f)
}

// Number returns the same feature key as a NumberFeature (decoded as float64).
//
// Note: this does not convert or validate values; it only reinterprets the key.
func (f JSONFeature) Number() NumberFeature {
	return NumberFeature(f)
}

// Boolean returns the same feature key as a BooleanFeature.
//
// Note: this does not convert or validate values; it only reinterprets the key.
func (f JSONFeature) Boolean() BooleanFeature {
	return BooleanFeature(f)
}

// Evaluate evaluates the feature using the provided GrowthBook client and optional attributes.
func (f JSONFeature) Evaluate(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[map[string]any], err error) {
	r, err := evaluateWithAttrs(ctx, client, string(f), attrs...)
	if err != nil {
		return result, err
	}

	if _, ok := r.Value.(map[string]any); !ok {
		return result, newTypeMismatchError(string(f), "map[string]any", r.Value)
	}

	return FeatureResult[map[string]any]{
		Raw:   r,
		Value: r.Value.(map[string]any),
		Valid: true,
	}, nil
}

// EvaluateAny evaluates the feature and returns the underlying decoded JSON value (any JSON shape).
// This differs from Evaluate, which is intentionally strict and only accepts JSON objects (map[string]any).
func (f JSONFeature) EvaluateAny(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[any], err error) {
	r, err := evaluateWithAttrs(ctx, client, string(f), attrs...)
	if err != nil {
		return result, err
	}

	return FeatureResult[any]{
		Raw:   r,
		Value: r.Value,
		Valid: true,
	}, nil
}

// Get evaluates the feature and returns (value, ok) for happy-path usage.
// It never returns an error; any error or type mismatch results in ok=false.
func (f JSONFeature) Get(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value map[string]any, ok bool) {
	res, err := f.Evaluate(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return nil, false
	}
	return res.Value, true
}

// GetOr evaluates the feature and returns defaultValue if evaluation fails or the value cannot be decoded.
func (f JSONFeature) GetOr(ctx context.Context, client *growthbook.Client, defaultValue map[string]any, attrs ...growthbook.Attributes) map[string]any {
	if v, ok := f.Get(ctx, client, attrs...); ok {
		return v
	}
	return defaultValue
}

// GetAny evaluates the feature and returns (value, ok) for happy-path usage (any JSON shape).
// It never returns an error; any error results in ok=false.
func (f JSONFeature) GetAny(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value any, ok bool) {
	res, err := f.EvaluateAny(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return nil, false
	}
	return res.Value, true
}

// GetAnyOr evaluates the feature and returns defaultValue if evaluation fails.
func (f JSONFeature) GetAnyOr(ctx context.Context, client *growthbook.Client, defaultValue any, attrs ...growthbook.Attributes) any {
	res, err := f.EvaluateAny(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return defaultValue
	}
	return res.Value
}
