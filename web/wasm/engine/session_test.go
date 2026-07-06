package engine

import (
	"encoding/json"
	"testing"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/sim"
)

func sessionFleetJSON(t *testing.T) []byte {
	t.Helper()
	fleet := []allocator.Device{
		{ID: "inf2-0", Node: 0, IslandID: allocator.NoIsland, Category: allocator.CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []allocator.Precision{allocator.PrecisionFP16, allocator.PrecisionINT8}},
		{ID: "h100-0", Node: 1, IslandID: 1, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16}},
	}
	b, err := EncodeFleet(fleet)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

const sessionTrainJob = `{"workload":{"name":"train-1","kind":"train","minMemoryGB":80,"requiredPrecisions":["bf16"],"deviceCount":1,"sameIsland":false,"gang":false,"latencySensitive":false,"costWeight":0.5},"duration":10}`

// The session snapshot must agree with driving sim.Cluster directly.
func TestSessionSnapshotMatchesSim(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Submit([]byte(sessionTrainJob)); err != nil {
		t.Fatal(err)
	}
	out, err := s.Advance(3)
	if err != nil {
		t.Fatal(err)
	}
	var snap map[string]any
	if err := json.Unmarshal(out, &snap); err != nil {
		t.Fatal(err)
	}

	// Reference: raw clusters, same ops.
	fleet, _ := decodeFleet(sessionFleetJSON(t))
	ref := sim.NewCluster(fleet, allocator.AllocateSift)
	var dto SubmitDTO
	_ = json.Unmarshal([]byte(sessionTrainJob), &dto)
	ref.Submit(dto.Workload.toWorkload(), dto.Duration)
	ref.Advance(3)

	if snap["clock"].(float64) != ref.Clock() {
		t.Errorf("clock = %v, want %v", snap["clock"], ref.Clock())
	}
	if got := len(snap["running"].([]any)); got != len(ref.RunningJobs()) {
		t.Errorf("running = %d, want %d", got, len(ref.RunningJobs()))
	}
	if snap["cost"].(float64) != ref.CostAccrued() {
		t.Errorf("cost = %v, want %v", snap["cost"], ref.CostAccrued())
	}
}

// The shadow runs legacy on identical traffic: the train job lands on the
// infer ASIC and shows up as wasted.
func TestSessionShadowDiverges(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Submit([]byte(sessionTrainJob)); err != nil {
		t.Fatal(err)
	}
	out, _ := s.Advance(1)
	var snap struct {
		Shadow ShadowDTO `json:"shadow"`
	}
	if err := json.Unmarshal(out, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Shadow.Wasted != 1 || snap.Shadow.Busy != 1 {
		t.Errorf("shadow = %+v, want busy=1 wasted=1 (first-fit onto inf2-0)", snap.Shadow)
	}
}

// Fleet mutations mirror to both clusters.
func TestSessionMutationsMirror(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	extra, _ := EncodeFleet([]allocator.Device{{ID: "h100-9", Node: 7, IslandID: 9, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []allocator.Precision{allocator.PrecisionBF16}}})
	if _, err := s.AddNode(extra); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DrainNode(0); err != nil {
		t.Fatal(err)
	}
	out, _ := s.Advance(0)
	var snap struct {
		Devices []map[string]any `json:"devices"`
	}
	if err := json.Unmarshal(out, &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Devices) != 2 { // inf2-0 drained away, h100-0 + h100-9 remain
		t.Fatalf("devices = %d, want 2", len(snap.Devices))
	}
	if len(s.legacy.Fleet()) != 2 {
		t.Errorf("legacy fleet = %d, want mirrored 2", len(s.legacy.Fleet()))
	}
}

// Wire-format keys, pinned like TestWireFormatKeys (ADR-0027).
func TestSessionWireFormatKeys(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	s.Submit([]byte(sessionTrainJob)) // one running
	s.Submit([]byte(sessionTrainJob)) // one queued (single trainable device)
	out, _ := s.Advance(1)
	var snap map[string]json.RawMessage
	if err := json.Unmarshal(out, &snap); err != nil {
		t.Fatal(err)
	}
	wantTop := []string{"clock", "devices", "queue", "running", "usefulDone", "cost", "events", "shadow"}
	for _, k := range wantTop {
		if _, ok := snap[k]; !ok {
			t.Errorf("snapshot missing key %q", k)
		}
	}
	var devs []map[string]json.RawMessage
	json.Unmarshal(snap["devices"], &devs)
	for _, k := range []string{"id", "node", "island", "memoryGB", "costPerHr", "jobID", "draining"} {
		if _, ok := devs[0][k]; !ok {
			t.Errorf("device missing key %q", k)
		}
	}
	var jobs []map[string]json.RawMessage
	json.Unmarshal(snap["running"], &jobs)
	for _, k := range []string{"id", "workload", "duration", "arrivedAt", "placedAt", "end", "deviceIDs", "useful", "costPerHr"} {
		if _, ok := jobs[0][k]; !ok {
			t.Errorf("job missing key %q", k)
		}
	}
	var shadow map[string]json.RawMessage
	json.Unmarshal(snap["shadow"], &shadow)
	for _, k := range []string{"busy", "wasted", "queue", "usefulDone", "cost"} {
		if _, ok := shadow[k]; !ok {
			t.Errorf("shadow missing key %q", k)
		}
	}
	var events []map[string]json.RawMessage
	json.Unmarshal(snap["events"], &events)
	if len(events) == 0 {
		t.Fatal("expected at least one placed event from the first Advance")
	}
	for _, k := range []string{"kind", "at", "jobID", "node", "deviceIDs"} {
		if _, ok := events[0][k]; !ok {
			t.Errorf("event missing key %q", k)
		}
	}
}

// A placed job explains against its placement-time snapshot; the trace must
// bind exactly the devices the job actually holds.
func TestSessionExplainPlacedJob(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	s.Submit([]byte(sessionTrainJob))
	s.Advance(0)
	out, err := s.Explain(0)
	if err != nil {
		t.Fatal(err)
	}
	var tr TraceDTO
	if err := json.Unmarshal(out, &tr); err != nil {
		t.Fatal(err)
	}
	j, _ := s.sift.Job(0)
	if len(tr.Bound) != 1 || tr.Bound[0] != j.DeviceIDs[0] {
		t.Errorf("trace bound %v, want %v", tr.Bound, j.DeviceIDs)
	}
}

// A queued job explains against what is taken right now (why nothing fits).
func TestSessionExplainQueuedJob(t *testing.T) {
	s, err := NewSession(sessionFleetJSON(t))
	if err != nil {
		t.Fatal(err)
	}
	s.Submit([]byte(sessionTrainJob)) // takes the only trainable device
	s.Submit([]byte(sessionTrainJob)) // queues
	s.Advance(0)
	out, err := s.Explain(1)
	if err != nil {
		t.Fatal(err)
	}
	var tr TraceDTO
	if err := json.Unmarshal(out, &tr); err != nil {
		t.Fatal(err)
	}
	if tr.Err == "" {
		t.Errorf("queued-job trace should carry a no-fit err, got bound=%v", tr.Bound)
	}
}

func TestSessionExplainUnknownJob(t *testing.T) {
	s, _ := NewSession(sessionFleetJSON(t))
	if _, err := s.Explain(99); err == nil {
		t.Error("want error for unknown job id")
	}
}
