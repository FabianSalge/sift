package allocator

// AllocateSift returns the device IDs Sift would bind for w against the devices
// already taken (allocated[id]==true), or ok=false if the whole request can't be
// satisfied. Stateless: it records nothing. The stateful SiftScheduler.Place
// wraps it, so the matching logic has one source.
func AllocateSift(devices []Device, w Workload, allocated map[string]bool) ([]string, bool) {
	n := w.DeviceCount
	if n < 1 {
		n = 1
	}
	var candidates []Device
	for _, d := range devices {
		if !allocated[d.ID] && Feasible(d, w) {
			candidates = append(candidates, d)
		}
	}
	chosen, ok := selectN(candidates, w, n)
	if !ok {
		return nil, false
	}
	ids := make([]string, len(chosen))
	for i, d := range chosen {
		ids[i] = d.ID
	}
	return ids, true
}

// AllocateLegacy returns the first n = max(1, DeviceCount) free devices in fleet
// order, blind to fit — the integer device-plugin behavior.
func AllocateLegacy(devices []Device, w Workload, allocated map[string]bool) ([]string, bool) {
	n := w.DeviceCount
	if n < 1 {
		n = 1
	}
	var ids []string
	for _, d := range devices {
		if !allocated[d.ID] {
			ids = append(ids, d.ID)
			if len(ids) == n {
				return ids, true
			}
		}
	}
	return nil, false
}

// sumCost totals the hourly cost of the named devices.
func sumCost(devices []Device, ids []string) float64 {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	var c float64
	for _, d := range devices {
		if want[d.ID] {
			c += d.CostPerHr
		}
	}
	return c
}
