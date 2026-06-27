package main

import (
	"fmt"
	"os"

	"github.com/FabianSalge/sift/config"
	"github.com/FabianSalge/sift/report"
)

func main() {
	fleet, err := config.LoadFleetFile("scenarios/realistic-2026.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "load fleet:", err)
		os.Exit(1)
	}
	fmt.Print(format(report.Run(fleet, benchWorkloads())))
}
