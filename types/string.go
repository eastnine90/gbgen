package types

import (
	"context"

	"github.com/growthbook/growthbook-golang"
)

// StringFeature is a typed wrapper for a string GrowthBook feature.
type StringFeature string

// Key returns the underlying GrowthBook feature key.
func (f StringFeature) Key() string {
	return string(f)
}

// Evaluate evaluates the feature using the provided GrowthBook client and optional attributes.
func (f StringFeature) Evaluate(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (result FeatureResult[string], err error) {
	r, err := evaluateWithAttrs(ctx, client, string(f), attrs...)
	if err != nil {
		return result, err
	}

	if _, ok := r.Value.(string); !ok {
		return result, newTypeMismatchError(string(f), "string", r.Value)
	}

	return FeatureResult[string]{
		Raw:   r,
		Value: r.Value.(string),
		Valid: true,
	}, nil
}

// Get evaluates the feature and returns (value, ok) for happy-path usage.
// It never returns an error; any error or type mismatch results in ok=false.
func (f StringFeature) Get(ctx context.Context, client *growthbook.Client, attrs ...growthbook.Attributes) (value string, ok bool) {
	res, err := f.Evaluate(ctx, client, attrs...)
	if err != nil || !res.Valid {
		return "", false
	}
	return res.Value, true
}

// GetOr evaluates the feature and returns defaultValue if evaluation fails or the value cannot be decoded.
func (f StringFeature) GetOr(ctx context.Context, client *growthbook.Client, defaultValue string, attrs ...growthbook.Attributes) string {
	if v, ok := f.Get(ctx, client, attrs...); ok {
		return v
	}
	return defaultValue
}
