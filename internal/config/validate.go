package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError is a user-facing error that lists config problems.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	if len(e.Problems) == 0 {
		return "invalid configuration"
	}
	var b strings.Builder
	b.WriteString("invalid configuration:\n")
	for _, p := range e.Problems {
		b.WriteString(" - ")
		b.WriteString(p)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// Validate validates the config using struct tags and returns a user-friendly error.
func (c Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(c); err != nil {
		return humanizeValidationError(err)
	}
	return nil
}

func humanizeValidationError(err error) error {
	var ve validator.ValidationErrors
	if !strings.Contains(err.Error(), "ValidationErrors") {
		// Fall back to the raw error if it isn't a standard validation error.
		if ok := asValidationErrors(err, &ve); !ok {
			return err
		}
	}
	if ok := asValidationErrors(err, &ve); !ok {
		return err
	}

	problems := make([]string, 0, len(ve))
	for _, fe := range ve {
		path := toConfigPath(fe.StructNamespace())
		switch fe.Tag() {
		case "required":
			problems = append(problems, fmt.Sprintf("%s is required", path))
		case "url":
			problems = append(problems, fmt.Sprintf("%s must be a valid URL (e.g. https://api.growthbook.io)", path))
		default:
			problems = append(problems, fmt.Sprintf("%s is invalid (%s)", path, fe.Tag()))
		}
	}
	return &ValidationError{Problems: problems}
}

func asValidationErrors(err error, out *validator.ValidationErrors) bool {
	ve, ok := err.(validator.ValidationErrors)
	if ok {
		*out = ve
		return true
	}
	return false
}

func toConfigPath(structNamespace string) string {
	// Example: "Config.GrowthBook.APIBaseURL"
	s := strings.TrimPrefix(structNamespace, "Config.")
	s = strings.TrimPrefix(s, "Config.")
	s = strings.ReplaceAll(s, ".", ".")

	// Map struct field names to config key casing.
	s = strings.ReplaceAll(s, "GrowthBook.", "growthbook.")
	s = strings.ReplaceAll(s, "Generator.", "generator.")
	s = strings.ReplaceAll(s, "APIBaseURL", "apiBaseURL")
	s = strings.ReplaceAll(s, "APIKey", "apiKey")
	s = strings.ReplaceAll(s, "ProjectID", "projectID")
	s = strings.ReplaceAll(s, "OutputDir", "outputDir")
	s = strings.ReplaceAll(s, "PackageName", "packageName")

	return s
}
