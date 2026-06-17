# AFP Sidecar（中文说明）

**面向自主 Agent 网络的运行时协调层。**

`AFP Sidecar` 是本仓库的可运行实现：在请求进入本地执行环境前，执行物理层控制（熵压、信誉、递归安全）与策略裁决。

> **"TCP governs packets. AFP governs optimizers."**
> *(TCP 治理数据包，AFP 治理优化器)*

传统互联网基础设施（如 TCP 与 Istio）的设计前提是：网络参与者是外部指令的被动执行者。然而在自主 AI Agent 时代，节点已成为**主动优化器**——它们制定计划、调整策略，并递归地委派任务以实现本地目标。

随着网络演变为意图驱动环境，传统的限流器与服务网格已力不从心。若缺乏控制论层面的约束，不受限制的 Agent 优化必然导致**重试级联、资源滥用、递归委派风暴，以及全局协作崩溃。**

**Aegis Fabric Protocol（AFP）** 引入了*后果持久层（Consequence Persistence Layer, CPL）*。通过将物理约束与自适应摩擦直接嵌入带外（out-of-band）Sidecar，AFP 使网络稳定性源于去中心化的后果执行，而非中心化管控。

### 实证铁证 — 蒙特卡洛压测

我们在高频互动的智能体网格中，对 AFP 控制论机制实施了严格的蒙特卡洛仿真：**1,000 次大数定律聚合** · **500 个自主节点** · **5% 恶意/失控节点**（持续输出递归委托爆炸与高负载外部性）· **100 个物理纪元（Epoch）**。

| 网络 | 100 Epoch 后结果 |
|------|----------------|
| **传统 Baseline**（无 CPL） | 存活节点从 **500 塌缩至 2.05**（约 **0.4%**）。恶意负载榨干共享算力与上下文内存，健康节点卷入 OOM 无限重试潮 — **全面协同坍塌（Coordination Collapse）**。 |
| **AFP 网络** | 健康节点 **500.00** 全部存活（**100% 拓扑存活率**）。预防性熔断与非对称滞后惩罚迅速隔离异常熵压，恶意节点被封入断连子图 — 且不牺牲健康网络吞吐。 |

本地复现：`go run ./cmd/simulator` · 完整流程见下方 [实证验证](#实证验证empirical-proof)。

---

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

## 实证验证（Empirical Proof）

AFP 不仅在[白皮书](https://zenodo.org/records/20674352)中有理论阐述——本仓库还提供**可复现的测试 Demo**，端到端验证 CPL 机制，并导出机器可读证据。

### Demo 验证了什么

| 层级 | 被测机制 | 运行方式 | 预期信号 |
|------|---------|---------|---------|
| **L7 黑盒** | 快路径准入、递归断路、陌生人税 | `docker compose up -d --build` 后执行 `./scripts/verify_http_gateway.sh` | 正常流量 `200` · 深度 > 10 返回 `508` · 不可信陌生人 `403` |
| **治理内核** | 运行模式开关 + Ingress FSM | `./scripts/verify_modes.sh` · `./scripts/verify_recursion_loop.sh` | 企业网格绕过陌生人税；开放交换拦截陌生人；递归深度触发断路 |
| **Monte Carlo** | 恶意负载下 Baseline vs AFP | `go run ./cmd/simulator` | 1,000 次推演 × 500 节点（5% 恶意），AFP 维持更高存活节点数与更低平均负载 |
| **遥测证据包** | Prometheus 指标 + Grafana 面板 | `make demo-report` | 导出 `artifacts/report/`：PromQL JSON、面板截图、HTML 报告 |

### 一键复现

```bash
# 1) 启动双模式 HTTP Gateway 节点
docker compose up -d --build

# 2) 黑盒集成验证（200 / 508 / 403）
./scripts/verify_http_gateway.sh

# 3) 治理内核行为验证
./scripts/verify_modes.sh
./scripts/verify_recursion_loop.sh

# 4) Baseline vs AFP 韧性对比（Monte Carlo）
go run ./cmd/simulator

# 5) 生成可观测性证据包（Grafana + Prometheus）
make demo-report
```

### 已提交的证据产物

`artifacts/` 目录包含一次 Demo 运行的快照：

- `artifacts/report/report.html` — 可读性摘要
- `artifacts/report/prometheus/*.json` — PromQL 原始查询结果（入口拒绝率、快路径吞吐、CVP 惩罚、注入延迟 p95）
- `artifacts/screenshots/*.png` — Grafana 面板导出（下次 `make demo-report` 时生成）

捕获的核心遥测维度：**入口拒绝率**、**快路径吞吐**、**CVP 惩罚事件**、**自适应摩擦（注入延迟 p95）**——即去中心化后果执行的可观测签名。

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

> 完整复现流程见上方 [实证验证](#实证验证empirical-proof)。

```bash
docker compose up -d --build
```

会启动两个节点：
- `afp-mesh-node` · `localhost:8082`（`enterprise-mesh`）
- `afp-open-node` · `localhost:8083`（`open-exchange`）

预期行为：
- `8082` 正常请求返回 `200`
- `8082` 递归深度超限返回 `508`
- `8083` 陌生人低信誉请求返回 `403`

## 自动化验证脚本

> 一键运行全部验证见上方 [实证验证](#实证验证empirical-proof) 中的命令。

- 模式开关验证：`./scripts/verify_modes.sh`
- 递归断路验证：`./scripts/verify_recursion_loop.sh`
- HTTP 黑盒集成验证（针对已运行容器）：`./scripts/verify_http_gateway.sh`
- Protobuf 生成：`./scripts/gen_proto.sh`

## Monte Carlo 推演

> Baseline vs AFP 对比方法见上方 [实证验证](#实证验证empirical-proof)。

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

> 证据包结构与遥测维度见上方 [实证验证](#实证验证empirical-proof)。

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
