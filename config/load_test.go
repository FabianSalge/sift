package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// A minimal one-class fleet should expand its count into N devices, map every
// catalog field onto allocator.Device, generate per-class IDs in file order,
// and default IslandID to NoIsland when no island is given.
func TestLoadFleetExpandsCountAndMapsFields(t *testing.T) {
	const yaml = `
catalog:
  h100:
    vendor: nvidia
    category: gpu
    memoryGB: 80
    precisions: [bf16, fp16, fp8]
    interconnect: nvlink
    costPerHr: 2.5
    trainable: true
fleet:
  - { class: h100, count: 2, node: 1 }
`
	got, err := LoadFleet(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadFleet returned error: %v", err)
	}

	want := []allocator.Device{
		{
			ID:           "h100-0",
			Node:         1,
			IslandID:     allocator.NoIsland,
			Vendor:       allocator.VendorNVIDIA,
			Category:     allocator.CategoryGPU,
			MemoryGB:     80,
			Precisions:   []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16, allocator.PrecisionFP8},
			Interconnect: allocator.InterconnectNVLink,
			CostPerHr:    2.5,
			Trainable:    true,
		},
		{
			ID:           "h100-1",
			Node:         1,
			IslandID:     allocator.NoIsland,
			Vendor:       allocator.VendorNVIDIA,
			Category:     allocator.CategoryGPU,
			MemoryGB:     80,
			Precisions:   []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16, allocator.PrecisionFP8},
			Interconnect: allocator.InterconnectNVLink,
			CostPerHr:    2.5,
			Trainable:    true,
		},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d devices, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !deviceEqual(got[i], want[i]) {
			t.Errorf("device %d:\n got  %+v\n want %+v", i, got[i], want[i])
		}
	}
}

// Placements are expanded in file order; the per-class ID counter continues
// across separate placements of the same class; island is read when present and
// defaults to NoIsland when absent.
func TestLoadFleetOrderingIslandsAndPerClassIDs(t *testing.T) {
	const yaml = `
catalog:
  h100:
    vendor: nvidia
    category: gpu
    memoryGB: 80
    precisions: [bf16]
    interconnect: nvlink
    costPerHr: 2.5
    trainable: true
  mi300x:
    vendor: amd
    category: gpu
    memoryGB: 192
    precisions: [bf16, fp8]
    interconnect: infinity-fabric
    costPerHr: 1.9
    trainable: true
fleet:
  - { class: h100,   count: 2, node: 0, island: 0 }
  - { class: mi300x, count: 1, node: 1, island: 5 }
  - { class: h100,   count: 1, node: 2 }
`
	got, err := LoadFleet(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadFleet returned error: %v", err)
	}

	type place struct {
		id     string
		node   int
		island int
	}
	want := []place{
		{"h100-0", 0, 0},
		{"h100-1", 0, 0},
		{"mi300x-0", 1, 5},
		{"h100-2", 2, allocator.NoIsland},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d devices, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].ID != w.id || got[i].Node != w.node || got[i].IslandID != w.island {
			t.Errorf("device %d: got (id=%s node=%d island=%d), want (id=%s node=%d island=%d)",
				i, got[i].ID, got[i].Node, got[i].IslandID, w.id, w.node, w.island)
		}
	}
}

// Invalid scenarios must be rejected with an error naming the offending
// field/value, rather than silently producing wrong devices.
func TestLoadFleetRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name      string
		yaml      string
		wantInErr string
	}{
		{
			name:      "unknown vendor",
			yaml:      base(`vendor: intel`, ``),
			wantInErr: "vendor",
		},
		{
			name:      "unknown category",
			yaml:      base(`category: dpu`, ``),
			wantInErr: "category",
		},
		{
			name:      "unknown interconnect",
			yaml:      base(`interconnect: pcie`, ``),
			wantInErr: "interconnect",
		},
		{
			name:      "unknown precision",
			yaml:      base(`precisions: [bf16, fp32]`, ``),
			wantInErr: "precision",
		},
		{
			name:      "empty precisions",
			yaml:      base(`precisions: []`, ``),
			wantInErr: "precision",
		},
		{
			name:      "non-positive memory",
			yaml:      base(`memoryGB: 0`, ``),
			wantInErr: "memoryGB",
		},
		{
			name:      "negative cost",
			yaml:      base(`costPerHr: -1`, ``),
			wantInErr: "costPerHr",
		},
		{
			name:      "fleet references unknown class",
			yaml:      base(``, `class: a100`),
			wantInErr: "a100",
		},
		{
			name:      "count below one",
			yaml:      base(``, `count: 0`),
			wantInErr: "count",
		},
		{
			name:      "negative node",
			yaml:      base(``, `node: -1`),
			wantInErr: "node",
		},
		{
			name:      "negative island",
			yaml:      base(``, `island: -1`),
			wantInErr: "island",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadFleet(strings.NewReader(tc.yaml))
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantInErr)
			}
		})
	}
}

// The shipped realistic-2026 scenario must load cleanly and arm all three
// failure modes with the traps placed early (ADR-0014). This test is the
// fixture's witness: if an edit disarms a trap, it fails here.
func TestRealistic2026FixtureArmsAllTraps(t *testing.T) {
	devices, err := LoadFleetFile("../scenarios/realistic-2026.yaml")
	if err != nil {
		t.Fatalf("LoadFleetFile: %v", err)
	}
	if len(devices) == 0 {
		t.Fatal("scenario produced no devices")
	}

	// Type-rejection trap: a non-trainable inference ASIC sits first, so a
	// first-fit legacy scheduler is tempted to put a training job on it.
	if devices[0].Trainable || devices[0].Category != allocator.CategoryInferASIC {
		t.Errorf("trap 1 (type): want a non-trainable infer-asic first, got %+v", devices[0])
	}

	// Cost trap: the first *trainable* GPU is not the cheapest trainable GPU, so
	// first-fit overspends relative to the cheapest fitting device.
	firstTrainableCost := -1.0
	minTrainableCost := 0.0
	for _, d := range devices {
		if !d.Trainable {
			continue
		}
		if firstTrainableCost < 0 {
			firstTrainableCost = d.CostPerHr
			minTrainableCost = d.CostPerHr
		}
		if d.CostPerHr < minTrainableCost {
			minTrainableCost = d.CostPerHr
		}
	}
	if firstTrainableCost <= minTrainableCost {
		t.Errorf("trap 2 (cost): first trainable GPU cost %.2f is not above cheapest %.2f", firstTrainableCost, minTrainableCost)
	}

	// Topology trap: at least two distinct interconnect islands of trainable
	// devices exist, so a same-island multi-device job can be fragmented.
	islands := map[int]int{}
	for _, d := range devices {
		if d.Trainable && d.IslandID != allocator.NoIsland {
			islands[d.IslandID]++
		}
	}
	if len(islands) < 2 {
		t.Errorf("trap 3 (topology): want >=2 trainable islands, got %d (%v)", len(islands), islands)
	}
}

// base builds a minimal-but-valid scenario YAML, applying an optional override
// line to the single catalog class and to the single fleet placement. An empty
// override leaves that section at its valid default.
func base(classOverride, placementOverride string) string {
	class := map[string]string{
		"vendor":       "nvidia",
		"category":     "gpu",
		"memoryGB":     "80",
		"precisions":   "[bf16]",
		"interconnect": "nvlink",
		"costPerHr":    "2.5",
		"trainable":    "true",
	}
	place := map[string]string{
		"class": "h100",
		"count": "1",
		"node":  "0",
	}
	apply := func(m map[string]string, override string) {
		if override == "" {
			return
		}
		k, v, _ := strings.Cut(override, ":")
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	apply(class, classOverride)
	apply(place, placementOverride)

	var sb strings.Builder
	sb.WriteString("catalog:\n  h100:\n")
	for _, k := range []string{"vendor", "category", "memoryGB", "precisions", "interconnect", "costPerHr", "trainable"} {
		fmt.Fprintf(&sb, "    %s: %s\n", k, class[k])
	}
	sb.WriteString("fleet:\n  - {")
	parts := []string{}
	for _, k := range []string{"class", "count", "node", "island"} {
		if v, ok := place[k]; ok {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		}
	}
	sb.WriteString(strings.Join(parts, ", "))
	sb.WriteString("}\n")
	return sb.String()
}

// deviceEqual compares two devices including their precision slices.
func deviceEqual(a, b allocator.Device) bool {
	if a.ID != b.ID || a.Node != b.Node || a.IslandID != b.IslandID ||
		a.Vendor != b.Vendor || a.Category != b.Category || a.MemoryGB != b.MemoryGB ||
		a.Interconnect != b.Interconnect || a.CostPerHr != b.CostPerHr || a.Trainable != b.Trainable {
		return false
	}
	if len(a.Precisions) != len(b.Precisions) {
		return false
	}
	for i := range a.Precisions {
		if a.Precisions[i] != b.Precisions[i] {
			return false
		}
	}
	return true
}
