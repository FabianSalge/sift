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

// Place binds the first free device, regardless of whether it fits.
func (s *LegacyScheduler) Place(w Workload) (Placement, error) {
	for _, d := range s.devices {
		if !s.allocated[d.ID] {
			s.allocated[d.ID] = true
			return Placement{Workload: w.Name, DeviceIDs: []string{d.ID}, CostPerHr: d.CostPerHr}, nil
		}
	}
	return Placement{}, ErrNoFeasibleDevice
}
