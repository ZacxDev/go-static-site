package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-static-site",
	Short: "go-static-site - generate static sites from a single manifest file",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("manifest", "m", "manifest.star", "Path to the manifest file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output with timing information")
}

// RegisterCommand allows external packages to add commands to the root command
func RegisterCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
