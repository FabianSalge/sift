package allocator

import (
	"errors"
	"testing"
)

// feasible is the hard filter: a device must satisfy trainability (for training
// jobs), memory, and the required-precision subset.
func TestFeasible(t *testing.T) {
	trainJob := Workload{
		Kind:               KindTrain,
		MinMemoryGB:        80,
		RequiredPrecisions: []Precision{PrecisionBF16, PrecisionFP8},
	}
	gpu := Device{
		Trainable:  true,
		MemoryGB:   80,
		Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8},
	}

	cases := []struct {
		name string
		dev  Device
		work Workload
		want bool
	}{
		{"matching trainable gpu", gpu, trainJob, true},
		{"non-trainable rejected for training", withTrainable(gpu, false), trainJob, false},
		{"insufficient memory", withMemory(gpu, 40), trainJob, false},
		{"missing required precision", withPrecisions(gpu, PrecisionBF16, PrecisionFP16), trainJob, false},
		{"inference job ignores trainability", withTrainable(gpu, false), Workload{Kind: KindInfer, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := feasible(tc.dev, tc.work); got != tc.want {
				t.Errorf("feasible = %v, want %v", got, tc.want)
			}
		})
	}
}

// SiftScheduler.Place binds the best feasible free device, marks it allocated so
// it can't be reused, and reports ErrNoFeasibleDevice when nothing fits.
func TestSiftSchedulerPlace(t *testing.T) {
	pool := []Device{
		{ID: "h100-0", Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16}},
		{ID: "b200-0", Trainable: true, MemoryGB: 192, CostPerHr: 6.0, Precisions: []Precision{PrecisionBF16}},
	}
	job := Workload{Name: "train", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 1}

	s := NewSiftScheduler(pool)

	p1, err := s.Place(job)
	if err != nil {
		t.Fatalf("place 1: %v", err)
	}
	if len(p1.DeviceIDs) != 1 || p1.DeviceIDs[0] != "h100-0" || p1.CostPerHr != 2.5 {
		t.Errorf("place 1 = %+v, want cheapest h100-0 @ 2.5", p1)
	}

	p2, err := s.Place(job)
	if err != nil {
		t.Fatalf("place 2: %v", err)
	}
	if p2.DeviceIDs[0] != "b200-0" {
		t.Errorf("place 2 = %+v, want b200-0 (h100 already allocated)", p2)
	}

	if _, err := s.Place(job); !errors.Is(err, ErrNoFeasibleDevice) {
		t.Errorf("place 3 err = %v, want ErrNoFeasibleDevice", err)
	}
}

// A multi-device workload with no island constraint binds the n cheapest fitting
// devices, summing their cost.
func TestSiftSchedulerMultiDevice(t *testing.T) {
	pool := []Device{
		{ID: "mi300x-0", Trainable: true, MemoryGB: 192, CostPerHr: 1.9, Precisions: []Precision{PrecisionBF16}},
		{ID: "h100-0", Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16}},
		{ID: "b200-0", Trainable: true, MemoryGB: 192, CostPerHr: 6.0, Precisions: []Precision{PrecisionBF16}},
	}
	job := Workload{Name: "train", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, DeviceCount: 2, CostWeight: 1}

	p, err := NewSiftScheduler(pool).Place(job)
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if len(p.DeviceIDs) != 2 {
		t.Fatalf("bound %d devices, want 2: %+v", len(p.DeviceIDs), p)
	}
	bound := map[string]bool{p.DeviceIDs[0]: true, p.DeviceIDs[1]: true}
	if !bound["mi300x-0"] || !bound["h100-0"] {
		t.Errorf("bound %v, want the two cheapest mi300x-0 + h100-0", p.DeviceIDs)
	}
	if p.CostPerHr != 4.4 {
		t.Errorf("cost %.2f, want 4.40 (1.9 + 2.5)", p.CostPerHr)
	}
}

// A SameIsland gang must land wholly within one island, even when cheaper
// devices exist in an island too small to hold the whole gang.
func TestSiftSchedulerSameIsland(t *testing.T) {
	pool := []Device{
		{ID: "h100-0", IslandID: 1, Trainable: true, MemoryGB: 80, CostPerHr: 1.0, Precisions: []Precision{PrecisionBF16}},
		{ID: "h100-1", IslandID: 1, Trainable: true, MemoryGB: 80, CostPerHr: 1.0, Precisions: []Precision{PrecisionBF16}},
		{ID: "mi300x-0", IslandID: 2, Trainable: true, MemoryGB: 192, CostPerHr: 2.0, Precisions: []Precision{PrecisionBF16}},
		{ID: "mi300x-1", IslandID: 2, Trainable: true, MemoryGB: 192, CostPerHr: 2.0, Precisions: []Precision{PrecisionBF16}},
		{ID: "mi300x-2", IslandID: 2, Trainable: true, MemoryGB: 192, CostPerHr: 2.0, Precisions: []Precision{PrecisionBF16}},
		{ID: "mi300x-3", IslandID: 2, Trainable: true, MemoryGB: 192, CostPerHr: 2.0, Precisions: []Precision{PrecisionBF16}},
	}
	dev := byID(pool)
	job := Workload{Name: "gang", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, DeviceCount: 4, SameIsland: true, CostWeight: 1}

	p, err := NewSiftScheduler(pool).Place(job)
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if len(p.DeviceIDs) != 4 {
		t.Fatalf("bound %d devices, want 4: %+v", len(p.DeviceIDs), p)
	}
	for _, id := range p.DeviceIDs {
		if dev[id].IslandID != 2 {
			t.Errorf("bound %v — all should be in island 2 (the only island holding 4)", p.DeviceIDs)
			break
		}
	}
}

func TestSiftSchedulerInfeasible(t *testing.T) {
	s := NewSiftScheduler([]Device{{ID: "h100-0", Trainable: true, MemoryGB: 80}})
	if _, err := s.Place(Workload{Kind: KindTrain, MinMemoryGB: 999}); !errors.Is(err, ErrNoFeasibleDevice) {
		t.Errorf("err = %v, want ErrNoFeasibleDevice", err)
	}
}

func withTrainable(d Device, v bool) Device          { d.Trainable = v; return d }
func withMemory(d Device, gb float64) Device         { d.MemoryGB = gb; return d }
func withPrecisions(d Device, p ...Precision) Device { d.Precisions = p; return d }

// bestDevice scores feasible candidates by (CostWeight*CostPerHr, memory waste,
// ID): cheapest fitting when the job is cost-sensitive, best-fit otherwise, with
// a deterministic ID tiebreak.
func TestBestDevice(t *testing.T) {
	b200 := Device{ID: "b200-0", MemoryGB: 192, CostPerHr: 6.0}
	h100 := Device{ID: "h100-0", MemoryGB: 80, CostPerHr: 2.5}
	mi300x := Device{ID: "mi300x-0", MemoryGB: 192, CostPerHr: 1.9}

	t.Run("cost-sensitive picks cheapest fitting", func(t *testing.T) {
		got, ok := bestDevice([]Device{b200, h100, mi300x}, Workload{MinMemoryGB: 80, CostWeight: 1})
		if !ok || got.ID != "mi300x-0" {
			t.Errorf("got %q (ok=%v), want mi300x-0", got.ID, ok)
		}
	})

	t.Run("cost-indifferent picks best-fit", func(t *testing.T) {
		got, ok := bestDevice([]Device{b200, h100, mi300x}, Workload{MinMemoryGB: 80, CostWeight: 0})
		if !ok || got.ID != "h100-0" {
			t.Errorf("got %q (ok=%v), want h100-0 (least wasted memory)", got.ID, ok)
		}
	})

	t.Run("full tie breaks on ID", func(t *testing.T) {
		a := Device{ID: "h100-1", MemoryGB: 80, CostPerHr: 2.5}
		b := Device{ID: "h100-0", MemoryGB: 80, CostPerHr: 2.5}
		got, ok := bestDevice([]Device{a, b}, Workload{MinMemoryGB: 80, CostWeight: 1})
		if !ok || got.ID != "h100-0" {
			t.Errorf("got %q (ok=%v), want h100-0", got.ID, ok)
		}
	})

	t.Run("no candidates", func(t *testing.T) {
		if _, ok := bestDevice(nil, Workload{}); ok {
			t.Error("expected ok=false for empty candidates")
		}
	})
}
