# AFP Sidecar（中文说明）

`AFP Sidecar` 是一个面向 Agent 网络的治理运行时。它在请求进入本地执行环境前，执行物理层控制（熵压、信誉、递归安全）与策略裁决。

> **“TCP 路由数据包，AFP 治理优化器。”**

英文文档：`README.md`

**白皮书：** [AFP 技术白皮书](https://zenodo.org/records/20674352)

## 当前状态

### 已实现（可运行、可验证）
- TCP 数据面（LV 帧，支持粘包/半包处理）
- ACC + FSM 治理内核
- 运行模式特征开关（`enterprise-mesh` / `open-exchange`）
- 递归拓扑寻址 + 递归断路器
- HTTP 中间件包装层（L7 兼容验证）
- Prometheus 指标 + Grafana 看板
- Docker / Compose / K8s Sidecar 示例

### 尚需继续硬化（当前为原型/简化）
- 密码学签名校验仍是简化实现
- cgroups 内存读取当前采用 runtime fallback（未完整接 `/sys/fs/cgroup`）
- 拓扑 referral / gossip 传输为测试桩形态

## 技术栈

- Go（`go 1.23+`）
- Protobuf（`proto3`，`buf` + `protoc-gen-go`）
- TCP + HTTP 双路径兼容
- Prometheus + Grafana
- Docker / Docker Compose / Kubernetes

## 项目结构

```text
afp-sidecar/
├─ cmd/
│  ├─ sidecar/         # 主服务入口（ingress + egress + metrics）
│  ├─ http_gateway/    # HTTP 中间件包装层（L7 兼容）
│  ├─ egressclient/    # 本地 egress 协议测试桩
│  ├─ modetester/      # 模式开关验证流量注入
│  ├─ looptester/      # 递归深度断路验证
│  ├─ node_z/          # 远端节点测试桩
│  ├─ testclient/      # TCP 帧行为验证
│  ├─ loadgen/         # 压测流量生成
│  └─ simulator/       # Monte Carlo 推演
├─ internal/
│  ├─ config/          # 运行模式与引导配置
│  ├─ control/         # FSM + ACC 公式内核
│  ├─ core/            # 熵压监控与 OS 指标接口
│  ├─ dataplane/       # codec / ingress / egress
│  ├─ topology/        # DID 解析与 gossip 原型
│  └─ telemetry/       # Prometheus 指标定义
├─ api/afp/v1/         # protobuf 源定义
├─ pkg/protocol/v1/    # protobuf 生成代码
├─ deploy/             # k8s + monitor 配置
├─ scripts/            # 自动化验证脚本
└─ artifacts/          # 报告与截图产物
```

## 运行模式

通过环境变量 `AFP_RUN_MODE` 控制：
- `enterprise-mesh`：仅启用 AFP-Core（内网拥塞/熵压/递归安全）
- `open-exchange`：启用 AFP-Core + 零信任网络策略（陌生人税等）

默认模式：`enterprise-mesh`。

## 快速启动

启动 Sidecar：

```bash
go run ./cmd/sidecar
```

启动 HTTP Gateway（L7 包装层）：

```bash
go run ./cmd/http_gateway
```

运行递归寻址客户端：

```bash
go run ./cmd/egressclient
```

## 黑盒双模演示（Docker Compose）

```bash
docker-compose up -d --build
```

会启动两个节点：
- `localhost:8082`：`enterprise-mesh`
- `localhost:8083`：`open-exchange`

预期行为：
- `8082` 正常请求返回 `200`
- `8082` 递归深度超限返回 `508`
- `8083` 陌生人低信誉请求返回 `403`

## 自动化验证脚本

- 模式开关验证：`./scripts/verify_modes.sh`
- 递归断路验证：`./scripts/verify_recursion_loop.sh`
- HTTP 黑盒集成验证（针对已运行容器）：`./scripts/verify_http_gateway.sh`
- Protobuf 生成：`./scripts/gen_proto.sh`

## Monte Carlo 推演

本地运行：

```bash
go run ./cmd/simulator
```

容器内运行（非交互）：

```bash
docker exec afp-mesh-node simulator -runs 1000
```

## 可观测性

指标端点：
- `http://127.0.0.1:9090/metrics`（本地 sidecar 直跑时）

监控栈配置：
- `deploy/monitor/prometheus.yml`
- `deploy/monitor/docker-compose.yml`
- `deploy/monitor/grafana-dashboard-afp.json`

启动：

```bash
cd deploy/monitor
docker-compose up -d
```

Grafana：
- `http://127.0.0.1:3000`
- 默认账号密码：`admin / admin`

## Demo 与证据导出

```bash
make demo-snapshots
make demo-report
```

输出目录：
- `artifacts/report/README.md`
- `artifacts/report/report.html`
- `artifacts/report/prometheus/*.json`
- `artifacts/screenshots/*.png`

## 许可证

本项目采用 Apache License 2.0。
详见 `LICENSE`。
