// Command sealtag publishes packages to several registries safely.
package main

import (
	"fmt"
	"os"

	"github.com/thusalapi/sealtag/internal/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
