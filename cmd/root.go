package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sk",
	Short: "sk — agentskills.io package manager",
	Long: `sk is a package manager for agentskills.io-compliant skills.

It lets you discover, install, and publish AI agent skills across
25+ coding tools from a unified command line interface.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(versionCmd)
}
