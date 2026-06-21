# Modeling the Fleet vs. Real Hardware Discovery

## What it is
How a DRA driver learns what devices exist — and why Sift *models* the fleet
instead of detecting it.

A real DRA resource driver runs as a DaemonSet: one pod per node. On each node,
its kubelet plugin **inspects the hardware physically present in that box** — for
example, NVIDIA's GPU driver queries NVML to enumerate the real GPUs — and
publishes a `ResourceSlice` describing them. The function that does this is the
driver's device-enumeration entry point (`EnumerateDevices()` in
dra-example-driver).

Two things are worth pinning down, because they're easy to get backwards:

- **Devices flow from the driver to the cluster, not the other way.** Kubernetes
  knows *nothing* about accelerators until a driver advertises them via
  ResourceSlices. The driver is the source of truth.
- **A driver reads the node's *hardware*, not "the cluster."** There is no
  cluster-side inventory of accelerators to "infer" from — the driver creates that
  inventory by looking at local hardware.

## Why it matters for Sift
Sift is **$0-hardware** by design. The kind "nodes" are Docker containers with no
accelerators inside, so there is literally nothing real to detect. Sift therefore
**models** the fleet: a scenario file (`scenarios/realistic-2026.yaml`) stands in
for "what hardware exists, and where."

This is not a shortcut hiding real detection. The upstream example driver is
*already* a mock — with no hardware to find, it fabricates 8 identical fake GPUs.
Sift simply makes that mock *representative*: a heterogeneous fleet of real-2026
accelerators (NVIDIA H100/B200, AMD MI300X, AWS Inferentia2) instead of clones.

## How Sift uses it
`EnumerateDevices()` is the **seam** where a real driver would query hardware.
Sift plugs its model in at exactly that point:

```
scenarios/realistic-2026.yaml → config.LoadFleet → dra.Describe → EnumerateDevices → ResourceSlice
```

The crucial property: **the model and real hardware share one interface.** Hand
Sift to a machine with real accelerators and you would swap that single function
for hardware-detection code — and *everything downstream is identical*: the same
attributes get published, the same CEL selects them, the same kube-scheduler binds
them. Nothing else changes. The model is a faithful stand-in precisely because it
lives behind the function a real driver uses to read hardware.

Per-node realism is staged:
- **Single-node (3a):** the whole fleet is advertised by one node — a
  simplification to prove the publishing pipeline end to end.
- **Mirrored (3b):** N kind workers, each driver publishing only the devices its
  node "has" (selected from the scenario by a node label). This matches the real
  shape — each node enumerates its own hardware — even though *which* devices a
  node has still comes from the model, not silicon.

## What stays real
Only the device *source* is modeled. The rest is genuine and runs unmodified:
the ResourceSlices are real Kubernetes objects, the CEL selectors are evaluated by
the real **kube-scheduler**, and allocation/binding follow the real DRA
structured-parameters path. Sift's honest framing: measure what can be measured,
model the rest, and say which is which — here, the hardware is modeled and the DRA
mechanism is real.

## Gotcha / what confused me
- **"Shouldn't the driver infer devices from the cluster?"** A real driver infers
  them from the **node's hardware**, not from the cluster — and there is no
  hardware here, so we model. Direction is driver → cluster.
- **The mock isn't a cheat.** It sits behind the exact function (`EnumerateDevices`)
  a production driver uses to read real hardware, so swapping in real detection
  changes nothing downstream.

## See also
- `docs/concepts/dra.md`, `docs/concepts/resourceslice-resourceclaim.md`,
  `docs/concepts/structured-parameteres.md`
- `dra/` — the pure `Device → attributes` translation.
- `driver/internal/profiles/gpu/sift.go` — the `EnumerateDevices` customization.
