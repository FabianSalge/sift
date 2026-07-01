package sim

import (
	"strconv"
	"testing"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/config"
)

// On the real realistic-2026 fleet, a training-heavy stream must diverge: Sift
// completes more useful work than legacy, and legacy wastes capacity on
// placements that can't run (its first-fit grabs the non-trainable Inferentia2s
// at the head of the fleet). This is the headline the Stream view exists to show
// (spec §9); it guards the mechanism against sim/allocator/fleet regressions.
func TestRealisticStreamDiverges(t *testing.T) {
	fleet, err := config.LoadFleetFile("../scenarios/realistic-2026.yaml")
	if err != nil {
		t.Fatal(err)
	}

	trainJob := func(name string) allocator.Workload {
		return allocator.Workload{
			Name: name, Kind: allocator.KindTrain, MinMemoryGB: 80,
			RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.5,
		}
	}
	var stream Stream
	for i := 0; i < 8; i++ {
		stream = append(stream, Arrival{At: float64(i), Workload: trainJob("train-" + strconv.Itoa(i)), Duration: 12})
	}

	res := Run(fleet, stream)
	usefulSift := usefulDone(res.Sift)
	usefulLegacy := usefulDone(res.Legacy)
	wastedLegacy := wastedPlacements(res.Legacy)

	if usefulSift <= usefulLegacy {
		t.Errorf("Sift useful=%d not greater than legacy useful=%d — no divergence", usefulSift, usefulLegacy)
	}
	if wastedLegacy == 0 {
		t.Errorf("legacy wasted=%d, want > 0 (it should squander the Inferentia2s on training jobs)", wastedLegacy)
	}
}

func usefulDone(r SchedulerResult) int {
	n := 0
	for _, a := range r.Arrivals {
		if a.Useful && a.End >= 0 {
			n++
		}
	}
	return n
}

func wastedPlacements(r SchedulerResult) int {
	n := 0
	for _, a := range r.Arrivals {
		if a.PlacedAt >= 0 && !a.Useful {
			n++
		}
	}
	return n
}
