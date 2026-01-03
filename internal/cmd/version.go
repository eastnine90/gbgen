package cmd

import (
	"fmt"

	"github.com/eastnine90/gbgen/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			// Keep output stable/simple for scripting, but include commit metadata when available.
			if buildinfo.Commit != "" {
				fmt.Printf("gbgen %s (%s)\n", buildinfo.Version, buildinfo.Commit)
				return
			}
			fmt.Printf("gbgen %s\n", buildinfo.Version)
		},
	}
}
