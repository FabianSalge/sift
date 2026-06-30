package sim

import (
	"reflect"
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

func gpu(id string, island int) allocator.Device {
	return allocator.Device{ID: id, IslandID: island, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []allocator.Precision{allocator.PrecisionBF16}}
}
func train(name string) allocator.Workload {
	return allocator.Workload{Name: name, Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.5}
}

func find(r SchedulerResult, name string) ArrivalResult {
	for _, a := range r.Arrivals {
		if a.Workload == name {
			return a
		}
	}
	return ArrivalResult{PlacedAt: -1}
}

// Same input → identical Result (no clocks, no RNG).
func TestRunDeterministic(t *testing.T) {
	fleet := []allocator.Device{gpu("h100-0", 1)}
	stream := Stream{{At: 0, Workload: train("a"), Duration: 5}, {At: 6, Workload: train("b"), Duration: 5}}
	if !reflect.DeepEqual(Run(fleet, stream), Run(fleet, stream)) {
		t.Error("Run is not deterministic")
	}
}

// A device frees on completion and a later arrival reuses it.
func TestDeviceFreedOnCompletion(t *testing.T) {
	fleet := []allocator.Device{gpu("h100-0", 1)}
	stream := Stream{{At: 0, Workload: train("a"), Duration: 5}, {At: 6, Workload: train("b"), Duration: 5}}
	r := Run(fleet, stream).Sift
	b := find(r, "b")
	if b.PlacedAt != 6 || len(b.DeviceIDs) != 1 || b.DeviceIDs[0] != "h100-0" {
		t.Errorf("b placed at %v on %v, want 6 on [h100-0]", b.PlacedAt, b.DeviceIDs)
	}
}

// A job that arrives while the only device is busy waits until it frees.
func TestQueueWaitsForCapacity(t *testing.T) {
	fleet := []allocator.Device{gpu("h100-0", 1)}
	stream := Stream{{At: 0, Workload: train("a"), Duration: 10}, {At: 1, Workload: train("b"), Duration: 5}}
	b := find(Run(fleet, stream).Sift, "b")
	if b.PlacedAt != 10 {
		t.Errorf("b placed at %v, want 10 (waits for a to free h100-0)", b.PlacedAt)
	}
}

// Sift places a training job on a capable GPU (useful); legacy grabs the first
// free device — a non-trainable inference ASIC — which is occupied but wasted.
func TestUsefulVsWasted(t *testing.T) {
	fleet := []allocator.Device{
		{ID: "inf2-0", IslandID: allocator.NoIsland, Category: allocator.CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []allocator.Precision{allocator.PrecisionINT8}},
		gpu("h100-0", 1),
	}
	stream := Stream{{At: 0, Workload: train("t"), Duration: 5}}
	res := Run(fleet, stream)

	s := find(res.Sift, "t")
	if !s.Useful || s.DeviceIDs[0] != "h100-0" {
		t.Errorf("Sift t = %+v, want useful on h100-0", s)
	}
	l := find(res.Legacy, "t")
	if l.Useful || l.DeviceIDs[0] != "inf2-0" {
		t.Errorf("Legacy t = %+v, want wasted on inf2-0", l)
	}
}

// Horizon is the last completion time across both schedulers.
func TestHorizon(t *testing.T) {
	fleet := []allocator.Device{gpu("h100-0", 1)}
	stream := Stream{{At: 0, Workload: train("a"), Duration: 7}}
	if h := Run(fleet, stream).Horizon; h != 7 {
		t.Errorf("Horizon = %v, want 7", h)
	}
}
