package cmd

import (
	"fmt"
	"os"

	"github.com/ZacxDev/go-static-site/handlers"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build artifacts and output directory",
	Run: func(cmd *cobra.Command, args []string) {
		showStats, _ := cmd.Flags().GetBool("stats")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Load manifest to get configured output directory
		manifestPath := "manifest.star"
		manifest, err := handlers.LoadManifest(manifestPath)

		// Determine output directory: flag takes precedence, then config, then default
		outputDir, _ := cmd.Flags().GetString("output")
		if cmd.Flags().Changed("output") {
			// Flag was explicitly set, use it
		} else if err == nil && manifest != nil && manifest.OutputDir != "" {
			// Use config value if manifest loaded successfully
			outputDir = manifest.OutputDir
		}
		// Otherwise use flag's default value of "public"

		if showStats {
			// Load and display artifact registry stats
			registry, err := utils.LoadArtifactRegistry(outputDir)
			if err != nil {
				fmt.Printf("No artifacts to clean (registry not found): %v\n", err)
				return
			}

			stats := registry.GetBuildStats()
			fmt.Printf("Build artifacts in %s:\n", outputDir)
			fmt.Printf("  Generated files: %d\n", stats["generated_files"])
			fmt.Printf("  Static files: %d\n", stats["static_files"])
			fmt.Printf("  Asset files: %d\n", stats["asset_files"])
			fmt.Printf("  Total files: %d\n", stats["total_files"])
			fmt.Printf("  Last build: %s\n", stats["build_timestamp"])
			fmt.Printf("  Manifest: %s\n", stats["manifest_path"])
			return
		}

		if dryRun {
			fmt.Printf("Would clean output directory: %s\n", outputDir)
			if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
				registry, err := utils.LoadArtifactRegistry(outputDir)
				if err == nil {
					stats := registry.GetBuildStats()
					fmt.Printf("Would remove %d files\n", stats["total_files"])
				}
			}
			return
		}

		fmt.Printf("Cleaning output directory: %s\n", outputDir)
		err = os.RemoveAll(outputDir)
		if err != nil {
			fmt.Printf("Error cleaning output directory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully cleaned %s\n", outputDir)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringP("output", "o", "public", "Output directory to clean")
	cleanCmd.Flags().Bool("stats", false, "Show build artifact statistics instead of cleaning")
	cleanCmd.Flags().Bool("dry-run", false, "Show what would be cleaned without actually cleaning")
}
