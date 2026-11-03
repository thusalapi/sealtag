// Command shipwright publishes packages to several registries safely.
package main

import (
	"fmt"
	"os"

	"github.com/thusalapi/shipwright/internal/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
