package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ZacxDev/go-static-site/handlers"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		manifestPath, _ := cmd.Flags().GetString("manifest")

		// Load manifest to check for default port
		manifest, err := handlers.LoadManifest(manifestPath)
		if err != nil {
			log.Fatalf("Error loading manifest: %v", err)
		}

		// Determine port: --port flag takes precedence over manifest default_port
		port, _ := cmd.Flags().GetString("port")
		flagChanged := cmd.Flags().Changed("port")

		if !flagChanged && manifest.DefaultPort != "" {
			port = manifest.DefaultPort
		}

		fmt.Printf("Starting server on port %s\n", port)
		fmt.Printf("Using manifest: %s\n", manifestPath)

		router, err := handlers.SetupRouterWithManifest(manifestPath)
		if err != nil {
			log.Fatalf("Error setting up router: %v", err)
		}

		log.Fatal(http.ListenAndServe(":"+port, router))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "9010", "Port to run the server on")
}
