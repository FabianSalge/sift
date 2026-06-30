// Package sim forward-simulates a stream of workload arrivals against both
// schedulers: each placement occupies its devices for a duration then frees
// them; a job is "useful" only if its placement is feasible (legacy's blind
// placements occupy capacity but may do no work). It is pure and deterministic —
// a function of (fleet, stream) — and holds no scheduling logic of its own: every
// placement comes from allocator.Allocate{Sift,Legacy}.
package sim

import (
	"math"

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

type allocFn func(devices []allocator.Device, w allocator.Workload, allocated map[string]bool) ([]string, bool)

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

type running struct {
	idx int
	end float64
	ids []string
}

func runOne(name string, fleet []allocator.Device, stream Stream, place allocFn) SchedulerResult {
	byID := indexByID(fleet)
	res := make([]ArrivalResult, len(stream))
	for i, a := range stream {
		res[i] = ArrivalResult{Index: i, Workload: a.Workload.Name, ArrivedAt: a.At, PlacedAt: -1, End: -1}
	}

	allocated := map[string]bool{}
	var jobs []running
	var queue []int // arrival indices waiting, in arrival order
	next := 0       // index of the next arrival not yet admitted

	for {
		if next >= len(stream) && len(queue) == 0 && len(jobs) == 0 {
			break
		}

		// next event time = min(next arrival, next completion)
		t := math.Inf(1)
		if next < len(stream) {
			t = stream[next].At
		}
		for _, j := range jobs {
			if j.end < t {
				t = j.end
			}
		}
		if math.IsInf(t, 1) {
			break // queued arrivals can never be placed (left PlacedAt = -1)
		}

		// 1. complete jobs that end at or before t (free their devices)
		kept := jobs[:0]
		for _, j := range jobs {
			if j.end <= t {
				for _, id := range j.ids {
					delete(allocated, id)
				}
			} else {
				kept = append(kept, j)
			}
		}
		jobs = kept

		// 2. admit arrivals at or before t
		for next < len(stream) && stream[next].At <= t {
			queue = append(queue, next)
			next++
		}

		// 3. drain the queue in arrival order (backfill: skip what won't fit)
		var stillQueued []int
		for _, idx := range queue {
			a := stream[idx]
			ids, ok := place(fleet, a.Workload, allocated)
			if !ok {
				stillQueued = append(stillQueued, idx)
				continue
			}
			for _, id := range ids {
				allocated[id] = true
			}
			feas, island := evaluate(byID, a.Workload, ids)
			res[idx].PlacedAt = t
			res[idx].End = t + a.Duration
			res[idx].DeviceIDs = ids
			res[idx].Feasible = feas
			res[idx].SameIslandOK = island
			res[idx].Useful = feas && island
			res[idx].CostPerHr = costOf(byID, ids)
			jobs = append(jobs, running{idx: idx, end: t + a.Duration, ids: ids})
		}
		queue = stillQueued
	}

	return SchedulerResult{Name: name, Arrivals: res}
}

// evaluate reports whether a placement is feasible and (for a same-island gang)
// whole — mirrors the contrast outcome accounting.
func evaluate(byID map[string]allocator.Device, w allocator.Workload, ids []string) (feasible, sameIsland bool) {
	feasible = true
	islands := map[int]bool{}
	for _, id := range ids {
		d := byID[id]
		if !allocator.Feasible(d, w) {
			feasible = false
		}
		islands[d.IslandID] = true
	}
	sameIsland = true
	if w.SameIsland && w.DeviceCount > 1 {
		sameIsland = len(islands) == 1
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
