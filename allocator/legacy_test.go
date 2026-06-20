package allocator

import (
	"errors"
	"testing"
)

// LegacyScheduler grabs the first free device in fleet order, ignoring every
// capability — so a training job lands on the non-trainable trap placed first.
func TestLegacyScheduler(t *testing.T) {
	pool := []Device{
		{ID: "inf-0", Trainable: false, MemoryGB: 32}, // the trap, placed first
		{ID: "h100-0", Trainable: true, MemoryGB: 80},
	}
	job := Workload{Name: "train", Kind: KindTrain, MinMemoryGB: 80}

	s := NewLegacyScheduler(pool)

	p1, err := s.Place(job)
	if err != nil {
		t.Fatalf("place 1: %v", err)
	}
	if p1.DeviceIDs[0] != "inf-0" {
		t.Errorf("place 1 = %+v, want inf-0 (first free, capabilities ignored)", p1)
	}

	p2, _ := s.Place(job)
	if p2.DeviceIDs[0] != "h100-0" {
		t.Errorf("place 2 = %+v, want h100-0", p2)
	}

	if _, err := s.Place(job); !errors.Is(err, ErrNoFeasibleDevice) {
		t.Errorf("place 3 err = %v, want ErrNoFeasibleDevice", err)
	}
}

// Legacy multi-device grabs the first n free devices in fleet order, ignoring the
// same-island constraint — so a gang gets fragmented across islands.
func TestLegacyMultiDevice(t *testing.T) {
	pool := []Device{
		{ID: "a", IslandID: 0}, {ID: "b", IslandID: 0},
		{ID: "c", IslandID: 1}, {ID: "d", IslandID: 1},
	}
	job := Workload{Name: "gang", DeviceCount: 3, SameIsland: true}

	p, err := NewLegacyScheduler(pool).Place(job)
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if len(p.DeviceIDs) != 3 || p.DeviceIDs[0] != "a" || p.DeviceIDs[1] != "b" || p.DeviceIDs[2] != "c" {
		t.Errorf("bound %v, want the first three free [a b c]", p.DeviceIDs)
	}

	if _, err := NewLegacyScheduler(pool).Place(Workload{DeviceCount: 5}); !errors.Is(err, ErrNoFeasibleDevice) {
		t.Errorf("err = %v, want ErrNoFeasibleDevice (only 4 devices)", err)
	}
}
