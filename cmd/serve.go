package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ZacxDev/go-static-site/handlers"
	"github.com/ZacxDev/go-static-site/utils"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		timer := utils.NewTimer(verbose)

		manifestPath, _ := cmd.Flags().GetString("manifest")

		// Load manifest to check for default port
		manifest, err := handlers.LoadManifest(manifestPath)
		if err != nil {
			log.Fatalf("Error loading manifest: %v", err)
		}
		timer.Step("Loaded manifest")

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
		timer.Step("Set up router")

		fmt.Printf("Server ready\n")
		timer.PrintSummary()

		log.Fatal(http.ListenAndServe(":"+port, router))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "9010", "Port to run the server on")
}
