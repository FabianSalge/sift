package main

import (
	"fmt"
	"strings"

	"github.com/FabianSalge/sift/allocator"
)

// placer is the common shape of both schedulers — the interface ADR-0019
// deferred until a consumer needed it.
type placer interface {
	Place(allocator.Workload) (allocator.Placement, error)
}

// Outcome is one scheduler's result for one workload.
type Outcome struct {
	Workload     string
	DeviceIDs    []string
	CostPerHr    float64
	Feasible     bool // every bound device can actually run the workload
	SameIslandOK bool // for a same-island gang: all devices share one island
	Pending      bool
}

// Summary aggregates one scheduler's outcomes over the workload sequence.
type Summary struct {
	Name        string
	TotalCost   float64
	TypeCorrect int
	GangsWhole  int
	Fragmented  int
	Pending     int
	Outcomes    []Outcome
}

// Report is the full Sift-vs-legacy comparison for one fleet + workload set.
type Report struct {
	Fleet     int
	Workloads int
	Sift      Summary
	Legacy    Summary
}

func run(fleet []allocator.Device, workloads []allocator.Workload) Report {
	idx := indexByID(fleet)
	return Report{
		Fleet:     len(fleet),
		Workloads: len(workloads),
		Sift:      runOne("Sift", allocator.NewSiftScheduler(fleet), workloads, idx),
		Legacy:    runOne("Legacy", allocator.NewLegacyScheduler(fleet), workloads, idx),
	}
}

func runOne(name string, sched placer, workloads []allocator.Workload, idx map[string]allocator.Device) Summary {
	s := Summary{Name: name}
	for _, w := range workloads {
		p, err := sched.Place(w)
		if err != nil {
			s.Pending++
			s.Outcomes = append(s.Outcomes, Outcome{Workload: w.Name, Pending: true})
			continue
		}
		o := Outcome{Workload: w.Name, DeviceIDs: p.DeviceIDs, CostPerHr: p.CostPerHr, Feasible: true, SameIslandOK: true}
		islands := map[int]bool{}
		for _, id := range p.DeviceIDs {
			d := idx[id]
			if !allocator.Feasible(d, w) {
				o.Feasible = false
			}
			islands[d.IslandID] = true
		}
		if w.SameIsland && w.DeviceCount > 1 {
			o.SameIslandOK = len(islands) == 1
			if o.SameIslandOK {
				s.GangsWhole++
			} else {
				s.Fragmented++
			}
		}
		if o.Feasible {
			s.TypeCorrect++
		}
		s.TotalCost += p.CostPerHr
		s.Outcomes = append(s.Outcomes, o)
	}
	return s
}

func indexByID(fleet []allocator.Device) map[string]allocator.Device {
	m := make(map[string]allocator.Device, len(fleet))
	for _, d := range fleet {
		m[d.ID] = d
	}
	return m
}

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

func format(rep Report) string {
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

func cell(o Outcome) string {
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
