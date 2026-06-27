package main

import (
	"strings"
	"testing"

	"github.com/FabianSalge/sift/report"
)

// format must mark a mis-typed placement WRONG-TYPE and render device IDs and
// costs in the table.
func TestFormatTable(t *testing.T) {
	rep := report.Report{
		Fleet: 2, Workloads: 1,
		Sift: report.Summary{Name: "Sift", TotalCost: 2.5, TypeCorrect: 1, Outcomes: []report.Outcome{
			{Workload: "train", DeviceIDs: []string{"h100-0"}, CostPerHr: 2.5, Feasible: true, SameIslandOK: true},
		}},
		Legacy: report.Summary{Name: "Legacy", TotalCost: 0.75, Outcomes: []report.Outcome{
			{Workload: "train", DeviceIDs: []string{"inf-0"}, CostPerHr: 0.75, Feasible: false, SameIslandOK: true},
		}},
	}

	out := format(rep)
	for _, want := range []string{"Sift", "Legacy", "type-correct", "WRONG-TYPE", "h100-0", "$2.50"} {
		if !strings.Contains(out, want) {
			t.Errorf("format() missing %q:\n%s", want, out)
		}
	}
}
