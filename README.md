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
> kube-scheduler select against it — is in place. A browser demo is in progress.

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

`go run ./cmd/bench` puts all three on one fleet — Sift on the left, legacy
first-fit on the right:

```text
Sift vs Legacy — realistic-2026 (18 devices, 5 workloads)

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

Legacy takes the first free device every time: it spends int8-only inference
ASICs on training jobs they can't run, and splits a same-island gang across two
islands. Sift matches each job to a device that fits, picks the cheapest that
does, and keeps the gang whole — for less total cost. It is an illustration of
the three failure modes, not a benchmark evaluation.

## Reading a decision

The table shows *what* each scheduler did; `-explain` shows *why*. Sift's choice
is never a black box — `allocator.Explain` replays the same filter → score → bind
over the fleet and reports every device's verdict, reproducible from the CLI:

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
rejects all four Inferentia2s outright — they can't train, are too small, and
lack bf16 — even though they are the cheapest devices and listed first, exactly
the trap first-fit falls into. Among the survivors, the score prefers the
cheapest *fitting* device weighted by the job's cost sensitivity: the MI300X
wins at $1.90/hr over an equally capable H100 at $2.50 and a B200 at over three
times the price. Same mechanism runs behind every row of the table above.

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
go run ./cmd/bench                    # the Sift-vs-legacy contrast
go run ./cmd/bench -explain train-llm # trace one decision, filter → score → bind
```

## Docs

- [`docs/concepts`](docs/concepts) — notes on DRA, heterogeneous accelerators,
  topology islands, and capability-aware scheduling.
