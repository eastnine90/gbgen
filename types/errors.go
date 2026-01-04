package types

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrTypeMismatch is a sentinel error for quick checks via errors.Is.
var ErrTypeMismatch = errors.New("type mismatch")
var ErrMissingResult = errors.New("missing feature result from client")
var ErrMissingKey = errors.New("missing feature key")

// TypeMismatchError is returned when a feature value cannot be decoded into the expected Go type.
// It is intended to be used by Evaluate/decoder code paths instead of panicking on type assertions.
type TypeMismatchError struct {
	FeatureKey string
	Expected   string
	ActualType string
}

func (e *TypeMismatchError) Error() string {
	if e == nil {
		return "type mismatch"
	}
	if e.FeatureKey == "" {
		return fmt.Sprintf("type mismatch: expected %s, got %s", e.Expected, e.ActualType)
	}
	return fmt.Sprintf("type mismatch for feature %q: expected %s, got %s", e.FeatureKey, e.Expected, e.ActualType)
}

// Unwrap allows errors.Is(err, ErrTypeMismatch) to match TypeMismatchError.
func (e *TypeMismatchError) Unwrap() error {
	return ErrTypeMismatch
}

// newTypeMismatchError builds a TypeMismatchError with a best-effort actual type string.
func newTypeMismatchError(featureKey, expected string, actual any) *TypeMismatchError {
	actualType := "nil"
	if actual != nil {
		actualType = reflect.TypeOf(actual).String()
	}
	return &TypeMismatchError{
		FeatureKey: featureKey,
		Expected:   expected,
		ActualType: actualType,
	}
}
