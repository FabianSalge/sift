package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/config"
	"github.com/FabianSalge/sift/report"
)

func main() {
	explain := flag.String("explain", "", "trace one workload's filter → score → bind decision instead of the comparison")
	flag.Parse()

	fleet, err := config.LoadFleetFile("scenarios/realistic-2026.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "load fleet:", err)
		os.Exit(1)
	}

	if *explain != "" {
		w, ok := workloadByName(benchWorkloads(), *explain)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown workload %q; choose one of: %s\n", *explain, workloadNames(benchWorkloads()))
			os.Exit(1)
		}
		fmt.Print(formatExplain("realistic-2026", w, allocator.Explain(fleet, w, nil), fleet))
		return
	}

	fmt.Print(format(report.Run(fleet, benchWorkloads())))
}
