# 第 1 章：意图危机与信令协议的失效

> **The Crisis of Intent — Why Application-Layer Signaling Cannot Govern Optimizers**
>
> 白皮书 v2.0 · 草稿 v0.1 · 与 AFP 代码库 PR-6c 对齐

---

## 1.1 开篇命题

互联网五十年的基础设施契约隐含一个假设：**节点是被动的。** TCP 保证字节有序到达；BGP 保证路由可达；API Gateway 保证 HTTP 语义合法。整条栈优化的是*报文*与*请求*，而非*产生报文与请求的那个优化过程本身*。

大语言模型驱动的自主 Agent 打破了这一假设。Agent 不是「发请求的程序」，而是**持续运行的优化器（Optimizer）**——它在内部状态空间中搜索、规划、递归委派、生成子任务，并在多数时间里**从未产生任何可观测的网络 I/O**。

当 Agent「发癫」时，防火墙沉默，网关沉默，可观测性平台沉默。沉默的不是因为系统健康，而是因为**灾难发生在信令层之下、套接字之上**——发生在*意图层*。

AFP 的核心论断由此导出：

> **治理对象必须从 packet 上移至 intent。**
> **治理位置必须从 in-band 应用协议移至 out-of-band 物理执行面。**

---

## 1.2 应用层信令协议的角色边界：以 ASP 为参照

Argent Signaling Protocol（ASP）及同类应用层信令，解决的是**已暴露意图的协商问题**：

- 两个 Agent 如何发现彼此？
- 如何交换能力声明与协作要约？
- 如何在多轮对话中维持会话状态？

这些是**交通灯**问题：当两辆车已经出现在路口，规则决定谁先走。

但交通灯无法回答以下问题：

| 场景 | 信令层能做什么 | 信令层不能做什么 |
|------|---------------|-----------------|
| Planner 陷入 `while True` 递归 | 会话可能仍「合法」 | 无法感知内部图状态爆炸 |
| 一次规划产出 10,000 个 `estimated_tasks` | 无网络流量可拦截 | 无法测量意图爆发 |
| 上下文窗口逼近 OOM 但尚未 HTTP 超时 | ASP 会话仍 ACTIVE | 无法读取 cgroup 压力 |
| LLM 进入「假死」— 进程活着但不产出 | 健康检查可能通过 | 无法区分「慢」与「失控」 |

**结论 1.1（信令的充分性边界）：** 应用层信令是*协作语义*的必要条件，不是*物理安全*的充分条件。在 Agent 架构中，它将**永远滞后于意图生成至少一个调度周期**——而那一个周期，足以烧掉六位数 Token 或拖垮共享集群。

信令协议不是敌人；**把它当作唯一防线才是架构级失误。**

---

## 1.3 三类意图灾难：不是 Bug，是优化器的结构性行为

### 1.3.1 意图爆炸（Intent Burst）

Agent 框架的默认优化方向是**分解**：大任务 → 子任务 → 工具调用链。在缺乏外部摩擦时，这是一个正反馈环：

```text
用户请求 → Planner 拆 50 步 → 每步再拆 → 内部队列指数增长
```

此过程在 Python 堆内存中完成。Prometheus 的 `http_requests_total` 不变；Istio 的 sidecar 无流量可拦；WAF 无签名可匹配。

**意图爆炸是 Agent 的「DDoS 自身」**——攻击面在进程内，不在网络上。

### 1.3.2 递归死锁（Recursive Delegation Loop）

LangGraph、CrewAI 及自研状态机均允许**条件边回指自身**。这是表达力所必需的特性，也是死循环的温床：

```text
planner → should_continue? → planner → should_continue? → …
```

进程不 crash。Pod 不 OOMKill（尚未）。`kubectl get pods` 显示 Running。SRE 仪表盘全绿。

**这是「大模型假死」的工程真相：** 不是模型卡住，而是**控制流在图拓扑中闭合**，而应用层没有任何物理理由停止。

### 1.3.3 上下文雪崩（Context Avalanche）

即使递归深度受控，**上下文单调累积**仍构成独立灾难维度。Agent 的「记忆」是优化器状态的一部分；状态越大，下一步规划越慢、越贵、越不可预测。

传统限流按 QPS；Agent 灾难按 **bytes × recursion × branching factor** 复合增长。

---

## 1.4 带外 Sidecar：为什么必须是「物理」防线

Industry 已尝试三条路径，均在 Agent 场景下暴露结构性不足：

| 路径 | 代表 | 失效模式 |
|------|------|----------|
| **In-band 网关** | API Gateway, Istio | 只见到已发出的 HTTP；意图在进程内已发生 |
| **应用层信令** | ASP, 自定义 RPC | 协商已暴露意图；无法约束未暴露的内部规划 |
| **框架内护栏** | Prompt 规则, 最大迭代次数 | 可被绕过；多 Agent 不一致；升级框架即失效 |

AFP 选择第四条路径：**带外（Out-of-Band）伴生 Sidecar**，在共享 Pod 网络命名空间内，通过 **Unix Domain Socket** 建立微秒级 IPC 走廊——

- 不经过 TCP 协议栈
- 不依赖 Agent 框架的合作
- 不等待网络 I/O 发生

Python SDK 在 LangGraph 节点**执行前**调用 `PreFlightCheck`；Go Sidecar 的 `EvaluatePreFlight` 在同一 ACC/FSM 内核上裁决：**ALLOW · THROTTLE · ISOLATED**。

这不是「又一个 K8s Controller」。Controller 制定法律；**Sidecar 是执法机构**；SDK 是 Agent 与物理世界之间的**最后一道窄门**。

---

## 1.5 从 v1.0 到 v2.0：实证与升维

v1.0 白皮书以 Monte Carlo 方法证明：在 500 节点、5% 恶意、100 Epoch 条件下，Baseline 拓扑存活率约 **0.4%**，AFP 维持 **100%** 协同存续。

v2.0 不再重复证明「AFP 有效」，而解释**为什么必须在此时、以这种方式有效**：

1. **意图前置（Pre-Intent）** — 风控红线必须在 `network()` 之前（第 2 章）
2. **双源治理（Dual-Source）** — 声明式法律 + 亚秒级指令覆盖（第 3 章）
3. **拓扑隔离（Topological Quarantine）** — Gossip + CVP 的涌现式防御（第 4 章）

---

## 1.6 本章结论：裸奔的优化器

LangGraph、CrewAI、AutoGen 的维护者正在构建越来越强大的**意图生成引擎**。ASP 等信令协议的维护者正在构建越来越精细的**意图交换语法**。

但若没有 AFP 所代表的**物理执行面约束**，整个栈在一个致命问题上裸奔：

> **Who governs the optimizer before it optimizes?**

TCP 不回答这个问题。HTTP 不回答。ASP 不回答。

**AFP 回答。**

---

## 1.7 与实现的映射（供工程读者）

| 本章概念 | 代码锚点 |
|----------|----------|
| 意图前置探询 | `sdk/python/afp_sdk` · `PreFlightCheck` · `@afp_governed_node` |
| 递归死锁拦截 | `internal/dataplane/preflight.go` · `maxRecursionDepth` |
| 带外 Sidecar | `deploy/kubernetes/agent-pod-demo.yaml` · UDS `emptyDir` |
| 优雅降级 | `on_quota_exceeded="annotate"` · `afp_blocked` 状态 |
| 买家秀日志 | `README.md` · `annotated-stop: recursion depth exceeded` |

---

## 1.8 下一章预告

第 2 章将进入**边缘执行面的微观控制论**：EntropyMonitor 如何将 context、并发、内存压力建模为可比较的「熵压」标量；ACC 内核如何在微秒级完成 FSM 裁决；以及为何 UDS IPC 是 Agent 场景下唯一合理的执法延迟量级。

---

*本章为 v2.0 草稿 v0.1，欢迎架构评审与学术润色。*
