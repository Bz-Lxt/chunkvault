#!/usr/bin/env bash
set -uo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
WORKDIR="$(mktemp -d "${TMPDIR:-/tmp}/cv.XXXXXX")"
PORT=""
APP_PID=""
BIN=""

die_env() {
  echo "[EXPECT] service reachable"
  echo "[ACTUAL] $1"
  exit 2
}

cleanup() {
  if [[ -n "${APP_PID:-}" ]]; then
    kill "$APP_PID" >/dev/null 2>&1 || true
    wait "$APP_PID" 2>/dev/null || true
  fi
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

pick_port() {
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
}

start_srv() {
  mkdir -p "$WORKDIR/data"
  if [[ -z "$BIN" ]]; then
    BIN="$WORKDIR/chunkvault"
    (cd "$ROOT" && go build -o "$BIN" ./cmd/chunkvault) || die_env "build failed"
  fi
  PORT="$(pick_port)"
  "$BIN" -addr ":$PORT" -data "$WORKDIR/data" -chunk 32 -web "$ROOT/web" >"$WORKDIR/log" 2>&1 &
  APP_PID=$!
  for _ in $(seq 1 80); do
    if curl -sf "http://127.0.0.1:$PORT/health" >/dev/null; then
      return 0
    fi
    sleep 0.1
  done
  die_env "health timeout $(tr '\n' ' ' < "$WORKDIR/log")"
}

json_get() {
  python3 -c "import json,sys; print(json.load(sys.stdin)$1)"
}

export CHUNKVAULT_SLOW_MS=400
start_srv
BODY="$(python3 -c 'print("Q"*80)')"
curl -sS -m 0.25 -X PUT "http://127.0.0.1:$PORT/v1/blobs?name=c" --data-binary "$BODY" >/dev/null || true
sleep 2
STATS="$(curl -sS "http://127.0.0.1:$PORT/v1/stats")"
BLOBS="$(printf '%s' "$STATS" | json_get "['blobs']")"
ACTUAL="blobs=$BLOBS"

echo "[EXPECT] blobs=0 after canceled PUT"
echo "[ACTUAL] $ACTUAL"
if [[ "$BLOBS" == "0" ]]; then
  exit 0
fi
exit 1
