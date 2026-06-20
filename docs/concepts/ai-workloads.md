# AI Workload Shapes

## What it is
Different AI workloads stress different hardware properties — learn the
*properties*, not a rigid taxonomy. The main shapes:
- **Training / fine-tuning** — long-running, latency-insensitive, memory- and
  throughput-hungry; needs high precision (bf16) for stability; multi-device and
  communication-heavy (wants fast interconnect); often gang-scheduled.
- **Inference serving** — latency-sensitive, and it *splits in two*:
  - **Prefill** (process the prompt): compute-bound; wants raw FLOPs.
  - **Decode** (generate tokens): memory-bandwidth-bound and sequential; tolerates
    low precision (fp8/int8); KV-cache hungry.
  This split is the basis of "disaggregated serving".
- **Batch / offline** — throughput-oriented, latency-tolerant → cost-sensitive;
  happy on cheap or preemptible hardware.
- **Agentic** — bursty, many small steps, often idle, sometimes needs isolation.

## Why it matters for Sift
These appetites are the `Workload` attributes: precision required, memory,
device count + interconnect + gang, latency sensitivity, cost weight. The
scenario maps real workloads onto the catalog so the right one lands on the right
device (and the wrong one is rejected).

## How Sift uses it
The realistic scenario models a believable mix (interactive inference +
continuous fine-tuning + cost-sensitive batch) that genuinely *conflicts* over
the fleet — which is what arms the three failure modes.

## Gotcha / what confused me
_(your turn — e.g. "inference" hides two very different sub-workloads; precision
tolerance is what rejects an fp8-only inference ASIC for a bf16 training job)_

## See also
- `docs/concepts/heterogeneous-accelerators.md`
- `docs/concepts/capability-aware-scheduling.md`