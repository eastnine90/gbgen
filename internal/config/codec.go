package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
)

// DetectFormat resolves the config format from an explicit override or a file extension.
// If both are empty/unknown, it defaults to YAML.
func DetectFormat(path string, override string) (Format, error) {
	if strings.TrimSpace(override) != "" {
		return normalizeFormat(override)
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FormatJSON, nil
	case ".yml", ".yaml":
		return FormatYAML, nil
	case ".toml":
		return FormatTOML, nil
	case "":
		return FormatYAML, nil
	default:
		// Defaulting here is friendlier for init use-cases; callers can choose to error instead.
		return FormatYAML, nil
	}
}

func Marshal(cfg Config, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(cfg, "", "  ")
	case FormatYAML:
		return yaml.Marshal(cfg)
	case FormatTOML:
		return toml.Marshal(cfg)
	default:
		return nil, fmt.Errorf("unsupported format %q (expected json|yaml|toml)", format)
	}
}

func Unmarshal(b []byte, format Format) (Config, error) {
	var cfg Config
	switch format {
	case FormatJSON:
		if err := json.Unmarshal(b, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse json config: %w", err)
		}
	case FormatYAML:
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse yaml config: %w", err)
		}
	case FormatTOML:
		if err := toml.Unmarshal(b, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse toml config: %w", err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported format %q (expected json|yaml|toml)", format)
	}
	return cfg, nil
}

func normalizeFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "toml":
		return FormatTOML, nil
	default:
		return "", fmt.Errorf("unsupported format %q (expected json|yaml|toml)", s)
	}
}
