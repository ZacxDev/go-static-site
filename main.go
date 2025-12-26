package main

import (
	"github.com/ZacxDev/go-static-site/cmd"

	// Import pages to register gomponents with the component registry
	_ "github.com/ZacxDev/go-static-site/components/pages"
)

func main() {
	cmd.Execute()
}
