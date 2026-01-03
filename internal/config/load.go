package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Overrides represents explicit values (e.g., CLI flags) that override config/env/defaults.
// A nil pointer means "not set".
type Overrides struct {
	APIBaseURL        *string
	APIKey            *string
	ProjectID         *string
	OutputDir         *string
	PackageName       *string
	EmitTypedFeatures *bool
	EmitFeatureList   *bool
}

// Load builds the final config using the following precedence (highest wins):
// overrides > env > config file > defaults.
//
// This performs no validation. Call Config.Validate() separately if desired.
func Load(opts LoadOptions) (Config, error) {
	cfg := Defaults()

	// 1) Config file
	if opts.ConfigPath != "" {
		fileCfg, err := loadFromFile(opts.ConfigPath)
		if err != nil {
			return Config{}, err
		}
		cfg = merge(cfg, fileCfg)
	}

	// 2) Env
	cfg = applyEnv(cfg, opts.EnvPrefix)

	// 3) Explicit overrides (e.g. CLI flags)
	cfg = applyOverrides(cfg, opts.Overrides)

	return cfg, nil
}

type LoadOptions struct {
	ConfigPath string
	EnvPrefix  string
	Overrides  Overrides
}

func Defaults() Config {
	return Config{
		GrowthBook: GrowthBookConfig{
			APIBaseURL: "https://api.growthbook.io",
			APIKey:     "",
			ProjectID:  nil,
		},
		Generator: GeneratorConfig{
			OutputDir:         "./internal/growthbooktypes",
			PackageName:       "growthbooktypes",
			EmitTypedFeatures: false,
			EmitFeatureList:   false,
		},
	}
}

func loadFromFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var f Format
	switch ext {
	case ".json":
		f = FormatJSON
	case ".yml", ".yaml":
		f = FormatYAML
	case ".toml":
		f = FormatTOML
	default:
		return Config{}, fmt.Errorf("unsupported config extension %q (expected .json/.yaml/.yml/.toml)", ext)
	}

	return Unmarshal(b, f)
}

func merge(base, overlay Config) Config {
	out := base

	// GrowthBook
	if overlay.GrowthBook.APIBaseURL != "" {
		out.GrowthBook.APIBaseURL = overlay.GrowthBook.APIBaseURL
	}
	if overlay.GrowthBook.APIKey != "" {
		out.GrowthBook.APIKey = overlay.GrowthBook.APIKey
	}
	if overlay.GrowthBook.ProjectID != nil {
		out.GrowthBook.ProjectID = overlay.GrowthBook.ProjectID
	}

	// Generator
	if overlay.Generator.OutputDir != "" {
		out.Generator.OutputDir = overlay.Generator.OutputDir
	}
	if overlay.Generator.PackageName != "" {
		out.Generator.PackageName = overlay.Generator.PackageName
	}
	// Note: for config-file unmarshalling, we can't distinguish "unset" vs false.
	// We intentionally only treat "true" as an override here; env/overrides can turn it off explicitly.
	if overlay.Generator.EmitTypedFeatures {
		out.Generator.EmitTypedFeatures = true
	}
	if overlay.Generator.EmitFeatureList {
		out.Generator.EmitFeatureList = true
	}

	return out
}

func applyEnv(cfg Config, prefix string) Config {
	p := strings.TrimSpace(prefix)
	if p == "" {
		p = "GBGEN"
	}
	key := func(s string) string { return p + "_" + s }

	if v := os.Getenv(key("API_BASE_URL")); v != "" {
		cfg.GrowthBook.APIBaseURL = v
	}
	if v := os.Getenv(key("API_KEY")); v != "" {
		cfg.GrowthBook.APIKey = v
	}
	if v := os.Getenv(key("PROJECT_ID")); v != "" {
		tmp := v
		cfg.GrowthBook.ProjectID = &tmp
	}
	if v := os.Getenv(key("OUTPUT_DIR")); v != "" {
		cfg.Generator.OutputDir = v
	}
	if v := os.Getenv(key("PACKAGE_NAME")); v != "" {
		cfg.Generator.PackageName = v
	}
	if v := os.Getenv(key("EMIT_TYPED_FEATURES")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Generator.EmitTypedFeatures = b
		}
	}
	if v := os.Getenv(key("EMIT_FEATURE_LIST")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Generator.EmitFeatureList = b
		}
	}

	return cfg
}

func applyOverrides(cfg Config, o Overrides) Config {
	if o.APIBaseURL != nil {
		cfg.GrowthBook.APIBaseURL = *o.APIBaseURL
	}
	if o.APIKey != nil {
		cfg.GrowthBook.APIKey = *o.APIKey
	}
	if o.ProjectID != nil {
		cfg.GrowthBook.ProjectID = o.ProjectID
	}
	if o.OutputDir != nil {
		cfg.Generator.OutputDir = *o.OutputDir
	}
	if o.PackageName != nil {
		cfg.Generator.PackageName = *o.PackageName
	}
	if o.EmitTypedFeatures != nil {
		cfg.Generator.EmitTypedFeatures = *o.EmitTypedFeatures
	}
	if o.EmitFeatureList != nil {
		cfg.Generator.EmitFeatureList = *o.EmitFeatureList
	}
	return cfg
}
