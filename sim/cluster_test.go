package sim

import (
	"reflect"
	"testing"

	"github.com/FabianSalge/sift/allocator"
)

// One non-trainable trap first, then two trainable GPUs in one island.
func clusterFixture() []allocator.Device {
	return []allocator.Device{
		{ID: "inf2-0", Node: 0, IslandID: allocator.NoIsland, Category: allocator.CategoryInferASIC, MemoryGB: 32, CostPerHr: 0.75, Precisions: []allocator.Precision{allocator.PrecisionFP16, allocator.PrecisionINT8}},
		{ID: "h100-0", Node: 1, IslandID: 1, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16}},
		{ID: "h100-1", Node: 1, IslandID: 1, Category: allocator.CategoryGPU, Trainable: true, MemoryGB: 80, CostPerHr: 2.5, Precisions: []allocator.Precision{allocator.PrecisionBF16, allocator.PrecisionFP16}},
	}
}

func trainW(name string) allocator.Workload {
	return allocator.Workload{Name: name, Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, CostWeight: 0.5}
}

func TestClusterSubmitThenAdvancePlaces(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	id := c.Submit(trainW("train-1"), 10)
	events := c.Advance(0)

	j, ok := c.Job(id)
	if !ok || j.PlacedAt != 0 || j.End != 10 || len(j.DeviceIDs) != 1 {
		t.Fatalf("job = %+v, want placed at 0 until 10 on one device", j)
	}
	if !j.Useful || !j.Feasible {
		t.Errorf("sift placement must be useful; got %+v", j)
	}
	if len(events) != 1 || events[0].Kind != "placed" || events[0].JobID != id {
		t.Errorf("events = %+v, want one placed event for job %d", events, id)
	}
	if len(c.RunningJobs()) != 1 || len(c.QueuedJobs()) != 0 {
		t.Errorf("running=%d queued=%d, want 1/0", len(c.RunningJobs()), len(c.QueuedJobs()))
	}
}

// Two trainables, three train jobs: the third waits, places when one frees.
func TestClusterCompletionFreesAndBackfills(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	c.Submit(trainW("a"), 5)
	c.Submit(trainW("b"), 8)
	third := c.Submit(trainW("c"), 5)
	c.Advance(0)
	if j, _ := c.Job(third); j.PlacedAt != -1 {
		t.Fatalf("third job placed at %v, want queued (-1)", j.PlacedAt)
	}

	events := c.Advance(6) // a ends at 5 → c places at 5
	j, _ := c.Job(third)
	if j.PlacedAt != 5 || j.End != 10 {
		t.Fatalf("third job = %+v, want placed at 5 until 10", j)
	}
	var kinds []string
	for _, e := range events {
		kinds = append(kinds, e.Kind)
	}
	if !reflect.DeepEqual(kinds, []string{"completed", "placed"}) {
		t.Errorf("event kinds = %v, want [completed placed]", kinds)
	}
	if c.UsefulDone() != 1 {
		t.Errorf("usefulDone = %d, want 1", c.UsefulDone())
	}
}

func TestClusterAdvanceBackwardIsNoop(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	c.Advance(10)
	c.Advance(3)
	if c.Clock() != 10 {
		t.Errorf("clock = %v, want 10 (monotonic)", c.Clock())
	}
}

func TestClusterCostAccrued(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	c.Submit(trainW("a"), 4) // one h100 @ 2.5
	c.Advance(0)             // place at t=0
	c.Advance(2)
	if got := c.CostAccrued(); got != 5.0 { // 2.5 × 2 running
		t.Errorf("cost at t=2 = %v, want 5.0", got)
	}
	c.Advance(10)
	if got := c.CostAccrued(); got != 10.0 { // 2.5 × 4 completed
		t.Errorf("cost after completion = %v, want 10.0", got)
	}
}

// Identical call sequences yield identical job histories.
func TestClusterDeterministic(t *testing.T) {
	build := func() *Cluster {
		c := NewCluster(clusterFixture(), allocator.AllocateSift)
		c.Submit(trainW("a"), 5)
		c.Advance(1)
		c.Submit(trainW("b"), 3)
		c.Advance(9)
		return c
	}
	a, b := build(), build()
	ja, _ := a.Job(1)
	jb, _ := b.Job(1)
	if !reflect.DeepEqual(ja, jb) {
		t.Errorf("diverged: %+v vs %+v", ja, jb)
	}
}

// Legacy in the same harness wastes: first-fit lands train on the infer ASIC.
func TestClusterLegacyWastes(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateLegacy)
	c.Submit(trainW("a"), 5)
	c.Advance(0)
	j, _ := c.Job(0)
	if j.Useful || j.DeviceIDs[0] != "inf2-0" {
		t.Fatalf("legacy job = %+v, want wasted hold on inf2-0", j)
	}
	c.Advance(6)
	if c.WastedDone() != 1 || c.UsefulDone() != 0 {
		t.Errorf("wasted=%d useful=%d, want 1/0", c.WastedDone(), c.UsefulDone())
	}
}

// The placement snapshot excludes the job's own devices but includes prior holds.
func TestClusterAllocSnapshot(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	c.Submit(trainW("a"), 10)
	second := c.Submit(trainW("b"), 10)
	c.Advance(0)
	j, _ := c.Job(second)
	if len(j.AllocSnapshot) != 1 {
		t.Fatalf("snapshot = %v, want exactly the first job's device", j.AllocSnapshot)
	}
	for _, own := range j.DeviceIDs {
		if j.AllocSnapshot[own] {
			t.Errorf("snapshot contains the job's own device %s", own)
		}
	}
}

func TestClusterAddDevicesUnblocksQueue(t *testing.T) {
	// Fleet with only the infer ASIC: a train job can never place.
	c := NewCluster(clusterFixture()[:1], allocator.AllocateSift)
	id := c.Submit(trainW("a"), 5)
	c.Advance(1)
	if j, _ := c.Job(id); j.PlacedAt != -1 {
		t.Fatalf("job placed with no capable device: %+v", j)
	}
	c.AddDevices(clusterFixture()[1:2]) // h100-0 arrives
	events := c.Advance(2)
	j, _ := c.Job(id)
	if j.PlacedAt != 2 || j.DeviceIDs[0] != "h100-0" {
		t.Fatalf("job = %+v, want placed at 2 on h100-0", j)
	}
	if len(events) != 1 || events[0].Kind != "placed" {
		t.Errorf("events = %+v, want one placed", events)
	}
}

func TestClusterDrainIdleNodeRemovesImmediately(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	events := c.DrainNode(0) // inf2-0, idle
	if len(events) != 1 || events[0].Kind != "node-removed" || events[0].Node != 0 {
		t.Fatalf("events = %+v, want [node-removed node 0]", events)
	}
	if len(c.Fleet()) != 2 {
		t.Errorf("fleet size = %d, want 2", len(c.Fleet()))
	}
}

func TestClusterDrainBusyNodeWaitsForCompletion(t *testing.T) {
	c := NewCluster(clusterFixture(), allocator.AllocateSift)
	c.Submit(trainW("a"), 5) // lands on an h100 (node 1)
	c.Advance(0)
	if events := c.DrainNode(1); len(events) != 0 {
		t.Fatalf("busy node removed early: %+v", events)
	}
	// Draining node accepts no new work: a second train job must queue.
	id := c.Submit(trainW("b"), 5)
	c.Advance(1)
	if j, _ := c.Job(id); j.PlacedAt != -1 {
		t.Fatalf("job placed on a draining node: %+v", j)
	}
	events := c.Advance(6) // running job ends at 5 → node 1 fully leaves
	removed := false
	for _, e := range events {
		if e.Kind == "node-removed" && e.Node == 1 {
			removed = true
		}
	}
	if !removed {
		t.Errorf("no node-removed for node 1 in %+v", events)
	}
	if len(c.Fleet()) != 1 || c.Fleet()[0].ID != "inf2-0" {
		t.Errorf("fleet = %+v, want only inf2-0", c.Fleet())
	}
}
