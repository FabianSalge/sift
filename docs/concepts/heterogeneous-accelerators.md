# Heterogeneous Accelerators

## What it is
Real AI fleets mix *qualitatively different* compute, not just faster/slower
copies of one chip. The axes of variation:
- **type / vendor** — GPU (NVIDIA, AMD) vs ASICs (Google TPU, AWS Trainium /
  Inferentia); vendor is a proxy for software stack (CUDA vs ROCm).
- **memory** — capacity differs a lot (e.g. 80GB vs 192GB).
- **precision** — which numeric formats it supports (bf16 / fp16 / fp8 / int8).
- **interconnect** — NVLink vs Infinity Fabric vs ICI vs none.
- **cost** — wildly different $/hr (an inference ASIC can be far cheaper).
- **specialization** — some chips train *and* infer; some (Inferentia) only
  infer and cannot train.

## Why it matters for Sift
This variation *is* the capability model. Each axis becomes a `Device` attribute,
and the differences are what let workloads land on the *right kind* of chip — the
thing the integer model can't express at all.

## How Sift uses it
The catalog is **data** (a list of `Device` values), never logic — the scheduler
reasons only over generic attributes, so any new chip slots in by adding a row.
The single-cluster scenario uses a provider-coherent on-prem mix (NVIDIA GPU +
AMD MI300X + an inference ASIC).


## See also
- `docs/concepts/capability-aware-scheduling.md`
- ADR-0011 (dropped provider), ADR-0012 (catalog)