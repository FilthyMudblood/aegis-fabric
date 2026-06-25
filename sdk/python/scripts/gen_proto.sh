#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
OUT_DIR="${ROOT}/sdk/python/afp_sdk/pb"
PROTO_FILE="${ROOT}/api/afp/v1/sdk_ipc.proto"

if ! python3 -c "import grpc_tools.protoc" >/dev/null 2>&1; then
  echo "error: grpcio-tools is not installed" >&2
  echo "hint: pip install grpcio-tools" >&2
  exit 1
fi

mkdir -p "${OUT_DIR}"
rm -rf "${OUT_DIR}/afp"

python3 -m grpc_tools.protoc \
  -I "${ROOT}/api" \
  --python_out="${OUT_DIR}" \
  --grpc_python_out="${OUT_DIR}" \
  "${PROTO_FILE}"

mv "${OUT_DIR}/afp/v1/sdk_ipc_pb2.py" "${OUT_DIR}/sdk_ipc_pb2.py"
mv "${OUT_DIR}/afp/v1/sdk_ipc_pb2_grpc.py" "${OUT_DIR}/sdk_ipc_pb2_grpc.py"
rm -rf "${OUT_DIR}/afp"

python3 - "${OUT_DIR}/sdk_ipc_pb2_grpc.py" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
text = path.read_text(encoding="utf-8")
text = text.replace("from afp.v1 import sdk_ipc_pb2 as afp_dot_v1_dot_sdk__ipc__pb2", "from . import sdk_ipc_pb2 as afp_dot_v1_dot_sdk__ipc__pb2")
path.write_text(text, encoding="utf-8")
PY

echo "Generated Python stubs in ${OUT_DIR}"
