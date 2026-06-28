GOROOT := $(shell go env GOROOT)
WASM_OUT := web/app/public/wasm

.PHONY: wasm
wasm:
	mkdir -p $(WASM_OUT)
	GOOS=js GOARCH=wasm go build -trimpath -o $(WASM_OUT)/app.wasm ./web/wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $(WASM_OUT)/wasm_exec.js
	@echo "wrote $(WASM_OUT)/app.wasm + $(WASM_OUT)/wasm_exec.js"

.PHONY: test
test:
	go test ./...
