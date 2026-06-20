# NVLink Islands, NUMA & Topology

## What it is
The interconnect hierarchy inside and across machines — slower at each step out:
- **NVLink island** — a set of GPUs directly wired together with NVLink, talking
  at very high bandwidth. Any pair *inside* an island is fast.
- **NUMA domain** — a CPU socket plus its local memory and attached devices.
  Crossing sockets is slower than staying local.
- **Across nodes** — GPUs on different servers talk over the network (RDMA /
  InfiniBand / RoCE), slower still.

So a multi-GPU job that constantly exchanges data (tensor parallelism) wants its
GPUs in **one island**; spread across islands or nodes, it slows down.

## Why it matters for Sift
This is the "right *place*" half of Sift's thesis. A same-island multi-device
workload placed across islands is the **topology failure mode** — one of the
three Sift catches and the integer model misses.

## How Sift uses it
`Device.IslandID` (with `NoIsland = -1`) models the interconnect group;
`Workload.SameIsland` requires all devices in one island. The scheduler's
same-island feasibility check is the topology filter.

## Gotcha / what confused me
_(your turn — e.g. node = one server (not the whole cluster); island vs NUMA are
related but distinct attributes)_

## See also
- `docs/concepts/capability-aware-scheduling.md`
- ADR-0014 (scenarios engineered to witness the topology failure)