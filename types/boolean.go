package types

import (
	"context"

	"github.com/growthbook/growthbook-golang"
)

// BooleanFeature is a typed wrapper for a boolean GrowthBook feature.
type BooleanFeature string

// Key returns the underlying GrowthBook feature key.
func (f BooleanFeature) Key() string {
	return string(f)
}

// Evaluate evaluates the feature using the provided GrowthBook client and optional attributes.
func (f BooleanFeature) Evaluate(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[bool], err error) {
	r, err := evaluateWithAttrs(ctx, client, string(f), attrs...)
	if err != nil {
		return result, err
	}

	if r == nil {
		return result, ErrMissingResult
	}

	if _, ok := r.Value.(bool); !ok {
		return result, newTypeMismatchError(string(f), "bool", r.Value)
	}

	return FeatureResult[bool]{
		Raw:   r,
		Value: r.On,
		Valid: true,
	}, nil
}

// Get evaluates the feature and returns (value, ok) for happy-path usage.
// It never returns an error; any error or type mismatch results in ok=false.
func (f BooleanFeature) Get(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value bool, ok bool) {
	res, err := f.Evaluate(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return false, false
	}
	return res.Value, true
}

// GetOr evaluates the feature and returns defaultValue if evaluation fails or the value cannot be decoded.
func (f BooleanFeature) GetOr(ctx context.Context, client *growthbook.Client, defaultValue bool, attrs ...growthbook.Attributes) bool {
	if v, ok := f.Get(ctx, client, attrs...); ok {
		return v
	}
	return defaultValue
}
