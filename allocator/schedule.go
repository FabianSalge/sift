package allocator

import (
	"errors"
	"sort"
)

// ErrNoFeasibleDevice means no free device satisfied the workload — the
// scheduler's "Pending" outcome.
var ErrNoFeasibleDevice = errors.New("no feasible device available")

// Placement is the result of binding a workload. DeviceIDs is a slice so
// multi-device (gang) workloads reuse the same shape.
type Placement struct {
	Workload  string
	DeviceIDs []string
	CostPerHr float64
}

// SiftScheduler places workloads filter -> score -> bind, tracking which devices
// it has already handed out.
type SiftScheduler struct {
	devices   []Device
	allocated map[string]bool
}

func NewSiftScheduler(devices []Device) *SiftScheduler {
	return &SiftScheduler{devices: devices, allocated: make(map[string]bool)}
}

// Place binds n = max(1, DeviceCount) free, feasible devices to w, or
// ErrNoFeasibleDevice if it cannot satisfy the whole request (no partial bind).
func (s *SiftScheduler) Place(w Workload) (Placement, error) {
	ids, ok := AllocateSift(s.devices, w, s.allocated)
	if !ok {
		return Placement{}, ErrNoFeasibleDevice
	}
	for _, id := range ids {
		s.allocated[id] = true
	}
	return Placement{Workload: w.Name, DeviceIDs: ids, CostPerHr: sumCost(s.devices, ids)}, nil
}

// selectN picks the n devices to bind. A same-island gang (n>1) must come from a
// single island; otherwise the n best are taken globally.
func selectN(candidates []Device, w Workload, n int) ([]Device, bool) {
	if w.SameIsland && n > 1 {
		return selectIsland(candidates, w, n)
	}
	return pickBestN(candidates, w, n)
}

// pickBestN returns the n best feasible devices by lessScore, or ok=false if
// fewer than n exist.
func pickBestN(candidates []Device, w Workload, n int) ([]Device, bool) {
	if len(candidates) < n {
		return nil, false
	}
	remaining := append([]Device(nil), candidates...)
	chosen := make([]Device, 0, n)
	for k := 0; k < n; k++ {
		best, ok := bestDevice(remaining, w)
		if !ok {
			return nil, false
		}
		chosen = append(chosen, best)
		remaining = removeByID(remaining, best.ID)
	}
	return chosen, true
}

// selectIsland returns the n best devices from the cheapest single island that
// can hold the whole gang, or ok=false if no island has n feasible devices.
func selectIsland(candidates []Device, w Workload, n int) ([]Device, bool) {
	byIsland := map[int][]Device{}
	var islands []int
	for _, d := range candidates {
		if d.IslandID == NoIsland {
			continue
		}
		if _, seen := byIsland[d.IslandID]; !seen {
			islands = append(islands, d.IslandID)
		}
		byIsland[d.IslandID] = append(byIsland[d.IslandID], d)
	}
	sort.Ints(islands) // deterministic; ties resolve to the lower island ID

	var best []Device
	var bestCost, bestWaste float64
	found := false
	for _, id := range islands {
		group, ok := pickBestN(byIsland[id], w, n)
		if !ok {
			continue
		}
		cost, waste := groupScore(group, w)
		if !found || cost < bestCost || (cost == bestCost && waste < bestWaste) {
			best, bestCost, bestWaste, found = group, cost, waste, true
		}
	}
	return best, found
}

// groupScore sums an island group's cost-weighted price and memory waste — the
// keys used to choose between islands that can each hold the gang.
func groupScore(group []Device, w Workload) (cost, waste float64) {
	for _, d := range group {
		cost += w.CostWeight * d.CostPerHr
		waste += d.MemoryGB - w.MinMemoryGB
	}
	return cost, waste
}

func removeByID(devs []Device, id string) []Device {
	var out []Device
	for _, d := range devs {
		if d.ID != id {
			out = append(out, d)
		}
	}
	return out
}

// Feasible reports whether a device satisfies a workload's hard constraints:
// trainability (for training jobs), memory, and the required-precision subset.
func Feasible(d Device, w Workload) bool {
	if w.Kind == KindTrain && !d.Trainable {
		return false
	}
	if d.MemoryGB < w.MinMemoryGB {
		return false
	}
	return supportsAll(d.Precisions, w.RequiredPrecisions)
}

// bestDevice returns the highest-scoring candidate, or ok=false if none. It
// assumes candidates already passed feasible.
func bestDevice(candidates []Device, w Workload) (Device, bool) {
	var best Device
	found := false
	for _, d := range candidates {
		if !found || lessScore(d, best, w) {
			best, found = d, true
		}
	}
	return best, found
}

// lessScore orders devices by the soft-preference key
// (CostWeight*CostPerHr, memory waste, ID): cost dominates when the job is
// cost-sensitive, else best-fit packing wins, with ID as a stable tiebreak.
func lessScore(a, b Device, w Workload) bool {
	if ca, cb := w.CostWeight*a.CostPerHr, w.CostWeight*b.CostPerHr; ca != cb {
		return ca < cb
	}
	if wa, wb := a.MemoryGB-w.MinMemoryGB, b.MemoryGB-w.MinMemoryGB; wa != wb {
		return wa < wb
	}
	return a.ID < b.ID
}

// supportsAll reports whether every required precision is in have.
func supportsAll(have, required []Precision) bool {
	return len(missingPrecisions(have, required)) == 0
}
