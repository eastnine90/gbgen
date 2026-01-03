package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gbgen",
	Short:        "Generate Go-safe types from GrowthBook features",
	SilenceUsage: true,
}

var (
	flagConfigPath string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagConfigPath, "config", "", "Path to config file (json|yaml|toml)")

	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newInitCmd())
}
