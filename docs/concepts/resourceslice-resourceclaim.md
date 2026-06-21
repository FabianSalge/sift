# ResourceSlice vs ResourceClaim

## What they are
The two core DRA objects, and they sit on opposite sides of the match:
- **ResourceSlice** — the *supply* side. A driver publishes one per node listing
  the devices available there and each device's **attributes** (memory, type,
  precision, interconnect, etc.). It is a static *advertisement* of capability.
- **ResourceClaim** — the *demand* side. A workload's request for hardware,
  expressing requirements (often as CEL expressions over device attributes).

Roughly: ResourceSlice = "here's what I have and what it can do." ResourceClaim
= "here's what I need."

## Why it matters for Sift
This supply/demand split is exactly Sift's `Device` (supply) vs `Workload`
(demand) split. Understanding that a ResourceSlice is a *static* advertisement is
why Sift's `Device` carries no free/used state (allocation lives elsewhere).

## How Sift uses it
`allocator.Device` fields → attributes on a published ResourceSlice.
`allocator.Workload` requirements → CEL in a ResourceClaim. The allocator is the
clean domain model; the driver translates it to/from these real objects.


## See also
- `docs/concepts/dra.md`
- ADR-0005 (no allocation state on Device), ADR-0013