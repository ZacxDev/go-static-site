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
		port, _ := cmd.Flags().GetString("port")
		fmt.Printf("Starting server on port %s\n", port)

		router, err := handlers.SetupRouter()
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
