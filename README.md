# Sift

Sift schedules GPU and accelerator workloads onto a mixed hardware fleet on
Kubernetes. It matches each workload to a device that can actually run it, in the
right place on the interconnect, at a reasonable cost. Standard Kubernetes treats
devices as interchangeable integers (`nvidia.com/gpu: 4`) and can express none of
that.

It's built on Dynamic Resource Allocation, the API that replaced device plugins
in Kubernetes 1.34. Production DRA drivers such as NVIDIA's ComputeDomains already
do this for one vendor; Sift is a smaller, runnable version of the same idea
across several vendors and device types. It needs no real hardware, so you can
read it, run it, and see why it makes each decision. It's a learning project, not
a production driver.

> **Status:** active development. The single-cluster core is in place: the
> allocator, a fleet loader, the benchmark, a real DRA driver that publishes the
> fleet for the kube-scheduler to select against, and a browser demo of the
> scheduler running live.

## How it works

All the scheduling logic lives in one package, `allocator`, with no Kubernetes or
YAML dependencies. It runs in three steps: filter, score, bind. Hard constraints
(device type, memory, precision, topology) filter the fleet down to the devices
that can run the workload. Soft preferences (cost, locality) score whatever
survives. The best-scoring device gets bound.

A legacy scheduler sits alongside it and does what the old device-plugin model
did: take the first free device, with no capability check and no scoring. Running
the two on the same fleet is how Sift shows what capability-aware placement buys
you.

That one package is the whole design. Three things import it, and none of them
fork it:

- the benchmark (`cmd/bench`): a text comparison of Sift against the legacy
  scheduler over a fleet.
- the DRA driver (`driver/`, a fork of dra-example-driver): it publishes the fleet
  as `ResourceSlice`s and lets the kube-scheduler select devices using CEL
  generated from the same allocator. A parity test keeps the CEL and the allocator
  from drifting apart.
- a browser demo (`web/`): the allocator compiled to WASM and driven live in a
  page, no server involved.

## What it shows

Three placement mistakes the integer device-plugin model can't prevent. Each one
has a test asserting that Sift avoids it and the legacy scheduler does not:

1. **Type-rejection.** A training job must not land on a non-trainable,
   inference-only ASIC.
2. **Cost.** A cost-sensitive job should take the cheapest device that fits, not
   the first free expensive one.
3. **Topology.** A multi-device job that needs a single interconnect island must
   not get split across two. In-cluster, this is DRA's `matchAttribute: island`.

`go run ./cmd/bench` runs all three on one fleet, Sift on the left and the legacy
scheduler on the right:

```text
Sift vs Legacy · realistic-2026 (18 devices, 5 workloads)

  workload      Sift                              Legacy
  --------      ----                              ------
  train-llm     mi300x-0 ($1.90) ok               inferentia2-0 ($0.75) WRONG-TYPE
  train-big     mi300x-1 ($1.90) ok               inferentia2-1 ($0.75) WRONG-TYPE
  infer-fp8     mi300x-2 ($1.90) ok               inferentia2-2 ($0.75) WRONG-TYPE
  infer-int8    inferentia2-0 ($0.75) ok          inferentia2-3 ($0.75) ok
  gang-train    h100-0 ×4 ($10.00) ok             b200-0 ×4 ($17.00) FRAGMENTED

  totals        Sift          Legacy
  total $/hr    $16.45        $20.00
  type-correct  5/5           2/5
  fragmented    0             1
  pending       0             0
```

The legacy scheduler grabs the first free device every time. It hands int8-only
inference ASICs to training jobs that can't run on them, and splits a single-island
gang across two islands. Sift gives each job a device that fits, picks the cheapest
one that does, and keeps the gang on one island, for a lower total bill. This is a
demonstration of the three failure modes, not a performance benchmark.

## Reading a decision

The table shows what each scheduler did. The `-explain` flag shows why.
`allocator.Explain` replays the same filter, score, and bind over the fleet and
reports a verdict for every device. You can reproduce it from the command line:

```text
$ go run ./cmd/bench -explain train-llm
train-llm  ·  train · needs ≥80GB, bf16 · 1 device · cost-weight 0.5
realistic-2026 · 18 devices · filter → score → bind
score = $/hr × cost-weight, then memory waste, then ID (lower wins)

  rank  device          verdict
  ----  ------          -------
  —     inferentia2-0   reject: not trainable, 32GB<80GB, no bf16
  —     inferentia2-1   reject: not trainable, 32GB<80GB, no bf16
  —     inferentia2-2   reject: not trainable, 32GB<80GB, no bf16
  —     inferentia2-3   reject: not trainable, 32GB<80GB, no bf16
  13    b200-0          ok    cost 3.00   waste 112GB
  14    b200-1          ok    cost 3.00   waste 112GB
  5     h100-0          ok    cost 1.25   waste 0GB
  6     h100-1          ok    cost 1.25   waste 0GB
  7     h100-2          ok    cost 1.25   waste 0GB
  8     h100-3          ok    cost 1.25   waste 0GB
  9     h100-4          ok    cost 1.25   waste 0GB
  10    h100-5          ok    cost 1.25   waste 0GB
  11    h100-6          ok    cost 1.25   waste 0GB
  12    h100-7          ok    cost 1.25   waste 0GB
  1     mi300x-0        ok    cost 0.95   waste 112GB   BIND
  2     mi300x-1        ok    cost 0.95   waste 112GB
  3     mi300x-2        ok    cost 0.95   waste 112GB
  4     mi300x-3        ok    cost 0.95   waste 112GB

  bound   mi300x-0   $1.90/hr   (lowest score of 14 feasible; 4 rejected)
```

`train-llm` needs a trainable device with at least 80GB and bf16. The filter
rejects all four Inferentia2s: they can't train, they're too small, and they lack
bf16. They're also the cheapest devices on the fleet and listed first, which is
exactly what the legacy scheduler falls for. Among the devices that pass, the
score prefers the cheapest one that fits, weighted by how cost-sensitive the job
is. The MI300X wins at $1.90/hr, ahead of an equally capable H100 at $2.50 and a
B200 at more than three times the price. Every row in the table above is decided
the same way.

## Try it in the browser

`web/` compiles `allocator` (plus `config`, `report`, and a small `sim` package)
to WASM and runs the whole scheduler client-side — no backend, just a static
page. It opens on a cluster that's already running: seeded ambient traffic is
arriving and being placed before you touch anything.

From there you're operating it, not just watching it. Create a workload
template — kind, memory floor, required precisions, device count, same-island —
give it a rate, and it starts showing up in the queue on its own; the burst
buttons (`×1`, `×5`) fire a batch on demand. Add machines from a small catalog
(an H100 island, an MI300X island, a couple of Inferentia2s, a cheap batch node)
and the fleet grows mid-run; drain a node and it stops taking new work, finishes
what's running, and leaves, kubectl-style.

A shadow strip under the header runs the legacy first-fit scheduler on the exact
same arrivals and fleet edits, in the background — it's never drawn as a second
fleet, only as numbers next to Sift's: devices wasted on jobs they can't
actually run, queue depth, and cost. Same traffic, two outcomes. Click any
running or queued job in the cluster view and a panel opens with the filter →
score → bind trace behind its placement (or, for a job still waiting, why
nothing fits yet).

`?seed=` and `?speed=` are the only two URL params the demo reads: seed fixes the
ambient-traffic generator so a run is reproducible, speed controls how fast
simulated time moves. Time only runs forward — there's no scrubbing back.

```sh
npm --prefix web/app run dev     # local dev server
npm --prefix web/app run build   # production bundle in web/app/dist
```

## Layout

| Path | Purpose |
|------|---------|
| `allocator/` | Pure Go decision logic and data model |
| `config/` | Loads scenario YAML into `[]allocator.Device` |
| `dra/` | Maps the model to DRA attributes and CEL selectors |
| `scenarios/` | Fleet definitions |
| `cmd/bench/` | Sift-vs-legacy benchmark |
| `driver/` | DRA driver fork that publishes the fleet (git submodule) |
| `web/` | Browser demo: the allocator compiled to WASM, driven live |
| `docs/concepts/` | Notes on the concepts behind the project |

## Build & test

```sh
go build ./...
go test ./...
go run ./cmd/bench                    # Sift vs the legacy scheduler
go run ./cmd/bench -explain train-llm # explain one placement, step by step
```

## Docs

- [`docs/concepts`](docs/concepts): notes on DRA, heterogeneous accelerators,
  topology islands, and capability-aware scheduling.
