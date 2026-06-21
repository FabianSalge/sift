// Package dra describes allocator devices in DRA's vocabulary — attributes and
// capacity — without importing any Kubernetes types, so the pure core stays lean.
// The driver (a separate module) converts these neutral values into resourceapi
// ResourceSlice devices.
package dra

import (
	"strconv"

	"github.com/FabianSalge/sift/allocator"
)

// Device is the neutral DRA shape of an allocator.Device: attribute maps split by
// value kind (DRA attributes are scalar: string/bool/int), plus memory capacity.
type Device struct {
	Name        string
	StringAttrs map[string]string
	BoolAttrs   map[string]bool
	IntAttrs    map[string]int64
	MemoryGB    float64 // published as the "memory" capacity quantity
}

// Describe maps a device's static capabilities to attributes: vendor/category/
// interconnect and cost as strings, trainability and each supported precision as
// booleans, node/island as ints (island omitted when standalone), memory as
// capacity.
func Describe(d allocator.Device) Device {
	out := Device{
		Name:     d.ID,
		MemoryGB: d.MemoryGB,
		StringAttrs: map[string]string{
			"vendor":       string(d.Vendor),
			"category":     string(d.Category),
			"interconnect": string(d.Interconnect),
			"cost_per_hr":  strconv.FormatFloat(d.CostPerHr, 'f', -1, 64),
		},
		BoolAttrs: map[string]bool{"trainable": d.Trainable},
		IntAttrs:  map[string]int64{"node": int64(d.Node)},
	}
	if d.IslandID != allocator.NoIsland {
		out.IntAttrs["island"] = int64(d.IslandID)
	}
	for _, p := range d.Precisions {
		out.BoolAttrs["precision_"+string(p)] = true
	}
	return out
}
