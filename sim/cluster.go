package sim

import "github.com/FabianSalge/sift/allocator"

// Job is one submitted workload's lifecycle in a live Cluster.
type Job struct {
	ID           int
	Workload     allocator.Workload
	Duration     float64
	ArrivedAt    float64
	PlacedAt     float64 // -1 while queued
	End          float64 // -1 while queued
	DeviceIDs    []string
	Feasible     bool
	SameIslandOK bool
	Useful       bool
	CostPerHr    float64
	// AllocSnapshot is the busy/draining set the placement decision saw
	// (excluding this job's own devices) — replayable through Explain.
	AllocSnapshot map[string]bool
}

// Event is a state change emitted by Advance/DrainNode for UI animation. All
// cluster state is derivable from snapshots; events are additive.
type Event struct {
	Kind      string // "placed" | "completed" | "node-removed"
	At        float64
	JobID     int // -1 for node-removed
	Node      int // -1 unless node-removed
	DeviceIDs []string
}

// Cluster is an advanceable live cluster: submit workloads, advance time,
// grow and drain the fleet. Deterministic — a function of the call sequence;
// every placement comes from the injected PlaceFunc.
type Cluster struct {
	fleet         []allocator.Device
	byID          map[string]allocator.Device
	place         PlaceFunc
	allocated     map[string]bool
	draining      map[string]bool
	jobs          []*Job // indexed by ID
	queue         []int  // job IDs waiting, submit order
	running       []int  // job IDs running, placement order
	clock         float64
	usefulDone    int
	wastedDone    int
	completedCost float64 // ∑ cost×duration of completed jobs
}

func NewCluster(fleet []allocator.Device, place PlaceFunc) *Cluster {
	f := append([]allocator.Device(nil), fleet...)
	return &Cluster{
		fleet: f, byID: indexByID(f), place: place,
		allocated: map[string]bool{}, draining: map[string]bool{},
	}
}

func (c *Cluster) Clock() float64            { return c.clock }
func (c *Cluster) Fleet() []allocator.Device { return c.fleet }
func (c *Cluster) UsefulDone() int           { return c.usefulDone }
func (c *Cluster) WastedDone() int           { return c.wastedDone }
func (c *Cluster) Draining(id string) bool   { return c.draining[id] }
func (c *Cluster) Idle() bool                { return len(c.queue) == 0 && len(c.running) == 0 }

func (c *Cluster) Job(id int) (Job, bool) {
	if id < 0 || id >= len(c.jobs) {
		return Job{}, false
	}
	return *c.jobs[id], true
}

func (c *Cluster) QueuedJobs() []Job {
	out := make([]Job, len(c.queue))
	for i, id := range c.queue {
		out[i] = *c.jobs[id]
	}
	return out
}

func (c *Cluster) RunningJobs() []Job {
	out := make([]Job, len(c.running))
	for i, id := range c.running {
		out[i] = *c.jobs[id]
	}
	return out
}

// CostAccrued is completed cost plus the running jobs' cost so far.
func (c *Cluster) CostAccrued() float64 {
	cost := c.completedCost
	for _, id := range c.running {
		j := c.jobs[id]
		cost += j.CostPerHr * (c.clock - j.PlacedAt)
	}
	return cost
}

// Taken is the set a placement decision must avoid: busy plus draining.
func (c *Cluster) Taken() map[string]bool {
	m := make(map[string]bool, len(c.allocated)+len(c.draining))
	for id := range c.allocated {
		m[id] = true
	}
	for id := range c.draining {
		m[id] = true
	}
	return m
}

// Submit enqueues w at the current clock; it is placed on the next Advance.
func (c *Cluster) Submit(w allocator.Workload, duration float64) int {
	id := len(c.jobs)
	c.jobs = append(c.jobs, &Job{ID: id, Workload: w, Duration: duration, ArrivedAt: c.clock, PlacedAt: -1, End: -1})
	c.queue = append(c.queue, id)
	return id
}

// Advance moves the clock to `to`, completing jobs and placing queued work at
// each intermediate completion time (and at `to` itself). Clock is monotonic.
func (c *Cluster) Advance(to float64) []Event {
	if to < c.clock {
		to = c.clock
	}
	var events []Event
	for {
		t, ok := c.nextEnd()
		if !ok || t > to {
			break
		}
		c.clock = t
		events = append(events, c.complete(t)...)
		events = append(events, c.drainQueue()...)
	}
	c.clock = to
	events = append(events, c.drainQueue()...)
	return events
}

func (c *Cluster) nextEnd() (float64, bool) {
	var t float64
	ok := false
	for _, id := range c.running {
		if e := c.jobs[id].End; !ok || e < t {
			t, ok = e, true
		}
	}
	return t, ok
}

// complete finishes every running job with End <= t, freeing its devices.
func (c *Cluster) complete(t float64) []Event {
	var events []Event
	kept := c.running[:0]
	for _, id := range c.running {
		j := c.jobs[id]
		if j.End > t {
			kept = append(kept, id)
			continue
		}
		for _, d := range j.DeviceIDs {
			delete(c.allocated, d)
		}
		if j.Useful {
			c.usefulDone++
		} else {
			c.wastedDone++
		}
		c.completedCost += j.CostPerHr * j.Duration
		events = append(events, Event{Kind: "completed", At: t, JobID: id, Node: -1, DeviceIDs: j.DeviceIDs})
		events = append(events, c.reap(t, j.DeviceIDs)...)
	}
	c.running = kept
	return events
}

// reap removes freed draining devices from the fleet, reporting nodes that
// emptied as a result. (Task 2 wires DrainNode into this.)
func (c *Cluster) reap(t float64, freed []string) []Event {
	var events []Event
	for _, id := range freed {
		if !c.draining[id] {
			continue
		}
		node := c.byID[id].Node
		c.removeDevice(id)
		if !c.nodeExists(node) {
			events = append(events, Event{Kind: "node-removed", At: t, JobID: -1, Node: node})
		}
	}
	return events
}

func (c *Cluster) removeDevice(id string) {
	for i, d := range c.fleet {
		if d.ID == id {
			c.fleet = append(c.fleet[:i], c.fleet[i+1:]...)
			break
		}
	}
	delete(c.byID, id)
	delete(c.draining, id)
}

func (c *Cluster) nodeExists(node int) bool {
	for _, d := range c.fleet {
		if d.Node == node {
			return true
		}
	}
	return false
}

// drainQueue places queued jobs in submit order (backfill: skip what won't
// fit). Draining devices are excluded by passing them as taken.
func (c *Cluster) drainQueue() []Event {
	var events []Event
	taken := c.Taken()
	var still []int
	for _, id := range c.queue {
		j := c.jobs[id]
		ids, ok := c.place(c.fleet, j.Workload, taken)
		if !ok {
			still = append(still, id)
			continue
		}
		j.AllocSnapshot = copyBools(taken)
		for _, d := range ids {
			c.allocated[d] = true
			taken[d] = true
		}
		j.PlacedAt = c.clock
		j.End = c.clock + j.Duration
		j.DeviceIDs = ids
		j.Feasible, j.SameIslandOK = evaluate(c.byID, j.Workload, ids)
		j.Useful = j.Feasible && j.SameIslandOK
		j.CostPerHr = costOf(c.byID, ids)
		c.running = append(c.running, id)
		events = append(events, Event{Kind: "placed", At: c.clock, JobID: id, Node: -1, DeviceIDs: ids})
	}
	c.queue = still
	return events
}

// AddDevices appends devices to the fleet; queued work may use them on the
// next Advance. IDs are the caller's responsibility (must be unique).
func (c *Cluster) AddDevices(devs []allocator.Device) {
	c.fleet = append(c.fleet, devs...)
	for _, d := range devs {
		c.byID[d.ID] = d
	}
}

// DrainNode marks every device on node as draining: no new work; free devices
// leave immediately, busy ones when their running job completes (kubectl
// drain semantics). Emits node-removed if the node empties at once.
func (c *Cluster) DrainNode(node int) []Event {
	var free []string
	for _, d := range c.fleet {
		if d.Node != node {
			continue
		}
		c.draining[d.ID] = true
		if !c.allocated[d.ID] {
			free = append(free, d.ID)
		}
	}
	return c.reap(c.clock, free)
}

func copyBools(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
