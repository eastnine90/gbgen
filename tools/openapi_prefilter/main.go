package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var httpMethods = map[string]struct{}{
	"get":     {},
	"put":     {},
	"post":    {},
	"delete":  {},
	"options": {},
	"head":    {},
	"patch":   {},
	"trace":   {},
}

func main() {
	inPath := flag.String("in", "", "input OpenAPI YAML (bundled)")
	outPath := flag.String("out", "", "output OpenAPI YAML (minimized)")
	tag := flag.String("tag", "features", "only keep operations with this OpenAPI tag")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -in and -out are required")
		os.Exit(2)
	}

	inBytes, err := os.ReadFile(*inPath)
	must(err)

	var doc map[string]any
	must(yaml.Unmarshal(inBytes, &doc))

	pathsAny, ok := doc["paths"].(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "ERROR: openapi.paths missing or invalid")
		os.Exit(1)
	}

	filteredPaths := filterPathsByTag(pathsAny, *tag)

	// Build minimized document skeleton.
	minDoc := map[string]any{}
	copyKeyIfPresent(minDoc, doc, "openapi")
	copyKeyIfPresent(minDoc, doc, "info")
	copyKeyIfPresent(minDoc, doc, "servers")
	copyKeyIfPresent(minDoc, doc, "security")
	minDoc["paths"] = filteredPaths

	// Keep only the matching tag metadata (nice-to-have).
	if tagsAny, ok := doc["tags"].([]any); ok {
		var kept []any
		for _, t := range tagsAny {
			if m, ok := t.(map[string]any); ok {
				if name, _ := m["name"].(string); name == *tag {
					kept = append(kept, m)
				}
			}
		}
		if len(kept) > 0 {
			minDoc["tags"] = kept
		}
	}

	componentsAny, _ := doc["components"].(map[string]any)
	minComponents := buildMinComponents(filteredPaths, componentsAny, doc)
	if len(minComponents) > 0 {
		minDoc["components"] = minComponents
	}

	// Some upstream schemas define enums inline on properties (e.g. Feature.valueType).
	// The Go openapi-generator commonly does NOT emit a dedicated enum type + constants for
	// inline property enums, but it does for named schemas. Promote the known inline enum
	// into a named component schema and replace inline definitions with a $ref across the
	// entire minimized document (including inline request/response body schemas).
	promoteValueTypeEnum(minDoc)

	outBytes, err := yaml.Marshal(minDoc)
	must(err)
	must(os.WriteFile(*outPath, outBytes, 0o644))
}

func filterPathsByTag(paths map[string]any, tag string) map[string]any {
	out := map[string]any{}
	for p, v := range paths {
		pathItem, ok := v.(map[string]any)
		if !ok {
			continue
		}
		newPathItem := map[string]any{}
		// Preserve path-level parameters, if present (they might be required by kept ops).
		if params, ok := pathItem["parameters"]; ok {
			newPathItem["parameters"] = params
		}
		for k, opAny := range pathItem {
			lk := strings.ToLower(k)
			if _, isMethod := httpMethods[lk]; !isMethod {
				continue
			}
			op, ok := opAny.(map[string]any)
			if !ok {
				continue
			}
			if operationHasTag(op, tag) {
				newPathItem[lk] = op
			}
		}
		if len(newPathItem) > 0 {
			out[p] = newPathItem
		}
	}
	return out
}

func operationHasTag(op map[string]any, tag string) bool {
	tagsAny, ok := op["tags"].([]any)
	if !ok {
		return false
	}
	for _, t := range tagsAny {
		if s, ok := t.(string); ok && s == tag {
			return true
		}
	}
	return false
}

func buildMinComponents(filteredPaths map[string]any, components map[string]any, fullDoc map[string]any) map[string]any {
	if len(components) == 0 {
		return map[string]any{}
	}

	minComponents := map[string]any{}

	// Seed refs from paths tree.
	needRefs := map[string]struct{}{}
	queue := []string{}
	addRef := func(ref string) {
		if !strings.HasPrefix(ref, "#/components/") {
			return
		}
		if _, ok := needRefs[ref]; ok {
			return
		}
		needRefs[ref] = struct{}{}
		queue = append(queue, ref)
	}

	walkAny(filteredPaths, func(ref string) {
		addRef(ref)
	})

	// Also include securitySchemes referenced by "security" objects (these are not $ref-based).
	needSchemes := referencedSecuritySchemes(fullDoc, filteredPaths)

	// BFS: resolve each component ref and scan inside it for more refs.
	seen := map[string]struct{}{}
	for len(queue) > 0 {
		ref := queue[0]
		queue = queue[1:]
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}

		section, name, ok := parseComponentRef(ref)
		if !ok {
			continue
		}

		sectionAny, ok := components[section].(map[string]any)
		if !ok {
			continue
		}
		obj, ok := sectionAny[name]
		if !ok {
			continue
		}

		// Copy into minComponents
		dstSection, _ := minComponents[section].(map[string]any)
		if dstSection == nil {
			dstSection = map[string]any{}
			minComponents[section] = dstSection
		}
		dstSection[name] = obj

		// Recurse for additional refs inside this object
		walkAny(obj, func(inner string) {
			addRef(inner)
		})
	}

	if len(needSchemes) > 0 {
		if secAny, ok := components["securitySchemes"].(map[string]any); ok {
			dst, _ := minComponents["securitySchemes"].(map[string]any)
			if dst == nil {
				dst = map[string]any{}
				minComponents["securitySchemes"] = dst
			}
			for name := range needSchemes {
				if v, ok := secAny[name]; ok {
					dst[name] = v
				}
			}
		}
	}

	return minComponents
}

func parseComponentRef(ref string) (section string, name string, ok bool) {
	// "#/components/<section>/<name>"
	parts := strings.Split(ref, "/")
	if len(parts) != 4 {
		return "", "", false
	}
	if parts[0] != "#" || parts[1] != "components" {
		return "", "", false
	}
	return parts[2], parts[3], true
}

func walkAny(v any, onRef func(ref string)) {
	switch t := v.(type) {
	case map[string]any:
		// `$ref` nodes
		if r, ok := t["$ref"].(string); ok {
			onRef(r)
		}
		for _, vv := range t {
			walkAny(vv, onRef)
		}
	case []any:
		for _, vv := range t {
			walkAny(vv, onRef)
		}
	}
}

func referencedSecuritySchemes(fullDoc map[string]any, filteredPaths map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	collect := func(secAny any) {
		// security: [ { bearerAuth: [] }, { basicAuth: [] } ]
		arr, ok := secAny.([]any)
		if !ok {
			return
		}
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			for k := range m {
				out[k] = struct{}{}
			}
		}
	}

	if topSec, ok := fullDoc["security"]; ok {
		collect(topSec)
	}

	// Scan per-operation security under kept paths.
	walkAny(filteredPaths, func(_ string) {})
	// Manual scan for security arrays (not $ref).
	var scanSecurity func(any)
	scanSecurity = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			if sec, ok := t["security"]; ok {
				collect(sec)
			}
			for _, vv := range t {
				scanSecurity(vv)
			}
		case []any:
			for _, vv := range t {
				scanSecurity(vv)
			}
		}
	}
	scanSecurity(filteredPaths)

	// Deterministic (not required, but nice for debugging).
	if len(out) > 0 {
		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		_ = keys
	}

	return out
}

func copyKeyIfPresent(dst, src map[string]any, key string) {
	if v, ok := src[key]; ok {
		dst[key] = v
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func promoteValueTypeEnum(doc map[string]any) {
	components, ok := doc["components"].(map[string]any)
	if !ok || len(components) == 0 {
		return
	}
	schemas, _ := components["schemas"].(map[string]any)
	if schemas == nil {
		schemas = map[string]any{}
		components["schemas"] = schemas
	}

	// If the schema already exists, just rewrite inline occurrences to $ref it.
	const enumSchemaName = "FeatureValueType"
	ref := "#/components/schemas/" + enumSchemaName

	var enumVals []string
	if existing, ok := schemas[enumSchemaName].(map[string]any); ok {
		enumVals = enumStrings(existing)
	}

	// Discover enum values from any inline `properties.valueType` enum in the minimized doc.
	if len(enumVals) == 0 {
		enumVals = findInlineValueTypeEnumValues(doc)
	}

	// Nothing to do.
	if len(enumVals) == 0 {
		return
	}

	// Ensure component schema exists (so generators can emit a named enum type + constants).
	if _, ok := schemas[enumSchemaName]; !ok {
		schemas[enumSchemaName] = map[string]any{
			"type":        "string",
			"description": "The data type of the feature payload.",
			"enum":        stringSliceToAny(enumVals),
		}
	}

	// Rewrite every inline properties.valueType enum in the minimized doc to be a $ref.
	rewriteInlineValueTypeEnums(doc, ref)
}

func findInlineValueTypeEnumValues(doc map[string]any) []string {
	var found []string
	var walk func(any)
	walk = func(v any) {
		if len(found) > 0 {
			return
		}
		switch t := v.(type) {
		case map[string]any:
			if props, ok := t["properties"].(map[string]any); ok {
				if vtAny, ok := props["valueType"].(map[string]any); ok {
					if typ, _ := vtAny["type"].(string); typ == "string" {
						if enums := enumStrings(vtAny); len(enums) > 0 {
							found = enums
							return
						}
					}
				}
			}
			for _, vv := range t {
				walk(vv)
				if len(found) > 0 {
					return
				}
			}
		case []any:
			for _, vv := range t {
				walk(vv)
				if len(found) > 0 {
					return
				}
			}
		}
	}
	walk(doc)
	return found
}

func rewriteInlineValueTypeEnums(v any, ref string) {
	switch t := v.(type) {
	case map[string]any:
		if props, ok := t["properties"].(map[string]any); ok && props != nil {
			if vtAny, ok := props["valueType"].(map[string]any); ok {
				if typ, _ := vtAny["type"].(string); typ == "string" && len(enumStrings(vtAny)) > 0 {
					props["valueType"] = map[string]any{"$ref": ref}
				}
			}
		}
		for _, vv := range t {
			rewriteInlineValueTypeEnums(vv, ref)
		}
	case []any:
		for _, vv := range t {
			rewriteInlineValueTypeEnums(vv, ref)
		}
	}
}

func enumStrings(schema map[string]any) []string {
	enumAny, ok := schema["enum"].([]any)
	if !ok || len(enumAny) == 0 {
		return nil
	}
	out := make([]string, 0, len(enumAny))
	for _, v := range enumAny {
		s, ok := v.(string)
		if !ok || s == "" {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringSliceToAny(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}
