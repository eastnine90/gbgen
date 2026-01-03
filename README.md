# gbgen

`gbgen` is a CLI tool that generates Go code from GrowthBook feature definitions.

It supports:
- Generating **feature key constants** (default)
- Generating **typed feature helpers** (optional) that evaluate via `github.com/growthbook/growthbook-golang`

## Install

```bash
go install github.com/eastnine90/gbgen@latest
```

## Quickstart

Create a config file:

```bash
gbgen init --config gbgen.yaml
```

Generate code:

```bash
gbgen generate --config gbgen.yaml
```

Output:
- Always writes a single file named **`features.gen.go`** into `generator.outputDir` (overwritten on every run).

## Configuration

Configuration sources are applied with this precedence:

**CLI flags > environment variables > config file > defaults**

Supported config formats: **JSON**, **YAML**, **TOML**.

Environment variable prefix: `GBGEN_` (e.g. `GBGEN_API_KEY`).

## Typed vs keys-only generation

- **Keys-only (default)**: emits `type FeatureKey string` and constants like `FeatureCheckoutRedesign`.
- **Typed (`generator.emitTypedFeatures=true`)**: emits typed vars like `FeatureCheckoutRedesign = types.BooleanFeature("checkout-redesign")`.
- **Feature list**: `generator.emitFeatureList=true` also emits `FeatureList` containing all feature keys.

## Using typed features

When `generator.emitTypedFeatures=true`, the generated file contains typed feature variables that can be evaluated using the GrowthBook Go SDK.
All helpers accept optional per-evaluation attributes (`growthbook.Attributes`) to apply on top of the client's base attributes.

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/eastnine90/gbgen/types"
	"github.com/growthbook/growthbook-golang"

	// Replace with your generated package path (generator.packageName).
	// Example: "github.com/you/yourrepo/internal/features"
	features "path/to/your/generated/package"
)

func main() {
	ctx := context.Background()

	// Initialize GrowthBook client however you normally do.
	var client *growthbook.Client

	res, err := features.FeatureCheckoutRedesign.Evaluate(ctx, client, growthbook.Attributes{"id": "u1"})
	if err != nil {
		if errors.Is(err, types.ErrMissingKey) {
			// Feature key is not present in the loaded feature definitions.
			// (Generated code may be ahead of the GrowthBook environment you are running against.)
			fmt.Println(err)
			return
		}
		if errors.Is(err, types.ErrTypeMismatch) {
			// Feature valueType changed in GrowthBook (or unexpected data) and can’t be decoded.
			fmt.Println(err)
			return
		}
		// Transport/config errors from the SDK (or client attribute errors).
		panic(err)
	}

	if res.IsValid() {
		fmt.Println("checkout redesign enabled:", res.GetValue())

		// Optional: access evaluation metadata (rule id, source, experiment info).
		if raw := res.GetRaw(); raw != nil {
			_ = raw.Source
			_ = raw.RuleId
			_ = raw.ExperimentResult
		}
	}
}
```

If you prefer a "happy path" API (no errors), use `Get`/`GetOr`:

```go
enabled := features.FeatureCheckoutRedesign.GetOr(ctx, client, false, growthbook.Attributes{"id": "u1"})
if v, ok := features.FeatureCheckoutRedesign.Get(ctx, client, growthbook.Attributes{"id": "u1"}); ok {
	_ = v
}
```

### JSON features: object-only vs any JSON

GrowthBook's `JSON` value type can represent **any JSON shape**.

In gbgen:
- `types.JSONFeature` is **strict** and expects a JSON object (`map[string]any`). If the value is a string/array/number/etc, it returns `types.ErrTypeMismatch`.
- `types.JSONFeature` also provides `EvaluateAny` / `GetAny` / `GetAnyOr` to accept any JSON value (`any`).

If you want to decode a JSON feature into a struct/slice/map of your choice, use `types.WithType[T]`:

```go
type CheckoutConfig struct {
	Currency string `json:"currency"`
	MaxItems int    `json:"maxItems"`
}

cfg := types.WithType[CheckoutConfig](types.JSONFeature("checkout-config")).GetOr(ctx, client, CheckoutConfig{})
```

More `WithType[T]` patterns:

- Decode with explicit error handling:

```go
type CheckoutConfig struct {
	Currency string `json:"currency"`
	MaxItems int    `json:"maxItems"`
}

res, err := types.WithType[CheckoutConfig](types.JSONFeature("checkout-config")).Evaluate(ctx, client)
if err != nil {
	if errors.Is(err, types.ErrMissingKey) {
		// The feature key is not present in the loaded definitions.
	}
	if errors.Is(err, types.ErrTypeMismatch) {
		// JSON shape doesn't match CheckoutConfig (or the feature type drifted).
	}
}
if res.IsValid() {
	_ = res.Value // CheckoutConfig
}
```

- Happy-path decode (no errors):

```go
cfg, ok := types.WithType[CheckoutConfig](types.JSONFeature("checkout-config")).Get(ctx, client)
_ = ok

cfg = types.WithType[CheckoutConfig](types.JSONFeature("checkout-config")).GetOr(ctx, client, CheckoutConfig{})
```

- Decode JSON arrays / maps:

```go
items := types.WithType[[]string](types.JSONFeature("allowed-items")).GetOr(ctx, client, nil)
weights := types.WithType[map[string]float64](types.JSONFeature("weights")).GetOr(ctx, client, nil)
```

### Number features

GrowthBook numeric feature values are decoded as `float64` by the GrowthBook Go SDK, so `types.NumberFeature` evaluates to `float64`.

### Type mismatch errors (when can it happen?)

Typed features assume that the feature's **value type** in GrowthBook matches what was generated.

In GrowthBook, the `valueType` is generally **not meant to change for an existing feature key** (the server enforces this in practice), so a `types.ErrTypeMismatch` usually indicates one of:

- The feature was **deleted and re-created** with the same key but a different type.
- An **override/forced value** (in GrowthBook rules or in SDK overrides) returns a different JSON type than expected.
- For `JSONFeature`, the underlying JSON value is **not an object** (e.g. it's an array/string/number), since `JSONFeature` exposes `map[string]any`.
- Your code is running against a **different GrowthBook instance/environment** than the one you generated from (stale generated code or mismatched config).

Recommended handling:
- Use `Evaluate(...)` when you want to **detect and log** mismatches.
- Use `GetOr(...)` when you want a **default on failure** and don't care why it failed.

If `generator.emitFeatureList=true`, the generated file also includes `FeatureList` with all feature keys.

## Development

Generate the GrowthBook API client (committed under `internal/growthbookapi`):

```bash
make gen-growthbookapi
```

Run unit tests:

```bash
go test ./...
```

Run local end-to-end tests (no external dependency; uses Docker Compose):

```bash
make test-e2e
```

## License

This project is licensed under the **MIT License** (see `LICENSE`).

The generated GrowthBook API client (`internal/growthbookapi`) is derived from GrowthBook's OpenAPI specification in the GrowthBook repository, which is licensed under **MIT Expat** for non-enterprise paths.


