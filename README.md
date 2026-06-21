# Sift

A capability- and topology-aware accelerator scheduler for Kubernetes Dynamic
Resource Allocation (DRA). Given a heterogeneous pool of accelerators, Sift
places each workload on the right *kind* of device, in the right *place*, at the
right *cost* — something the legacy integer device-plugin model
(`nvidia.com/gpu: N`) cannot express.

Sift is an educational, hardware-free implementation and visualization of a real,
shipping idea — the mechanism behind production drivers like NVIDIA's
ComputeDomains, generalized to a multi-vendor, beyond-GPU fleet. It is a model to
learn from and run, not a production driver.

> **Status:** active development. The single-cluster core — allocator, fleet
> loader, benchmark, and a real DRA driver that publishes the fleet and lets the
> kube-scheduler select against it — is in place. A browser demo is next.

## How it works

Every scheduling decision lives in one pure, dependency-free package, `allocator`,
shaped as **filter → score → bind**: hard constraints (device type, memory,
precision, topology) filter the pool to the feasible devices; soft preferences
(cost, locality) score the survivors; the best is bound. A legacy first-fit
scheduler — no scoring, ignores capabilities — is implemented alongside it. The
contrast between the two is the point.

That one core is the design. It is imported, never forked, by:

- the **benchmark** (`cmd/bench`) — a text contrast of Sift vs. legacy over a fleet;
- the **DRA driver** (`driver/`, a fork of dra-example-driver) — publishes the
  fleet as `ResourceSlice`s and lets the kube-scheduler select devices via CEL
  derived from the same allocator, pinned by a parity test so the two can't drift;
- *(planned)* a **WASM demo** — the allocator compiled to the browser.

## What it shows

Three placement failures the integer device-plugin model can't prevent — each
with a test asserting Sift catches it where first-fit does not:

1. **Type-rejection** — a training job must not land on a non-trainable,
   inference-only ASIC.
2. **Cost** — a cost-sensitive job takes the cheapest *fitting* device, not the
   first free expensive one.
3. **Topology** — a same-island multi-device job must not fragment across islands
   (in-cluster, this is DRA's `matchAttribute: island`).

## Layout

| Path | Purpose |
|------|---------|
| `allocator/` | Pure Go decision logic and data model |
| `config/` | Loads scenario YAML into `[]allocator.Device` |
| `dra/` | Maps the model into DRA's vocabulary — attributes and CEL selectors |
| `scenarios/` | Fleet definitions |
| `cmd/bench/` | Sift-vs-legacy benchmark |
| `driver/` | DRA driver fork that publishes the fleet (git submodule) |
| `docs/concepts/` | Concept notes written along the way |

## Build & test

```sh
go build ./...
go test ./...
go run ./cmd/bench
```

## Docs

- [`docs/concepts`](docs/concepts) — notes on DRA, heterogeneous accelerators,
  topology islands, and capability-aware scheduling.
