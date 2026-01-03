package types

import (
	"context"

	"github.com/growthbook/growthbook-golang"
)

// NumberFeature is a typed wrapper for a numeric GrowthBook feature.
// GrowthBook numeric feature values are decoded as float64 by the SDK.
type NumberFeature string

// Key returns the underlying GrowthBook feature key.
func (f NumberFeature) Key() string {
	return string(f)
}

// Evaluate evaluates the feature using the provided GrowthBook client and optional attributes.
func (f NumberFeature) Evaluate(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[float64], err error) {
	r, err := evaluateWithAttrs(ctx, client, string(f), attrs...)
	if err != nil {
		return result, err
	}

	if _, ok := r.Value.(float64); !ok {
		return result, newTypeMismatchError(string(f), "float64", r.Value)
	}

	return FeatureResult[float64]{
		Raw:   r,
		Value: r.Value.(float64),
		Valid: true,
	}, nil
}

// Get evaluates the feature and returns (value, ok) for happy-path usage.
// It never returns an error; any error or type mismatch results in ok=false.
func (f NumberFeature) Get(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value float64, ok bool) {
	res, err := f.Evaluate(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return 0, false
	}
	return res.Value, true
}

// GetOr evaluates the feature and returns defaultValue if evaluation fails or the value cannot be decoded.
func (f NumberFeature) GetOr(ctx context.Context, client *growthbook.Client, defaultValue float64, attrs ...growthbook.Attributes) float64 {
	if v, ok := f.Get(ctx, client, attrs...); ok {
		return v
	}
	return defaultValue
}
