# Aegis Fabric Protocol 白皮书 v2.0

## 复杂智能体网络的分布式物理防线

> **状态：** 起草中（代码库已于 PR-6c 冻结）
>
> **v1.0 参考：** [AFP Technical Whitepaper (Zenodo)](https://zenodo.org/records/20674352)
>
> **实现对照：** 本仓库 `main` 分支 · [`README.md`](../README.md) · [`ROADMAP.md`](../ROADMAP.md)

---

## 目录

| 章节 | 标题 | 状态 |
|------|------|------|
| 摘要 | Executive Summary — 从 TCP 到 AFP 的范式转移 | 待写 |
| **第 1 章** | [意图危机与信令协议的失效](chapter-01-crisis-of-intent.md) | **草稿 v0.1** |
| 第 2 章 | 物理刹车：边缘执行面的微观控制论 | 待写 |
| 第 3 章 | 双源合并治理模型：企业网格的宏观控制 | 待写 |
| 第 4 章 | 拓扑隔离与协同存续权（CVP） | 待写 |
| 第 5 章 | 生产级演进路线图 | 待写（见 ROADMAP.md） |

---

## 一句话命题

**TCP 治理数据包，AFP 治理优化器。**

在自主 Agent 网络中，灾难不再表现为丢包或 508，而表现为**意图的不可撤销生成**——AFP 将物理红线推进到意图酝酿期，在 Sidecar 带外平面完成预防性熔断。

---

## 读者

- 平台 / Infra 架构师（K8s、Service Mesh、多 Agent 编排）
- Agent 框架维护者（LangGraph、CrewAI、自定义 Planner）
- 信令与协作协议设计者（ASP 及同类应用层协议）
- 金融 / 高密场景合规与 SRE

---

## 版本记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v2.0-draft | 2026 | 与 PR-6c 实现对齐的理论升维 |
| v1.0 | 2025 | Zenodo 首发 · 图论与 Monte Carlo 实证 |
