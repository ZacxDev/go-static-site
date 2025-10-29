package cmd

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ZacxDev/go-static-site/handlers"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a static version of the site",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building static site...")

		manifestPath, _ := cmd.Flags().GetString("manifest")
		cleanOutput, _ := cmd.Flags().GetBool("clean")
		enableTracking, _ := cmd.Flags().GetBool("track-artifacts")

		manifest, err := handlers.LoadManifest(manifestPath)
		if err != nil {
			fmt.Printf("error loading manifest: %v", err)
			os.Exit(1)
		}

		// Determine output directory: flag takes precedence, then config, then default
		outputDir, _ := cmd.Flags().GetString("output")
		if cmd.Flags().Changed("output") {
			// Flag was explicitly set, use it
		} else if manifest.OutputDir != "" {
			// Use config value
			outputDir = manifest.OutputDir
		}
		// Otherwise use flag's default value of "public"

		router, err := handlers.SetupRouterWithManifest(manifestPath)
		if err != nil {
			fmt.Printf("Error setting up router: %v\n", err)
			os.Exit(1)
		}

		// Load existing artifact registry for cleanup
		var oldRegistry *utils.ArtifactRegistry
		if enableTracking && !cleanOutput {
			oldRegistry, err = utils.LoadArtifactRegistry(outputDir)
			if err != nil {
				fmt.Printf("Warning: Failed to load artifact registry: %v\n", err)
			}
		}

		// Clean output directory if requested
		if cleanOutput {
			fmt.Printf("Cleaning output directory: %s\n", outputDir)
			err = os.RemoveAll(outputDir)
			if err != nil {
				fmt.Printf("Error cleaning output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Create output directory
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Initialize new artifact registry
		newRegistry := &utils.ArtifactRegistry{
			GeneratedFiles: []string{},
			StaticFiles:    []string{},
			AssetFiles:     []string{},
			OutputDir:      outputDir,
			ManifestPath:   manifestPath,
			Metadata:       make(map[string]string),
		}

		// Determine static directory from config or default
		staticDir := manifest.StaticDir
		if staticDir == "" {
			staticDir = "static"
		}

		// Copy static files
		err = filepath.Walk(filepath.Join(".", staticDir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				destPath := filepath.Join(outputDir, path)
				err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
				if err != nil {
					return err
				}
				fmt.Printf("%+v %s\n", path, destPath)
				err = copyFile(path, destPath)
				if err == nil && enableTracking {
					newRegistry.AddStaticFile(destPath)
				}
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error copying static files: %v\n", err)
			os.Exit(1)
		}

		// Generate static pages
		server := httptest.NewServer(router)
		defer server.Close()

		langPattern := regexp.MustCompile(`\/\{lang:([^}]+)\}\/`)

		done := make(chan struct{})

		err = handlers.RenderAllPages(server, router, langPattern, true, outputDir, done)
		if err != nil {
			log.Fatalf("Error rendering: %v\n", err)
		}

		<-done

		// Track generated pages and assets for artifact tracking
		if enableTracking {
			err = trackGeneratedPages(outputDir, newRegistry)
			if err != nil {
				fmt.Printf("Warning: Failed to track generated pages: %v\n", err)
			}

			err = trackAssetFiles(outputDir, newRegistry)
			if err != nil {
				fmt.Printf("Warning: Failed to track asset files: %v\n", err)
			}
		}

		// Generate sitemaps
		err = utils.GenerateSitemaps(handlers.GetRegisteredRoutes(), manifest.AppOrigin, manifest.Routes)
		if err != nil {
			fmt.Printf("Error generating sitemap: %s\n", err.Error())
		}

		// Perform artifact cleanup and tracking
		if enableTracking {
			// Cleanup orphaned files from previous builds
			if oldRegistry != nil {
				err = oldRegistry.CleanupOrphanedFiles(newRegistry)
				if err != nil {
					fmt.Printf("Warning: Failed to cleanup orphaned files: %v\n", err)
				}
			}

			// Remove empty directories
			err = newRegistry.RemoveEmptyDirectories()
			if err != nil {
				fmt.Printf("Warning: Failed to remove empty directories: %v\n", err)
			}

			// Save artifact registry for next build
			err = newRegistry.SaveArtifactRegistry()
			if err != nil {
				fmt.Printf("Warning: Failed to save artifact registry: %v\n", err)
			} else {
				stats := newRegistry.GetBuildStats()
				fmt.Printf("Build tracking: %d generated, %d static, %d asset files\n",
					stats["generated_files"], stats["static_files"], stats["asset_files"])
			}
		}

		fmt.Printf("Static site generated successfully in the %s directory\n", outputDir)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringP("output", "o", "public", "Output directory for generated site")
	buildCmd.Flags().Bool("clean", false, "Clean output directory before build")
	buildCmd.Flags().Bool("track-artifacts", true, "Track generated artifacts for incremental cleanup")
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// trackGeneratedPages recursively finds and tracks all generated HTML files
func trackGeneratedPages(outputDir string, registry *utils.ArtifactRegistry) error {
	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the artifact registry file itself
		if filepath.Base(path) == ".build-artifacts.json" {
			return nil
		}

		// Track HTML files as generated content
		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".xml")) {
			registry.AddGeneratedFile(path)
		}

		return nil
	})
}

// trackAssetFiles recursively finds and tracks all JavaScript and CSS assets
func trackAssetFiles(outputDir string, registry *utils.ArtifactRegistry) error {
	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the artifact registry file itself
		if filepath.Base(path) == ".build-artifacts.json" {
			return nil
		}

		// Track JS and CSS files as asset content
		if !info.IsDir() && (strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".map")) {
			registry.AddAssetFile(path)
		}

		return nil
	})
}
