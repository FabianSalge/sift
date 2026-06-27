package allocator

import (
	"fmt"
	"sort"
)

// RejectReason is why a device fails a workload's hard filter (string-backed,
// ADR-0004).
type RejectReason string

const (
	ReasonNotTrainable       RejectReason = "not-trainable"
	ReasonInsufficientMemory RejectReason = "insufficient-memory"
	ReasonMissingPrecision   RejectReason = "missing-precision"
)

// Reason is one failed hard constraint, with a human-renderable detail.
type Reason struct {
	Code   RejectReason
	Detail string
}

// ScoreKey exposes the soft-preference ordering Sift uses (ADR-0019), so the UI
// can show why one feasible device outranks another.
type ScoreKey struct {
	CostComponent float64 // CostWeight * CostPerHr
	MemoryWaste   float64 // MemoryGB - MinMemoryGB
}

// DeviceVerdict is one device's place in the trace. Feasible is a pure
// capability check (== Feasible); Allocated marks a capable device already taken
// by an earlier workload, so it cannot be bound.
type DeviceVerdict struct {
	DeviceID  string
	Feasible  bool
	Allocated bool
	Reasons   []Reason // empty iff Feasible
	Score     ScoreKey // meaningful only if Feasible
	Rank      int      // 1-based among bindable devices in score order; 0 otherwise
}

// Trace is a full filter -> score -> bind explanation for one workload.
type Trace struct {
	Workload string
	Verdicts []DeviceVerdict // one per device, fleet order
	Bound    []string        // chosen device IDs (== SiftScheduler.Place)
	IslandID int             // island the bind landed in; NoIsland if n/a
	Err      string          // "" or the placement error message
}

// Explain runs filter -> score -> bind over a fleet for one workload and a
// current allocation set, returning a per-device trace. It reuses Feasible,
// lessScore and selectN, so its verdicts and bind cannot diverge from the
// schedulers (enforced by invariant tests).
func Explain(devices []Device, w Workload, allocated map[string]bool) Trace {
	n := w.DeviceCount
	if n < 1 {
		n = 1
	}

	verdicts := make([]DeviceVerdict, 0, len(devices))
	var bindable []Device
	for _, d := range devices {
		reasons := rejectReasons(d, w)
		v := DeviceVerdict{DeviceID: d.ID, Feasible: len(reasons) == 0, Allocated: allocated[d.ID], Reasons: reasons}
		if v.Feasible {
			v.Score = ScoreKey{CostComponent: w.CostWeight * d.CostPerHr, MemoryWaste: d.MemoryGB - w.MinMemoryGB}
			if !v.Allocated {
				bindable = append(bindable, d)
			}
		}
		verdicts = append(verdicts, v)
	}

	// Rank bindable devices by the same key the scheduler scores with.
	ranked := append([]Device(nil), bindable...)
	sort.SliceStable(ranked, func(i, j int) bool { return lessScore(ranked[i], ranked[j], w) })
	rank := make(map[string]int, len(ranked))
	for i, d := range ranked {
		rank[d.ID] = i + 1
	}
	for i := range verdicts {
		verdicts[i].Rank = rank[verdicts[i].DeviceID]
	}

	tr := Trace{Workload: w.Name, Verdicts: verdicts, IslandID: NoIsland}
	chosen, ok := selectN(bindable, w, n)
	if !ok {
		tr.Err = ErrNoFeasibleDevice.Error()
		return tr
	}
	ids := make([]string, len(chosen))
	for i, d := range chosen {
		ids[i] = d.ID
	}
	tr.Bound = ids
	tr.IslandID = chosen[0].IslandID
	return tr
}

// rejectReasons returns the hard constraints a device fails for w — empty iff
// Feasible(d, w) is true. It must stay in lockstep with Feasible (invariant
// test TestExplainFeasibilityMatchesFeasible).
func rejectReasons(d Device, w Workload) []Reason {
	var rs []Reason
	if w.Kind == KindTrain && !d.Trainable {
		rs = append(rs, Reason{Code: ReasonNotTrainable, Detail: "workload trains; device is not trainable"})
	}
	if d.MemoryGB < w.MinMemoryGB {
		rs = append(rs, Reason{Code: ReasonInsufficientMemory, Detail: fmt.Sprintf("%gGB < %gGB", d.MemoryGB, w.MinMemoryGB)})
	}
	for _, p := range missingPrecisions(d.Precisions, w.RequiredPrecisions) {
		rs = append(rs, Reason{Code: ReasonMissingPrecision, Detail: "missing " + string(p)})
	}
	return rs
}

// missingPrecisions returns required precisions absent from have.
func missingPrecisions(have, required []Precision) []Precision {
	set := make(map[Precision]struct{}, len(have))
	for _, p := range have {
		set[p] = struct{}{}
	}
	var miss []Precision
	for _, p := range required {
		if _, ok := set[p]; !ok {
			miss = append(miss, p)
		}
	}
	return miss
}
