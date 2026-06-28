package main

import (
	"strings"
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// formatExplain must render a real filter -> score -> bind trace: the rejected
// trap with its reasons, the feasible survivors ranked, and the bound winner
// with its true hourly price.
func TestFormatExplain(t *testing.T) {
	fleet := []allocator.Device{
		{ID: "inferentia2-0", IslandID: allocator.NoIsland, Vendor: allocator.VendorAWS, Category: allocator.CategoryInferASIC, MemoryGB: 32, Precisions: []allocator.Precision{allocator.PrecisionFP16, allocator.PrecisionINT8}, CostPerHr: 0.75, Trainable: false},
		{ID: "h100-0", IslandID: 1, Vendor: allocator.VendorNVIDIA, Category: allocator.CategoryGPU, MemoryGB: 80, Precisions: []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16, allocator.PrecisionFP8}, CostPerHr: 2.5, Trainable: true},
		{ID: "mi300x-0", IslandID: 3, Vendor: allocator.VendorAMD, Category: allocator.CategoryGPU, MemoryGB: 192, Precisions: []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16, allocator.PrecisionFP8}, CostPerHr: 1.9, Trainable: true},
	}
	w := allocator.Workload{Name: "train-llm", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.5}

	out := formatExplain("realistic-2026", w, allocator.Explain(fleet, w, nil), fleet)

	// Header carries the workload's hard requirements and soft weight.
	for _, want := range []string{"train-llm", "bf16", "cost-weight 0.5", "filter", "score", "bind"} {
		if !strings.Contains(out, want) {
			t.Errorf("header missing %q:\n%s", want, out)
		}
	}

	line := traceLine(out, "inferentia2-0")
	for _, want := range []string{"reject", "not trainable", "bf16"} {
		if !strings.Contains(line, want) {
			t.Errorf("inferentia2-0 line missing %q: %q", want, line)
		}
	}

	// mi300x-0 is the cheapest feasible device: rank 1, the bound winner.
	bind := traceLine(out, "mi300x-0")
	for _, want := range []string{"ok", "BIND", "1"} {
		if !strings.Contains(bind, want) {
			t.Errorf("mi300x-0 line missing %q: %q", want, bind)
		}
	}

	// h100-0 fits but is pricier: feasible, not bound.
	if h := traceLine(out, "h100-0"); !strings.Contains(h, "ok") || strings.Contains(h, "BIND") {
		t.Errorf("h100-0 should be feasible but not bound: %q", h)
	}

	// Footer states the bound device and its true hourly price (not the weighted score).
	if !strings.Contains(out, "$1.90") {
		t.Errorf("footer missing bound price $1.90:\n%s", out)
	}
}

// traceLine returns the first output line containing id, for per-device assertions.
func traceLine(out, id string) string {
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, id) {
			return l
		}
	}
	return ""
}
