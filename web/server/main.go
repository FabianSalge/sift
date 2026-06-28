// Command server is a tiny static-file server for the Sift web demo: it serves a
// built bundle (the Svelte app + the WASM engine) with the right wasm MIME,
// precompressed (.gz) serving, asset caching, an SPA fallback, security headers,
// and a /healthz endpoint. It is the only thing in the runtime container.
package main

import (
	"flag"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	dir := flag.String("dir", "dist", "directory of static files to serve")
	healthcheck := flag.Bool("healthcheck", false, "probe the local server's /healthz and exit (for container HEALTHCHECK)")
	flag.Parse()

	port := env("PORT", "8080")
	if *healthcheck {
		os.Exit(probe(port))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.WriteString(w, "ok\n")
	})
	mux.Handle("/", &static{root: *dir})

	addr := "0.0.0.0:" + port
	log.Printf("sift demo: serving %q on %s", *dir, addr)
	s := &http.Server{Addr: addr, Handler: secure(mux), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}

type static struct{ root string }

func (s *static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if upath == "/" {
		upath = "/index.html"
	}
	clean := filepath.Clean(upath)
	if strings.Contains(clean, "..") {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}

	full := filepath.Join(s.root, clean)
	if info, err := os.Stat(full); err != nil || info.IsDir() {
		clean, full = "/index.html", filepath.Join(s.root, "index.html") // SPA fallback
	}

	w.Header().Set("Cache-Control", cacheControl(clean))
	w.Header().Set("Content-Type", contentType(full))

	// Prefer a precompressed sibling when the client accepts gzip.
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		if gz, err := os.Open(full + ".gz"); err == nil {
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			_, _ = io.Copy(w, gz)
			return
		}
	}

	f, err := os.Open(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	_, _ = io.Copy(w, f)
}

func cacheControl(p string) string {
	switch {
	case strings.HasPrefix(p, "/assets/"): // vite content-hashed bundles
		return "public, max-age=31536000, immutable"
	case strings.HasPrefix(p, "/wasm/"): // large, not hashed — cache moderately
		return "public, max-age=3600"
	default: // index.html, scenarios — always revalidate
		return "no-cache"
	}
}

func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".wasm":
		return "application/wasm"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "text/yaml; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	default:
		if t := mime.TypeByExtension(filepath.Ext(path)); t != "" {
			return t
		}
		return "application/octet-stream"
	}
}

// secure adds defensive headers. WASM needs 'wasm-unsafe-eval' to instantiate.
func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func probe(port string) int {
	c := &http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
