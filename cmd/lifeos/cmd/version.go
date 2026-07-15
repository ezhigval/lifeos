package cmd

import (
	"fmt"
	goruntime "runtime"

	"github.com/spf13/cobra"
)

// Set via -ldflags at package time, e.g.
// -X github.com/valentinezhov/lifeos/cmd/lifeos/cmd.Version=LifeOS_alpha_1.0.0
var (
	Version = "dev"
	Commit  = "unknown"
	BuiltAt = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("lifeos %s\n", Version)
		fmt.Printf("  commit:  %s\n", Commit)
		fmt.Printf("  built:   %s\n", BuiltAt)
		fmt.Printf("  go:      %s %s/%s\n", goruntime.Version(), goruntime.GOOS, goruntime.GOARCH)
	},
}
