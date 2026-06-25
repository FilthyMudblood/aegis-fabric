# AFP Kubernetes Deployment

Standard sidecar co-deployment for enterprise agent nodes.

## Topology

```text
┌────────────────────────────────────────────── Pod (shared netns) ──────────────────────────────────────────────┐
│                                                                                                                │
│  ┌─────────────────────────┐      emptyDir UDS       ┌──────────────────────────┐                             │
│  │ agent-core              │  /var/run/afp/agent.sock │ afp-sidecar              │                             │
│  │ AFP_SDK_FAIL_MODE=closed│ ◄──────────────────────► │ AFP_ENTROPY_LIMIT=0.95   │                             │
│  │ LangGraph + afp_sdk     │                          │ ACC + IPC gRPC + ingress   │                             │
│  └───────────┬─────────────┘                          └────────────┬─────────────┘                             │
│              │ localhost:8081 egress                               │ :8080 mesh ingress                         │
└──────────────┼────────────────────────────────────────────────────┼────────────────────────────────────────────┘
               │                                                      │
               └──────── outbound mesh calls ─────────────────────────┘
```

## Apply

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap-afp.yaml
kubectl apply -f deploy/kubernetes/agent-pod-template.yaml
```

For local images built from this repo:

```bash
docker build -t ghcr.io/filthymudblood/aegis-fabric-sidecar:latest .
# kind/minikube: load image into cluster, then set imagePullPolicy: Never in the manifest
```

## Key contracts

| Concern | Agent container | Sidecar container |
|--------|-----------------|-------------------|
| IPC socket | `AFP_IPC_SOCKET=/var/run/afp/agent.sock` | same path |
| Fail mode | `AFP_SDK_FAIL_MODE=closed` | n/a |
| Entropy limit | via SDK pre-flight | `AFP_ENTROPY_LIMIT=0.95` |
| Shared volume | `afp-ipc-socket` → `/var/run/afp` | same mount |
| Network hijack | n/a | `NET_ADMIN` capability (iptables/eBPF hook point) |

## Helm skeleton

`deploy/kubernetes/helm/` contains `Chart.yaml` and `values.yaml` as the starting point for a chart that templates `agent-pod-template.yaml`.

## Notes

- Pod network namespace is shared: agent outbound traffic should target `127.0.0.1:8081` (sidecar egress).
- SDK IPC readiness is probed via `test -S /var/run/afp/agent.sock`.
- iptables/eBPF enforcement is not yet automated in the sidecar binary; `NET_ADMIN` is reserved for the next network hijack PR.
