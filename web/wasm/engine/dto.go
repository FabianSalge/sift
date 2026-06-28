package engine

import (
	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/report"
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
