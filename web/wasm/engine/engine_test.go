package engine

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/config"
	"github.com/FabianSalge/sift/report"
	"github.com/FabianSalge/sift/sim"
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
		w := want[i]
		precs := make([]string, len(w.Precisions))
		for j, p := range w.Precisions {
			precs[j] = string(p)
		}
		wantDTO := DeviceDTO{
			ID: w.ID, Node: w.Node, Island: w.IslandID, Vendor: string(w.Vendor),
			Category: string(w.Category), MemoryGB: w.MemoryGB, Precisions: precs,
			Interconnect: string(w.Interconnect), CostPerHr: w.CostPerHr, Trainable: w.Trainable,
		}
		if !reflect.DeepEqual(got[i], wantDTO) {
			t.Errorf("device %d = %+v, want %+v", i, got[i], wantDTO)
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
	wlJSON, err := json.Marshal(dtos)
	if err != nil {
		t.Fatal(err)
	}

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
	assertOutcomes := func(name string, got []OutcomeDTO, want []report.Outcome) {
		if len(got) != len(want) {
			t.Fatalf("%s: %d outcomes, want %d", name, len(got), len(want))
		}
		for i := range want {
			if len(got[i].DeviceIDs) != len(want[i].DeviceIDs) {
				t.Fatalf("%s outcome %d: %d devices, want %d", name, i, len(got[i].DeviceIDs), len(want[i].DeviceIDs))
			}
			for j := range want[i].DeviceIDs {
				if got[i].DeviceIDs[j] != want[i].DeviceIDs[j] {
					t.Errorf("%s outcome %d device %d = %s, want %s", name, i, j, got[i].DeviceIDs[j], want[i].DeviceIDs[j])
				}
			}
		}
	}
	assertOutcomes("sift", got.Sift.Outcomes, want.Sift.Outcomes)
	assertOutcomes("legacy", got.Legacy.Outcomes, want.Legacy.Outcomes)
}

// Explain must equal allocator.Explain for the same inputs.
func TestExplainMatchesAllocator(t *testing.T) {
	fleet, err := config.LoadFleet(bytes.NewReader([]byte(fixtureYAML)))
	if err != nil {
		t.Fatal(err)
	}
	w := fixtureWorkloads()[2] // gang-train
	fleetJSON, err := EncodeFleet(fleet)
	if err != nil {
		t.Fatal(err)
	}
	wJSON, err := json.Marshal(workloadToDTO(w))
	if err != nil {
		t.Fatal(err)
	}

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

// TestWireFormatKeys pins the camelCase JSON contract the browser frontend reads
// (spec §6). The parity tests round-trip through the same DTO, so a renamed tag
// would pass them silently; these assert the literal wire keys.
func TestWireFormatKeys(t *testing.T) {
	fleetJSON, err := LoadScenario([]byte(fixtureYAML))
	if err != nil {
		t.Fatal(err)
	}
	wls := fixtureWorkloads()
	dtos := make([]WorkloadDTO, len(wls))
	for i, w := range wls {
		dtos[i] = workloadToDTO(w)
	}
	wlJSON, err := json.Marshal(dtos)
	if err != nil {
		t.Fatal(err)
	}
	runJSON, err := Run(fleetJSON, wlJSON)
	if err != nil {
		t.Fatal(err)
	}
	wJSON, err := json.Marshal(workloadToDTO(wls[2]))
	if err != nil {
		t.Fatal(err)
	}
	explainJSON, err := Explain(fleetJSON, wJSON, []byte("null"))
	if err != nil {
		t.Fatal(err)
	}

	streamDTO := []ArrivalDTO{{At: 0, Workload: workloadToDTO(wls[0]), Duration: 5}}
	streamJSON2, err := json.Marshal(streamDTO)
	if err != nil {
		t.Fatal(err)
	}
	simJSON, err := Simulate(fleetJSON, streamJSON2)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		blob []byte
		keys []string
	}{
		{"LoadScenario", fleetJSON, []string{`"memoryGB"`, `"costPerHr"`, `"id"`, `"island"`, `"precisions"`, `"trainable"`}},
		{"Run", runJSON, []string{`"totalCost"`, `"typeCorrect"`, `"deviceIDs"`, `"sameIslandOK"`, `"gangsWhole"`, `"fragmented"`}},
		{"Explain", explainJSON, []string{`"deviceID"`, `"costComponent"`, `"memoryWaste"`, `"rank"`, `"bound"`, `"island"`}},
		{"Simulate", simJSON, []string{`"horizon"`, `"arrivedAt"`, `"placedAt"`, `"deviceIDs"`, `"useful"`, `"costPerHr"`}},
	}
	for _, c := range cases {
		s := string(c.blob)
		for _, k := range c.keys {
			if !strings.Contains(s, k) {
				t.Errorf("%s JSON missing wire key %s:\n%s", c.name, k, s)
			}
		}
	}
}

func TestSimulateMatchesSim(t *testing.T) {
	fleet, err := config.LoadFleet(bytes.NewReader([]byte(fixtureYAML)))
	if err != nil {
		t.Fatal(err)
	}
	stream := sim.Stream{
		{At: 0, Workload: fixtureWorkloads()[0], Duration: 5},
		{At: 1, Workload: fixtureWorkloads()[1], Duration: 3},
	}
	fleetJSON, _ := EncodeFleet(fleet)
	arrivals := make([]ArrivalDTO, len(stream))
	for i, a := range stream {
		arrivals[i] = ArrivalDTO{At: a.At, Workload: workloadToDTO(a.Workload), Duration: a.Duration}
	}
	streamJSON, err := json.Marshal(arrivals)
	if err != nil {
		t.Fatal(err)
	}

	gotJSON, err := Simulate(fleetJSON, streamJSON)
	if err != nil {
		t.Fatal(err)
	}
	var got ResultDTO
	if err := json.Unmarshal(gotJSON, &got); err != nil {
		t.Fatal(err)
	}
	want := sim.Run(fleet, stream)

	if got.Horizon != want.Horizon || len(got.Sift.Arrivals) != len(want.Sift.Arrivals) {
		t.Fatalf("summary mismatch: got horizon %v, %d arrivals; want %v, %d", got.Horizon, len(got.Sift.Arrivals), want.Horizon, len(want.Sift.Arrivals))
	}
	for i := range want.Sift.Arrivals {
		if got.Sift.Arrivals[i].PlacedAt != want.Sift.Arrivals[i].PlacedAt ||
			got.Sift.Arrivals[i].Useful != want.Sift.Arrivals[i].Useful {
			t.Errorf("sift arrival %d: got placedAt=%v useful=%v, want placedAt=%v useful=%v", i,
				got.Sift.Arrivals[i].PlacedAt, got.Sift.Arrivals[i].Useful,
				want.Sift.Arrivals[i].PlacedAt, want.Sift.Arrivals[i].Useful)
		}
	}
}
