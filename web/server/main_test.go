package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentType(t *testing.T) {
	cases := map[string]string{
		"app.wasm":  "application/wasm",
		"index.js":  "text/javascript; charset=utf-8",
		"index.css": "text/css; charset=utf-8",
		"index.html": "text/html; charset=utf-8",
		"f.yaml":    "text/yaml; charset=utf-8",
	}
	for in, want := range cases {
		if got := contentType(in); got != want {
			t.Errorf("contentType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCacheControl(t *testing.T) {
	cases := map[string]string{
		"/assets/index-abc123.js": "public, max-age=31536000, immutable",
		"/wasm/app.wasm":          "public, max-age=3600",
		"/index.html":             "no-cache",
		"/scenarios/x.yaml":       "no-cache",
	}
	for in, want := range cases {
		if got := cacheControl(in); got != want {
			t.Errorf("cacheControl(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStaticServesWasmAndFallsBack(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>sift</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "wasm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "wasm", "app.wasm"), []byte("\x00asm"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &static{root: dir}

	// .wasm is served as application/wasm
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/wasm/app.wasm", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/wasm" {
		t.Errorf("wasm Content-Type = %q, want application/wasm", ct)
	}

	// an unknown (SPA) route falls back to index.html
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/contrast/deep/link", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "sift") {
		t.Errorf("SPA fallback = %d %q, want 200 containing sift", rec.Code, rec.Body.String())
	}
}
