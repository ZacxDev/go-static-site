package cmd

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"

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
		manifest, err := handlers.LoadManifest(manifestPath)
		if err != nil {
			fmt.Printf("error loading manifest: %v", err)
			os.Exit(1)
		}

		router, err := handlers.SetupRouterWithManifest(manifestPath)
		if err != nil {
			fmt.Printf("Error setting up router: %v\n", err)
			os.Exit(1)
		}

		// Create public directory
		err = os.MkdirAll("./public", os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating public directory: %v\n", err)
			os.Exit(1)
		}

		// Copy static files
		err = filepath.Walk("./static", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				destPath := filepath.Join("public", path)
				err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
				if err != nil {
					return err
				}
				fmt.Printf("%+v %s\n", path, destPath)
				return copyFile(path, destPath)
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

		err = handlers.RenderAllPages(server, router, langPattern, true, done)
		if err != nil {
			log.Fatalf("Error rendering: %v\n", err)
		}

		<-done

		// Generate sitemaps
		err = utils.GenerateSitemaps(handlers.GetRegisteredRoutes(), manifest.AppOrigin, manifest.Routes)
		if err != nil {
			fmt.Printf("Error generating sitemap: %s\n", err.Error())
		}

		fmt.Println("Static site generated successfully in the ./public directory")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
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
