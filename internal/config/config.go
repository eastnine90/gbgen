package config

// Config is the single merged configuration for gbgen.
//
// NOTE: This is intentionally "definition-only" for now (no validation, no loading).
// It is tagged to support unmarshalling from JSON, YAML, or TOML.
type Config struct {
	GrowthBook GrowthBookConfig `json:"growthbook" yaml:"growthbook" toml:"growthbook"`
	Generator  GeneratorConfig  `json:"generator"  yaml:"generator"  toml:"generator"`
}

type GrowthBookConfig struct {
	APIBaseURL string  `json:"apiBaseURL" yaml:"apiBaseURL" toml:"apiBaseURL" validate:"required,url"`
	APIKey     string  `json:"apiKey"     yaml:"apiKey"     toml:"apiKey"     validate:"required"`
	ProjectID  *string `json:"projectID"  yaml:"projectID"  toml:"projectID"`
}

type GeneratorConfig struct {
	OutputDir         string `json:"outputDir"         yaml:"outputDir"         toml:"outputDir"         validate:"required"`
	PackageName       string `json:"packageName"       yaml:"packageName"       toml:"packageName"       validate:"required"`
	EmitTypedFeatures bool   `json:"emitTypedFeatures" yaml:"emitTypedFeatures" toml:"emitTypedFeatures"`
	EmitFeatureList   bool   `json:"emitFeatureList"   yaml:"emitFeatureList"   toml:"emitFeatureList"`
}
