package main

import (
	"fmt"
	"os"

	"github.com/akme/gh-stars-watcher/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
