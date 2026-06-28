//go:build !(js && wasm)

package main

import "fmt"

// This binary targets the browser:
//   GOOS=js GOARCH=wasm go build -o app.wasm ./web/wasm   (see `make wasm`)
// The host stub exists only so `go build ./...` succeeds on the dev platform,
// where the js/wasm main is excluded by build tags.
func main() {
	fmt.Println("sift wasm engine: build with GOOS=js GOARCH=wasm (make wasm)")
}
