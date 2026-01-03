package generator

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/eastnine90/gbgen/internal/buildinfo"
)

func renderFeatureKeysGo(pkg string, features []featureMeta, emitList bool) ([]byte, error) {
	lines := nameAndDedupe(features)

	pkgName := pkg
	if pkgName == "" {
		pkgName = "features"
	}

	preamble, err := renderPreamble(preambleOptions{
		PackageName:  pkgName,
		GBGenVersion: buildinfo.Version,
		DocLines: []string{
			fmt.Sprintf("Package %s contains generated GrowthBook feature keys.", pkgName),
			"",
			"Example:",
			fmt.Sprintf("\timport %q", "path/to/your/generated/"+pkgName),
			"",
			"\t// Use the generated keys with your GrowthBook SDK wrapper / evaluator.",
			"\t_ = " + "FeatureKey(\"example\")",
		},
	})
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.Write(preamble)

	fmt.Fprintf(&b, "type FeatureKey string\n\n")

	fmt.Fprintf(&b, "const (\n")
	for _, l := range lines {
		desc := strings.TrimSpace(l.Description)
		if desc != "" {
			for _, line := range strings.Split(desc, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				fmt.Fprintf(&b, "\t// %s\n", line)
			}
		}
		if l.NoActiveEnvs {
			fmt.Fprintf(&b, "\t// Deprecated: no active environments\n")
		}
		fmt.Fprintf(&b, "\t%s FeatureKey = %q\n", l.Name, l.ID)
	}
	fmt.Fprintf(&b, ")\n\n")

	if emitList {
		fmt.Fprintf(&b, "var FeatureList = []FeatureKey{\n")
		for _, l := range lines {
			fmt.Fprintf(&b, "\t%s,\n", l.Name)
		}
		fmt.Fprintf(&b, "}\n\n")
	}

	return formatGo(b.Bytes())
}

func toExportedIdentifier(featureID string) string {
	// Split on non-alphanumeric and PascalCase the parts.
	parts := splitNonAlnum(featureID)
	var out strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		out.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			out.WriteRune(r)
		}
	}
	if out.Len() == 0 {
		return "Unknown"
	}
	return out.String()
}

func splitNonAlnum(s string) []string {
	var parts []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			parts = append(parts, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return parts
}
