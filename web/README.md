# Sift web demo

A browser visualization of the Sift scheduler. The Go decision core
(`allocator`) is compiled to WebAssembly (`web/wasm`) and driven by a Svelte UI
(`web/app`) — Contrast, Explain, and Sandbox views over one accelerator fleet.
Everything runs client-side; there is no backend.

## Layout

| Path | Purpose |
|------|---------|
| `web/wasm/` | Go → WASM engine: `siftLoadScenario` / `siftRun` / `siftExplain` (JSON) |
| `web/app/` | Svelte + Vite UI |
| `web/server/` | Tiny static-file server for the self-hosted container |
| `web/Dockerfile` | Multi-stage distroless image |

## Develop

```sh
make demo-assets                 # build app.wasm + wasm_exec.js + copy scenarios
npm --prefix web/app install
npm --prefix web/app run dev     # http://localhost:5173
```

`make demo-assets` regenerates the WASM and copies the scenario YAML into
`web/app/public/` (both gitignored — the scenarios stay single-sourced from
`/scenarios`).

## Build & test

```sh
( cd web/app && bash vercel-build.sh )   # full build → web/app/dist
make smoke                                # headless DOM smoke test of each mode
```

Deep-linkable views: `?mode=explain&wl=train-llm&stage=score`,
`?mode=contrast&preset=topology&show=legacy`, `?mode=sandbox`.

## Deploy

**Vercel (primary).** The bundle is fully static. Point a Vercel project at this
repo with **Root Directory = `web/app`**; `web/app/vercel.json` does the rest —
it installs Go in the build (Vercel's image is Node-only), compiles the WASM via
`make demo-assets`, then runs the Vite build. Push to deploy.

**Container (self-host anywhere).** A hardened distroless image serves the same
bundle:

```sh
make demo-image                  # docker build -f web/Dockerfile -t sift-demo .
make demo-run                    # docker compose up → http://localhost:8080
```

The runtime image is `distroless/static:nonroot` (no shell, no package manager),
runs read-only as non-root, and exposes `/healthz`. The server (`web/server`)
sets the correct `application/wasm` MIME, serves precompressed `.gz` assets,
caches hashed bundles immutably, and applies a CSP that permits WASM
instantiation.
