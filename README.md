# gbgen

`gbgen` is a CLI tool that generates Go code from GrowthBook feature definitions.

It supports:
- Generating **feature key constants** (default)
- Generating **typed feature helpers** (optional) that evaluate via `github.com/growthbook/growthbook-golang`

## GrowthBook links

- **GrowthBook docs**: [growthbook.io/docs](https://growthbook.io/docs)
- **GrowthBook GitHub**: [github.com/growthbook/growthbook](https://github.com/growthbook/growthbook)
- **GrowthBook Go SDK**: [github.com/growthbook/growthbook-golang](https://github.com/growthbook/growthbook-golang)

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

## Prerequisite: GrowthBook API Secret Key (read-only)

To let `gbgen` read feature definitions from GrowthBook, you need an API key.

Recommended: create an **organization Secret Key** with the **`readonly`** role (least privilege).

- **Where to create**: GrowthBook UI → `Settings` → `API Keys` → create a new **Secret Key**
- **Permissions**: choose **read-only** (`readonly`)
- **How it’s used**: the key is sent to the GrowthBook REST API via either:
  - **Bearer auth**: `Authorization: Bearer <secret_key>`
  - **HTTP Basic auth**: username=`<secret_key>`, empty password

For the official authentication options and examples, see [GrowthBook API Authentication docs](https://docs.growthbook.io/api/#section/Authentication).

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
- `types.JSONFeature` provides convenience helpers `Object()` / `Array()` / `String()` / `Number()` / `Boolean()` that reinterpret the same feature key as other typed wrappers (no conversion; mismatches fail at evaluation time).

If you want to decode a JSON feature into a struct/slice/map of your choice, use `types.AsType[T]`:

```go
type CheckoutConfig struct {
	Currency string `json:"currency"`
	MaxItems int    `json:"maxItems"`
}

// Recommended: pass the generated typed feature variable directly.
// (This requires generator.emitTypedFeatures=true)
cfg := types.AsType[CheckoutConfig](features.FeatureCheckoutConfig).GetOr(ctx, client, CheckoutConfig{})
```

More `AsType[T]` patterns:

- Decode with explicit error handling:

```go
type CheckoutConfig struct {
	Currency string `json:"currency"`
	MaxItems int    `json:"maxItems"`
}

res, err := types.AsType[CheckoutConfig](features.FeatureCheckoutConfig).Evaluate(ctx, client)
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
cfg, ok := types.AsType[CheckoutConfig](features.FeatureCheckoutConfig).Get(ctx, client)
_ = ok

cfg = types.AsType[CheckoutConfig](features.FeatureCheckoutConfig).GetOr(ctx, client, CheckoutConfig{})
```

- Decode JSON arrays / maps:

```go
items := types.AsType[[]string](features.FeatureAllowedItems).GetOr(ctx, client, nil)
weights := types.AsType[map[string]float64](features.FeatureWeights).GetOr(ctx, client, nil)
```

If you just want the SDK-decoded JSON array shape (`[]any`), you can also use:

```go
itemsAny := features.FeatureAllowedItems.Array().GetOr(ctx, client, nil)
_ = itemsAny
```

If you don't have generated typed feature variables available, you can still use `AsType[T]` with a manual key:

```go
cfg := types.AsType[CheckoutConfig](types.JSONFeature("checkout-config")).GetOr(ctx, client, CheckoutConfig{})
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

## Release & build metadata

- **Go install**: tagged releases support `go install github.com/eastnine90/gbgen@latest`.
- **Version stamping**: the binary supports build-time version metadata via `-ldflags`:

```bash
go build -ldflags "\
  -X github.com/eastnine90/gbgen/internal/buildinfo.Version=v0.1.0 \
  -X github.com/eastnine90/gbgen/internal/buildinfo.Commit=$(git rev-parse HEAD)"
./gbgen version
```

- **SLSA (prepared)**: this repo includes a `.slsa-goreleaser.yml` build config for the SLSA Go builder workflow (to be wired in later).

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


