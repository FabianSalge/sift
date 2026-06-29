package allocator

import "testing"

func allocFixture() []Device {
	return []Device{
		{ID: "inf2-0", IslandID: NoIsland, Category: CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []Precision{PrecisionFP16, PrecisionINT8}},
		{ID: "h100-0", IslandID: 1, Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8}},
		{ID: "h100-1", IslandID: 1, Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8}},
		{ID: "mi300x-0", IslandID: 2, Category: CategoryGPU, Trainable: true, MemoryGB: 192, CostPerHr: 1.9, Precisions: []Precision{PrecisionBF16, PrecisionFP16}},
	}
}

// AllocateSift on an empty allocation must equal what the stateful Place binds.
func TestAllocateSiftMatchesPlace(t *testing.T) {
	fleet := allocFixture()
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	want, err := NewSiftScheduler(fleet).Place(w)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := AllocateSift(fleet, w, map[string]bool{})
	if !ok || len(got) != len(want.DeviceIDs) {
		t.Fatalf("AllocateSift = %v,%v; want %v", got, ok, want.DeviceIDs)
	}
	for i := range want.DeviceIDs {
		if got[i] != want.DeviceIDs[i] {
			t.Errorf("AllocateSift[%d] = %s, want %s", i, got[i], want.DeviceIDs[i])
		}
	}
}

// AllocateLegacy on an empty allocation must equal what the stateful Place binds.
func TestAllocateLegacyMatchesPlace(t *testing.T) {
	fleet := allocFixture()
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	want, err := NewLegacyScheduler(fleet).Place(w)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := AllocateLegacy(fleet, w, map[string]bool{})
	if !ok || len(got) != len(want.DeviceIDs) || got[0] != want.DeviceIDs[0] {
		t.Fatalf("AllocateLegacy = %v,%v; want %v", got, ok, want.DeviceIDs)
	}
}

// A non-empty allocation is respected (taken devices are skipped).
func TestAllocateRespectsAllocated(t *testing.T) {
	fleet := allocFixture()
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	// mi300x-0 is Sift's cheapest fit; mark it taken → Sift must pick an h100.
	got, ok := AllocateSift(fleet, w, map[string]bool{"mi300x-0": true})
	if !ok || got[0] != "h100-0" {
		t.Fatalf("AllocateSift with mi300x taken = %v, want [h100-0]", got)
	}
	// Legacy takes the first free in fleet order → inf2-0.
	gl, ok := AllocateLegacy(fleet, w, map[string]bool{})
	if !ok || gl[0] != "inf2-0" {
		t.Fatalf("AllocateLegacy = %v, want [inf2-0]", gl)
	}
}
