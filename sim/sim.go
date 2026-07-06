// Package sim forward-simulates a stream of workload arrivals against both
// schedulers: each placement occupies its devices for a duration then frees
// them; a job is "useful" only if its placement is feasible (legacy's blind
// placements occupy capacity but may do no work). It is pure and deterministic —
// a function of (fleet, stream) — and holds no scheduling logic of its own: every
// placement comes from allocator.Allocate{Sift,Legacy}.
package sim

import (
	"math"
	"strconv"

	"github.com/FabianSalge/sift/allocator"
)

type Arrival struct {
	At       float64
	Workload allocator.Workload
	Duration float64
}

type Stream []Arrival

// ArrivalResult is one arrival's fate under one scheduler. The UI derives all
// time-t state (occupancy, queue depth, completions, cost) from these.
type ArrivalResult struct {
	Index        int
	Workload     string
	ArrivedAt    float64
	PlacedAt     float64 // -1 if never placed within the horizon
	End          float64 // PlacedAt + Duration, or -1
	DeviceIDs    []string
	Feasible     bool
	SameIslandOK bool
	Useful       bool
	CostPerHr    float64
}

type SchedulerResult struct {
	Name     string
	Arrivals []ArrivalResult
}

type Result struct {
	Fleet   int
	Stream  int
	Horizon float64
	Sift    SchedulerResult
	Legacy  SchedulerResult
}

// PlaceFunc is one stateless placement decision (allocator.AllocateSift or
// AllocateLegacy) — the only door through which sim places anything.
type PlaceFunc func(devices []allocator.Device, w allocator.Workload, allocated map[string]bool) ([]string, bool)

// Run simulates the stream against both schedulers.
func Run(fleet []allocator.Device, stream Stream) Result {
	sift := runOne("Sift", fleet, stream, allocator.AllocateSift)
	legacy := runOne("Legacy", fleet, stream, allocator.AllocateLegacy)
	return Result{
		Fleet:   len(fleet),
		Stream:  len(stream),
		Horizon: math.Max(horizonOf(sift), horizonOf(legacy)),
		Sift:    sift,
		Legacy:  legacy,
	}
}

func runOne(name string, fleet []allocator.Device, stream Stream, place PlaceFunc) SchedulerResult {
	c := NewCluster(fleet, place)
	for _, a := range stream {
		// Advance twice on purpose: the first flushes completions due at or
		// before a.At (older queued work grabs freed devices first), the
		// second places this arrival at exactly a.At. Collapsing them would
		// change same-instant ordering.
		c.Advance(a.At)
		c.Submit(a.Workload, a.Duration)
		c.Advance(a.At)
	}
	// Drain: process remaining completions; anything still queued when the
	// last job completes can never place (PlacedAt stays -1) — the old
	// infinite-time break, same semantics.
	for len(c.running) > 0 {
		t, _ := c.nextEnd()
		c.Advance(t)
	}

	res := make([]ArrivalResult, len(c.jobs))
	for i, j := range c.jobs {
		res[i] = ArrivalResult{
			Index: j.ID, Workload: j.Workload.Name, ArrivedAt: j.ArrivedAt,
			PlacedAt: j.PlacedAt, End: j.End, DeviceIDs: j.DeviceIDs,
			Feasible: j.Feasible, SameIslandOK: j.SameIslandOK, Useful: j.Useful, CostPerHr: j.CostPerHr,
		}
	}
	return SchedulerResult{Name: name, Arrivals: res}
}

// evaluate reports whether a placement is feasible and (for a same-island gang)
// whole — mirrors the contrast outcome accounting.
func evaluate(byID map[string]allocator.Device, w allocator.Workload, ids []string) (feasible, sameIsland bool) {
	feasible = true
	for _, id := range ids {
		if !allocator.Feasible(byID[id], w) {
			feasible = false
		}
	}
	sameIsland = true
	if w.SameIsland && w.DeviceCount > 1 {
		// Group by island; a standalone device (NoIsland) shares no interconnect,
		// so it is its own group (keyed by id) — standalones never collapse into one.
		groups := map[string]bool{}
		for _, id := range ids {
			d := byID[id]
			if d.IslandID == allocator.NoIsland {
				groups["s:"+d.ID] = true
			} else {
				groups["i:"+strconv.Itoa(d.IslandID)] = true
			}
		}
		sameIsland = len(groups) == 1
	}
	return feasible, sameIsland
}

func costOf(byID map[string]allocator.Device, ids []string) float64 {
	var c float64
	for _, id := range ids {
		c += byID[id].CostPerHr
	}
	return c
}

func indexByID(fleet []allocator.Device) map[string]allocator.Device {
	m := make(map[string]allocator.Device, len(fleet))
	for _, d := range fleet {
		m[d.ID] = d
	}
	return m
}

func horizonOf(r SchedulerResult) float64 {
	h := 0.0
	for _, a := range r.Arrivals {
		if a.End > h {
			h = a.End
		}
	}
	return h
}
