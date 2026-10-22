package main

import (
	"fmt"
	"ghostops/internal/cli"
	"os"
)

func main() {
	cfg, err := cli.ParseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(1)
	}

	if err := cli.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
