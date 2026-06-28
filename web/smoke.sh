#!/usr/bin/env bash
# Headless smoke test for the web demo: builds it, serves the production bundle,
# and asserts the WASM-rendered DOM contains the expected content for each mode.
# No Playwright/Puppeteer — just a Chrome/Chromium --dump-dom. Exits non-zero on
# any missing assertion.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Locate a Chrome/Chromium binary across platforms.
CHROME=""
for c in \
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  "/Applications/Chromium.app/Contents/MacOS/Chromium" \
  "$(command -v google-chrome 2>/dev/null || true)" \
  "$(command -v chromium 2>/dev/null || true)" \
  "$(command -v chromium-browser 2>/dev/null || true)"; do
  if [ -n "$c" ] && [ -x "$c" ]; then CHROME="$c"; break; fi
done
if [ -z "$CHROME" ]; then echo "smoke: no Chrome/Chromium found; skipping"; exit 2; fi

echo "smoke: building demo assets + bundle"
make demo-assets >/dev/null
npm --prefix web/app run build >/dev/null

PORT="${SMOKE_PORT:-4178}"
npm --prefix web/app run preview -- --port "$PORT" >/tmp/sift-smoke-preview.log 2>&1 &
PV=$!
trap 'kill "$PV" 2>/dev/null || true' EXIT
for _ in $(seq 1 30); do curl -sf -o /dev/null "http://localhost:$PORT/" && break; sleep 1; done

dom() { "$CHROME" --headless=new --disable-gpu --no-sandbox --virtual-time-budget=16000 --dump-dom "$1" 2>/dev/null; }

fail=0
check() {
  local url="$1"; shift
  local html; html="$(dom "$url")"
  echo "$url"
  for n in "$@"; do
    if grep -qF -- "$n" <<<"$html"; then echo "  ok  $n"; else echo "  MISS $n"; fail=1; fi
  done
}

check "http://localhost:$PORT/?mode=contrast" \
  "18 devices" "INFERENTIA2" "H100" "type-correct" "16.45"
check "http://localhost:$PORT/?mode=explain&wl=train-llm&stage=score" \
  "by score rank" "mi300x-0" "rejected" "trainable"
check "http://localhost:$PORT/?mode=sandbox" \
  "author a workload" "feasible"

if [ "$fail" -ne 0 ]; then echo "smoke: FAILED"; exit 1; fi
echo "smoke: PASSED"
