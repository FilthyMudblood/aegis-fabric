# cmd — executable entrypoints

Binaries are grouped by **role**, not by build order.

```text
cmd/
├── dataplane/          # Production: sidecar data path + IPC clients
│   ├── sidecar/        # Main AFP sidecar (ingress, egress, UDS IPC, metrics)
│   ├── preflightclient/# CLI: probe PreFlightCheck over UDS
│   ├── egressclient/   # CLI: local egress dispatch test
│   └── testclient/     # CLI: ingress frame test
│
├── controlplane/       # Production: K8s + gRPC policy stream
│   ├── operator/       # AFPClusterPolicy → ConfigMap + stream publish
│   ├── policy-controller/ # StreamPolicyUpdates hub (Kill Switch)
│   ├── policyctl/      # CLI: emergency policy override
│   └── gencerts/       # Dev: generate mTLS material for kind
│
└── demo/               # Verification, Monte Carlo, L7 demos (non-production)
    ├── simulator/      # Monte Carlo mesh survival
    ├── http_gateway/   # L7 blackbox demo
    ├── loadgen/        # Load generator
    ├── looptester/     # Recursion loop harness
    ├── modetester/     # Run-mode verification
    └── node_z/         # Topology test node
```

**Build production stack:** `make build` · **Run sidecar:** `go run ./cmd/dataplane/sidecar`

See [CODEBASE.md](../CODEBASE.md) for the full repository map.
