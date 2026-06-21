package dra

import (
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// The domain the driver publishes bare attributes under (its name).
const testDomain = "gpu.example.com"

// A workload with no hard constraints (inference, no memory floor, no required
// precisions) must select every device — the "true" expression.
func TestWorkloadSelectorNoConstraints(t *testing.T) {
	w := allocator.Workload{Name: "loose-infer", Kind: allocator.KindInfer}
	got := WorkloadSelector(w, testDomain)
	if got.Expression != "true" {
		t.Errorf("Expression = %q, want \"true\"", got.Expression)
	}
	if got.Count != 1 {
		t.Errorf("Count = %d, want 1", got.Count)
	}
	if got.MatchAttribute != "" {
		t.Errorf("MatchAttribute = %q, want empty", got.MatchAttribute)
	}
}

// A training workload mirrors Feasible's trainability check: require the
// (always-published) trainable bool to be true.
func TestWorkloadSelectorTrainAddsTrainable(t *testing.T) {
	got := WorkloadSelector(allocator.Workload{Name: "t", Kind: allocator.KindTrain}, testDomain)
	want := "device.attributes['gpu.example.com'].trainable"
	if got.Expression != want {
		t.Errorf("Expression = %q, want %q", got.Expression, want)
	}
}

// A memory floor mirrors Feasible's MemoryGB >= MinMemoryGB via a Quantity
// compareTo against the published capacity.
func TestWorkloadSelectorMemoryFloor(t *testing.T) {
	got := WorkloadSelector(allocator.Workload{Name: "m", Kind: allocator.KindInfer, MinMemoryGB: 80}, testDomain)
	want := "device.capacity['gpu.example.com'].memory.compareTo(quantity('80Gi')) >= 0"
	if got.Expression != want {
		t.Errorf("Expression = %q, want %q", got.Expression, want)
	}
}

// Each required precision becomes a membership test, not field access — Describe
// publishes precision_<p> only when supported, and the real DRA CEL compiler
// errors on direct access to an absent inner key.
func TestWorkloadSelectorPrecisionMembership(t *testing.T) {
	w := allocator.Workload{Name: "p", Kind: allocator.KindInfer, RequiredPrecisions: []allocator.Precision{allocator.PrecisionINT8}}
	got := WorkloadSelector(w, testDomain)
	want := "'precision_int8' in device.attributes['gpu.example.com']"
	if got.Expression != want {
		t.Errorf("Expression = %q, want %q", got.Expression, want)
	}
}

// All clauses AND together in a fixed order (trainable, memory, then precisions
// sorted) so the emitted CEL is deterministic.
func TestWorkloadSelectorANDsClausesDeterministically(t *testing.T) {
	w := allocator.Workload{
		Name: "train-big", Kind: allocator.KindTrain, MinMemoryGB: 141,
		RequiredPrecisions: []allocator.Precision{allocator.PrecisionFP8, allocator.PrecisionBF16},
	}
	got := WorkloadSelector(w, testDomain)
	want := "device.attributes['gpu.example.com'].trainable" +
		" && device.capacity['gpu.example.com'].memory.compareTo(quantity('141Gi')) >= 0" +
		" && 'precision_bf16' in device.attributes['gpu.example.com']" +
		" && 'precision_fp8' in device.attributes['gpu.example.com']"
	if got.Expression != want {
		t.Errorf("Expression =\n  %q\nwant\n  %q", got.Expression, want)
	}
}

// Count is max(1, DeviceCount), mirroring SiftScheduler.Place.
func TestWorkloadSelectorCount(t *testing.T) {
	if got := WorkloadSelector(allocator.Workload{DeviceCount: 4}, testDomain); got.Count != 4 {
		t.Errorf("Count = %d, want 4", got.Count)
	}
	if got := WorkloadSelector(allocator.Workload{DeviceCount: 0}, testDomain); got.Count != 1 {
		t.Errorf("Count = %d, want 1 for DeviceCount<=0", got.Count)
	}
}

// A same-island multi-device gang sets the cross-device matchAttribute on island
// (the GA DRA mechanism for co-location), qualified by domain.
func TestWorkloadSelectorSameIslandGangSetsMatchAttribute(t *testing.T) {
	w := allocator.Workload{Name: "gang", Kind: allocator.KindTrain, DeviceCount: 4, SameIsland: true}
	got := WorkloadSelector(w, testDomain)
	if got.MatchAttribute != "gpu.example.com/island" {
		t.Errorf("MatchAttribute = %q, want gpu.example.com/island", got.MatchAttribute)
	}
}

// SameIsland with a single device is vacuous — there is nothing to co-locate, so
// no matchAttribute is emitted.
func TestWorkloadSelectorSameIslandSingleDeviceIsVacuous(t *testing.T) {
	w := allocator.Workload{Name: "solo", DeviceCount: 1, SameIsland: true}
	got := WorkloadSelector(w, testDomain)
	if got.MatchAttribute != "" {
		t.Errorf("MatchAttribute = %q, want empty for single-device SameIsland", got.MatchAttribute)
	}
}
