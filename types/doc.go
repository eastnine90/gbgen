// Package types contains the public runtime types used by gbgen's generated code.
//
// Generated feature variables are thin wrappers around GrowthBook feature keys that provide:
//   - Typed evaluation helpers (e.g. BooleanFeature, StringFeature, NumberFeature, JSONFeature, ArrayFeature)
//   - Structured type mismatch errors (TypeMismatchError) instead of panics
//   - Optional "happy-path" helpers (Get / GetOr) that never return errors
//
// JSON features:
//   - JSONFeature is strict and expects a JSON object (map[string]any).
//   - JSONFeature also provides EvaluateAny/GetAny/GetAnyOr helpers for any JSON shape (map/slice/string/number/bool/nil).
//   - JSONFeature provides convenience "cast" helpers (Object/Array/String/Number/Boolean) that reinterpret the same feature key
//     as other typed wrappers. These helpers do not convert values; mismatches are surfaced at evaluation time.
//   - TypedFeature (AsType[T]) can decode a feature value into a caller-provided type parameter T.
//     Missing feature keys return ErrMissingKey. The Get/GetOr helpers treat ErrMissingKey as a normal failure.
//
// All Evaluate/Get/GetOr helpers accept optional per-evaluation attributes (growthbook.Attributes) to apply on top of the client's base attributes.
// However, by the design of growthbook-golang, any prior attributes of the client would be ignored if optional attributes were passed.
//
// Note: numeric values in JSON are decoded as float64 by the GrowthBook Go SDK.
//
// Example:
//
//	res, err := FeatureMyFlag.Evaluate(ctx, client)
//	if err != nil {
//	  // errors.Is(err, types.ErrMissingKey) can be used to detect missing keys
//	  // errors.Is(err, types.ErrTypeMismatch) can be used to detect decode failures
//	}
//	if res.IsValid() {
//	  _ = res.GetValue()
//	  // Extra metadata (rule id, source, experiment info) is available via res.GetRaw().
//	}
//
// TypeMismatchError can happen if the feature's value type does not match what your code expects.
// In practice, GrowthBook usually treats a feature's valueType as immutable, so mismatches commonly
// indicate stale generated code (or a feature being deleted/re-created with a different type),
// overrides returning an unexpected type, or JSON values that aren't objects.
//
// For convenience in code paths that prefer defaults over errors:
//
//	v := FeatureMyFlag.GetOr(ctx, client, false)
//
// Decoding JSON into a struct:
//
//	type Config struct {
//	  Currency string `json:"currency"`
//	  MaxItems int    `json:"maxItems"`
//	}
//	cfg := AsType[Config](JSONFeature("checkout-config")).GetOr(ctx, client, Config{})
package types
