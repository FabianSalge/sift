package engine

import (
	"encoding/json"
	"fmt"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/sim"
)

// Session holds the live demo's paired clusters — Sift and the legacy shadow —
// fed identical mutations. Stateful by design (it IS the wasm boundary); all
// scheduling lives in sim/allocator.
type Session struct {
	sift   *sim.Cluster
	legacy *sim.Cluster
}

func NewSession(fleetJSON []byte) (*Session, error) {
	fleet, err := decodeFleet(fleetJSON)
	if err != nil {
		return nil, err
	}
	return &Session{
		sift:   sim.NewCluster(fleet, allocator.AllocateSift),
		legacy: sim.NewCluster(fleet, allocator.AllocateLegacy),
	}, nil
}

// Submit enqueues one job on both clusters; returns {"jobID": n} (IDs match
// across the pair because every mutation is mirrored).
func (s *Session) Submit(jobJSON []byte) ([]byte, error) {
	var dto SubmitDTO
	if err := json.Unmarshal(jobJSON, &dto); err != nil {
		return nil, err
	}
	w := dto.Workload.toWorkload()
	id := s.sift.Submit(w, dto.Duration)
	s.legacy.Submit(w, dto.Duration)
	return json.Marshal(map[string]int{"jobID": id})
}

func (s *Session) AddNode(devicesJSON []byte) ([]byte, error) {
	devs, err := decodeFleet(devicesJSON)
	if err != nil {
		return nil, err
	}
	s.sift.AddDevices(devs)
	s.legacy.AddDevices(devs)
	return []byte(`{}`), nil
}

func (s *Session) DrainNode(node int) ([]byte, error) {
	s.sift.DrainNode(node)
	s.legacy.DrainNode(node)
	return []byte(`{}`), nil
}

// Advance moves both clusters to t and returns the snapshot: full Sift state,
// Sift events since the last call, legacy reduced to shadow metrics.
func (s *Session) Advance(to float64) ([]byte, error) {
	events := s.sift.Advance(to)
	s.legacy.Advance(to)
	return json.Marshal(s.snapshot(events))
}

// Explain replays jobID's decision. A placed job traces against its stored
// placement-time allocation snapshot (over the current fleet); a queued job
// traces against what is taken right now — why nothing fits.
func (s *Session) Explain(jobID int) ([]byte, error) {
	j, ok := s.sift.Job(jobID)
	if !ok {
		return nil, fmt.Errorf("no job %d", jobID)
	}
	alloc := j.AllocSnapshot
	if j.PlacedAt < 0 {
		alloc = s.sift.Taken()
	}
	return json.Marshal(traceToDTO(allocator.Explain(s.sift.Fleet(), j.Workload, alloc)))
}

func (s *Session) snapshot(events []sim.Event) ClusterSnapshotDTO {
	c := s.sift
	jobOn := map[string]int{}
	running := c.RunningJobs()
	for _, j := range running {
		for _, id := range j.DeviceIDs {
			jobOn[id] = j.ID
		}
	}
	devs := make([]ClusterDeviceDTO, len(c.Fleet()))
	for i, d := range c.Fleet() {
		jid, busy := jobOn[d.ID]
		if !busy {
			jid = -1
		}
		devs[i] = ClusterDeviceDTO{DeviceDTO: deviceToDTO(d), JobID: jid, Draining: c.Draining(d.ID)}
	}

	queued := c.QueuedJobs()
	queue := make([]ClusterJobDTO, len(queued))
	for i, j := range queued {
		queue[i] = jobToDTO(j)
	}
	run := make([]ClusterJobDTO, len(running))
	for i, j := range running {
		run[i] = jobToDTO(j)
	}
	evts := make([]EventDTO, len(events))
	for i, e := range events {
		evts[i] = eventToDTO(e)
	}

	l := s.legacy
	var busy, wasted int
	for _, j := range l.RunningJobs() {
		busy += len(j.DeviceIDs)
		if !j.Useful {
			wasted += len(j.DeviceIDs)
		}
	}
	return ClusterSnapshotDTO{
		Clock: c.Clock(), Devices: devs, Queue: queue, Running: run,
		UsefulDone: c.UsefulDone(), Cost: c.CostAccrued(), Events: evts,
		Shadow: ShadowDTO{Busy: busy, Wasted: wasted, Queue: len(l.QueuedJobs()), UsefulDone: l.UsefulDone(), Cost: l.CostAccrued()},
	}
}
