//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/FabianSalge/sift/web/wasm/engine"
)

// jsonFunc adapts an engine function (JSON []byte in/out) to a JS callback that
// takes string args and returns { ok: bool, data?: string, error?: string }.
func jsonFunc(arity int, fn func(args []js.Value) ([]byte, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < arity {
			return map[string]any{"ok": false, "error": "too few arguments"}
		}
		out, err := fn(args)
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}
		}
		return map[string]any{"ok": true, "data": string(out)}
	})
}

func main() {
	js.Global().Set("siftLoadScenario", jsonFunc(1, func(a []js.Value) ([]byte, error) {
		return engine.LoadScenario([]byte(a[0].String()))
	}))
	js.Global().Set("siftRun", jsonFunc(2, func(a []js.Value) ([]byte, error) {
		return engine.Run([]byte(a[0].String()), []byte(a[1].String()))
	}))
	js.Global().Set("siftExplain", jsonFunc(3, func(a []js.Value) ([]byte, error) {
		return engine.Explain([]byte(a[0].String()), []byte(a[1].String()), []byte(a[2].String()))
	}))
	select {} // keep the Go runtime alive so the exported funcs stay callable
}
