package allocator

import "testing"

// byID indexes a fleet so a test can inspect the device a scheduler bound.
func byID(fleet []Device) map[string]Device {
	m := make(map[string]Device, len(fleet))
	for _, d := range fleet {
		m[d.ID] = d
	}
	return m
}

// Failure mode 1 — type rejection. A training job must not land on a
// non-trainable inference ASIC. The ASIC is placed first, so legacy grabs it.
func TestFailureModeTypeRejection(t *testing.T) {
	fleet := []Device{
		{ID: "inf-0", Category: CategoryInferASIC, Trainable: false, MemoryGB: 32, CostPerHr: 0.75, Precisions: []Precision{PrecisionFP16, PrecisionINT8}},
		{ID: "h100-0", Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8}},
	}
	dev := byID(fleet)
	job := Workload{Name: "train", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}

	sift, err := NewSiftScheduler(fleet).Place(job)
	if err != nil {
		t.Fatalf("sift: %v", err)
	}
	if got := dev[sift.DeviceIDs[0]]; !feasible(got, job) {
		t.Errorf("sift bound infeasible device %q", got.ID)
	}

	legacy, _ := NewLegacyScheduler(fleet).Place(job)
	if feasible(dev[legacy.DeviceIDs[0]], job) {
		t.Errorf("legacy was supposed to mis-type, but bound a feasible device %q", legacy.DeviceIDs[0])
	}
	if legacy.DeviceIDs[0] != "inf-0" {
		t.Errorf("legacy bound %q, expected the non-trainable trap inf-0", legacy.DeviceIDs[0])
	}
}

// Failure mode 2 — cost. A cost-sensitive job must take the cheapest fitting
// device, not the first free expensive one (placed first to tempt legacy).
func TestFailureModeCost(t *testing.T) {
	fleet := []Device{
		{ID: "b200-0", Category: CategoryGPU, Trainable: true, MemoryGB: 192, CostPerHr: 6.0, Precisions: []Precision{PrecisionBF16, PrecisionFP8}},
		{ID: "h100-0", Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP8}},
		{ID: "mi300x-0", Category: CategoryGPU, Trainable: true, MemoryGB: 192, CostPerHr: 1.9, Precisions: []Precision{PrecisionBF16, PrecisionFP8}},
	}
	job := Workload{Name: "batch", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 1}

	sift, _ := NewSiftScheduler(fleet).Place(job)
	legacy, _ := NewLegacyScheduler(fleet).Place(job)

	if sift.DeviceIDs[0] != "mi300x-0" {
		t.Errorf("sift bound %q @ %.2f, want cheapest fitting mi300x-0 @ 1.90", sift.DeviceIDs[0], sift.CostPerHr)
	}
	if legacy.DeviceIDs[0] != "b200-0" {
		t.Errorf("legacy bound %q, want the first/expensive b200-0", legacy.DeviceIDs[0])
	}
	if sift.CostPerHr >= legacy.CostPerHr {
		t.Errorf("expected Sift cheaper than legacy: sift %.2f vs legacy %.2f", sift.CostPerHr, legacy.CostPerHr)
	}
}
