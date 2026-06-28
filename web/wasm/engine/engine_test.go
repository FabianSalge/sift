package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/config"
	"github.com/FabianSalge/sift/report"
)

const fixtureYAML = `
catalog:
  h100:
    vendor: nvidia
    category: gpu
    memoryGB: 80
    precisions: [bf16, fp16, fp8]
    interconnect: nvlink
    costPerHr: 2.5
    trainable: true
  inferentia2:
    vendor: aws
    category: infer-asic
    memoryGB: 32
    precisions: [fp16, int8]
    interconnect: neuronlink
    costPerHr: 0.75
    trainable: false
fleet:
  - class: inferentia2
    count: 2
    node: 0
  - class: h100
    count: 4
    node: 1
    island: 1
`

func fixtureWorkloads() []allocator.Workload {
	return []allocator.Workload{
		{Name: "train-llm", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 1, CostWeight: 0.7},
		{Name: "infer-int8", Kind: allocator.KindInfer, MinMemoryGB: 16, RequiredPrecisions: []allocator.Precision{allocator.PrecisionINT8}, DeviceCount: 1, CostWeight: 1},
		{Name: "gang-train", Kind: allocator.KindTrain, MinMemoryGB: 80, RequiredPrecisions: []allocator.Precision{allocator.PrecisionBF16}, DeviceCount: 4, SameIsland: true, CostWeight: 0.7},
	}
}

// LoadScenario must reproduce exactly what config.LoadFleet parses.
func TestLoadScenarioMatchesConfig(t *testing.T) {
	fleetJSON, err := LoadScenario([]byte(fixtureYAML))
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	var got []DeviceDTO
	if err := json.Unmarshal(fleetJSON, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want, err := config.LoadFleet(bytes.NewReader([]byte(fixtureYAML)))
	if err != nil {
		t.Fatalf("config.LoadFleet: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Category != string(want[i].Category) ||
			got[i].MemoryGB != want[i].MemoryGB || got[i].Island != want[i].IslandID {
			t.Errorf("device %d = %+v, want id/category/mem/island from %+v", i, got[i], want[i])
		}
	}
}

// Run must equal report.Run for the same inputs.
func TestRunMatchesReport(t *testing.T) {
	fleet, err := config.LoadFleet(bytes.NewReader([]byte(fixtureYAML)))
	if err != nil {
		t.Fatal(err)
	}
	wls := fixtureWorkloads()
	fleetJSON, err := EncodeFleet(fleet)
	if err != nil {
		t.Fatal(err)
	}
	dtos := make([]WorkloadDTO, len(wls))
	for i, w := range wls {
		dtos[i] = workloadToDTO(w)
	}
	wlJSON, _ := json.Marshal(dtos)

	gotJSON, err := Run(fleetJSON, wlJSON)
	if err != nil {
		t.Fatal(err)
	}
	var got ReportDTO
	if err := json.Unmarshal(gotJSON, &got); err != nil {
		t.Fatal(err)
	}
	want := report.Run(fleet, wls)

	if got.Sift.TotalCost != want.Sift.TotalCost || got.Sift.TypeCorrect != want.Sift.TypeCorrect ||
		got.Legacy.Fragmented != want.Legacy.Fragmented || got.Legacy.TypeCorrect != want.Legacy.TypeCorrect {
		t.Errorf("summary mismatch: got sift{%.2f,%d} legacy{frag %d,tc %d}, want sift{%.2f,%d} legacy{frag %d,tc %d}",
			got.Sift.TotalCost, got.Sift.TypeCorrect, got.Legacy.Fragmented, got.Legacy.TypeCorrect,
			want.Sift.TotalCost, want.Sift.TypeCorrect, want.Legacy.Fragmented, want.Legacy.TypeCorrect)
	}
	for i := range want.Sift.Outcomes {
		if len(got.Sift.Outcomes[i].DeviceIDs) != len(want.Sift.Outcomes[i].DeviceIDs) {
			t.Errorf("sift outcome %d device count mismatch", i)
		}
	}
}

// Explain must equal allocator.Explain for the same inputs.
func TestExplainMatchesAllocator(t *testing.T) {
	fleet, err := config.LoadFleet(bytes.NewReader([]byte(fixtureYAML)))
	if err != nil {
		t.Fatal(err)
	}
	w := fixtureWorkloads()[2] // gang-train
	fleetJSON, _ := EncodeFleet(fleet)
	wJSON, _ := json.Marshal(workloadToDTO(w))

	gotJSON, err := Explain(fleetJSON, wJSON, []byte("null"))
	if err != nil {
		t.Fatal(err)
	}
	var got TraceDTO
	if err := json.Unmarshal(gotJSON, &got); err != nil {
		t.Fatal(err)
	}
	want := allocator.Explain(fleet, w, nil)

	if len(got.Bound) != len(want.Bound) {
		t.Fatalf("bound = %v, want %v", got.Bound, want.Bound)
	}
	for i := range want.Bound {
		if got.Bound[i] != want.Bound[i] {
			t.Errorf("bound[%d] = %s, want %s", i, got.Bound[i], want.Bound[i])
		}
	}
	if len(got.Verdicts) != len(want.Verdicts) {
		t.Fatalf("verdicts = %d, want %d", len(got.Verdicts), len(want.Verdicts))
	}
	for i := range want.Verdicts {
		if got.Verdicts[i].Feasible != want.Verdicts[i].Feasible || got.Verdicts[i].Rank != want.Verdicts[i].Rank {
			t.Errorf("verdict %s: got feasible=%v rank=%d, want feasible=%v rank=%d",
				got.Verdicts[i].DeviceID, got.Verdicts[i].Feasible, got.Verdicts[i].Rank,
				want.Verdicts[i].Feasible, want.Verdicts[i].Rank)
		}
	}
}
