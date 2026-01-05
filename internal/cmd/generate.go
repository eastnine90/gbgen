package cmd

import (
	"os"
	"path/filepath"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/generator"
	"github.com/spf13/cobra"
)

const outputFileName = "features.gen.go"

func newGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate Go types from GrowthBook features",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.LoadOptions{
				ConfigPath: flagConfigPath,
				EnvPrefix:  "GBGEN",
			})
			if err != nil {
				return err
			}
			// validate config
			if err := cfg.Validate(); err != nil {
				return err
			}

			ctx := cmd.Context()
			g, err := generator.NewGenerator(cfg)
			if err != nil {
				return err
			}

			src, err := g.Generate(ctx)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(cfg.Generator.OutputDir, 0o755); err != nil {
				return err
			}

			outPath := filepath.Join(cfg.Generator.OutputDir, outputFileName)

			return os.WriteFile(outPath, src, 0o644)
		},
	}
}
