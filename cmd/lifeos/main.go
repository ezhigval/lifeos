package main

import (
	"os"

	"github.com/valentinezhov/lifeos/cmd/lifeos/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
