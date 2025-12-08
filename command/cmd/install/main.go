package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/braiphub/go-core/command/installer"
)

func main() {
	force := flag.Bool("force", false, "Force overwrite existing files")
	flag.Parse()

	inst, err := installer.New(*force)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := inst.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
