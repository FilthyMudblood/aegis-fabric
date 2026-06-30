# AFP Kubernetes Deployment

Standard sidecar co-deployment for enterprise agent nodes.

> **Start here after reading the root [README.md](../../README.md).**

## Topology

```text
┌────────────────────────────────────── Pod (shared netns) ──────────────────────────────────────┐
│                                                                                                │
│  ┌──────────────────────┐   emptyDir UDS    ┌───────────────────────────┐                      │
│  │ agent-core           │ /var/run/afp/     │ afp-sidecar               │                      │
│  │ AFP_SDK_FAIL_MODE=   │ agent.sock        │ AFP_ENTROPY_LIMIT=0.95    │                      │
│  │ closed               │ ◄───────────────► │ ACC + IPC + hot-reload    │                      │
│  └──────────┬───────────┘                   └─────────────┬─────────────┘                      │
│             │ localhost:8081 egress                       │ :8080 ingress                      │
└─────────────┼─────────────────────────────────────────────┼────────────────────────────────────┘
              │                                             │
              └──────────── mesh egress ────────────────────┘

Cluster-scoped AFPClusterPolicy
        ↓ operator reconcile
afp-sidecar-config / afp-agent-config (ConfigMaps)
        ↓ volume mount as files
/etc/afp/policy/*  ──fsnotify──►  RuntimePolicy (atomic hot-reload)
```

## 10-Minute Apply (kind)

```bash
./scripts/kind-quickstart.sh
```

## Manual Apply

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap-afp.yaml
kubectl apply -f deploy/kubernetes/crd/afpclusterpolicy.yaml
kubectl apply -f deploy/kubernetes/examples/afpclusterpolicy-enterprise.yaml
kubectl apply -f deploy/kubernetes/agent-pod-template.yaml
kubectl apply -f deploy/kubernetes/operator-deployment.yaml   # optional in-cluster operator
```

Build and load local image:

```bash
docker build -t ghcr.io/filthymudblood/aegis-fabric-sidecar:latest .
kind load docker-image ghcr.io/filthymudblood/aegis-fabric-sidecar:latest --name afp
```

## Key Contracts

| Concern | Agent container | Sidecar container |
|--------|-----------------|-------------------|
| IPC socket | `AFP_IPC_SOCKET=/var/run/afp/agent.sock` | same |
| Fail mode | `AFP_SDK_FAIL_MODE=closed` | n/a |
| Entropy limit | via SDK pre-flight | `AFP_ENTROPY_LIMIT` + hot-reload |
| Shared UDS volume | `afp-ipc-socket` → `/var/run/afp` | same |
| Policy files | n/a | `afp-policy` → `/etc/afp/policy` |
| Network hijack | n/a | `NET_ADMIN` (iptables/eBPF hook point) |

## Policy Operator

```bash
kubectl apply -f deploy/kubernetes/crd/afpclusterpolicy.yaml
kubectl apply -f deploy/kubernetes/examples/afpclusterpolicy-enterprise.yaml

# out-of-cluster dev loop
go run ./cmd/operator

# or in-cluster
kubectl apply -f deploy/kubernetes/operator-deployment.yaml
```

Patch policy at runtime:

```bash
kubectl patch afpclusterpolicy enterprise-default --type merge \
  -p '{"spec":{"entropyLimit":0.80,"maxRecursionDepth":8}}'
```

Verify ConfigMap propagation:

```bash
kubectl -n afp-system get configmap afp-sidecar-config -o yaml
kubectl -n afp-system exec deploy/afp-agent-node -c afp-sidecar -- ls -la /etc/afp/policy
```

## Policy Controller (Phase 2)

```bash
kubectl apply -f deploy/kubernetes/policy-controller-deployment.yaml
kubectl -n afp-system port-forward svc/afp-policy-controller 8090:8090

# Flip Kill Switch from your workstation
go run ./cmd/policyctl --controller 127.0.0.1:8090 --kill-switch
go run ./cmd/policyctl --controller 127.0.0.1:8090 --clear
```

Sidecars connect when `AFP_POLICY_CONTROLLER_ADDR` is set (demo deployment enables this).

## Demo Agent

```bash
make demo-agent-docker
kubectl apply -f deploy/kubernetes/agent-pod-demo.yaml
kubectl -n afp-system logs -f deploy/afp-agent-node -c agent-core
```

`Dockerfile.demo-agent` packages `sdk/python/examples/langgraph_planner.py` with `--loop` for periodic recursion-breaker demos over UDS.

### Expected agent-core logs

```text
afp-demo-agent: waiting for sidecar IPC at /var/run/afp/agent.sock
afp-demo-agent: sidecar socket ready
--- langgraph planner demo (initial_depth=10) ---
[AFP SDK] LangGraph node blocked: afp-core: recursion depth exceeded physical limit, intent loop detected
annotated-stop: afp-core: recursion depth exceeded physical limit, intent loop detected
```

## In-Pod Demos

```bash
POD=$(kubectl -n afp-system get pod -l app.kubernetes.io/component=agent-node -o jsonpath='{.items[0].metadata.name}')

# Recursion breaker
kubectl -n afp-system exec "$POD" -c afp-sidecar -- \
  preflightclient --recursion-depth 12

# Intent burst
kubectl -n afp-system exec "$POD" -c afp-sidecar -- \
  preflightclient --estimated-tasks 10000
```

## Helm Skeleton

`helm/` contains `Chart.yaml` and `values.yaml` as a starting point for templating `agent-pod-template.yaml`.

## Notes

- Pod containers share a network namespace; agent egress targets `127.0.0.1:8081`.
- ConfigMap hot-reload avoids pod restarts for threshold tuning; expect ~60s kubelet sync latency.
- Sub-second Kill Switch is planned via Phase 2 gRPC `StreamPolicyUpdates`.
- iptables/eBPF automation is not yet in the sidecar binary; `NET_ADMIN` is reserved.
