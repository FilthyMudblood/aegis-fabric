#!/usr/bin/env bash

set -euo pipefail

GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
GRAFANA_USER="${GRAFANA_USER:-admin}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-admin}"
FROM_MS="${FROM_MS:-$(( ( $(date +%s) - 900 ) * 1000 ))}"
TO_MS="${TO_MS:-$(( $(date +%s) * 1000 ))}"
OUT_DIR="${OUT_DIR:-artifacts/screenshots}"

mkdir -p "${OUT_DIR}"

render_panel() {
  local panel_id="$1"
  local name="$2"
  local out_file="${OUT_DIR}/${name}.png"
  local url="${GRAFANA_URL}/render/d-solo/afp-core-demo/afp-core-demo?orgId=1&panelId=${panel_id}&from=${FROM_MS}&to=${TO_MS}&width=1600&height=900&tz=Asia%2FShanghai"

  echo "Rendering panel ${panel_id} -> ${out_file}"
  curl -fsSL -u "${GRAFANA_USER}:${GRAFANA_PASSWORD}" "${url}" -o "${out_file}"
}

echo "Exporting AFP dashboard snapshots from ${GRAFANA_URL}"
sleep 12
render_panel 1 "01-ingress-reject-rate"
render_panel 2 "02-fast-path-throughput"
render_panel 3 "03-cvp-penalty-events"
render_panel 4 "04-injected-delay-p95"
echo "Done. Screenshots saved to ${OUT_DIR}"
