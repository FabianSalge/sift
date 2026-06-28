# syntax=docker/dockerfile:1
#
# Multi-stage build for the Sift web demo. Build from the repo root:
#   docker build -f web/Dockerfile -t sift-demo:latest .
# Runtime is distroless/static (no shell, no package manager): just the static
# server binary + the built bundle. Pin the base images by @sha256 digest for
# reproducible builds.

# ---- stage 1: Go — the WASM engine + the static server ----
FROM golang:1.26.4-bookworm AS gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY allocator ./allocator
COPY config ./config
COPY report ./report
COPY web/wasm ./web/wasm
COPY web/server ./web/server
COPY scenarios ./scenarios
RUN mkdir -p /out/wasm /out/scenarios \
 && GOOS=js GOARCH=wasm go build -trimpath -o /out/wasm/app.wasm ./web/wasm \
 && cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" /out/wasm/wasm_exec.js \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./web/server \
 && cp scenarios/*.yaml /out/scenarios/

# ---- stage 2: Node — the Svelte bundle (+ precompressed assets) ----
FROM node:22-bookworm-slim AS webbuild
WORKDIR /app
COPY web/app/package.json web/app/package-lock.json ./
RUN npm ci
COPY web/app/ ./
COPY --from=gobuild /out/wasm ./public/wasm
COPY --from=gobuild /out/scenarios ./public/scenarios
RUN npm run build \
 && find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.wasm' -o -name '*.svg' -o -name '*.json' \) \
      -exec gzip -9 -k {} \;

# ---- stage 3: runtime — distroless, nonroot, read-only-friendly ----
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
LABEL org.opencontainers.image.title="sift-demo" \
      org.opencontainers.image.description="Capability- & topology-aware accelerator scheduling — browser demo" \
      org.opencontainers.image.licenses="MIT"
WORKDIR /app
COPY --from=gobuild /out/server /server
COPY --from=webbuild /app/dist /app/dist
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=2s --retries=3 \
  CMD ["/server", "-healthcheck"]
ENTRYPOINT ["/server", "-dir", "/app/dist"]
