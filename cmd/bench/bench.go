package main

import (
	"fmt"
	"strings"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/report"
)

// benchWorkloads is the demo mix: it exercises all three failure modes and one
// positive capability-match (int8 inference belongs on the cheap Inferentia2).
func benchWorkloads() []allocator.Workload {
	return []allocator.Workload{
		{Name: "train-llm", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.5},
		{Name: "train-big", Kind: allocator.KindTrain, MinMemoryGB: 150, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.5},
		{Name: "infer-fp8", Kind: allocator.KindInfer, MinMemoryGB: 16, RequiredPrecisions: []allocator.Precision{allocator.PrecisionFP8}, DeviceCount: 1, CostWeight: 1},
		{Name: "infer-int8", Kind: allocator.KindInfer, MinMemoryGB: 16, RequiredPrecisions: []allocator.Precision{allocator.PrecisionINT8}, DeviceCount: 1, CostWeight: 1},
		{Name: "gang-train", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 4, SameIsland: true, CostWeight: 0.7},
	}
}

func format(rep report.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Sift vs Legacy — realistic-2026 (%d devices, %d workloads)\n\n", rep.Fleet, rep.Workloads)
	fmt.Fprintf(&b, "  %-12s  %-32s  %s\n", "workload", "Sift", "Legacy")
	fmt.Fprintf(&b, "  %-12s  %-32s  %s\n", "--------", "----", "------")
	for i := range rep.Sift.Outcomes {
		fmt.Fprintf(&b, "  %-12s  %-32s  %s\n",
			rep.Sift.Outcomes[i].Workload, cell(rep.Sift.Outcomes[i]), cell(rep.Legacy.Outcomes[i]))
	}
	w := rep.Workloads
	fmt.Fprintf(&b, "\n  %-12s  %-12s  %s\n", "totals", "Sift", "Legacy")
	fmt.Fprintf(&b, "  %-12s  $%-11.2f  $%.2f\n", "total $/hr", rep.Sift.TotalCost, rep.Legacy.TotalCost)
	fmt.Fprintf(&b, "  %-12s  %-12s  %s\n", "type-correct", frac(rep.Sift.TypeCorrect, w), frac(rep.Legacy.TypeCorrect, w))
	fmt.Fprintf(&b, "  %-12s  %-12d  %d\n", "fragmented", rep.Sift.Fragmented, rep.Legacy.Fragmented)
	fmt.Fprintf(&b, "  %-12s  %-12d  %d\n", "pending", rep.Sift.Pending, rep.Legacy.Pending)
	return b.String()
}

func cell(o report.Outcome) string {
	if o.Pending {
		return "Pending"
	}
	mark := "ok"
	if !o.Feasible {
		mark = "WRONG-TYPE"
	} else if !o.SameIslandOK {
		mark = "FRAGMENTED"
	}
	devs := o.DeviceIDs[0]
	if len(o.DeviceIDs) > 1 { // abbreviate a gang so the table stays aligned
		devs = fmt.Sprintf("%s ×%d", o.DeviceIDs[0], len(o.DeviceIDs))
	}
	return fmt.Sprintf("%s ($%.2f) %s", devs, o.CostPerHr, mark)
}

func frac(n, total int) string { return fmt.Sprintf("%d/%d", n, total) }
