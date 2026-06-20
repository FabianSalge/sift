// Package config translates scenario YAML into the pure allocator model.
// The allocator package itself stays free of YAML/reflect/k8s deps (ADR-0008);
// all parsing and enum validation lives here.
package config

import (
	"fmt"
	"io"
	"os"

	"github.com/FabianSalge/sift/allocator"
	"gopkg.in/yaml.v3"
)

// rawFleet mirrors the on-disk YAML shape: a catalog of device classes plus an
// ordered fleet of placements (counts of a class onto a node/island).
type rawFleet struct {
	Catalog map[string]rawClass `yaml:"catalog"`
	Fleet   []rawPlacement      `yaml:"fleet"`
}

// rawClass is one catalog entry — the static capabilities of a device class.
type rawClass struct {
	Vendor       string   `yaml:"vendor"`
	Category     string   `yaml:"category"`
	MemoryGB     float64  `yaml:"memoryGB"`
	Precisions   []string `yaml:"precisions"`
	Interconnect string   `yaml:"interconnect"`
	CostPerHr    float64  `yaml:"costPerHr"`
	Trainable    bool     `yaml:"trainable"`
}

// rawPlacement places count copies of a catalog class onto a node, optionally in
// an island. Island is a pointer so an absent value means "standalone" rather
// than island 0.
type rawPlacement struct {
	Class  string `yaml:"class"`
	Count  int    `yaml:"count"`
	Node   int    `yaml:"node"`
	Island *int   `yaml:"island"`
}

// String->const maps for the model's enums. A missing key is an unknown value;
// these maps double as the validator (see resolve* helpers, added with their
// tests).
var (
	vendors = map[string]allocator.Vendor{
		"nvidia": allocator.VendorNVIDIA,
		"amd":    allocator.VendorAMD,
		"google": allocator.VendorGoogle,
		"aws":    allocator.VendorAWS,
	}
	categories = map[string]allocator.DeviceCategory{
		"gpu":        allocator.CategoryGPU,
		"train-asic": allocator.CategoryTrainASIC,
		"infer-asic": allocator.CategoryInferASIC,
	}
	interconnects = map[string]allocator.Interconnect{
		"nvlink":          allocator.InterconnectNVLink,
		"infinity-fabric": allocator.InterconnectInfinityFabric,
		"ici":             allocator.InterconnectICI,
		"neuronlink":      allocator.InterconnectNeuronLink,
		"none":            allocator.InterconnectNone,
	}
	precisions = map[string]allocator.Precision{
		"bf16": allocator.PrecisionBF16,
		"fp16": allocator.PrecisionFP16,
		"fp8":  allocator.PrecisionFP8,
		"fp4":  allocator.PrecisionFP4,
		"int8": allocator.PrecisionINT8,
	}
)

// resolvedClass is a catalog class after its enums have been validated and
// mapped onto the allocator's typed constants.
type resolvedClass struct {
	vendor       allocator.Vendor
	category     allocator.DeviceCategory
	interconnect allocator.Interconnect
	precisions   []allocator.Precision
	memoryGB     float64
	costPerHr    float64
	trainable    bool
}

// LoadFleetFile opens a scenario YAML file and loads it via LoadFleet.
func LoadFleetFile(path string) ([]allocator.Device, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadFleet(f)
}

// LoadFleet parses a scenario YAML stream into a flat slice of devices, in file
// order. Each placement's count is expanded into individual devices with
// generated, per-class IDs (e.g. h100-0, h100-1). Unknown enum values, missing
// catalog references, and out-of-range numbers are rejected with an error that
// names the offending field.
func LoadFleet(r io.Reader) ([]allocator.Device, error) {
	var raw rawFleet
	if err := yaml.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	// Resolve and validate every catalog class once up front.
	classes := make(map[string]resolvedClass, len(raw.Catalog))
	for name, c := range raw.Catalog {
		rc, err := resolveClass(name, c)
		if err != nil {
			return nil, err
		}
		classes[name] = rc
	}

	var devices []allocator.Device
	perClass := map[string]int{}
	for i, p := range raw.Fleet {
		rc, ok := classes[p.Class]
		if !ok {
			return nil, fmt.Errorf("fleet placement %d: unknown class %q (not in catalog)", i, p.Class)
		}
		if p.Count < 1 {
			return nil, fmt.Errorf("fleet placement %d (class %q): count must be >= 1, got %d", i, p.Class, p.Count)
		}
		if p.Node < 0 {
			return nil, fmt.Errorf("fleet placement %d (class %q): node must be >= 0, got %d", i, p.Class, p.Node)
		}
		island := allocator.NoIsland
		if p.Island != nil {
			if *p.Island < 0 {
				return nil, fmt.Errorf("fleet placement %d (class %q): island must be >= 0, got %d", i, p.Class, *p.Island)
			}
			island = *p.Island
		}

		for j := 0; j < p.Count; j++ {
			id := fmt.Sprintf("%s-%d", p.Class, perClass[p.Class])
			perClass[p.Class]++
			devices = append(devices, allocator.Device{
				ID:           id,
				Node:         p.Node,
				IslandID:     island,
				Vendor:       rc.vendor,
				Category:     rc.category,
				MemoryGB:     rc.memoryGB,
				Precisions:   rc.precisions,
				Interconnect: rc.interconnect,
				CostPerHr:    rc.costPerHr,
				Trainable:    rc.trainable,
			})
		}
	}
	return devices, nil
}

// resolveClass validates one catalog entry and maps its string enums onto the
// allocator's typed constants. A missing map key is an unknown enum value.
func resolveClass(name string, c rawClass) (resolvedClass, error) {
	var rc resolvedClass

	vendor, ok := vendors[c.Vendor]
	if !ok {
		return rc, fmt.Errorf("catalog class %q: unknown vendor %q", name, c.Vendor)
	}
	category, ok := categories[c.Category]
	if !ok {
		return rc, fmt.Errorf("catalog class %q: unknown category %q", name, c.Category)
	}
	interconnect, ok := interconnects[c.Interconnect]
	if !ok {
		return rc, fmt.Errorf("catalog class %q: unknown interconnect %q", name, c.Interconnect)
	}
	if len(c.Precisions) == 0 {
		return rc, fmt.Errorf("catalog class %q: precisions must not be empty", name)
	}
	precs := make([]allocator.Precision, len(c.Precisions))
	for i, pn := range c.Precisions {
		p, ok := precisions[pn]
		if !ok {
			return rc, fmt.Errorf("catalog class %q: unknown precision %q", name, pn)
		}
		precs[i] = p
	}
	if c.MemoryGB <= 0 {
		return rc, fmt.Errorf("catalog class %q: memoryGB must be > 0, got %v", name, c.MemoryGB)
	}
	if c.CostPerHr < 0 {
		return rc, fmt.Errorf("catalog class %q: costPerHr must be >= 0, got %v", name, c.CostPerHr)
	}

	return resolvedClass{
		vendor:       vendor,
		category:     category,
		interconnect: interconnect,
		precisions:   precs,
		memoryGB:     c.MemoryGB,
		costPerHr:    c.CostPerHr,
		trainable:    c.Trainable,
	}, nil
}
