package allocator

type Vendor string

const (
	VendorNVIDIA Vendor = "nvidia"
	VendorAMD    Vendor = "amd"
	VendorGoogle Vendor = "google"
	VendorAWS    Vendor = "aws"
)

type DeviceCategory string

const (
	CategoryGPU       DeviceCategory = "gpu"
	CategoryTrainASIC DeviceCategory = "train-asic"
	CategoryInferASIC DeviceCategory = "infer-asic"
)

type Interconnect string

const (
	InterconnectNVLink         Interconnect = "nvlink"
	InterconnectInfinityFabric Interconnect = "infinity-fabric"
	InterconnectICI            Interconnect = "ici"
	InterconnectNeuronLink     Interconnect = "neuronlink"
	InterconnectNone           Interconnect = "none"
)

type Precision string

const (
	PrecisionBF16 Precision = "bf16"
	PrecisionFP16 Precision = "fp16"
	PrecisionFP8  Precision = "fp8"
	PrecisionFP4  Precision = "fp4"
	PrecisionINT8 Precision = "int8"
)

// NoIsland marks a device with no fast-interconnect group.
const NoIsland = -1

// ---- the two domain types ----

// Device is a static capability advertisement — the allocator's analogue of a
// DRA ResourceSlice. It carries NO allocation/free state by design.
type Device struct {
	ID           string
	Node         int
	IslandID     int // interconnect group; NoIsland if standalone
	Vendor       Vendor
	Category     DeviceCategory
	MemoryGB     float64
	Precisions   []Precision // the SET this device supports
	Interconnect Interconnect
	CostPerHr    float64
	Trainable    bool
}

// WorkloadKind is what the job does; distinct from DeviceCategory.
type WorkloadKind string

const (
	KindTrain WorkloadKind = "train"
	KindInfer WorkloadKind = "infer"
)

// Workload is a request — the allocator's analogue of a DRA ResourceClaim.
type Workload struct {
	Name               string
	Kind               WorkloadKind
	MinMemoryGB        float64
	RequiredPrecisions []Precision // device.Precisions must be a superset
	DeviceCount        int
	SameIsland         bool
	Gang               bool
	LatencySensitive   bool
	CostWeight         float64 // 0..1, how strongly to prefer cheaper devices
}
