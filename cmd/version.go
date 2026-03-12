package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print sk version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sk %s\n", Version)
	},
}
