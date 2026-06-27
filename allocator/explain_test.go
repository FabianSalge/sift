package allocator

import "testing"

func explainFixture() []Device {
	return []Device{
		{ID: "inf2-0", IslandID: NoIsland, Category: CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []Precision{PrecisionFP16, PrecisionINT8}},
		{ID: "h100-0", IslandID: 1, Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8}},
		{ID: "h100-1", IslandID: 1, Category: CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []Precision{PrecisionBF16, PrecisionFP16, PrecisionFP8}},
		{ID: "mi300x-0", IslandID: 2, Category: CategoryGPU, Trainable: true, MemoryGB: 192, CostPerHr: 1.9, Precisions: []Precision{PrecisionBF16, PrecisionFP16}},
	}
}

func verdictByID(tr Trace, id string) (DeviceVerdict, bool) {
	for _, v := range tr.Verdicts {
		if v.DeviceID == id {
			return v, true
		}
	}
	return DeviceVerdict{}, false
}

// Invariant 1: Explain's per-device feasibility equals Feasible for every device.
func TestExplainFeasibilityMatchesFeasible(t *testing.T) {
	fleet := explainFixture()
	workloads := []Workload{
		{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5},
		{Name: "infer-int8", Kind: KindInfer, MinMemoryGB: 16, RequiredPrecisions: []Precision{PrecisionINT8}, CostWeight: 1},
		{Name: "gang", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, DeviceCount: 2, SameIsland: true, CostWeight: 1},
	}
	for _, w := range workloads {
		tr := Explain(fleet, w, nil)
		for _, d := range fleet {
			v, ok := verdictByID(tr, d.ID)
			if !ok {
				t.Fatalf("%s: no verdict for %s", w.Name, d.ID)
			}
			if v.Feasible != Feasible(d, w) {
				t.Errorf("%s/%s: verdict.Feasible=%v, Feasible()=%v", w.Name, d.ID, v.Feasible, Feasible(d, w))
			}
		}
	}
}

// Invariant 2: Explain's bound set equals SiftScheduler.Place for the same state.
func TestExplainBindMatchesPlace(t *testing.T) {
	fleet := explainFixture()
	workloads := []Workload{
		{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5},
		{Name: "gang", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, DeviceCount: 2, SameIsland: true, CostWeight: 1},
	}
	for _, w := range workloads {
		want, err := NewSiftScheduler(fleet).Place(w)
		tr := Explain(fleet, w, nil)
		if err != nil {
			if tr.Err == "" {
				t.Errorf("%s: Place errored but Trace.Err empty", w.Name)
			}
			continue
		}
		if len(tr.Bound) != len(want.DeviceIDs) {
			t.Fatalf("%s: Bound=%v, want %v", w.Name, tr.Bound, want.DeviceIDs)
		}
		for i := range want.DeviceIDs {
			if tr.Bound[i] != want.DeviceIDs[i] {
				t.Errorf("%s: Bound[%d]=%s, want %s", w.Name, i, tr.Bound[i], want.DeviceIDs[i])
			}
		}
	}
}

// A non-trainable, too-small, wrong-precision device fails for all three reasons.
func TestExplainReasons(t *testing.T) {
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	tr := Explain(explainFixture(), w, nil)
	v, _ := verdictByID(tr, "inf2-0")
	if v.Feasible {
		t.Fatalf("inf2-0 should be infeasible for %s", w.Name)
	}
	got := map[RejectReason]bool{}
	for _, r := range v.Reasons {
		got[r.Code] = true
	}
	for _, want := range []RejectReason{ReasonNotTrainable, ReasonInsufficientMemory, ReasonMissingPrecision} {
		if !got[want] {
			t.Errorf("inf2-0 reasons missing %q; got %+v", want, v.Reasons)
		}
	}
}

// The cheapest feasible device (cost-weighted) ranks first.
func TestExplainRanksByCost(t *testing.T) {
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	tr := Explain(explainFixture(), w, nil)
	// mi300x-0: 0.5*1.9=0.95 beats h100: 0.5*2.5=1.25 -> rank 1.
	if v, _ := verdictByID(tr, "mi300x-0"); v.Rank != 1 {
		t.Errorf("mi300x-0 rank=%d, want 1", v.Rank)
	}
}

// An already-allocated but capable device is marked Allocated and not bound.
func TestExplainAllocated(t *testing.T) {
	w := Workload{Name: "train-llm", Kind: KindTrain, MinMemoryGB: 80, RequiredPrecisions: []Precision{PrecisionBF16}, CostWeight: 0.5}
	tr := Explain(explainFixture(), w, map[string]bool{"mi300x-0": true})
	v, _ := verdictByID(tr, "mi300x-0")
	if !v.Feasible || !v.Allocated {
		t.Errorf("mi300x-0: Feasible=%v Allocated=%v, want true/true", v.Feasible, v.Allocated)
	}
	if len(tr.Bound) != 1 || tr.Bound[0] != "h100-0" {
		t.Errorf("Bound=%v, want [h100-0] (mi300x taken)", tr.Bound)
	}
}
