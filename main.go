package main

import (
	"fmt"
	"os"

	"github.com/josegonzalez/aws-eks-calculator/internal/tui"
)

var (
	tuiRun = tui.Run
	osExit = os.Exit
)

func run() error {
	return tuiRun()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
	}
}
