# AFP Sidecar

**A runtime coordination layer for autonomous agent networks.**

> **"TCP routes packets. AFP governs optimizers."**

中文文档：`README.zh-CN.md`

**Whitepaper:** [AFP Technical Whitepaper](https://zenodo.org/records/20674352)

## Why This Project Exists

Modern service infrastructure (L4/L7 proxies, service mesh, RPC gateways) is optimized for deterministic request/response systems. Autonomous agents are different: they are optimizers that can recursively delegate work, amplify mistakes, and create systemic externalities (tool storms, context explosion, deadlock loops).

AFP Sidecar introduces an out-of-band governance runtime to control those failure modes before they propagate through the network.

## Current Implementation Status

Implemented and runnable:
- TCP ingress/egress data plane with LV framing (sticky/partial packet safe)
- ACC + FSM governance engine
- run-mode feature gates (`enterprise-mesh` / `open-exchange`)
- recursive topology discovery + recursion depth breaker
- HTTP middleware wrapper (`cmd/http_gateway`) for L7 compatibility proof
- Prometheus metrics + Grafana dashboard
- Docker / Compose / K8s sidecar deployment examples

Not yet fully production-complete (planned/hardenable):
- cryptographic signature verification is currently simplified
- cgroups memory source is currently runtime fallback, not full `/sys/fs/cgroup` reader
- topology referral and gossip transport are scaffolded with test stubs

## Tech Stack

- Go (`go 1.23+`)
- Protobuf (`proto3`) + `buf` + `protoc-gen-go`
- TCP framing + HTTP wrapper
- Prometheus (`client_golang`) + Grafana
- Docker / Docker Compose / Kubernetes sidecar model

## Project Layout

```text
afp-sidecar/
├─ cmd/
│  ├─ sidecar/         # main TCP runtime (ingress + egress + metrics)
│  ├─ http_gateway/    # HTTP middleware wrapper (L7 compatibility)
│  ├─ egressclient/    # local egress protocol tester
│  ├─ modetester/      # feature-gate validation payload sender
│  ├─ looptester/      # recursion-depth loop breaker tester
│  ├─ node_z/          # remote node test stub
│  ├─ testclient/      # TCP frame behavior tester
│  ├─ loadgen/         # synthetic load generator
│  └─ simulator/       # Monte Carlo resilience simulation
├─ internal/
│  ├─ config/          # run mode + bootstrap config
│  ├─ control/         # FSM + ACC kernel
│  ├─ core/            # entropy monitor + OS metrics provider
│  ├─ dataplane/       # codec / ingress / egress logic
│  ├─ topology/        # DID resolver + gossip scaffolding
│  └─ telemetry/       # Prometheus metrics definitions
├─ api/afp/v1/         # protobuf source
├─ pkg/protocol/v1/    # generated protobuf Go code
├─ deploy/
│  ├─ k8s/             # pod manifest examples
│  └─ monitor/         # local Prometheus + Grafana stack
├─ scripts/            # verification and helper scripts
└─ artifacts/          # generated reports and screenshots
```

## Run Modes

Set by `AFP_RUN_MODE`:
- `enterprise-mesh`: AFP-Core only (internal congestion + recursion safety)
- `open-exchange`: AFP-Core + zero-trust network policy (stranger tax, stronger ingress checks)

Default mode: `enterprise-mesh`.

## Quick Start

Run sidecar:

```bash
go run ./cmd/sidecar
```

Run HTTP gateway wrapper:

```bash
go run ./cmd/http_gateway
```

Run egress recursive test client:

```bash
go run ./cmd/egressclient
```

## Blackbox Dual-Mode Demo (Docker Compose)

```bash
docker-compose up -d --build
```

This starts:
- `afp-mesh-node` on `localhost:8082` (`enterprise-mesh`)
- `afp-open-node` on `localhost:8083` (`open-exchange`)

Smoke expectations:
- `8082` accepts normal requests (`200`)
- `8082` trips recursion breaker on depth > 10 (`508`)
- `8083` rejects untrusted stranger requests (`403`)

## Automated Verification

- mode feature-gate test:
  - `./scripts/verify_modes.sh`
- recursion breaker loop test:
  - `./scripts/verify_recursion_loop.sh`
- HTTP L7 blackbox test (against running containers):
  - `./scripts/verify_http_gateway.sh`
- protobuf generation:
  - `./scripts/gen_proto.sh`

## Monte Carlo Validation

Run simulation locally:

```bash
go run ./cmd/simulator
```

Run inside compose node:

```bash
docker exec afp-mesh-node simulator -runs 1000
```

## Observability

Metrics endpoint:
- `http://127.0.0.1:9090/metrics` (when sidecar is running directly)

Monitor stack files:
- `deploy/monitor/prometheus.yml`
- `deploy/monitor/docker-compose.yml`
- `deploy/monitor/grafana-dashboard-afp.json`

Start local monitor stack:

```bash
cd deploy/monitor
docker-compose up -d
```

Grafana:
- `http://127.0.0.1:3000`
- default credentials: `admin / admin`

## Demo Evidence Artifacts

```bash
make demo-snapshots
make demo-report
```

Artifacts:
- `artifacts/report/README.md`
- `artifacts/report/report.html`
- `artifacts/report/prometheus/*.json`
- `artifacts/screenshots/*.png`

## License

This project is licensed under the Apache License 2.0.
See `LICENSE`.
