package allocator

import "errors"

// ErrNoFeasibleDevice means no free device satisfied the workload — the
// scheduler's "Pending" outcome.
var ErrNoFeasibleDevice = errors.New("no feasible device available")

// Placement is the result of binding a workload. DeviceIDs is a slice so
// multi-device workloads (increment 2) reuse the same shape.
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

// Place binds the best free, feasible device to w, or ErrNoFeasibleDevice.
func (s *SiftScheduler) Place(w Workload) (Placement, error) {
	var candidates []Device
	for _, d := range s.devices {
		if !s.allocated[d.ID] && feasible(d, w) {
			candidates = append(candidates, d)
		}
	}
	best, ok := bestDevice(candidates, w)
	if !ok {
		return Placement{}, ErrNoFeasibleDevice
	}
	s.allocated[best.ID] = true
	return Placement{Workload: w.Name, DeviceIDs: []string{best.ID}, CostPerHr: best.CostPerHr}, nil
}

// feasible reports whether a device satisfies a workload's hard constraints:
// trainability (for training jobs), memory, and the required-precision subset.
func feasible(d Device, w Workload) bool {
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
	set := make(map[Precision]struct{}, len(have))
	for _, p := range have {
		set[p] = struct{}{}
	}
	for _, p := range required {
		if _, ok := set[p]; !ok {
			return false
		}
	}
	return true
}
