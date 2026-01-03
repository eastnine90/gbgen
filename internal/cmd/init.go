package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/eastnine90/gbgen/internal/config"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		outPath string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a sample config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outPath == "" {
				outPath = "gbgen.yaml"
			}

			f := strings.ToLower(strings.TrimSpace(format))
			if f == "" {
				detected, err := config.DetectFormat(outPath, "")
				if err != nil {
					return err
				}
				f = string(detected)
			}

			cfg := config.Sample()

			format, err := config.DetectFormat(outPath, f)
			if err != nil {
				return err
			}
			b, err := config.Marshal(cfg, format)
			if err != nil {
				return err
			}

			dir := filepath.Dir(outPath)
			if dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
			}
			return os.WriteFile(outPath, b, 0o644)
		},
	}

	cmd.Flags().StringVarP(&outPath, "out", "o", "gbgen.yaml", "Output path for the config file")
	cmd.Flags().StringVar(&format, "format", "", "Config format: json|yaml|toml (defaults from file extension)")

	return cmd
}
