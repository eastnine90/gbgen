// Package generator fetches feature definitions from the GrowthBook API and renders Go source.
//
// Output behavior:
// - Always writes a single file named "features.gen.go" (the file content varies by config).
// - Generated identifiers are derived from feature IDs and deduplicated when needed.
package generator


