package allocator

// LegacyScheduler models the integer device-plugin: it hands out the first free
// device in fleet order, blind to type, memory, precision, and cost. That
// blindness is why early-placed traps (ADR-0014) mis-assign it.
type LegacyScheduler struct {
	devices   []Device
	allocated map[string]bool
}

func NewLegacyScheduler(devices []Device) *LegacyScheduler {
	return &LegacyScheduler{devices: devices, allocated: make(map[string]bool)}
}

// Place binds the first n = max(1, DeviceCount) free devices in fleet order,
// regardless of whether they fit or share an island.
func (s *LegacyScheduler) Place(w Workload) (Placement, error) {
	ids, ok := AllocateLegacy(s.devices, w, s.allocated)
	if !ok {
		return Placement{}, ErrNoFeasibleDevice
	}
	for _, id := range ids {
		s.allocated[id] = true
	}
	return Placement{Workload: w.Name, DeviceIDs: ids, CostPerHr: sumCost(s.devices, ids)}, nil
}
