#!/bin/sh
set -eu

SOCKET="${AFP_IPC_SOCKET:-/var/run/afp/agent.sock}"

echo "afp-demo-agent: waiting for sidecar IPC at ${SOCKET}"
i=0
while [ ! -S "${SOCKET}" ]; do
  i=$((i + 1))
  if [ "${i}" -ge 120 ]; then
    echo "afp-demo-agent: timeout waiting for ${SOCKET}" >&2
    exit 1
  fi
  sleep 1
done

echo "afp-demo-agent: sidecar socket ready"
exec python /app/examples/langgraph_planner.py "$@"
