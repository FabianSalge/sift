# Sift

A capability- and topology-aware accelerator scheduler for Kubernetes Dynamic
Resource Allocation (DRA). Given a heterogeneous pool of accelerators, Sift
places each workload on the right *kind* of device, in the right *place*, at the
right *cost* — something the legacy integer device-plugin model
(`nvidia.com/gpu: N`) cannot express.

> **Status:** early, in active development. I'm building this as an educational, hardware-free
> project and not a production driver.

## How it works

Every scheduling decision lives in one pure, dependency-free package,
`allocator`, shaped as **filter → score → bind**: hard constraints (device type,
memory, precision, topology) filter the pool to feasible devices; soft
preferences (cost, locality) score the survivors; the best is bound. The legacy
scheduler look for a first fit and have no scoring. This project aims to compare
the legacy method to a more advanced scheduler.

## Layout

| Path | Purpose |
|------|---------|
| `allocator/` |  Go decision logic and data model |
| `config/` | Loads scenario YAML into `[]allocator.Device` |
| `scenarios/` | Fleet definitions |
| `docs/` | Decisions (ADRs), worklog, backlog, and concept notes. |

## Build & test

```sh
go build ./...
go test ./...
```

## Docs

- [`docs/concepts`](docs/concepts) — conceptual findings made along the way