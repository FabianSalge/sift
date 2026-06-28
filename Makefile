GOROOT := $(shell go env GOROOT)
WASM_OUT := web/app/public/wasm
SCEN_OUT := web/app/public/scenarios

.PHONY: wasm
wasm:
	mkdir -p $(WASM_OUT)
	GOOS=js GOARCH=wasm go build -trimpath -o $(WASM_OUT)/app.wasm ./web/wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $(WASM_OUT)/wasm_exec.js
	@echo "wrote $(WASM_OUT)/app.wasm + $(WASM_OUT)/wasm_exec.js"

# Copy the canonical scenario YAML into the app as fetchable static assets, so the
# demo loads them through the real config.LoadFleet (no second source of truth).
.PHONY: scenarios
scenarios:
	mkdir -p $(SCEN_OUT)
	cp scenarios/*.yaml $(SCEN_OUT)/
	@echo "copied scenarios -> $(SCEN_OUT)"

# Everything the web app needs at runtime (regenerated; gitignored).
.PHONY: demo-assets
demo-assets: wasm scenarios

# Headless smoke test of the rendered demo (builds + serves + asserts the DOM).
.PHONY: smoke
smoke:
	./web/smoke.sh

.PHONY: test
test:
	go test ./...
