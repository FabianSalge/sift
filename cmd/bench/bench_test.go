package main

import (
	"strings"
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// run must show the contrast: Sift never mis-types or fragments; legacy does
// both (it grabs the non-trainable trap first, and spills the gang across
// islands).
func TestRunContrast(t *testing.T) {
	gpu := func(id string, island int, cost float64) allocator.Device {
		return allocator.Device{ID: id, IslandID: island, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: cost, Precisions: []allocator.Precision{allocator.PrecisionBF16}}
	}
	fleet := []allocator.Device{
		{ID: "inf-0", Category: allocator.CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []allocator.Precision{allocator.PrecisionINT8}}, // trap, first
		gpu("h100-0", 1, 2.5), gpu("mi300x-0", 2, 1.9), gpu("h100-1", 1, 2.5), gpu("mi300x-1", 2, 1.9),
	}
	workloads := []allocator.Workload{
		{Name: "train", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, CostWeight: 1},
		{Name: "gang", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 2, SameIsland: true, CostWeight: 1},
	}

	rep := run(fleet, workloads)

	if rep.Sift.TypeCorrect != 2 {
		t.Errorf("Sift type-correct = %d, want 2", rep.Sift.TypeCorrect)
	}
	if rep.Legacy.TypeCorrect >= 2 {
		t.Errorf("Legacy type-correct = %d, want < 2 (it mis-types)", rep.Legacy.TypeCorrect)
	}
	if rep.Sift.Fragmented != 0 {
		t.Errorf("Sift fragmented = %d, want 0", rep.Sift.Fragmented)
	}
	if rep.Legacy.Fragmented < 1 {
		t.Errorf("Legacy fragmented = %d, want >= 1", rep.Legacy.Fragmented)
	}
	if rep.Sift.GangsWhole != 1 {
		t.Errorf("Sift gangs-whole = %d, want 1", rep.Sift.GangsWhole)
	}

	out := format(rep)
	for _, want := range []string{"Sift", "Legacy", "type-correct"} {
		if !strings.Contains(out, want) {
			t.Errorf("format() missing %q:\n%s", want, out)
		}
	}
}
