// Package engine wraps the pure allocator/config/report core in camelCase JSON so
// the WASM browser build (and its native parity tests) can call it. It marshals;
// it does not decide — golden rule (ADR-0003). Every result comes from the core.
package engine

import (
	"bytes"
	"encoding/json"

	"github.com/FabianSalge/sift/allocator"
	"github.com/FabianSalge/sift/config"
	"github.com/FabianSalge/sift/report"
	"github.com/FabianSalge/sift/sim"
)

// EncodeFleet renders a fleet as camelCase DeviceDTO JSON.
func EncodeFleet(devs []allocator.Device) ([]byte, error) {
	dtos := make([]DeviceDTO, len(devs))
	for i, d := range devs {
		dtos[i] = deviceToDTO(d)
	}
	return json.Marshal(dtos)
}

// LoadScenario parses scenario YAML with the real config loader into fleet JSON.
func LoadScenario(yamlText []byte) ([]byte, error) {
	fleet, err := config.LoadFleet(bytes.NewReader(yamlText))
	if err != nil {
		return nil, err
	}
	return EncodeFleet(fleet)
}

func decodeFleet(fleetJSON []byte) ([]allocator.Device, error) {
	var dtos []DeviceDTO
	if err := json.Unmarshal(fleetJSON, &dtos); err != nil {
		return nil, err
	}
	devs := make([]allocator.Device, len(dtos))
	for i, dto := range dtos {
		devs[i] = dto.toDevice()
	}
	return devs, nil
}

// Run places the workload sequence on both schedulers and returns the contrast.
func Run(fleetJSON, workloadsJSON []byte) ([]byte, error) {
	fleet, err := decodeFleet(fleetJSON)
	if err != nil {
		return nil, err
	}
	var dtos []WorkloadDTO
	if err := json.Unmarshal(workloadsJSON, &dtos); err != nil {
		return nil, err
	}
	wls := make([]allocator.Workload, len(dtos))
	for i, dto := range dtos {
		wls[i] = dto.toWorkload()
	}
	return json.Marshal(reportToDTO(report.Run(fleet, wls)))
}

// Explain returns the filter->score->bind trace for one workload. allocatedJSON
// is a JSON object of deviceID->bool (or "null"/empty for a fresh fleet).
func Explain(fleetJSON, workloadJSON, allocatedJSON []byte) ([]byte, error) {
	fleet, err := decodeFleet(fleetJSON)
	if err != nil {
		return nil, err
	}
	var dto WorkloadDTO
	if err := json.Unmarshal(workloadJSON, &dto); err != nil {
		return nil, err
	}
	allocated := map[string]bool{}
	if len(allocatedJSON) > 0 && string(allocatedJSON) != "null" {
		if err := json.Unmarshal(allocatedJSON, &allocated); err != nil {
			return nil, err
		}
	}
	return json.Marshal(traceToDTO(allocator.Explain(fleet, dto.toWorkload(), allocated)))
}

// Simulate forward-runs an arrival stream against both schedulers.
func Simulate(fleetJSON, streamJSON []byte) ([]byte, error) {
	fleet, err := decodeFleet(fleetJSON)
	if err != nil {
		return nil, err
	}
	var dtos []ArrivalDTO
	if err := json.Unmarshal(streamJSON, &dtos); err != nil {
		return nil, err
	}
	stream := make(sim.Stream, len(dtos))
	for i, a := range dtos {
		stream[i] = sim.Arrival{At: a.At, Workload: a.Workload.toWorkload(), Duration: a.Duration}
	}
	return json.Marshal(resultToDTO(sim.Run(fleet, stream)))
}
