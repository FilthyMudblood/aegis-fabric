#!/usr/bin/env bash

set -euo pipefail

PROM_URL="${PROM_URL:-http://localhost:9090}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
REPORT_DIR="${REPORT_DIR:-artifacts/report}"
SCREENSHOT_DIR="${SCREENSHOT_DIR:-artifacts/screenshots}"
DEMO_AUTO_UP="${DEMO_AUTO_UP:-0}"
DEMO_TRIGGER_LOADGEN="${DEMO_TRIGGER_LOADGEN:-1}"
WAIT_TIMEOUT_SECONDS="${WAIT_TIMEOUT_SECONDS:-120}"
NOW_UTC="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

mkdir -p "${REPORT_DIR}/prometheus"
mkdir -p "${SCREENSHOT_DIR}"

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "Missing required command: ${cmd}" >&2
    exit 1
  fi
}

is_http_ready() {
  local url="$1"
  curl -fsS -o /dev/null --max-time 3 "${url}"
}

wait_for_http() {
  local url="$1"
  local name="$2"
  local timeout="$3"
  local start
  start="$(date +%s)"
  while ! is_http_ready "${url}"; do
    local now
    now="$(date +%s)"
    if (( now - start >= timeout )); then
      echo "Timeout waiting for ${name} at ${url}" >&2
      return 1
    fi
    sleep 2
  done
}

ensure_demo_stack() {
  if is_http_ready "${GRAFANA_URL}/api/health" && is_http_ready "${PROM_URL}/-/healthy"; then
    return 0
  fi

  if [[ "${DEMO_AUTO_UP}" != "1" ]]; then
    echo "Grafana or Prometheus is not reachable." >&2
    echo "Start the demo stack first: make demo-up" >&2
    echo "Or rerun with auto-start: DEMO_AUTO_UP=1 make demo-report" >&2
    exit 1
  fi

  require_cmd docker
  echo "Starting demo stack via docker compose..."
  if ! docker compose up --build -d; then
    echo "Failed to start docker compose. Ensure Docker Desktop is running." >&2
    exit 1
  fi

  echo "Waiting for Prometheus and Grafana to become healthy..."
  wait_for_http "${PROM_URL}/-/healthy" "Prometheus" "${WAIT_TIMEOUT_SECONDS}"
  wait_for_http "${GRAFANA_URL}/api/health" "Grafana" "${WAIT_TIMEOUT_SECONDS}"
}

require_cmd curl
require_cmd python3
if [[ "${DEMO_TRIGGER_LOADGEN}" == "1" || "${DEMO_AUTO_UP}" == "1" ]]; then
  require_cmd docker
fi
ensure_demo_stack

if [[ "${DEMO_TRIGGER_LOADGEN}" == "1" ]]; then
  echo "Triggering load generator to warm metrics..."
  if ! docker compose run --rm -e AFP_REQUESTS=1500 -e AFP_INTERVAL=10ms loadgen >/tmp/afp-loadgen-report.log 2>&1; then
    echo "Warning: load generator warmup failed. Continuing with current metrics." >&2
  fi
fi

# Ensure screenshots exist for the report bundle.
OUT_DIR="${SCREENSHOT_DIR}" GRAFANA_URL="${GRAFANA_URL}" ./scripts/export_demo_snapshots.sh

query_and_store() {
  local name="$1"
  local expr="$2"
  local out_json="${REPORT_DIR}/prometheus/${name}.json"
  curl -fsSL --get \
    --data-urlencode "query=${expr}" \
    "${PROM_URL}/api/v1/query" > "${out_json}"
}

extract_scalar() {
  local file="$1"
  python3 - "$file" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)

results = data.get("data", {}).get("result", [])
if not results:
    print("n/a")
    raise SystemExit(0)

value = results[0].get("value", [])
if len(value) < 2:
    print("n/a")
    raise SystemExit(0)

print(value[1])
PY
}

REJECT_EXPR='sum(rate(afp_ingress_actions_total{action=~"drop_.*|stranger_tax_reject|circuit_breaker_oom"}[1m]))'
FAST_PATH_EXPR='sum(rate(afp_ingress_actions_total{action="fast_path"}[1m]))'
CVP_PENALTY_EXPR='sum(rate(afp_cvp_penalty_events_total[1m]))'
INJECTED_P95_EXPR='histogram_quantile(0.95, sum by (le) (rate(afp_injected_delay_milliseconds_bucket[5m])))'

query_and_store "ingress_reject_rate_1m" "${REJECT_EXPR} or on() vector(0)"
query_and_store "fast_path_rate_1m" "${FAST_PATH_EXPR} or on() vector(0)"
query_and_store "cvp_penalty_events_rate_1m" "${CVP_PENALTY_EXPR} or on() vector(0)"
query_and_store "injected_delay_p95_5m" "${INJECTED_P95_EXPR} or on() vector(0)"

REJECT_VALUE="$(extract_scalar "${REPORT_DIR}/prometheus/ingress_reject_rate_1m.json")"
FAST_PATH_VALUE="$(extract_scalar "${REPORT_DIR}/prometheus/fast_path_rate_1m.json")"
CVP_PENALTY_VALUE="$(extract_scalar "${REPORT_DIR}/prometheus/cvp_penalty_events_rate_1m.json")"
INJECTED_P95_VALUE="$(extract_scalar "${REPORT_DIR}/prometheus/injected_delay_p95_5m.json")"

cat > "${REPORT_DIR}/README.md" <<EOF
# AFP Demo Report

Generated at (UTC): ${NOW_UTC}

## Key Signals
- Ingress reject rate (1m): \`${REJECT_VALUE}\` req/s
- Fast path throughput (1m): \`${FAST_PATH_VALUE}\` req/s
- CVP penalty events rate (1m): \`${CVP_PENALTY_VALUE}\` events/s
- Injected delay p95 (5m): \`${INJECTED_P95_VALUE}\` ms

## Included Screenshots
- \`${SCREENSHOT_DIR}/01-ingress-reject-rate.png\`
- \`${SCREENSHOT_DIR}/02-fast-path-throughput.png\`
- \`${SCREENSHOT_DIR}/03-cvp-penalty-events.png\`
- \`${SCREENSHOT_DIR}/04-injected-delay-p95.png\`

## Raw Prometheus Query Evidence
- \`${REPORT_DIR}/prometheus/ingress_reject_rate_1m.json\`
- \`${REPORT_DIR}/prometheus/fast_path_rate_1m.json\`
- \`${REPORT_DIR}/prometheus/cvp_penalty_events_rate_1m.json\`
- \`${REPORT_DIR}/prometheus/injected_delay_p95_5m.json\`
EOF

cat > "${REPORT_DIR}/report.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>AFP Demo Report</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
      margin: 32px auto;
      max-width: 1080px;
      line-height: 1.5;
      color: #1f2937;
      padding: 0 16px;
    }
    h1, h2 { margin-bottom: 8px; }
    .muted { color: #6b7280; margin-top: 0; }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 12px;
      margin: 16px 0 28px;
    }
    .card {
      border: 1px solid #e5e7eb;
      border-radius: 10px;
      padding: 12px 14px;
      background: #f9fafb;
    }
    .label { font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: .04em; }
    .value { font-size: 20px; font-weight: 600; margin-top: 4px; }
    .shots {
      display: grid;
      grid-template-columns: 1fr;
      gap: 20px;
    }
    figure {
      margin: 0;
      border: 1px solid #e5e7eb;
      border-radius: 10px;
      padding: 10px;
      background: #fff;
    }
    img { width: 100%; height: auto; border-radius: 6px; }
    figcaption { font-size: 14px; color: #374151; margin-top: 8px; }
    code { background: #f3f4f6; padding: 2px 6px; border-radius: 6px; }
  </style>
</head>
<body>
  <h1>AFP Demo Report</h1>
  <p class="muted">Generated at (UTC): ${NOW_UTC}</p>

  <h2>Key Signals</h2>
  <div class="grid">
    <div class="card">
      <div class="label">Ingress Reject Rate (1m)</div>
      <div class="value">${REJECT_VALUE} req/s</div>
    </div>
    <div class="card">
      <div class="label">Fast Path Throughput (1m)</div>
      <div class="value">${FAST_PATH_VALUE} req/s</div>
    </div>
    <div class="card">
      <div class="label">CVP Penalty Events Rate (1m)</div>
      <div class="value">${CVP_PENALTY_VALUE} events/s</div>
    </div>
    <div class="card">
      <div class="label">Injected Delay p95 (5m)</div>
      <div class="value">${INJECTED_P95_VALUE} ms</div>
    </div>
  </div>

  <h2>Dashboard Snapshots</h2>
  <div class="shots">
    <figure>
      <img src="../screenshots/01-ingress-reject-rate.png" alt="Ingress Reject Rate (1m)" />
      <figcaption>Ingress Reject Rate (1m)</figcaption>
    </figure>
    <figure>
      <img src="../screenshots/02-fast-path-throughput.png" alt="Fast Path Throughput (1m)" />
      <figcaption>Fast Path Throughput (1m)</figcaption>
    </figure>
    <figure>
      <img src="../screenshots/03-cvp-penalty-events.png" alt="CVP Penalty Events (1m)" />
      <figcaption>CVP Penalty Events (1m)</figcaption>
    </figure>
    <figure>
      <img src="../screenshots/04-injected-delay-p95.png" alt="Injected Delay p95 (5m)" />
      <figcaption>Injected Delay p95 (5m)</figcaption>
    </figure>
  </div>

  <h2>Raw Prometheus Evidence</h2>
  <ul>
    <li><code>prometheus/ingress_reject_rate_1m.json</code></li>
    <li><code>prometheus/fast_path_rate_1m.json</code></li>
    <li><code>prometheus/cvp_penalty_events_rate_1m.json</code></li>
    <li><code>prometheus/injected_delay_p95_5m.json</code></li>
  </ul>
</body>
</html>
EOF

echo "Demo report generated: ${REPORT_DIR}/README.md"
echo "Demo report generated: ${REPORT_DIR}/report.html"
