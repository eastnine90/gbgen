package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eastnine90/gbgen/internal/buildinfo"
	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

func renderTypedFeaturesGo(pkg string, features []featureMeta, emitList bool) ([]byte, error) {
	pkgName := pkg
	if pkgName == "" {
		pkgName = "features"
	}

	lines := nameAndDedupe(features)

	preamble, err := renderPreamble(preambleOptions{
		PackageName:  pkgName,
		GBGenVersion: buildinfo.Version,
		DocLines: []string{
			fmt.Sprintf("Package %s contains generated GrowthBook typed feature helpers.", pkgName),
			"",
			"Example:",
			fmt.Sprintf("\timport %q", "path/to/your/generated/"+pkgName),
			"\timport \"github.com/growthbook/growthbook-golang\"",
			"",
			"\tres, err := " + "FeatureExample.Evaluate(ctx, client)",
			"\t_ = res; _ = err",
		},
		Imports: []string{"github.com/eastnine90/gbgen/types"},
	})
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.Write(preamble)

	if emitList {
		fmt.Fprintf(&b, "type FeatureKey string\n\n")
	}

	fmt.Fprintf(&b, "var (\n")
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

		typeExpr, err := typedFeatureTypeExpr(l.ValueType)
		if err != nil {
			return nil, fmt.Errorf("feature %q: %w", l.ID, err)
		}

		fmt.Fprintf(&b, "\t%s = %s(%q)\n", l.Name, typeExpr, l.ID)
	}
	fmt.Fprintf(&b, ")\n")

	if emitList {
		fmt.Fprintf(&b, "\nvar FeatureList = []FeatureKey{\n")
		for _, l := range lines {
			fmt.Fprintf(&b, "\tFeatureKey(%q),\n", l.ID)
		}
		fmt.Fprintf(&b, "}\n")
	}

	return formatGo(b.Bytes())
}

func typedFeatureTypeExpr(vt growthbookapi.FeatureValueType) (string, error) {
	switch vt {
	case growthbookapi.FEATUREVALUETYPE_BOOLEAN:
		return "types.BooleanFeature", nil
	case growthbookapi.FEATUREVALUETYPE_STRING:
		return "types.StringFeature", nil
	case growthbookapi.FEATUREVALUETYPE_NUMBER:
		return "types.NumberFeature", nil
	case growthbookapi.FEATUREVALUETYPE_JSON:
		return "types.JSONFeature", nil
	default:
		return "", fmt.Errorf("unsupported valueType %q", string(vt))
	}
}
