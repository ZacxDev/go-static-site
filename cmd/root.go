package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mylinksprofile",
	Short: "MyLinksProfile - Showcase all your links in one place",
	Long:  `MyLinksProfile is a platform that allows you to create a single page featuring links to your content on different platforms.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
