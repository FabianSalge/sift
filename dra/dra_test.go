package dra

import (
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// Describe maps a Device's capabilities to neutral DRA attributes: strings, the
// trainable bool, node/island ints, and one boolean per supported precision, plus
// memory as capacity.
func TestDescribe(t *testing.T) {
	d := allocator.Device{
		ID: "h100-0", Node: 2, IslandID: 1,
		Vendor: allocator.VendorNVIDIA, Category: allocator.CategoryGPU,
		MemoryGB: 80, CostPerHr: 2.5,
		Precisions:   []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP8},
		Interconnect: allocator.InterconnectNVLink, Trainable: true,
	}
	got := Describe(d)

	if got.Name != "h100-0" {
		t.Errorf("Name = %q, want h100-0", got.Name)
	}
	if got.MemoryGB != 80 {
		t.Errorf("MemoryGB = %v, want 80", got.MemoryGB)
	}
	for k, want := range map[string]string{"vendor": "nvidia", "category": "gpu", "interconnect": "nvlink", "cost_per_hr": "2.5"} {
		if got.StringAttrs[k] != want {
			t.Errorf("StringAttrs[%q] = %q, want %q", k, got.StringAttrs[k], want)
		}
	}
	for k, want := range map[string]bool{"trainable": true, "precision_bf16": true, "precision_fp8": true} {
		if got.BoolAttrs[k] != want {
			t.Errorf("BoolAttrs[%q] = %v, want %v", k, got.BoolAttrs[k], want)
		}
	}
	if _, ok := got.BoolAttrs["precision_fp16"]; ok {
		t.Error("precision_fp16 should be absent (device doesn't support it)")
	}
	if got.IntAttrs["node"] != 2 || got.IntAttrs["island"] != 1 {
		t.Errorf("node/island = %d/%d, want 2/1", got.IntAttrs["node"], got.IntAttrs["island"])
	}
}

// A standalone device (NoIsland) must not get an island attribute — otherwise a
// same-island matchAttribute constraint could wrongly group standalone devices.
func TestDescribeOmitsIslandWhenStandalone(t *testing.T) {
	got := Describe(allocator.Device{ID: "inf-0", Node: 0, IslandID: allocator.NoIsland, Vendor: allocator.VendorAWS, Category: allocator.CategoryInferASIC, MemoryGB: 32, Interconnect: allocator.InterconnectNeuronLink})
	if _, ok := got.IntAttrs["island"]; ok {
		t.Errorf("island attr should be omitted for a NoIsland device, got %d", got.IntAttrs["island"])
	}
	if got.IntAttrs["node"] != 0 {
		t.Error("node attr should still be present")
	}
}
