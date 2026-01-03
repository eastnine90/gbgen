package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromYAML_ThenEnvOverrides(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "gbgen.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
growthbook:
  apiBaseURL: "https://from-file.example"
  apiKey: "file-key"
generator:
  outputDir: "./file-out"
  packageName: "filepkg"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GBGEN_API_KEY", "env-key")
	t.Setenv("GBGEN_PACKAGE_NAME", "envpkg")

	got, err := Load(LoadOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if got.GrowthBook.APIBaseURL != "https://from-file.example" {
		t.Fatalf("apiBaseURL = %q", got.GrowthBook.APIBaseURL)
	}
	if got.GrowthBook.APIKey != "env-key" {
		t.Fatalf("apiKey = %q", got.GrowthBook.APIKey)
	}
	if got.Generator.PackageName != "envpkg" {
		t.Fatalf("packageName = %q", got.Generator.PackageName)
	}
}

func TestLoad_OverridesWin(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "gbgen.json")
	if err := os.WriteFile(cfgPath, []byte(`{"growthbook":{"apiBaseURL":"https://file"},"generator":{"outputDir":"./file","packageName":"file"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GBGEN_OUTPUT_DIR", "./env")
	overrideOut := "./override"

	got, err := Load(LoadOptions{
		ConfigPath: cfgPath,
		Overrides: Overrides{
			OutputDir: &overrideOut,
		},
	})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got.Generator.OutputDir != "./override" {
		t.Fatalf("outputDir = %q", got.Generator.OutputDir)
	}
}

func TestLoad_EmitTypedFeatures_FileTrue_EnvFalseWins(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "gbgen.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
generator:
  emitTypedFeatures: true
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GBGEN_EMIT_TYPED_FEATURES", "false")

	got, err := Load(LoadOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got.Generator.EmitTypedFeatures != false {
		t.Fatalf("emitTypedFeatures = %v", got.Generator.EmitTypedFeatures)
	}
}

func TestLoad_EmitFeatureList_FileTrue_EnvFalseWins(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "gbgen.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
generator:
  emitFeatureList: true
`), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GBGEN_EMIT_FEATURE_LIST", "false")

	got, err := Load(LoadOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got.Generator.EmitFeatureList != false {
		t.Fatalf("emitFeatureList = %v", got.Generator.EmitFeatureList)
	}
}
