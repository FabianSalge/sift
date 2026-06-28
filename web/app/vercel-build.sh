#!/usr/bin/env bash
# Vercel build for the Sift demo. Vercel's Root Directory is web/app, and its
# build image is Node-only — so we fetch Go to compile the WASM engine, then run
# the normal Vite build. Runs locally too (skips the Go download if go is already
# on PATH), so the same script verifies the pipeline.
set -euo pipefail

GO_VERSION="${GO_VERSION:-1.26.4}"
if ! command -v go >/dev/null 2>&1; then
  echo "→ installing Go ${GO_VERSION} (Vercel image has no Go)…"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /tmp -xz
  export PATH="/tmp/go/bin:$PATH"
fi
export GOCACHE="${GOCACHE:-/tmp/.gocache}" GOPATH="${GOPATH:-/tmp/.gopath}"
go version

echo "→ building WASM engine + scenario assets"
( cd ../.. && make demo-assets )

echo "→ building static bundle"
npm run build
