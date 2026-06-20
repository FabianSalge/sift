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
	n := w.DeviceCount
	if n < 1 {
		n = 1
	}
	var chosen []Device
	for _, d := range s.devices {
		if !s.allocated[d.ID] {
			chosen = append(chosen, d)
			if len(chosen) == n {
				break
			}
		}
	}
	if len(chosen) < n {
		return Placement{}, ErrNoFeasibleDevice
	}
	ids := make([]string, n)
	var cost float64
	for i, d := range chosen {
		s.allocated[d.ID] = true
		ids[i] = d.ID
		cost += d.CostPerHr
	}
	return Placement{Workload: w.Name, DeviceIDs: ids, CostPerHr: cost}, nil
}
