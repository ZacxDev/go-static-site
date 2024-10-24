package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ZacxDev/go-static-site/handlers"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a static version of the site",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building static site...")

		router, err := handlers.SetupRouter()
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

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				return nil // Skip routes without a path template
			}

			// Skip sitemap because we generate that seperately
			if path == "/sitemap.xml" {
				return nil
			}

			matches := langPattern.FindStringSubmatch(path)
			if len(matches) > 1 {
				langs := strings.Split(matches[1], "|")

				// Base route without the language pattern
				baseRoute := langPattern.ReplaceAllString(path, "/")

				// Generate URLs for each supported language
				for _, lang := range langs {
					langPath := fmt.Sprintf("/%s%s", lang, baseRoute)
					err := generateStaticPage(server, langPath, lang)
					if err != nil {
						fmt.Printf("Error generating static page for %s: %v\n", langPath, err)
					}
				}
			} else {
				// Handle non-language specific routes
				err := generateStaticPage(server, path, "")
				if err != nil {
					fmt.Printf("Error generating static page for %s: %v\n", path, err)
				}
			}

			return nil
		})

		// Generate sitemaps
		err = utils.GenerateSitemaps(handlers.GetRegisteredRoutes())
		if err != nil {
			fmt.Printf("Error generating sitemap: %s\n", err.Error())
		}

		fmt.Println("Static site generated successfully in the ./public directory")
	},
}

func generateStaticPage(server *httptest.Server, route string, lang string) error {
	url := server.URL + route
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := filepath.Join("public", route[1:], "index.html")
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Generated %s\n", filePath)
	return nil
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
