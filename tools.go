//go:build tools

package tools

// This file tracks build-time tooling dependencies in go.mod.
// It is excluded from normal builds via the "tools" build tag.

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)


