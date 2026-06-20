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
