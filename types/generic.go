package types

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/growthbook/growthbook-golang"
)

// FeatureResult is the typed result of evaluating a feature.
// Valid indicates whether evaluation succeeded (it may be false if the value could not be decoded).
type FeatureResult[T any] struct {
	// Raw is the underlying GrowthBook SDK result for access to metadata such as
	// rule id, source, and experiment information.
	Raw   *growthbook.FeatureResult
	Value T
	Valid bool
}

// GetValue returns the typed feature value.
func (f *FeatureResult[T]) GetValue() T {
	return f.Value
}

// IsValid reports whether the result contains a valid typed value.
func (f *FeatureResult[T]) IsValid() bool {
	return f.Valid
}

// GetExperimentResult returns the underlying GrowthBook experiment result (if any).
func (f *FeatureResult[T]) GetExperimentResult() *growthbook.ExperimentResult {
	if f == nil || f.Raw == nil {
		return nil
	}
	return f.Raw.ExperimentResult
}

// GetRaw returns the underlying GrowthBook SDK result (if available).
func (f *FeatureResult[T]) GetRaw() *growthbook.FeatureResult {
	if f == nil {
		return nil
	}
	return f.Raw
}

func evaluateWithAttrs(ctx context.Context, client *growthbook.Client, key string, attrs ...growthbook.Attributes) (result *growthbook.FeatureResult, err error) {
	for _, attr := range attrs {
		if client, err = client.WithAttributes(attr); err != nil {
			return nil, err
		}
	}

	r := client.EvalFeature(ctx, key)

	if r == nil {
		return nil, ErrMissingResult
	}

	// Missing keys are returned as an explicit error so callers can differentiate
	// between "missing feature" and "type mismatch".
	if r.Source == growthbook.UnknownFeatureResultSource {
		return r, ErrMissingKey
	}

	return r, nil
}

type TypedFeature[T any] struct {
	featureKey
}

type featureKey interface {
	Key() string
}

func WithType[T any](f featureKey) TypedFeature[T] {
	return TypedFeature[T]{
		featureKey: f,
	}
}

func (f TypedFeature[T]) Key() string {
	return f.featureKey.Key()
}

// Evaluate evaluates the feature and decodes the underlying value into T using encoding/json.
//
// This is useful for JSON features when you want to decode into a caller-chosen struct/slice/etc.
// It performs a JSON round-trip (Marshal + Unmarshal).
//
// If the feature key is missing from the loaded definitions, Evaluate returns ErrMissingKey.
// Decode failures return TypeMismatchError (errors.Is(err, ErrTypeMismatch) == true).
func (f TypedFeature[T]) Evaluate(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[T], err error) {
	r, err := evaluateWithAttrs(ctx, client, f.Key(), attrs...)
	if err != nil {
		return result, err
	}
	result.Raw = r

	valueByte, err := json.Marshal(r.Value)
	if err != nil {
		return result, err
	}

	var value T
	if err := json.Unmarshal(valueByte, &value); err != nil {
		expected := reflect.TypeOf((*T)(nil)).Elem().String()
		return result, newTypeMismatchError(f.Key(), expected, r.Value)
	}

	return FeatureResult[T]{
		Raw:   r,
		Value: value,
		Valid: true,
	}, nil
}

// Get evaluates and decodes the feature and returns (value, ok) for happy-path usage.
// It never returns an error; any error (including ErrMissingKey) or decode failure results in ok=false.
func (f TypedFeature[T]) Get(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value T, ok bool) {
	res, err := f.Evaluate(ctx, client, attrs...)
	if err != nil || !res.Valid {
		var zero T
		return zero, false
	}
	return res.Value, true
}

// GetOr evaluates and decodes the feature and returns defaultValue on failure (including ErrMissingKey).
func (f TypedFeature[T]) GetOr(ctx context.Context, client *growthbook.Client, defaultValue T, attrs ...growthbook.Attributes) T {
	if v, ok := f.Get(ctx, client, attrs...); ok {
		return v
	}
	return defaultValue
}
