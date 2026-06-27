// Package report runs both schedulers over a workload sequence and aggregates
// the Sift-vs-legacy contrast. It is pure (imports only allocator) so the CLI
// benchmark and the WASM demo share one source for the comparison.
package report

import "github.com/FabianSalge/sift/allocator"

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

// Run places the workload sequence on both schedulers and aggregates the result.
func Run(fleet []allocator.Device, workloads []allocator.Workload) Report {
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
