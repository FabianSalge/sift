package engine

import (
	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/report"
	"github.com/FabianSalge/sift/sim"
)

// DeviceDTO is the camelCase JSON shape of an allocator.Device. Island = -1 for
// a standalone device (allocator.NoIsland).
type DeviceDTO struct {
	ID           string   `json:"id"`
	Node         int      `json:"node"`
	Island       int      `json:"island"`
	Vendor       string   `json:"vendor"`
	Category     string   `json:"category"`
	MemoryGB     float64  `json:"memoryGB"`
	Precisions   []string `json:"precisions"`
	Interconnect string   `json:"interconnect"`
	CostPerHr    float64  `json:"costPerHr"`
	Trainable    bool     `json:"trainable"`
}

func deviceToDTO(d allocator.Device) DeviceDTO {
	precs := make([]string, len(d.Precisions))
	for i, p := range d.Precisions {
		precs[i] = string(p)
	}
	return DeviceDTO{
		ID: d.ID, Node: d.Node, Island: d.IslandID,
		Vendor: string(d.Vendor), Category: string(d.Category), MemoryGB: d.MemoryGB,
		Precisions: precs, Interconnect: string(d.Interconnect), CostPerHr: d.CostPerHr, Trainable: d.Trainable,
	}
}

func (dto DeviceDTO) toDevice() allocator.Device {
	precs := make([]allocator.Precision, len(dto.Precisions))
	for i, p := range dto.Precisions {
		precs[i] = allocator.Precision(p)
	}
	return allocator.Device{
		ID: dto.ID, Node: dto.Node, IslandID: dto.Island,
		Vendor: allocator.Vendor(dto.Vendor), Category: allocator.DeviceCategory(dto.Category), MemoryGB: dto.MemoryGB,
		Precisions: precs, Interconnect: allocator.Interconnect(dto.Interconnect), CostPerHr: dto.CostPerHr, Trainable: dto.Trainable,
	}
}

// WorkloadDTO is the camelCase JSON shape of an allocator.Workload.
type WorkloadDTO struct {
	Name               string   `json:"name"`
	Kind               string   `json:"kind"`
	MinMemoryGB        float64  `json:"minMemoryGB"`
	RequiredPrecisions []string `json:"requiredPrecisions"`
	DeviceCount        int      `json:"deviceCount"`
	SameIsland         bool     `json:"sameIsland"`
	Gang               bool     `json:"gang"`
	LatencySensitive   bool     `json:"latencySensitive"`
	CostWeight         float64  `json:"costWeight"`
}

func (dto WorkloadDTO) toWorkload() allocator.Workload {
	precs := make([]allocator.Precision, len(dto.RequiredPrecisions))
	for i, p := range dto.RequiredPrecisions {
		precs[i] = allocator.Precision(p)
	}
	return allocator.Workload{
		Name: dto.Name, Kind: allocator.WorkloadKind(dto.Kind), MinMemoryGB: dto.MinMemoryGB,
		RequiredPrecisions: precs, DeviceCount: dto.DeviceCount, SameIsland: dto.SameIsland,
		Gang: dto.Gang, LatencySensitive: dto.LatencySensitive, CostWeight: dto.CostWeight,
	}
}

func workloadToDTO(w allocator.Workload) WorkloadDTO {
	precs := make([]string, len(w.RequiredPrecisions))
	for i, p := range w.RequiredPrecisions {
		precs[i] = string(p)
	}
	return WorkloadDTO{
		Name: w.Name, Kind: string(w.Kind), MinMemoryGB: w.MinMemoryGB,
		RequiredPrecisions: precs, DeviceCount: w.DeviceCount, SameIsland: w.SameIsland,
		Gang: w.Gang, LatencySensitive: w.LatencySensitive, CostWeight: w.CostWeight,
	}
}

// ---- report DTOs ----

type OutcomeDTO struct {
	Workload     string   `json:"workload"`
	DeviceIDs    []string `json:"deviceIDs"`
	CostPerHr    float64  `json:"costPerHr"`
	Feasible     bool     `json:"feasible"`
	SameIslandOK bool     `json:"sameIslandOK"`
	Pending      bool     `json:"pending"`
}

type SummaryDTO struct {
	Name        string       `json:"name"`
	TotalCost   float64      `json:"totalCost"`
	TypeCorrect int          `json:"typeCorrect"`
	GangsWhole  int          `json:"gangsWhole"`
	Fragmented  int          `json:"fragmented"`
	Pending     int          `json:"pending"`
	Outcomes    []OutcomeDTO `json:"outcomes"`
}

type ReportDTO struct {
	Fleet     int        `json:"fleet"`
	Workloads int        `json:"workloads"`
	Sift      SummaryDTO `json:"sift"`
	Legacy    SummaryDTO `json:"legacy"`
}

func reportToDTO(r report.Report) ReportDTO {
	return ReportDTO{Fleet: r.Fleet, Workloads: r.Workloads, Sift: summaryToDTO(r.Sift), Legacy: summaryToDTO(r.Legacy)}
}

func summaryToDTO(s report.Summary) SummaryDTO {
	outs := make([]OutcomeDTO, len(s.Outcomes))
	for i, o := range s.Outcomes {
		outs[i] = OutcomeDTO{Workload: o.Workload, DeviceIDs: o.DeviceIDs, CostPerHr: o.CostPerHr, Feasible: o.Feasible, SameIslandOK: o.SameIslandOK, Pending: o.Pending}
	}
	return SummaryDTO{Name: s.Name, TotalCost: s.TotalCost, TypeCorrect: s.TypeCorrect, GangsWhole: s.GangsWhole, Fragmented: s.Fragmented, Pending: s.Pending, Outcomes: outs}
}

// ---- trace DTOs ----

type ReasonDTO struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type ScoreDTO struct {
	CostComponent float64 `json:"costComponent"`
	MemoryWaste   float64 `json:"memoryWaste"`
}

type VerdictDTO struct {
	DeviceID  string      `json:"deviceID"`
	Feasible  bool        `json:"feasible"`
	Allocated bool        `json:"allocated"`
	Reasons   []ReasonDTO `json:"reasons"`
	Score     ScoreDTO    `json:"score"`
	Rank      int         `json:"rank"`
}

type TraceDTO struct {
	Workload string       `json:"workload"`
	Verdicts []VerdictDTO `json:"verdicts"`
	Bound    []string     `json:"bound"`
	Island   int          `json:"island"`
	Err      string       `json:"err"`
}

func traceToDTO(t allocator.Trace) TraceDTO {
	verdicts := make([]VerdictDTO, len(t.Verdicts))
	for i, v := range t.Verdicts {
		reasons := make([]ReasonDTO, len(v.Reasons))
		for j, r := range v.Reasons {
			reasons[j] = ReasonDTO{Code: string(r.Code), Detail: r.Detail}
		}
		verdicts[i] = VerdictDTO{
			DeviceID: v.DeviceID, Feasible: v.Feasible, Allocated: v.Allocated,
			Reasons: reasons, Score: ScoreDTO{CostComponent: v.Score.CostComponent, MemoryWaste: v.Score.MemoryWaste}, Rank: v.Rank,
		}
	}
	return TraceDTO{Workload: t.Workload, Verdicts: verdicts, Bound: t.Bound, Island: t.IslandID, Err: t.Err}
}

// ---- stream simulation DTOs ----

type ArrivalDTO struct {
	At       float64     `json:"at"`
	Workload WorkloadDTO `json:"workload"`
	Duration float64     `json:"duration"`
}

type ArrivalResultDTO struct {
	Index        int      `json:"index"`
	Workload     string   `json:"workload"`
	ArrivedAt    float64  `json:"arrivedAt"`
	PlacedAt     float64  `json:"placedAt"`
	End          float64  `json:"end"`
	DeviceIDs    []string `json:"deviceIDs"`
	Feasible     bool     `json:"feasible"`
	SameIslandOK bool     `json:"sameIslandOK"`
	Useful       bool     `json:"useful"`
	CostPerHr    float64  `json:"costPerHr"`
}

type SchedulerResultDTO struct {
	Name     string             `json:"name"`
	Arrivals []ArrivalResultDTO `json:"arrivals"`
}

type ResultDTO struct {
	Fleet   int                `json:"fleet"`
	Stream  int                `json:"stream"`
	Horizon float64            `json:"horizon"`
	Sift    SchedulerResultDTO `json:"sift"`
	Legacy  SchedulerResultDTO `json:"legacy"`
}

func resultToDTO(r sim.Result) ResultDTO {
	return ResultDTO{Fleet: r.Fleet, Stream: r.Stream, Horizon: r.Horizon, Sift: schedResultToDTO(r.Sift), Legacy: schedResultToDTO(r.Legacy)}
}

func schedResultToDTO(s sim.SchedulerResult) SchedulerResultDTO {
	arr := make([]ArrivalResultDTO, len(s.Arrivals))
	for i, a := range s.Arrivals {
		arr[i] = ArrivalResultDTO{
			Index: a.Index, Workload: a.Workload, ArrivedAt: a.ArrivedAt, PlacedAt: a.PlacedAt, End: a.End,
			DeviceIDs: a.DeviceIDs, Feasible: a.Feasible, SameIslandOK: a.SameIslandOK, Useful: a.Useful, CostPerHr: a.CostPerHr,
		}
	}
	return SchedulerResultDTO{Name: s.Name, Arrivals: arr}
}

// ---- live cluster session DTOs ----

type SubmitDTO struct {
	Workload WorkloadDTO `json:"workload"`
	Duration float64     `json:"duration"`
}

type ClusterDeviceDTO struct {
	DeviceDTO
	JobID    int  `json:"jobID"` // -1 when idle
	Draining bool `json:"draining"`
}

type ClusterJobDTO struct {
	ID        int         `json:"id"`
	Workload  WorkloadDTO `json:"workload"`
	Duration  float64     `json:"duration"`
	ArrivedAt float64     `json:"arrivedAt"`
	PlacedAt  float64     `json:"placedAt"`
	End       float64     `json:"end"`
	DeviceIDs []string    `json:"deviceIDs"`
	Useful    bool        `json:"useful"`
	CostPerHr float64     `json:"costPerHr"`
}

type EventDTO struct {
	Kind      string   `json:"kind"`
	At        float64  `json:"at"`
	JobID     int      `json:"jobID"`
	Node      int      `json:"node"`
	DeviceIDs []string `json:"deviceIDs"`
}

// ShadowDTO is the legacy cluster reduced to metrics — it is never rendered.
type ShadowDTO struct {
	Busy       int     `json:"busy"`
	Wasted     int     `json:"wasted"`
	Queue      int     `json:"queue"`
	UsefulDone int     `json:"usefulDone"`
	Cost       float64 `json:"cost"`
}

type ClusterSnapshotDTO struct {
	Clock      float64            `json:"clock"`
	Devices    []ClusterDeviceDTO `json:"devices"`
	Queue      []ClusterJobDTO    `json:"queue"`
	Running    []ClusterJobDTO    `json:"running"`
	UsefulDone int                `json:"usefulDone"`
	Cost       float64            `json:"cost"`
	Events     []EventDTO         `json:"events"`
	Shadow     ShadowDTO          `json:"shadow"`
}

func jobToDTO(j sim.Job) ClusterJobDTO {
	return ClusterJobDTO{
		ID: j.ID, Workload: workloadToDTO(j.Workload), Duration: j.Duration,
		ArrivedAt: j.ArrivedAt, PlacedAt: j.PlacedAt, End: j.End,
		DeviceIDs: j.DeviceIDs, Useful: j.Useful, CostPerHr: j.CostPerHr,
	}
}

func eventToDTO(e sim.Event) EventDTO {
	return EventDTO{Kind: e.Kind, At: e.At, JobID: e.JobID, Node: e.Node, DeviceIDs: e.DeviceIDs}
}
