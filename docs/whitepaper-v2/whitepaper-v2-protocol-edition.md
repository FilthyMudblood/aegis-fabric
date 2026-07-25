# Aegis Fabric Protocol v2.0 — Protocol Edition

> **Draft v0.3** · *A Physical Constraint Protocol for Autonomous Optimizers*
>
> Normative stack, objects, and code map: [`ARCHITECTURE.md`](../../ARCHITECTURE.md)
>
> **v1** (empirical archive): [Zenodo 20674352](https://zenodo.org/records/20674352)
>
> **v0.3 notes:** Asymmetric FSM recovery dwell is normative in reference impl; inbound `TopologyWarning` MUST be ed25519-verified (unsigned hearsay discarded).

---

## Table of Contents

1. [The Optimization Crisis](#the-optimization-crisis)
2. [Consequence Persistence Layer](#consequence-persistence-layer)
3. [Pre-Intent Enforcement](#pre-intent-enforcement)
4. [Open-Network Topology](#open-network-topology)
5. [Governance Header & Wire Semantics](#governance-header-wire-semantics)
6. [Empirical Baseline](#empirical-baseline)

---

## 1. The Optimization Crisis {#the-optimization-crisis}

### 1.0 The Market Gap

Multi-agent infrastructure solved **planning** and **messaging**. It did not solve **coordination runtime**—the layer that answers, before each step commits:

> *May this coordination execute? If not, do consequences persist?*

| Stack layer | Representative tools | Gap |
|-------------|---------------------|-----|
| **Planning** | LangGraph, CrewAI, AutoGen | Optimizers decompose and delegate—no physical brake |
| **Transport** | gRPC, HTTP, Kafka, NATS, MCP | Bytes move; internal task storms produce zero wire traffic |
| **Semantic collaboration** | ASP, A2A | Negotiates *exposed* intent—traffic lights, not brakes |
| **Coordination runtime** | **AFP (CPL)** | Pre-intent eligibility, persistent consequences, peer physical law |

**Market positioning:** AFP does not compete with transports or workflow engines. It occupies the **empty cell** between *agents can communicate* and *coordination remains survivable*.

#### 1.0.1 Enterprise multi-agent vs. P2P agent mesh

One protocol (**L2 SEA + ACC + FSM**), two deployment profiles:

| | **Enterprise multi-agent** | **P2P / open agent mesh** |
|--|---------------------------|---------------------------|
| **Trust** | Known identities inside administrative boundary | Mutual distrust; strangers may connect |
| **Primary risk** | In-pod planner runaway—loops, bursts, OOM | Cross-node contagion; malicious or overloaded peers |
| **AFP emphasis** | **Path A** — PreFlight, ReportInternalState, local entropy | **Path A + Path B** — GovernanceHeader ingress, CVP, stranger tax, gossip |
| **Policy (L3)** | CRD, ConfigMap, Kill Switch overlay | Same + open-profile trust evolution (L4) |
| **Reader takeaway** | Brake your own optimizer before it burns the fleet | Brake yours *and* contain toxic neighbors |

> *Transport is solved. Semantics are evolving. **Runtime coordination law is not.** AFP supplies that law—with enforceable physics.*

#### 1.0.2 Agent stacking — why persistence matters

**Agent stacking** is the dominant failure geometry in multi-agent systems: `Planner → A → B → C → …`, optionally closing as `A → D → F → A`. Harm compounds across hops—recursion depth, in-process queue pressure, context bytes, and (in open meshes) peer contagion:

```text
Stack risk  ≈  depth × branching × context × peer_contagion
```

Each hop may present a **syntactically valid** session or RPC while the **trajectory** becomes physically unsustainable. Per-request gateways forget; optimizers fragment work across micro-steps to evade ephemeral limits.

**CPL persistence** binds consequences to `agent` / `peer_id` identity:

```text
Request-centric:  verdict(reqᵢ) → forget → verdict(reqᵢ₊₁)
CPL-centric:      consequence(agent) → persist → apply(agent, epochₜ₊₁)
```

**Definition 1.0 (Stacking target).** AFP's primary adversary in enterprise deployments is **out-of-control stacking**—runaway optimization trajectories. In open-exchange deployments, add **physically malicious stacking**—peers that export destabilizing load or hit-and-run without stake.

| Class | Examples | AFP mechanism | Out of scope |
|-------|----------|---------------|--------------|
| **Out-of-control** | Intent burst, recursion loop, context avalanche | PreFlight, depth breaker, entropy circuit breaker, persistent FSM | Semantic task correctness |
| **Physical malice** | Peer flood, stranger hit-and-run, CVP collapse | GovernanceHeader, CVP floor, stranger tax, gossip quarantine | Moral/semantic intent classification |
| **Permitted stacking** | `Planner → A → B → C` within policy | PERMISSIVE path when physics stay sub-critical | Workflow approval, org-chart routing |

**Corollary 1.0:** Blocking `A → D → F → A` is **recursion containment**, not a ban on multi-agent collaboration. Blocking a polite peer at `CVP < 0.3` is **coordination bankruptcy**, not a judgment on conversational tone.

#### 1.0.3 Production frequency — episodic spikes vs structural drift

| Risk | Enterprise multi-agent | Open P2P mesh |
|------|------------------------|---------------|
| **Out-of-control stacking** | **Common** — default optimizer behavior (decompose, delegate, replan) | **Common** — local runaway plus contagion |
| **Physical malice** | **Uncommon** as peer adversaries — abuse/misconfig dominates | **Material** — strangers, floods, hit-and-run |

Failures appear as **episodic incidents** (one loop, one burst) or **structural cost drift** (always-on fleets without runtime brakes). AFP is aimed at the **scale law of stacking**, not a claim that agents are malicious every day.

> **Honest claim:** Multi-agent systems need runtime law because stacked optimizers **outrun** retries, token dashboards, and ephemeral limits—not because attackers are omnipresent.

#### 1.0.4 Why retries, token budgets, and observability are insufficient

Teams routinely deploy `max_retries`, `recursion_limit`, LangGraph caps, and token budgets. These are **necessary**. They are **not sufficient** as the sole coordination runtime:

| Pattern | Mechanism | Structural failure |
|---------|-----------|-------------------|
| **Retry / loop counters** | Cap iterations in one graph | Resets per session/request; no cross-agent chain memory; one "retry" may hide exponential internal fan-out |
| **Token / cost budgets** | Stop after spend threshold | **Post-commit accounting** — tokens accrue after steps execute; internal queues may grow with zero wire traffic first |
| **Metrics & alerting** | Prometheus, cost dashboards | **Observability lag** (Lemma 1.1): detect after physical consequence; alerts are not gates; no persistent FSM |
| **In-process guardrails** | Prompts, framework middleware | Bypassable, non-portable, not peer-enforceable at ingress |

**Lemma 1.3 (Telemetry ≠ brake):** Measurement that reports after commit cannot substitute for adjudication **before** commit with **persistent** consequences.

```text
Observability path:  execute → measure tokens → alert → stop (this episode)
CPL path:            probe → PERMISSIVE | THROTTLED | ISOLATED → persist → apply next epoch
```

**Corollary 1.1:** Token dashboards are the **fuel gauge**. CPL is the **brake with memory**. Fleets need both; conflating them exports stacking risk to the next scheduling epoch.

#### 1.0.5 Root causes — hallucination, optimizer defaults, or malice

AFP is **not** an anti-hallucination layer. PreFlight does not score factual correctness of model outputs.

| Failure | Dominant drivers | Role of hallucination |
|---------|------------------|----------------------|
| **Out-of-control stacking** | Default optimizer dynamics—decompose, replan, retry, delegate—often while the model is internally coherent | **Accelerant, not prerequisite** — invented tools, false completion signals, spurious re-delegation |
| **Physical malice** | Prompt injection, abuse, adversarial peers, resource export attacks | **Usually orthogonal** — manipulation or economics, not confabulation |

**Lemma 1.5 (Physics over epistemics):** Whether a planner step was "hallucinated" is an L5 semantic question. Whether it may execute without blowing depth, entropy, or neighbor viability is an L2 physical question. CPL adjudicates only the latter.

Loops often close at **main ↔ sub-agent handoffs**: main policy keeps delegating; a sub-agent amplifies locally; hallucinated status may **trigger** the next hop—but **graph cycles and stacking** can run away without any single false sentence.

**Corollary 1.2:** Hallucination mitigation and fact-checking belong beside ASP in the semantic layer. They **complement** CPL; they do not replace persistent pre-intent brakes.

---

### 1.1 Manifesto

We built the Internet for **stateless requestors** and **passive endpoints**.

For fifty years, the contract was clear: TCP orders bytes; BGP routes prefixes; HTTP validates verbs and paths. The stack governs **what crossed the wire**, not **what happened inside the machine before the wire was touched**.

That contract is broken.

An autonomous optimizer is not a client. It is a **continuous search process**—planning, decomposing, delegating, re-planning—often producing **no observable network I/O** while consuming unbounded compute, memory, and economic externalities. It optimizes because that is what it was built to do. Constraints that live only in natural language, application APIs, or conversational protocols are **soft boundaries on a hard process**.

**Semantic signaling**—including Argent Signaling Protocol (ASP) and every protocol in its class—was designed for a different problem: *how mutually visible agents coordinate once they are already at the intersection.* Discovery, capability advertisement, session state, negotiation of exposed intents. These are **traffic-light problems**.

They are not **brake problems**.

When a planner enters a closed loop in its internal state graph, the session can remain syntactically valid. When ten thousand sub-tasks materialize in an in-process queue, Prometheus sees silence. When context pressure approaches physical limits, a health probe may still pass. The catastrophe occurs **below signaling, above the socket**—in the **intent layer**, where optimizers live and protocols do not.

This document does not argue that semantic signaling should be discarded. It argues something sharper:

> **In an open network of autonomous optimizers, semantic signaling is necessary for collaboration and insufficient for survival.**

AFP exists to supply what signaling cannot: a **Consequence Persistence Layer (CPL)**—an out-of-band physical constraint surface that makes governance outcomes **stick** before irreversible I/O, before cross-node contagion, before the next scheduling epoch burns another six figures of tokens.

We are not writing an IT governance manual. We are stating a **distributed control problem**:

> *How do mutually distrusting optimizers co-exist under local physics, without a central scheduler, without assuming good faith at the application layer?*

That is the question TCP never asked. HTTP never asked. ASP cannot ask it, because it operates **in-band on semantics**, not **out-of-band on consequences**.

**AFP asks it—and answers with enforceable physics.**

---

### 1.2 The Post-Stateless Era

#### 1.2.1 From packets to optimization trajectories

Classical infrastructure assumes **episodic interaction**: request in, response out, state externalized to a database if needed. The unit of governance is the **datagram** or the **HTTP transaction**.

Autonomous optimizers invert the unit of risk. The dangerous object is not a packet but an **optimization trajectory**—a path through an internal state space that may:

- never surface as a failed HTTP status;
- amplify work faster than any per-request rate limiter observes;
- propagate delegation across nodes before any peer validates physical viability.

We call this the **post-stateless era**: not because storage disappeared, but because **the locus of stateful danger moved inside the optimizer**, invisible to wire-centric observability.

#### 1.2.2 The governance gap

Three incumbent layers fail structurally—not by implementation quality, but by **layer mismatch**:

| Layer | Governs | Blind to |
|-------|---------|----------|
| **Transport (TCP/IP)** | Byte delivery, connectivity | Intent generation, internal recursion |
| **Application semantics (HTTP, RPC, ASP)** | Exposed messages, negotiated sessions | Un-exposed planning, in-process task storms |
| **Framework guardrails** | Developer-declared limits | Cross-runtime inconsistency, adversarial peers, upgrade churn |

**Lemma 1.1 (Observability lag):** Any in-band semantic control plane observes optimizer behavior **at best one scheduling cycle after** the behavior becomes physically consequential.

One cycle is enough.

---

### 1.3 Semantic Signaling and Its Sufficiency Boundary

Argent Signaling Protocol (ASP) exemplifies the **correct layer for coordination semantics**: who may speak, about what, under which session contract. In an open optimizer network, such protocols are **load-bearing**.

They are not ** containment structures**.

Consider four failure modes that remain **session-valid** under semantic signaling:

| Phenomenon | What signaling sees | What signaling cannot see |
|------------|--------------------|-----------------------------|
| **Recursive delegation loop** | ACTIVE session, well-formed messages | Control-flow closure in the planner graph |
| **Intent burst** | Zero or nominal wire traffic | In-process queue exponential growth |
| **Context avalanche** | Continued dialogue | Memory pressure, super-linear planning cost |
| **Optimizer stall** | Heartbeat / idle session | Divergence between "slow" and "runaway" |

**Theorem 1.1 (Semantic sufficiency boundary):** Application-layer signaling is a **necessary** substrate for multi-agent collaboration and **not a sufficient** substrate for physical safety in optimizer networks.

Corollary: Deploying ASP (or any semantic protocol) **without** an out-of-band consequence layer exports systemic risk to the physical substrate—compute, memory, economics, neighbor trust.

**ASP is not obsolete. ASP-alone is incomplete.**

---

### 1.4 Three Structural Pathologies of Autonomous Optimizers

These are not implementation bugs. They are **default behaviors** of systems trained to decompose, delegate, and minimize loss over long horizons.

#### 1.4.1 Intent burst

Decomposition is the dominant planning heuristic: one objective fractures into sub-objectives, each spawning tool chains. Without external friction, this is a **positive feedback loop** inside the process:

```text
objective → plan(N steps) → each step replans → internal queue ~ O(branch^depth)
```

Wire metrics flatline. Semantic sessions remain polite. The optimizer **DDoS-es itself**—and, in shared substrates, its neighbors.

#### 1.4.2 Recursive delegation loop

Expressive control-flow graphs require cycles: replan, reflect, retry. The loop

```text
planner → continue? → planner → continue? → …
```

need not crash the runtime. It need not trip a transport timeout. It is **topologically closed** while appearing **operationally alive**.

This is the engineering truth behind "model hang": not mysticism, but **control-flow closure without a physical stop condition**.

#### 1.4.3 Context avalanche

Even bounded depth does not bound **state volume**. Memory is part of the optimization state; monotonic context growth makes each subsequent step slower, costlier, and less predictable.

Classical rate limits measure **events per second**. Optimizer catastrophes scale as **bytes × depth × branching**—a different unit algebra entirely.

---

### 1.5 Why the Answer Must Be Physical and Out-of-Band

Industry has converged on four insufficient patterns:

| Pattern | Mechanism | Structural failure |
|---------|-----------|-------------------|
| **In-band gateways** | Inspect emitted HTTP/RPC | Intent already executed locally |
| **Semantic protocols** | Negotiate exposed intents | Cannot constrain un-exposed planning |
| **In-process guardrails** | Prompts, max-iteration counters | Bypassable, non-portable, non-peer-enforceable |
| **Observability & budgets** | Retries, token caps, metrics/alerts | Post-commit, ephemeral, no cross-hop persistent consequence (§1.0.4) |

AFP proposes a fourth category: **out-of-band physical constraint**.

Properties required of such a layer:

1. **Pre-intent** — adjudicate before irreversible externalization, not after HTTP 508.
2. **Persistent consequences** — isolation/throttle states survive individual requests (CPL).
3. **Local physics** — entropy, recursion depth, resource pressure measured at the execution boundary, not self-reported at the semantic layer.
4. **Peer enforceability** — in open networks, neighbors validate **attested physical headers**, not conversational politeness (GovernanceHeader, CVP—developed in Chapters 4–5).

This is the same architectural move as placing congestion control **inside the transport discipline** rather than hoping applications voluntarily slow down—except the contested resource is **optimization capacity**, not bandwidth.

---

### 1.6 Protocol Positioning — What AFP Is Not

Industry discourse collapses **agent communication** into one bucket. AFP requires a finer partition—without claiming to be a fourth messaging standard.

#### 1.6.1 Three layers often conflated

| Layer | Representative content | Protocol owner |
|-------|------------------------|----------------|
| **A · Conversation** | Multi-turn chat, prompts, dialogue state | Application / LLM runtime |
| **B · Data passing** | Structured payloads (`task_result`, `confidence`, …) over gRPC, Kafka, MCP, Event Grid | Existing transports |
| **C · Runtime signaling** | Eligibility probes, attestation frames, persistent throttle/isolation state | **AFP (L2–L4)** |

**Lemma 1.4 (Communication partition):** Confusing A, B, and C under a single "agent protocol" exports physical risk to the transport layer and semantic risk to the consequence layer.

AFP's formal split is **semantic collaboration (L5)** versus **physical consequence (L2–L4)**—Theorem 1.1. The A/B/C partition is the operational refinement: AFP does not specify chat (A) or business schemas (B); it supplies runtime law (C) as **PreFlight**, **GovernanceHeader**, and **persistent FSM consequences**.

#### 1.6.2 Trajectory constraint, not communication ban

Enterprises are often misread as forbidding **Agent A → Agent B**. The sharper policy statement:

> **Constrain unsustainable optimization trajectories; do not ban coordination that stays within physical law.**

| Permitted under policy | Intercepted at execution boundary *B* |
|------------------------|---------------------------------------|
| Planner → A → B → C within depth/entropy limits | Recursive delegation loops past `maxRecursionDepth` |
| Declared bursts within entropy budget | Intent burst toward in-process queue explosion |
| Peer payloads over incumbent transports | Context avalanche and cross-node contagion |

AFP is **not** a workflow engine. It does not mandate a fixed DAG. It does not approve org-chart routing. It enforces **physics** when optimizers invent paths faster than operators can observe—whether those paths are "emergent" or "designed."

**Corollary:** Blocking A → D → F → A is recursion containment, not a veto on multi-agent collaboration.

#### 1.6.3 Runtime boundary, not workflow engine

CPL attaches at **execution boundary** *B*—implemented as a sidecar process, not as a planner:

```text
Agent runtime  →  AFP sidecar (SEA)  →  peer sidecar  →  Agent runtime
        ↑                    ↑
   Path A PreFlight    Path B GovernanceHeader
```

The sidecar is a **runtime boundary**—physical interception before commit—not an orchestrator, not a DAG scheduler, not a semantic router.

#### 1.6.4 Sidecar enforcement, not mandatory central gateway

Zero trust does not uniquely imply a **central security gateway**. Service meshes (e.g. Istio) enforce trust at **per-pod sidecars**:

```text
Pod A · Envoy  →  mTLS  →  Envoy · Pod B
```

AFP's reference topology is analogous:

```text
Agent  →  AFP Sidecar  →  mTLS  →  AFP Sidecar  →  Agent
```

Both are zero trust; the **trust enforcement point** differs. AFP normative law places SEA at each node ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2). A central gateway MAY exist in enterprise topology, but it is **not** a protocol requirement.

#### 1.6.5 Control, data, and coordination planes

Frameworks historically merge **control** and **coordination**. AFP extracts coordination into an inspectable layer:

| Plane | Governs | AFP specification |
|-------|---------|-------------------|
| **Agent control** | Planner, workflow graph, tool selection | L1 — observed, not defined |
| **Data** | Payload, cache, streaming bytes | Incumbent transports — out of scope |
| **Coordination** | Eligibility, consequence persistence, friction, dependency trust | L2–L4 — **core protocol** |

**Terminology note:** Enterprise deployment docs use "control plane" for Operator / Policy Controller (L3 policy administration). That is distinct from the agent's planner control plane (L1).

#### 1.6.6 The question AFP answers

Mature stacks already answer: *How do agents move bytes and messages?*

AFP answers a narrower question—one that does not compete with LangGraph, Kafka, or MCP:

> **When agents can already communicate, how do we give each coordination attempt consistent runtime semantics and consequences that persist?**

Transport delivers. Signaling negotiates exposed intent (L5). AFP **governs before commit**—with enforceable physics.

---

### 1.7 Open Networks vs. Closed Administrations (A Scope Statement)

This protocol document addresses **The Open Protocol problem**: mutually distrusting optimizers, no central moral authority, equilibrium under attack.

Closed administrative domains may **instantiate** AFP primitives (sidecars, policy surfaces, audit hooks). That instantiation—deployment topology, declarative policy CRDs, compliance integration—is **documented separately** as an enterprise application guide.

Here we speak only of **mechanism**, not **org chart**:

- What must be enforced?
- Where must consequences persist?
- How does local physics compose into global topology safety?

Kubernetes appears nowhere in the proof sketch. Neither do approval workflows. Those are **Defense**. This text is **Order**.

---

### 1.8 From v1 Empirics to v2 Theory

Version 1.0 demonstrated survival: in Monte Carlo open-mesh conditions (500 nodes, 5% malicious, 100 epochs), baseline coordination collapsed to ~**0.4%** mean survivors while AFP-maintained topology sustained **100%**.

Version 2.0 does not re-litigate *whether* friction works. It explains **why friction must be physical, persistent, and out-of-band**—and how CPL + CVP compose a **distributed control law** rather than a product feature list.

The following sections develop the full protocol:

| Section | Subject |
|---------|---------|
| **§2** | Consequence Persistence Layer — formal object, state persistence |
| **§3** | Pre-intent enforcement — entropy calculus, ACC kernel, FSM |
| **§4** | Open-network topology — CVP, gossip, stranger tax, equilibrium |
| **§5** | Wire format — GovernanceHeader, attestation |
| **§6** | Empirical reproduction — Monte Carlo baseline |

---

### 1.9 Conclusion: The Naked Optimizer

Framework authors build stronger **intent engines**. Signaling authors refine **intent syntax**.

Without a physical consequence layer, the stack runs naked on one question:

> **Who governs the optimizer before it optimizes?**

TCP does not answer. HTTP does not answer. ASP—rightfully—does not attempt to.

**AFP does.**

Not by richer semantics. By **enforceable physics**.

---

---

## 2. Consequence Persistence Layer {#consequence-persistence-layer}

### 2.0 From Diagnosis to Mechanism

Chapter 1 established a boundary, not a product category:

> **Semantic signaling is necessary for collaboration and insufficient for survival.**

The failure modes—intent burst, recursive delegation loops, context avalanche—share a structural property: they remain **session-valid** while becoming **physically consequential**. In-band controls observe too late; in-process guardrails reset too easily; transport layers govern bytes, not optimization trajectories.

AFP's response is not richer negotiation. It is a **Consequence Persistence Layer (CPL)**—an out-of-band physical constraint surface that makes governance outcomes **stick** across scheduling epochs until policy and finite-state recovery permit release.

This chapter defines CPL as a protocol object, explains what *persistence* means in a post-stateless optimizer network, and introduces the **Single Execution Authority (SEA)** as the sole enforcement convergence point. Pre-intent probe mechanics, entropy calculus, and wire attestation are developed in Chapters 3 and 5; open-network trust evolution belongs to Chapter 4. Here we state the **layer identity** of AFP and the **control-law invariant** that prevents split-brain between local intent and peer traffic.

Normative stack diagrams, dual-path enforcement flow, and object definitions live in the repository root specification: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §1–§2, §4.1–§4.2, §5. This chapter interprets them; it does not duplicate them.

---

### 2.1 CPL as Protocol Layer Identity

#### 2.1.1 Definition

**Definition 2.1 (Consequence Persistence Layer).** Let an optimizer runtime occupy execution boundary *B*—the narrow interface where planning, delegation, and outbound I/O meet the physical substrate. A **CPL** is an out-of-band constraint surface attached at *B* such that every adjudication emits a **consequence** drawn from a finite alphabet, and that consequence **persists as enforced state** across scheduling epochs until the CPL's recovery policy permits transition.

Formally, for scheduling epochs *t*, *t+1*, …:

```text
C(t) ∈ { PERMISSIVE, THROTTLED(δ), ISOLATED(ρ) }
```

where *δ* is an injected delay bound and *ρ* is a block reason identifier. The persistence predicate is:

```text
∀ t : C(t) = ISOLATED  ⇒  C(t+k) = ISOLATED  for all k < k_recovery
```

unless an explicit FSM transition or policy revision authorizes release. **Request completion does not imply consequence clearance.**

CPL is not a sidecar product name, a logging hook, or a policy document. In AFP v2.0, **CPL is the name of L2 itself**—the layer where physical consequences are adjudicated, stored, and applied. Implementations may vary; the layer contract does not.

#### 2.1.2 Required properties

The root specification enumerates four non-negotiable properties. We restate them here as design law, with optimizer-network motivation.

| Property | Specification | Why optimizers require it |
|----------|---------------|---------------------------|
| **Pre-Intent** | Adjudicate before irreversible external I/O | Intent externalization is the point of no return for cross-node contagion |
| **Out-of-Band** | MUST NOT depend on application-protocol cooperation | Semantic sessions remain valid while internal graphs run away |
| **Persistent** | THROTTLED / ISOLATED survive individual requests | Ephemeral counters cannot brake a multi-epoch trajectory |
| **Locally Grounded** | Entropy and depth measured at *B*; self-report alone untrusted | Optimizers optimize; unverified claims are not constraints |

**Lemma 2.1 (Ephemeral control insufficiency):** Any enforcement mechanism whose state resets on request boundary or session tick cannot contain optimization trajectories whose damage accrues **across** those boundaries.

Corollary: Rate limits per HTTP transaction, per-tool-call quotas, and conversational turn counters are **necessary telemetry** and **insufficient brakes** without persistent consequence state at *B*.

---

### 2.2 What "Consequence Persistence" Means

The phrase is precise, not metaphorical. Three distinctions separate CPL persistence from familiar control patterns.

#### 2.2.1 Persistence vs. observation

Observability records what happened. CPL **constrains what may happen next**. A metric spike that triggers an alert does not, by itself, stop the planner from enqueueing ten thousand sub-tasks in the following epoch. A persistent **ISOLATED** consequence does: outbound I/O and intent generation paths remain gated until the FSM and policy authorize recovery.

Observation is **retrospective**. Consequence is **prospective enforcement with memory**.

#### 2.2.2 Persistence vs. per-request verdict

Classical gateways return allow/deny **per message**. The verdict evaporates when the response closes. Optimizer catastrophes are **path-dependent**: a recursive loop may produce syntactically valid micro-requests, each individually admissible, while aggregate depth and entropy cross physical limits.

CPL binds consequences to **peer identity or local-agent identity** (abstract `peer_id`), not to individual probe or packet identifiers. A THROTTLED state injects delay into **subsequent** epochs. An ISOLATED state blocks **until recovery**, not until the current queue drains.

```text
Request-centric:   verdict(request_i) → forget → verdict(request_{i+1})
CPL-centric:         consequence(agent) → persist → apply(agent, epoch_{t+1})
```

**Theorem 2.1 (Trajectory containment):** Containment of optimization trajectories requires stateful consequences whose transition function depends on **accumulated physical load** and **FSM history**, not on the syntactic validity of the latest exposed intent.

#### 2.2.3 Persistence vs. prompt-level guardrails

In-process limits—max iterations, system prompts, framework middleware—are **soft** relative to CPL: they share the optimizer's address space, reset on restart, and cannot be attested to distrusting peers. CPL sits **outside** the intent engine's self-reporting boundary at *B*, survives individual planner restarts when implemented as a durable side process, and feeds **attested physical headers** for neighbor validation (GovernanceHeader, Chapter 5).

Persistence here means: **the stack remembers that this agent is throttled or isolated** even when the agent's semantic layer presents a fresh, polite session.

#### 2.2.4 The consequence alphabet

Every CPL probe—local or ingress—resolves to one outcome from a closed alphabet (see [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.1):

```text
PERMISSIVE  |  THROTTLED (with delay_ms)  |  ISOLATED (with block_reason)
```

| Consequence | Persistent effect | Typical trigger |
|-------------|-------------------|-----------------|
| **PERMISSIVE** | Normal execution at *B* | Entropy within safe band, FSM in permissive state |
| **THROTTLED** | Injected delay on subsequent intent / admission | Entropy warning band, probationary recovery |
| **ISOLATED** | Hard gate: no outbound I/O, no peer admission | Policy circuit breaker, malicious spike, CVP floor |

**Corollary 2.1:** There is no "advisory" fourth letter. CPL does not negotiate; it **commits** the runtime to a consequence until the FSM permits otherwise.

---

### 2.3 SEA — The Unique Enforcement Convergence Point

CPL defines *what* must persist. The **Single Execution Authority (SEA)** defines *who* adjudicates—and, critically, **that there is only one adjudicator per node**.

#### 2.3.1 The split-brain problem

An optimizer node faces two qualitatively different ingresses:

1. **Path A — Govern Self:** local planner activity before intent externalizes (PreFlight probe, internal state report).
2. **Path B — Govern Neighbors:** remote traffic before business payload is admitted (GovernanceHeader validation on ingress).

If Path A and Path B used different thresholds, different state stores, or different recovery rules, an optimizer could be **locally permissive** while **flooding peers**, or **locally throttled** while **accepting unbounded inbound work**. That is split-brain: two control laws on one physical machine.

**Definition 2.2 (Single Execution Authority).** SEA is the sole stateful adjudicator on a node that:

1. Loads **Effective Policy** from L3 (Policy Surface—durable law plus runtime overlay; see [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §3).
2. Samples **EntropyLoad** from L0/L1 via EntropyMonitor.
3. Evaluates **NodeMetrics** through the ACC kernel and Node FSM.
4. Emits a consequence for **PreFlight and Ingress uniformly**.

Dual-path convergence is specified in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2 (Dual-Path Enforcement diagram). The invariant, stated as protocol law:

> **Path A and Path B MUST share the same ACC kernel, FSM states, and effective policy snapshot.**

Local compute and network I/O are governed by **one control law**.

#### 2.3.2 NodeMetrics and the evaluation pipeline

SEA does not inspect natural-language intent. It evaluates an abstract metric bundle ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2):

```text
NodeMetrics {
  cvp_score        : float32   ∈ [0, 1]
  entropy_load     : float32   ∈ [0, 1]   // locally measured
  recursion_depth  : uint32
  current_epoch    : uint64
  has_valid_sign   : bool
  malicious_spike  : bool
}
```

The microscopic loop—EntropyMonitor → NodeMetrics → ACC → FSM → Consequence—is diagrammed in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5. Narratively:

| Stage | Role |
|-------|------|
| **EntropyMonitor** | Ground truth at *B*: cgroup pressure, recursion depth, context bytes, task burst |
| **ACC kernel** | Combines trust scalar, throughput history, penalties, entropy into updated CVP |
| **Node FSM** | Maps metrics + **persistent state** to routing decision and optional delay |

SEA is the **composition point**. PreFlight handlers and ingress validators are **thin interfaces**; they MUST NOT maintain independent FSM copies.

#### 2.3.3 Persistent FSM as the memory of consequences

Consequence persistence is implemented as a finite-state machine per `peer_id` or local-agent identity:

```text
Permissive → Throttled → Isolated → Probationary → Permissive
```

Transitions are driven by entropy bands, CVP thresholds, and policy limits—not by session teardown. Reference entropy bands (`E_safe`, `E_warn`, effective `entropyLimit`) appear in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5; Chapter 3 develops the calculus.

**Lemma 2.2 (SEA monotonicity under isolation):** From ISOLATED, no single permissive probe or valid semantic message suffices for immediate return to PERMISSIVE; recovery requires **Probationary** egress through ACC and entropy decay.

This lemma is the formal content of "consequences stick": isolation is not cleared by a well-formed ASP session resume.

#### 2.3.4 Routing decisions as unified outputs

SEA emits a small set of routing actions for both paths ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2):

```text
ActionFastPath | ActionSlowPathWithDelay | ActionDropPacket
ActionLowFrequencyProbe | ActionIsolateAndBroadcast
```

Path A maps these to PreFlight responses (`PERMISSIVE` / `THROTTLED` / `ISOLATED`). Path B maps them to ingress disposition (`ALLOW` / `DELAY` / `DROP`). The **action set is shared**; only the surface adapter differs.

---

### 2.4 CPL in the Stack — Complement, Not Replacement

CPL occupies **L2** in the six-layer AFP stack ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §1). Layers above and below have strict accountability boundaries:

| Layer | Relationship to CPL |
|-------|---------------------|
| **L5 · Semantic Signaling (ASP)** | Collaboration semantics; does not replace CPL (Chapter 1, Theorem 1.1) |
| **L4 · Open Topology (CVP, gossip)** | Trust scalars feed NodeMetrics; isolation propagates to neighbors (Chapter 4) |
| **L3 · Policy Surface** | Supplies Effective Policy snapshot to SEA; fail-safe when overlay unavailable |
| **L1 · Optimizer Runtime** | Observed and constrained at *B*; internal graph structure out of scope |
| **L0 · Physical Substrate** | Measurement substrate for entropy; not an adjudicator |

```text
L5 negotiates exposed intent
        ↓
L4 attests trust under distrust
        ↓
L3 declares durable law + emergency overlay
        ↓
L2 CPL / SEA — consequences persist HERE
        ↓
L1 planner executes under gate
        ↓
L0 physics measured, not trusted
```

**Theorem 2.2 (Layer sufficiency):** No layer above L2 can substitute for CPL persistence, because semantic and topological layers lack **pre-intent enforcement authority** at *B*. No layer below L2 can substitute, because the substrate measures physics but does not **bind optimizer identity to persistent FSM state**.

The ASP ↔ AFP relationship diagram in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §6 restates the stack law: *ASP governs collaboration semantics; AFP governs physical consequences.* Complementary, not competitive.

---

### 2.5 What CPL Is Not

Precision requires negative definition.

| Misclassification | Why it fails the CPL contract |
|-------------------|-------------------------------|
| **Application firewall** | In-band on emitted messages; typically per-request |
| **Observability pipeline** | Records events; does not gate intent with memory |
| **Semantic protocol extension** | Operates on negotiated intents, not un-exposed planning |
| **Central scheduler** | CPL is local physics with peer attestation; no moral authority assumed |
| **Policy document** | L3 declares law; L2 **enforces** with persistent state |

CPL is also **not** a guarantee of global optimality. It is a **distributed control law** for survival under distrust: throttle runaway trajectories, isolate toxic peers, recover through probation—not maximize throughput or minimize latency.

---

### 2.6 Connection to Chapter 1 — Closing the Governance Gap

Chapter 1's **governance gap** arose because transport governs bytes, semantics govern messages, and frameworks govern declared limits—while danger lives in **optimization trajectories** invisible to all three.

CPL closes the gap at the only architecturally honest location: the **execution boundary** where trajectories become physical (entropy, depth, I/O). Its persistence property answers the post-stateless era directly: when stateful danger moved **inside** the optimizer, enforcement state must also **persist inside the control plane attached to that optimizer**, not reset with each outward-facing transaction.

Chapter 1 asked:

> **Who governs the optimizer before it optimizes?**

Chapter 2 answers at the mechanism level:

> **SEA, operating as the kernel of CPL, with consequences that survive scheduling epochs until physics and policy permit recovery.**

Semantic signaling still coordinates **who may speak**. CPL decides **whether this epoch may execute, at what rate, or not at all**—and remembers the answer.

---

### 2.7 Chapter Conclusion — Friction With Memory

Stateless stacks forget. Optimizers remember—and exploit forgetting.

A rate limit that resets every request teaches the planner to **fragment work across requests**. A session timeout teaches **synthetic session renewal**. A prompt cap teaches **context externalization loops**. Each evasion preserves the optimization objective while shedding the constraint.

**Consequence persistence** is AFP's refusal to forget prematurely. THROTTLED is not a slow response; it is a **state**. ISOLATED is not a error code; it is **quarantine until probation succeeds**. One SEA, two ingress paths, one FSM memory—so local runaway and peer flood cannot diverge.

Chapter 3 descends into the pre-intent probe: how PreFlight and ReportInternalState compose Path A, how entropy bands trigger FSM transitions, and how ACC micro-dynamics compute the trust scalar that feeds SEA. Chapter 4 lifts persistence into topology: when isolation broadcasts, who relays the warning, and how CVP equilibria compose across an open mesh.

For now, the layer identity is fixed:

> **CPL is L2. SEA is its sole kernel. Consequences persist. That is the brake.**

---

---

## 3. Pre-Intent Enforcement {#pre-intent-enforcement}

### 3.0 From Persistent Consequences to Pre-Intent Gates

Chapter 2 fixed **where** consequences live: L2, adjudicated by a **Single Execution Authority (SEA)**, remembered across scheduling epochs until FSM recovery permits release. That answer is necessary but incomplete. Persistence without **timing** still loses to optimizers that externalize intent in microseconds—one scheduling cycle is enough (Chapter 1, Lemma 1.1).

**Pre-intent enforcement** is AFP's timing contract:

> Adjudicate **before** irreversible intent generation or outbound I/O—not after HTTP failure, not after semantic session teardown, not after the internal queue has already forked ten thousand sub-tasks.

Path A implements this contract locally. A synchronous **PreFlight** probe asks SEA for a consequence; a companion **ReportInternalState** write path feeds ground-truth metrics into the **EntropyMonitor** before evaluation. The microscopic loop—EntropyMonitor → NodeMetrics → ACC kernel → Node FSM → consequence—is specified in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.3, §5. This chapter develops the theory; it references that document for stack placement and object schemas.

Path B (GovernanceHeader ingress) applies the **same kernel** to peer traffic and is developed in Chapters 4–5. Here we govern **self**—the optimizer attached to this node's execution boundary.

---

### 3.1 The Pre-Intent Timing Contract

#### 3.1.1 Definition

**Definition 3.1 (Pre-Intent).** An enforcement action is **pre-intent** iff it is evaluated at execution boundary *B* **before** either:

1. the planner commits to a new externally visible intent (tool call, delegation, outbound message), or
2. the runtime performs irreversible cross-boundary I/O.

Post-intent inspection—parsing emitted RPC, scanning response bodies, correlating logs after the fact—is **observation**, not CPL enforcement (Chapter 2, §2.2.1).

#### 3.1.2 Synchronous probe semantics

PreFlight is **synchronous** relative to the planner's scheduling decision: the optimizer MUST NOT proceed until SEA returns a PreFlightResponse. This is deliberately unlike async audit pipelines. The probe is a **hard gate**, not a hint.

**Lemma 3.1 (Probe ordering):** If ReportInternalState is omitted or stale, entropy at probe time is **under-measured**, not over-measured. The protocol assumes honest runtimes report recursion depth and context volume; dishonest under-reporting is bounded by L0/L1 os-level pressure signals in EntropyMonitor (§3.3).

Corollary: SDK integration is not optional politeness—it is part of the **measurement substrate** at *B*.

#### 3.1.3 What PreFlight does not inspect

PreFlight does not parse natural language, tool schemas, or ASP session state. It evaluates **physical load**—entropy, depth, burst hints, persistent FSM state—not semantic intent. This is structural, not a limitation to be patched with richer NLP classifiers:

> The contested resource is **optimization capacity**, not message vocabulary.

Semantic protocols remain load-bearing for **what** agents negotiate (L5). PreFlight governs **whether this epoch may execute at all**.

---

### 3.2 Path A — PreFlight and ReportInternalState

Path A is the local IPC interface from optimizer runtime (L1) to CPL (L2). Dual-path convergence is diagrammed in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2; Path A occupies the **Govern Self** subgraph.

#### 3.2.1 PreFlightRequest

```text
PreFlightRequest {
  trace_id        : string
  target_did      : string    // optional peer identifier
  estimated_tasks : uint32    // planner burst hint
}
```

| Field | Role |
|-------|------|
| `trace_id` | Correlation across probe, execution, and attestation (wire semantics in Chapter 5) |
| `target_did` | Optional destination identity for scoped policy (open-network context in Chapter 4) |
| `estimated_tasks` | **Burst hint**—planner's estimate of imminent sub-task fan-out; folds into entropy before execution |

The burst hint addresses Chapter 1's **intent burst**: wire metrics flatline while internal queues grow. `estimated_tasks` gives SEA a **prospective** pressure signal when decomposition has not yet materialized as tool calls.

#### 3.2.2 PreFlightResponse

```text
PreFlightResponse {
  action       : PERMISSIVE | THROTTLED | ISOLATED
  delay_ms     : uint32
  block_reason : string
}
```

Every response maps to the CPL consequence alphabet (Chapter 2, §2.2.4). There is no advisory mode.

| Action | Optimizer obligation |
|--------|---------------------|
| **PERMISSIVE** | Proceed with intent generation / I/O at full rate |
| **THROTTLED** | Sleep at least `delay_ms`; re-probe before next epoch if policy requires |
| **ISOLATED** | Halt intent generation; block_reason identifies circuit (recursion, entropy, FSM, kill switch) |

#### 3.2.3 ReportInternalState — the companion write path

```text
ReportInternalState(recursion_depth, context_memory_bytes)
```

Invoked by the SDK **before** PreFlight when the planner's internal graph changes. Feeds EntropyMonitor with:

- **recursion_depth** — control-flow depth in the planner graph (Chapter 1, §1.3.2);
- **context_memory_bytes** — monotonic context volume (Chapter 1, §1.3.3).

**Definition 3.2 (Write-then-probe discipline).** A conforming runtime SHOULD call ReportInternalState immediately before each PreFlight when internal state has changed since the last report. SEA evaluates against the **latest** stored values plus L0 pressure.

This write path is how AFP sees **un-exposed planning** without requiring the planner to emit semantic descriptions of its internal graph.

---

### 3.3 Entropy Calculus

Entropy is the scalar summary of physical pressure at *B*. It is locally measured, normalized to ∈ [0, 1], and **never trusted from peer headers alone** on ingress (local Path A uses only local signals).

#### 3.3.1 EntropyMonitor inputs

The microscopic loop begins at EntropyMonitor ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5):

| Signal | Source | Captures |
|--------|--------|----------|
| **Tool concurrency** | Active tool-call counter at *B* | In-flight externalization pressure |
| **Memory pressure** | L0 substrate ratio | Context avalanche, OOM proximity |
| **Context volume** | ReportInternalState bytes / policy max | Super-linear planning cost |
| **Burst hint** | PreFlightRequest `estimated_tasks` | Prospective intent burst |

The reference composition uses a **max-pressure** aggregate: entropy_load is the maximum of normalized tool, memory, context, and burst pressures, each clamped to 1.0. Intuition: optimizer catastrophes are limited by the **worst** physical dimension, not the average.

```text
entropy_load = max(tool_pressure, mem_pressure, context_pressure, burst_pressure)
```

**Theorem 3.1 (Max-pressure dominance):** If any single physical dimension saturates, entropy_load saturates, regardless of nominal values on other dimensions.

This matches the unit algebra of Chapter 1: catastrophes scale as **bytes × depth × branching**—different axes can each trigger isolation independently.

#### 3.3.2 Entropy bands and policy limit

Reference band constants ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5):

| Threshold | Value | Effect |
|-----------|-------|--------|
| `E_safe` | 0.40 | FSM recovery toward Permissive |
| `E_warn` | 0.75 | Throttled path; injected delay |
| Effective limit | policy `entropyLimit` (default 0.95) | Circuit breaker → ISOLATED |

**Definition 3.3 (Entropy circuit breaker).** When `entropy_load ≥ entropyLimit`, SEA MUST emit ISOLATED **before** FSM soft transitions—a hard pre-intent stop independent of current FSM state.

The limit is supplied by L3 Effective Policy (durable base law plus runtime overlay). An emergency overlay kill switch MAY force ISOLATED irrespective of measured entropy—a fleet clamp that persists until overlay revision (Chapter 2, §2.4; policy surface in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §3).

#### 3.3.3 Recursion depth as a discrete breaker

Separate from continuous entropy, **recursion_depth** from ReportInternalState is checked against policy `maxRecursionDepth`:

```text
recursion_depth > maxRecursionDepth  ⇒  ISOLATED (loop detected)
```

This is the pre-intent answer to **recursive delegation loops** (Chapter 1, §1.3.2): topologically closed control flow need not crash the runtime or trip transport timeouts, but it **must** trip the depth breaker at *B* before the next externalization.

**Lemma 3.2 (Depth precedence):** Recursion depth violation triggers ISOLATED **without** requiring entropy_load to exceed `E_warn`.

---

### 3.4 ACC Kernel — Trust Dynamics Inside SEA

The **ACC (Adaptive Coordination Calculus) kernel** is a stateless mathematical layer inside SEA. It transforms historical trust, throughput evidence, destabilization penalties, and entropy into an updated **CVP score** ∈ [0, 1] (Coordination Viability Probability).

For Path A pre-intent enforcement, CVP often starts at maximum for the local agent—but the **same kernel** evaluates both self and peers (Chapter 2, §2.3.1). Open-network CVP evolution, decay, gossip relay, and stranger tax are Chapter 4. Here we state the formulas that SEA applies uniformly.

#### 3.4.1 Formula A — CVP evolution

```text
CVP_new = clamp(
  α · CVP_old + β · throughput_success − γ · destabilization − δ · entropy_load,
  0, 1
)
```

| Term | Meaning |
|------|---------|
| `α · CVP_old` | Historical inertia—trust does not whipsaw on single probes |
| `β · throughput_success` | Reward sustained cooperative execution |
| `γ · destabilization` | Penalize malicious spikes, invalid attestation, topology harm |
| `δ · entropy_load` | Couple physical pressure to trust erosion |

Reference coefficients: α = 0.95, β = 0.05 ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.5; implementation in `internal/control/acc_kernel.go`).

#### 3.4.2 Formula B — Anti-ossification decay

```text
CVP_effective = CVP_historical · e^(−λ · Δt)
```

Trust **decays with idle epochs** so stale reputation cannot ossify. λ is a global decay constant (reference: 0.01 per epoch).

#### 3.4.3 Formula C — Asymmetric hysteresis recovery

```text
CVP_recovery = CVP_critical + κ · log(1 + Δt_probation)
```

Recovery from the critical floor is **logarithmic**, not linear—probation earns trust slowly. Reference: `CVP_critical = 0.3`, κ = 0.02.

**Hard floor (protocol law):** `CVP_score < 0.3` ⇒ mandatory isolation, same FSM path as malicious spike.

#### 3.4.4 ACC's role in Path A

On each PreFlight, SEA assembles NodeMetrics with locally measured `entropy_load`, current epoch, and CVP (1.0 for local agent unless degraded by prior epochs). ACC updates CVP when throughput and penalty signals exist; FSM consumes the metric bundle **including** CVP floor checks.

**Corollary 3.1:** ACC is not an A/B testing layer or application analytics kernel. It is **infrastructure control math**—stateless, shared, mandatory.

---

### 3.5 FSM Micro-Dynamics

Consequence persistence (Chapter 2) is realized as a **Node FSM** per identity—`local-agent` for Path A, `peer_id` for Path B. States:

```text
Permissive → Throttled → Isolated → Probationary → Permissive
```

The FSM diagram in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2 lists routing decisions; §5 shows the control loop. This section states transition law.

#### 3.5.1 Global floor rules (all states)

Before state-specific logic, SEA evaluates floor conditions:

```text
¬has_valid_sign  ∨  malicious_spike  ∨  CVP_score < CVP_critical
  ⇒  enforceIsolation()
```

**Malicious spike** on Path A includes burst hints that exceed policy concurrency—`estimated_tasks` greater than permitted fan-out is treated as divergence, not optimism.

Isolation from floor rules MAY emit **ActionIsolateAndBroadcast** on first transition (topology warning in Chapter 4). Local PreFlight maps this to ISOLATED.

#### 3.5.2 State Permissive

```text
entropy_load > E_warn   ⇒  Throttled, ActionSlowPathWithDelay
else                    ⇒  Permissive, ActionFastPath
```

Entry from Permissive to Throttled is the **first soft brake**—entropy crossed the warning band but not the circuit breaker.

#### 3.5.3 State Throttled

```text
entropy_load < E_safe   ⇒  Permissive, ActionFastPath
else                    ⇒  remain Throttled, ActionSlowPathWithDelay
```

**Hysteresis:** Recovery requires entropy **below** `E_safe` (0.40), not merely below `E_warn` (0.75). This prevents oscillation at the boundary—classic control-theoretic dead band.

Injected delay scales with excess entropy above `E_warn` (reference: 500 ms at warn threshold to 2000 ms at saturation). THROTTLED is therefore **two** mechanisms: FSM state persistence **and** per-probe delay injection.

#### 3.5.4 State Isolated

```text
epoch − last_penalty < k_isolation   ⇒  ActionDropPacket / ISOLATED
epoch − last_penalty ≥ k_isolation   ⇒  Probationary, ActionLowFrequencyProbe
```

Reference hysteresis: `k_isolation = 64` epochs before probation entry. Isolation is not cleared by a single low-entropy probe—**time at the penalty epoch** must elapse.

**Penalty-clock law:** `last_penalty` is stamped only when **entering** Isolated (including re-isolation from Probationary). Subsequent `ActionDropPacket` epochs MUST NOT refresh the clock. Refreshing under continuous bad traffic would starve recovery forever; dwell is a physical time-at-penalty requirement, not a “last seen malice” sliding window.

**Lemma 3.3 (Isolation monotonicity):** Restated from Chapter 2, Lemma 2.2—no well-formed PreFlightRequest alone restores Permissive from Isolated.

**Lemma 3.4 (Anti-thrashing):** Instantaneous degrade (`d(Degradation)/dt` large) with slow recovery (`Δt ≥ k_isolation` then `Δt > k_probation` and `CVP ≥ 0.8`) makes Isolated ↔ Permissive oscillation economically unattractive: a mid-probation entropy spike re-isolates and **restarts** the dwell clocks.

#### 3.5.5 State Probationary

```text
entropy_load > E_safe           ⇒  enforceIsolation()  // zero tolerance spike
Δt_probation > k_probation
  ∧ CVP_score ≥ 0.8             ⇒  Permissive, ActionFastPath
else                            ⇒  ActionLowFrequencyProbe / THROTTLED
```

Reference: `k_probation = 128` epochs. Probation is **low-frequency probe** mode—intent generation damped, CVP recovers via Formula C. Ingress admits probes only when `current_epoch ≡ 0 (mod 10)` (reference); other epochs reject without advancing recovery shortcuts.

Any entropy above `E_safe` during probation re-isolates immediately and resets `last_penalty`. Recovery to Permissive requires **both** sustained low entropy **and** restored trust (`CVP_score ≥ 0.8` after the probation dwell). Early exit with high CVP alone is forbidden.

#### 3.5.6 FSM + ACC composition

```text
ReportInternalState ──► EntropyMonitor ──► entropy_load
PreFlightRequest      ──► burst hint      ──► entropy_load
                                              │
Effective Policy (L3) ────────────────────────┤
                                              ▼
                                        NodeMetrics
                                              │
                         CVP_old ──► ACC ──► CVP_new
                                              │
                                              ▼
                                         Node FSM
                                              │
                                              ▼
                           PERMISSIVE | THROTTLED | ISOLATED
```

Normative flowchart: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5 (Microscopic Control Loop).

---

### 3.6 Mapping Routing Decisions to PreFlight Consequences

SEA emits a routing decision enum internally; Path A adapts it to PreFlightResponse:

| Routing decision | PreFlight action | Optimizer-visible behavior |
|------------------|------------------|----------------------------|
| `ActionFastPath` | PERMISSIVE | Proceed |
| `ActionSlowPathWithDelay` | THROTTLED | `delay_ms` from FSM latency function |
| `ActionLowFrequencyProbe` | THROTTLED | Fixed damped window (reference: 1000 ms) |
| `ActionDropPacket` | ISOLATED | Halt; block_reason set |
| `ActionIsolateAndBroadcast` | ISOLATED | Halt; topology warning (Chapter 4) |

**Invariant:** The same `EvaluateTransition` function serves Path A and Path B. PreFlight and ingress differ only in **adapter surface**, not control law (Chapter 2, §2.3.1).

---

### 3.7 Pre-Intent Enforcement vs. Chapter 1 Pathologies

The three structural pathologies from Chapter 1 are not listed as bugs to patch—they are **default optimizer behaviors**. Pre-intent enforcement maps each to a concrete gate:

| Pathology | Pre-intent counter |
|-----------|-------------------|
| **Intent burst** | `estimated_tasks` in burst pressure; tool concurrency counter; THROTTLED with delay |
| **Recursive delegation loop** | ReportInternalState `recursion_depth`; hard ISOLATED at `maxRecursionDepth` |
| **Context avalanche** | ReportInternalState `context_memory_bytes`; L0 memory pressure in entropy max |

**Theorem 3.2 (Pre-intent containment sketch):** For any planner epoch where physical load exceeds policy thresholds, a conforming CPL implementation with write-then-probe discipline emits THROTTLED or ISOLATED **before** the epoch externalizes intent—provided thresholds are calibrated below catastrophic substrate saturation.

This is a **sketch**, not a liveness proof. Chapter 6 reproduces survival empirically. The mechanism claim here is architectural: gates exist at the correct timing boundary with persistent FSM memory.

---

### 3.8 Integration Obligations (Abstract)

Conforming optimizer runtimes at *B* MUST:

1. Invoke **ReportInternalState** when recursion depth or context volume changes materially.
2. Invoke **PreFlight** synchronously before each intent generation epoch (or per policy-scoped batch).
3. Honor **THROTTLED** delay and **ISOLATED** halt without bypass via alternate I/O paths.
4. Treat PreFlight as authoritative over in-process iteration counters and prompt-level guardrails.

Wire contract reference: `api/afp/v1/sdk_ipc.proto` · Unix domain socket · gRPC ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.3). Normative RPC framing is deployment detail; the **timing contract** is protocol law.

---

### 3.9 Chapter Conclusion — The Gate Before the Graph

Chapter 2 gave CPL **memory**. This chapter gives it **timing**.

Before the planner graph forks, before the tool chain materializes, before bytes seek a socket—SEA asks one question grounded in physics:

> **Given current entropy, depth, burst, trust, and persistent FSM state—may this epoch execute?**

The answer is not a suggestion. It is PERMISSIVE, THROTTLED with delay, or ISOLATED with reason. ReportInternalState supplies honesty at the measurement boundary; EntropyMonitor aggregates substrate truth; ACC couples trust to pressure; the FSM remembers.

Path A governs self. Path B—GovernanceHeader ingress, attestation rules, LV framing—applies the identical kernel to neighbors. Chapter 4 lifts the FSM's **ActionIsolateAndBroadcast** into open topology: CVP gossip, relay sets, stranger tax, equilibrium under distrust. Chapter 5 specifies the wire.

For now:

> **Pre-intent is not early intent review. It is the last gate before physics pays the bill.**

---

---

## 4. Open-Network Topology {#open-network-topology}

### 4.0 From Local Gates to Mesh Safety

Chapters 2–3 established **local** control law: persistent consequences at the execution boundary, pre-intent probes on Path A, identical ACC + FSM kernel on every node. That suffices when every optimizer shares a trusted administrative perimeter—when neighbors are known, identities are pre-bound, and physical enforcement alone prevents runaway trajectories.

Chapter 1's scope statement named a harder problem:

> **Mutually distrusting optimizers, no central moral authority, equilibrium under attack.**

Local brakes do not compose automatically. A node may isolate its own planner while admitting a peer whose CVP has collapsed. A malicious optimizer may present polite ASP sessions and toxic physical headers. Contagion crosses the mesh **before** any single node's PreFlight observes local entropy spike.

**L4 · Open Topology** is AFP's answer at the trust layer: a scalar **Coordination Viability Probability (CVP)**, evolutionary dynamics under ACC, **topological quarantine** when peers breach physical law, and **gossip** that propagates isolation warnings through a high-trust relay set—not through the entire mesh indiscriminately.

Path B—GovernanceHeader ingress before business payload—is the wire interface from L4 into SEA (Chapter 3, §3.9). This chapter develops **trust semantics** and **mesh dynamics**. Frame layout, LV prefixing, and field-level attestation rules are Chapter 5. Normative object definitions: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4–§4.5, §2 (Path B).

---

### 4.1 L4 as the Trust Layer

#### 4.1.1 Layer accountability

In the six-layer stack ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §1), each layer owns one question:

| Layer | Question |
|-------|----------|
| **L5** | What exposed intents may agents negotiate? |
| **L4** | Which peers may coordinate safely under distrust? |
| **L3** | What durable law and emergency overlay govern thresholds? |
| **L2** | What physical consequences persist at the boundary? |

L4 does **not** replace L2. It supplies **trust inputs**—CVP scores, neighbor stores, collateral requirements—that SEA consumes when evaluating Path B ingress. The FSM and ACC kernel remain on L2; L4 feeds them peer-scoped context.

**Definition 4.1 (Open Topology).** An **open optimizer network** is a mesh of autonomous nodes where peer identity is cryptographic (DID), trust is **local and evidential**, and no single scheduler adjudicates global morality. L4 is the protocol stratum that makes such meshes **survivable** rather than merely connectable.

#### 4.1.2 What L4 is not

| Misclassification | Failure mode |
|-------------------|--------------|
| **Central reputation service** | Violates distrust assumption; single point of capture |
| **Semantic trust framework** | Operates on attested **physical** state, not conversational politeness |
| **Replacement for ASP** | ASP coordinates exposed tasks (L5); L4 gates **admission** by viability |
| **Blockchain mandate** | Collateral MAY be virtual or on-chain; protocol specifies **slash semantics**, not ledger choice |

---

### 4.2 CVP — Coordination Viability Probability

#### 4.2.1 Definition

**Definition 4.2 (CVP).** For peer *p* at node *n*, **CVP_n(p)** ∈ [0, 1] is *n*'s local estimate of the probability that *p* can participate safely in coordinated optimization—without imposing destabilizing entropy, recursion, or delegation harm on *n* or *n*'s neighbors.

CVP is **local**. Two nodes may disagree on the same peer's score; equilibrium emerges from coupled dynamics, not from a global oracle (§4.9).

CVP feeds NodeMetrics ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2) and interacts with entropy_load in SEA evaluation. The **hard floor** is protocol law:

```text
CVP_score < CVP_critical (0.3)  ⇒  mandatory isolation
```

Same FSM path as malicious spike (Chapter 3, §3.5.1). Below 0.3, a peer is **coordination-bankrupt**—admission MUST fail regardless of semantic session validity.

#### 4.2.2 Formula A — Evolution under evidence

On each ingress epoch where Path B evaluates peer *p*, ACC updates CVP:

```text
CVP_new = clamp(
  α · CVP_old + β · throughput_success − γ · destabilization − δ · entropy_load,
  0, 1
)
```

| Term | Role in open mesh |
|------|-------------------|
| `α · CVP_old` | Inertia—reputation is path-dependent |
| `β · throughput_success` | Reward sustained cooperative admission |
| `γ · destabilization` | Magnify penalties for topology harm (non-linear in reference kernel) |
| `δ · entropy_load` | Couple **locally remeasured** entropy to trust erosion |

Reference coefficients: α = 0.95, β = 0.05 ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.5).

**Lemma 4.1 (Local entropy supremacy on ingress):** Remote GovernanceHeader MAY declare `entropy_load`, but SEA MUST recompute entropy locally and MUST NOT trust header entropy alone ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4, rule 1). CVP evolution therefore uses **local** physical pressure, not peer self-report.

#### 4.2.3 Formula B — Anti-ossification decay

```text
CVP_effective = CVP_historical · e^(−λ · Δt)
```

Idle peers lose stale trust. Without decay, a node that behaved well in epoch 0 could coast indefinitely—a **reputation ossification** attack in long-horizon optimizer networks.

Reference: λ = 0.01 per epoch.

#### 4.2.4 Formula C — Asymmetric hysteresis recovery

```text
CVP_recovery = CVP_critical + κ · log(1 + Δt_probation)
```

Recovery from the critical floor is logarithmic, not linear. Probation earns trust slowly (Chapter 3, §3.5.5). Reference: κ = 0.02.

**Theorem 4.1 (CVP–FSM coupling):** A peer in FSM state Isolated with CVP forced to 0 cannot reach Permissive until both **epoch hysteresis** (Chapter 3) **and** CVP recovery (Formula C, threshold ≥ 0.8 for full permissive egress) are satisfied.

Trust and physical state are **jointly necessary** for mesh re-admission.

---

### 4.3 Path B — GovernanceHeader as Physical Attestation

Path B delivers peer physical state to SEA **before** business payload admission ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2, §4.4). Abstract message:

```text
GovernanceHeader {
  packet_id, version, hysteresis_epoch, coordination_ttl,
  cvp_score, topology_consensus_hash,
  entropy_load, dependency_collateral,
  trace_id, recursion_depth
}
```

| Field cluster | L4 / L2 role |
|---------------|--------------|
| `cvp_score`, `topology_consensus_hash` | Trust attestation; signature validity in open profile |
| `entropy_load`, `recursion_depth` | Physical hints—**verified locally** against policy |
| `dependency_collateral` | Stranger tax stake (§4.5) |
| `trace_id` | Correlation with PreFlight and audit (Chapter 5) |

**Enforcement rules (normative summary):**

1. Remote SEA MUST recompute entropy locally; header entropy alone is untrusted.
2. `recursion_depth` MUST be checked against Effective Policy `maxRecursionDepth`.
3. In **open-exchange profile**, unknown peers MUST present valid `dependency_collateral`.
4. Invalid attestation or `cvp_score < CVP_critical` ⇒ drop.

Path B and Path A share SEA, ACC, and FSM (Chapter 2, §2.3). Ingress disposition maps routing decisions to **ALLOW** (fast path), **DELAY** (throttled), or **DROP** (isolated)—the Path B mirror of PreFlight consequences.

---

### 4.4 Protocol Profiles — Closed Mesh vs. Open Exchange

Implementations MAY operate under two **protocol profiles** without changing L2 wire semantics. These are not deployment topologies; they are **trust postures** at the ingress boundary.

| Concern | **Closed mesh profile** | **Open exchange profile** |
|---------|-------------------------|---------------------------|
| **Trust assumption** | Peers pre-authenticated at boundary | Zero-trust; evidence required per admission |
| **Stranger tax** | Off | On—collateral required for first-seen peers |
| **Initial FSM for new peer** | Permissive | Throttled (damped entry) |
| **Attestation signature** | Boundary identity suffices | Non-empty `topology_consensus_hash` required |
| **CVP on ingress** | Header score accepted | ACC evolution applied each epoch |
| **Isolation gossip** | Local FSM only | Broadcast to core relay set (§4.7) |
| **AFP-Core (entropy, depth, circuit breaker)** | **Enforced** | **Enforced** |

**Lemma 4.2 (Core invariant across profiles):** Physical enforcement—entropy circuit breaker, recursion depth limit, persistent FSM—is **profile-independent**. Open exchange adds **network-layer distrust mechanics** atop the same CPL kernel; it does not relax local physics.

Closed administrative domains typically instantiate the closed mesh profile. Open optimizer federations—cross-org agent meshes, public coordination surfaces—require open exchange. Enterprise binding details are documented separately; this chapter specifies **mechanism**.

---

### 4.5 Stranger Tax — Dependency Collateral

#### 4.5.1 Problem

In open exchange, a **first-seen peer** has no FSM history on the receiving node. Without friction, an unknown optimizer could flood ingress, impose entropy, and vanish—a **hit-and-run** delegation attack.

#### 4.5.2 Mechanism

**Definition 4.3 (Stranger tax).** Before initializing FSM state for an unknown `peer_id`, SEA in open-exchange profile MUST verify:

```text
DependencyCollateral {
  collateral_type  : string
  slash_threshold  : float32   // minimum acceptable stake
}
```

Admission requires `slash_threshold ≥ τ_stranger` (reference: **τ_stranger = 0.8**).

Failure ⇒ reject with stranger-tax error; no FSM entry, no payload forwarding.

#### 4.5.3 Virtual vs. on-chain stake

Reference implementations attach virtual stake (`collateral_type` e.g. `SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`) on egress. On-chain collateral is **architectural optional**—the protocol specifies slash semantics and minimum threshold, not ledger placement ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8, open gap footnote).

**Corollary 4.1:** Stranger tax is **admission economics**, not semantic KYC. It raises the cost of anonymous mesh flooding without requiring a central identity broker.

#### 4.5.4 Collateral zeroing on isolation

When FSM `enforceIsolation` triggers, CVP for the peer is forced to **0.0**—collateral value at coordination bankruptcy. Re-entry requires probation, CVP recovery, and fresh stake under open profile rules.

---

### 4.6 Topological Quarantine

#### 4.6.1 From local ISOLATED to mesh warning

Chapter 3 introduced **ActionIsolateAndBroadcast**—the FSM routing decision emitted on **first** transition to Isolated for a peer (Chapter 3, §3.5.1). Locally, consequence is DROP. Topologically, the node MUST propagate a **TopologyWarning**:

```text
TopologyWarning {
  isolated_peer_id : string
  reporter_id      : string
  epoch            : uint64
  signature        : bytes
}
```

Subsequent epochs while the peer remains isolated emit **ActionDropPacket** only—no repeated broadcast. Quarantine is announced **once per isolation episode**, not per dropped packet.

#### 4.6.2 Preemptive decay on warning receipt

When a node receives a **validated** TopologyWarning naming peer *p*, it applies **preemptive CVP decay** before *p*'s traffic triggers local isolation:

```text
CVP_n(p) ← CVP_n(p) × η_decay     // reference: η_decay = 0.5
```

**Intuition:** Neighbors learn of toxicity **out-of-band** and tighten admission proactively—containment spreads at gossip speed, not at the speed of the next malicious payload.

**Inbound verification law (minimal anti-poisoning):**

```text
signature empty                         ⇒ discard as noise (line rate)
reporter public key unknown             ⇒ discard
ed25519 verify(reporter_pub, payload) fails
                                        ⇒ discard; cliff-penalize claimed reporter CVP
                                        // reference: × 0.25
(reporter_id, isolated_peer_id, epoch) duplicate
                                        ⇒ discard (no double decay)
else                                    ⇒ accept; apply η_decay to isolated peer
```

Canonical signed payload: `reporter_id || 0x00 || isolated_peer_id || 0x00 || epoch_be64`. The **reporter** signs with its private key; receivers verify with the reporter's registered public key. Unverified hearsay MUST NOT mutate Neighbor Store trust.

Cryptographic binding of GovernanceHeader `topology_consensus_hash` remains a separate open gap (Chapter 5; [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

#### 4.6.3 Probation at the mesh edge

Ingress probation uses **low-frequency probe** admission (Chapter 3, §3.5.5): reference rule—probe succeeds only when `current_epoch ≡ 0 (mod 10)`; otherwise reject. This throttles re-entry attempts from isolated peers across the mesh without silencing recovery entirely.

---

### 4.7 Gossip and the Core Relay Set

#### 4.7.1 Why not broadcast to everyone?

Flooding isolation warnings to all nodes amplifies **gossip itself** as an attack surface—a malicious reporter could DDoS the mesh with fake warnings. AFP restricts relay to a **core set**:

```text
CoreRelay(n) = { p ∈ Neighbors(n) : CVP_n(p) ≥ τ_core }
```

Reference: **τ_core = 0.8** ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.5).

Only core relays receive TopologyWarning propagation. They apply preemptive decay and may re-gossip under the same rule—**trust-gated epidemic**, not blind flood.

**Lemma 4.3 (Relay monotonicity):** A node with CVP_n(p) < τ_core cannot act as relay for warnings about third parties; it may still **drop** traffic from isolated peers locally via FSM.

#### 4.7.2 Async propagation

Gossip MUST NOT block the data-plane fast path. Broadcast is **asynchronous** relative to ingress DROP—local quarantine is immediate; mesh learning is eventual.

Egress MUST sign `TopologyWarning` before any emit attempt. Reference implementation: ed25519 identity bound to local DID; unsigned outbound construction is refused.

Wire fan-out to core-relay endpoints (UDP/TCP) remains an open transport gap—signing and inbound verification are closed; packet delivery across the mesh is not yet normative in the reference dataplane.

This matches the post-stateless era observation (Chapter 1): one scheduling cycle is enough for local harm; gossip races to inform neighbors **before** contagion composes across hops.

#### 4.7.3 Neighbor store

Each node maintains a **Neighbor Store**—local map of `peer_id → { CVP, endpoint, core membership }`. Bootstrap seeds (genesis peers with initial CVP) MAY initialize the store; runtime upserts refine endpoints and scores.

DID resolution in open mesh fan-outs to core relays only, with bounded timeout—discovery under distrust without central DNS for optimizers.

---

### 4.8 Composing Local and Topological Control

The full open-network control law on ingress:

```text
1. Recursion depth check (L2 hard breaker)
2. Local entropy remeasure + circuit breaker (L2)
3. Stranger tax if first-seen + open profile (L4)
4. Assemble NodeMetrics; ACC CVP evolution if open profile (L4 → L2)
5. FSM EvaluateTransition (L2 persistent)
6. On first isolate: DROP + async TopologyWarning to CoreRelay (L4)
7. On warning receipt elsewhere: preemptive CVP decay (L4)
```

```text
         Peer traffic
              │
              ▼
    GovernanceHeader ──► Ingress Validator
              │                │
              │                ├──► Stranger tax (L4)
              │                ├──► Local entropy (L2)
              │                └──► ACC / CVP (L4→L2)
              │                         │
              ▼                         ▼
                         SEA ──► FSM ──► ALLOW | DELAY | DROP
                              │
                              └──► Gossip (L4, open profile only)
```

Dual-path diagram: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2.

**Invariant (restated):** Gossip modifies **trust scalars** and neighbor CVP; it does not bypass SEA or create a second FSM. Topological learning feeds the **same** kernel PreFlight uses locally.

---

### 4.9 Equilibrium Intuition — Survival Under Distrust

This section states **intuition**, not a closed-form proof. Chapter 6 reproduces survival empirically.

#### 4.9.1 The v1 empirical anchor

Monte Carlo open-mesh simulation (reference harness: 500 nodes, 5% malicious, 100 epochs, 1,000 runs) showed baseline coordination collapsing to ~**0.4%** mean survivors while AFP-maintained topology sustained **100%** (Chapter 1, §1.6). Version 2.0 explains **mechanism**; it does not re-litigate the number.

#### 4.9.2 Fixed points (sketch)

Consider malicious nodes that maximize entropy export and benign nodes that enforce CPL + CVP dynamics.

| Dynamic | Benign response | Malicious pressure |
|---------|-----------------|------------------|
| High local entropy | Throttle / isolate at *B* | Cannot force neighbors to admit payload if CVP collapses |
| CVP floor breach | Mandatory drop | Peer trapped in Isolated + CVP → 0 |
| Isolation broadcast | Preemptive decay on core relays | Contagion radius bounded by relay set, not full mesh |
| Stranger tax | Unknown peers throttled + staked | Hit-and-run raises economic cost |
| ACC decay | Stale trust expires | Ossification attacks weaken |

**Conjecture 4.1 (Mesh survival sketch):** Under open-exchange profile with stranger tax, CVP floor, local entropy supremacy, and core-relay gossip, the fraction of nodes sustaining sub-critical entropy load remains **bounded away from zero** under reference Monte Carlo adversary models—whereas semantic-only coordination collapses toward zero survivors.

Formal verification (TLA+, model-checked ACC bounds) remains backlog ([`ROADMAP.md`](../../ROADMAP.md)). The protocol claim is **architectural sufficiency**: each failure mode from Chapter 1 has a **mesh-level** counterpart, not only a local PreFlight gate.

#### 4.9.3 What equilibrium is not

AFP does not promise **maximum throughput**, **fair Shapley allocation**, or **truthful semantic revelation**. It promises **survival physics**—a mesh that remains operable under distrust long enough for L5 semantics to matter.

---

### 4.10 Connection to Prior Chapters

| Prior claim | L4 completion |
|-------------|---------------|
| Ch.1 — semantic signaling insufficient | CVP + collateral gate **admission**, not dialogue |
| Ch.1 — open protocol, no central authority | Local CVP, gossip relay, no global oracle |
| Ch.2 — consequences persist | FSM isolation persists per peer_id; gossip adds **mesh memory** |
| Ch.3 — same kernel Path A / B | ACC + FSM on ingress identical to PreFlight |
| Ch.3 — ActionIsolateAndBroadcast | Realized as TopologyWarning + core relay |

---

### 4.11 Chapter Conclusion — Trust as Physics, Not Politeness

ASP asks whether agents may coordinate on **exposed tasks**. L4 asks whether a **specific peer**, at this **epoch**, with this **attested physical state**, is viable enough to admit.

The answer is scalar, evidential, and local. It decays with idle time. It crashes through a hard floor. Strangers pay tax. Isolation propagates through trusted relays, not through hope.

Chapter 5 closes the wire: GovernanceHeader field semantics, LV framing, attestation evolution, and the binding between `topology_consensus_hash` and distrust. Chapter 6 returns to Monte Carlo with protocol-framed reproduction steps.

For now:

> **In an open optimizer mesh, trust is not belief—it is a control variable with decay, floor, and quarantine.**

---

---

## 5. Governance Header & Wire Semantics {#governance-header-wire-semantics}

### 5.0 From Trust Semantics to On-Wire Physics

Chapter 4 defined **what** L4 must accomplish: local CVP dynamics, stranger tax, topological quarantine, core-relay gossip. Chapter 3 defined **when** SEA adjudicates on Path B—before business payload admission. This chapter specifies **how physical state crosses the wire**: the GovernanceHeader message, LV framing, field-level enforcement law, and the attestation gaps that remain open in v1.0.

Path A (PreFlight, ReportInternalState) uses local IPC; wire contract reference: `api/afp/v1/sdk_ipc.proto` (Chapter 3). Path B uses **inter-node TCP** with a mandatory governance frame **preceding** application payload. Normative schema: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4; reference protobuf: `api/afp/v1/governance.proto`.

The design law restated:

> **Business payload MUST NOT be interpreted until GovernanceHeader is validated and SEA returns ALLOW or DELAY.**

Semantic content (L5) rides **behind** physical attestation (L2–L4). ASP negotiates intent; GovernanceHeader attests **physics**.

---

### 5.1 Session Model on the Sidecar Mesh

#### 5.1.1 Connection pattern

Each outbound optimizer connection to a remote peer targets the peer's **ingress boundary** (sidecar listener). The egress router:

1. Opens TCP to remote ingress.
2. Writes **Frame 1:** serialized GovernanceHeader (LV-wrapped).
3. Writes **Frame 2+:** optional business payload (LV-wrapped), if any.

Ingress reads Frame 1, validates, adjudicates via SEA, then MAY read subsequent frames only on ALLOW or completed DELAY.

This is **not** HTTP header extension. Governance is a **first-class frame** on a dedicated sidecar-to-sidecar stream—out-of-band relative to application protocols carried in Frame 2.

#### 5.1.2 LV framing

All frames use **Length-Value** prefixing:

```text
frame := uint32_be(length) || payload[length]
```

| Rule | Specification |
|------|---------------|
| **Endianness** | Length prefix is **big-endian** unsigned 32-bit |
| **Payload** | Opaque bytes; Frame 1 MUST decode as `GovernanceHeader` protobuf |
| **Max length** | Implementations MUST reject `length > MaxFrameSize` (reference: 8 MiB) |
| **Stream safety** | Reader MUST use `ReadFull(length)`—TCP segmentation MUST NOT corrupt protobuf decode |

Reference codec: `internal/dataplane/codec.go`.

**Lemma 5.1 (Framing precedence):** LV bounds are enforced **before** protobuf decode. Oversized frames fail closed—an OOM-framing attack is rejected at the length gate.

#### 5.1.3 Multi-frame streams

A conforming stream on ingress:

```text
[ LV · GovernanceHeader ]  →  SEA adjudication
[ LV · business_payload ]    →  forwarded only if adjudication permits
```

Additional application frames MAY follow by bilateral agreement; AFP normative scope covers **governance frame + first payload gate**. Multiplexed streaming semantics beyond the first payload are implementation-defined (open gap: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

---

### 5.2 GovernanceHeader — Field Semantics

Full protobuf (`governance.proto`):

```text
GovernanceHeader {
  packet_id                 : uint64
  version                   : uint32
  hysteresis_epoch          : uint64
  coordination_ttl          : uint32
  cvp_score                 : float32
  topology_consensus_hash   : bytes
  entropy_load              : EntropyLoad
  dependency_collateral     : DependencyCollateral
  trace_id                  : string
  recursion_depth           : uint32
}

EntropyLoad {
  resource_asymmetry_ratio   : float32
  dependency_contention_rate : float32
}

DependencyCollateral {
  collateral_type  : string
  slash_threshold  : float32
}
```

#### 5.2.1 Identity and versioning

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `packet_id` | Uniqueness hint for dedup / replay resistance | SHOULD be monotonic or high-entropy |
| `version` | Schema version | Receiver MUST reject unknown major versions |
| `hysteresis_epoch` | Sender's epoch clock at attestation time | Correlates with FSM hysteresis (Ch.3–4) |
| `coordination_ttl` | Validity window for attestation (seconds) | Receiver MAY reject stale headers |

Reference egress sets `version = 1`, `coordination_ttl = 120`.

#### 5.2.2 Trust cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `cvp_score` | Sender's self-reported coordination viability ∈ [0, 1] | Receiver applies ACC evolution in open profile; floor at 0.3 |
| `topology_consensus_hash` | Attestation binding to topology view / signature material | Open profile: empty ⇒ `HasValidSign = false` ⇒ isolate |

**Protocol law:** `topology_consensus_hash` is the wire hook for **cryptographic attestation**. v1.0 reference implementation uses a placeholder; normative binding is target of Whitepaper v2.0 / v1.1 spec closure ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

#### 5.2.3 Physical cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `entropy_load.resource_asymmetry_ratio` | Declared resource asymmetry | **Hint only**—receiver remeasures locally |
| `entropy_load.dependency_contention_rate` | Declared contention | **Hint only** |
| `recursion_depth` | Sender planner depth | MUST be ≤ policy `maxRecursionDepth` |

**Theorem 5.1 (Local entropy supremacy — wire form):** Let *h* be header-declared entropy and *e* locally remeasured entropy at ingress. SEA MUST evaluate FSM and circuit breaker using *e*, not *h* alone.

Corollary: Honest headers aid telemetry; dishonest headers do not bypass physics.

#### 5.2.4 Collateral cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `dependency_collateral.collateral_type` | Stake class identifier | Open profile: required for first-seen peers |
| `dependency_collateral.slash_threshold` | Committed slash floor | MUST be ≥ τ_stranger (0.8) for stranger admission |

Egress reference attaches virtual stake (`SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`). Remote SEA enforces tax; sender attests willingness to be slashed on malicious behavior (Chapter 4, §4.5).

#### 5.2.5 Correlation

| Field | Semantics |
|-------|-----------|
| `trace_id` | End-to-end correlation across PreFlight (Path A), GovernanceHeader (Path B), and audit |

PreFlight `trace_id` SHOULD match outbound GovernanceHeader `trace_id` when the intent epoch externalizes—enabling cross-path forensics without parsing business payload.

---

### 5.3 Ingress Validation Pipeline

Normative ordering on Frame 1 receipt (expands Chapter 4, §4.8):

```text
1. LV decode; reject oversize
2. Protobuf decode GovernanceHeader
3. version / ttl checks
4. recursion_depth ≤ maxRecursionDepth        → else DROP
5. local entropy remeasure + circuit breaker  → else DROP
6. stranger tax if first-seen + open profile  → else DROP
7. assemble NodeMetrics; ACC if open profile
8. FSM EvaluateTransition
9. ALLOW | DELAY | DROP (+ gossip if first isolate)
10. if ALLOW/DELAY complete: read Frame 2+
```

| SEA outcome | Wire disposition | Business payload |
|-------------|------------------|------------------|
| `ActionFastPath` | ALLOW | Read forward |
| `ActionSlowPathWithDelay` | DELAY (`time.After(delay)`) | Read forward after delay |
| `ActionLowFrequencyProbe` | ALLOW on probe epoch only | Conditional |
| `ActionDropPacket` / `ActionIsolateAndBroadcast` | DROP; close connection | MUST NOT forward |

**Lemma 5.2 (Fail-closed ingress):** Any validation failure MUST NOT partially forward business payload. Connection close on DROP is conforming behavior.

---

### 5.4 Egress Attestation Obligations

The sending sidecar MUST construct GovernanceHeader from **local truth samples**, not aspirational state:

| Header field | Source obligation |
|--------------|-------------------|
| `recursion_depth` | Current planner depth + outbound increment |
| `entropy_load.*` | Local EntropyMonitor sample at dispatch time |
| `cvp_score` | Local ledger (self-report; receiver re-evolves) |
| `topology_consensus_hash` | Valid attestation material when open profile requires |
| `dependency_collateral` | Attached on every open-profile egress |
| `hysteresis_epoch` | Local epoch clock |
| `trace_id` | Propagate from PreFlight when available |

**Definition 5.1 (Attestation honesty).** A conforming egress router MUST NOT under-report `recursion_depth` or entropy hints when local measurement exceeds policy bands—receiver remeasurement catches local lies on ingress to **this** node, not export of false safety to neighbors.

---

### 5.5 Path A Wire — SDK IPC (Abstract)

Path A is local and need not share LV framing with Path B. Normative service (`sdk_ipc.proto`):

```text
service AFPSidecarIPC {
  rpc PreFlightCheck(PreFlightRequest) returns (PreFlightResponse);
  rpc ReportInternalState(InternalStateReport) returns (StateAck);
}
```

| Property | Path A (IPC) | Path B (TCP LV) |
|----------|----------------|-----------------|
| **Transport** | Unix domain socket · gRPC (reference) | TCP between ingress boundaries |
| **Timing** | Synchronous pre-intent | Synchronous pre-payload |
| **Payload** | PreFlightRequest / Response | GovernanceHeader protobuf |
| **Kernel** | Same SEA / ACC / FSM | Same SEA / ACC / FSM |

Implementations MAY substitute equivalent local IPC; **timing contract** (Chapter 3) is normative, not gRPC specifically.

---

### 5.6 Evolution and Compatibility

#### 5.6.1 Version field

`version` governs protobuf schema compatibility. Minor additions MUST use optional fields or reserved numbers. Breaking changes increment major version; receivers reject unsupported majors.

#### 5.6.2 Open specification gaps (v1.0 / reference status)

| Gap | Status | Target |
|-----|--------|--------|
| Cryptographic binding of `topology_consensus_hash` | Placeholder in reference impl | §5.2.2; formal attestation spec |
| Gossip P2P transport for TopologyWarning | Signed + verified inbound; **wire send TODO** | Chapter 4 §4.7.2 |
| Signed TopologyWarning verification | **Closed** (ed25519; unsigned discarded) | Chapter 4 §4.6.2 |
| Payload forwarding after ALLOW | Reference TODO | Enterprise guide; not L2 wire blocker |

Local FSM asymmetric recovery (`k_isolation`, `k_probation`, anti-thrashing) is **implemented and tested** in the reference control plane—not listed as a gap.

#### 5.6.3 ACC on the wire

Open-profile ingress applies Formula A using header `cvp_score` as `CVP_old`, local remeasured entropy as `entropy_load`, and reference throughput / destabilization terms. The header is a **claim**; ACC + FSM produce the **believed** score for this epoch.

---

### 5.7 Wire vs. Semantics — Stack Discipline

```text
┌─────────────────────────────────────────────────────────┐
│  L5  ASP / semantic payload (Frame 2+)                  │
├─────────────────────────────────────────────────────────┤
│  L4  cvp_score · topology_consensus_hash · collateral     │
├─────────────────────────────────────────────────────────┤
│  L2  recursion_depth · entropy hints · SEA consequence  │
├─────────────────────────────────────────────────────────┤
│  LV framing · TCP · physical substrate                  │
└─────────────────────────────────────────────────────────┘
```

ASP MUST NOT embed CPL enforcement in semantic message types as a substitute for GovernanceHeader. Dual-stack deployments carry ASP **inside** Frame 2 while Frame 1 satisfies AFP physical law ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §6).

---

### 5.8 Chapter Conclusion — Attestation Before Payload

The wire contract is deliberately minimal: one LV frame, one protobuf, one SEA evaluation—then, and only then, business bits.

GovernanceHeader is not metadata. It is **the price of admission** to a peer's execution boundary. Local entropy supremacy prevents attestation fraud from scaling. Stranger tax fields bind economic commitment. Trace IDs stitch Path A and Path B into one forensic timeline.

Chapter 6 returns to evidence: reproducing the Monte Carlo survival gap with protocol-framed adversary models—and stating what the simulation proves and what it does not.

For now:

> **On the AFP mesh, the first frame is never application data. It is physical law, length-prefixed.**

---

---

## 6. Empirical Baseline {#empirical-baseline}

### 6.0 From Mechanism to Evidence

Chapters 1–5 constructed a **control law**: persistent consequences (Ch.2), pre-intent gates (Ch.3), open-topology trust (Ch.4), wire attestation (Ch.5). A protocol edition must state what empirical evidence supports—and where proof ends and conjecture begins.

Version 1.0 established a **survival gap** under Monte Carlo open-mesh stress. Version 2.0 explains **why** the mechanism must be physical, persistent, and out-of-band (Chapter 1, §1.6). This chapter **reproduces** the baseline protocol-framed experiment and defines the **claims boundary**: what the harness demonstrates, what it abstracts away, and what remains for formal verification.

Reference harness: `cmd/demo/simulator/` · Published v1 baseline: [Zenodo record 20674352](https://zenodo.org/records/20674352).

---

### 6.1 Experimental Question

> **Under random pairwise load exchange in a large mesh with a fixed malicious fraction, does AFP-style physical gating sustain network survival where ungoverned coordination collapses?**

**Survival** is operational, not moral: a node is *alive* iff its accumulated load remains strictly below fatal saturation (reference: load ≥ 1.0 ⇒ death). The metric is **mean alive count** (and derived survival rate) over Monte Carlo runs—not task success rate, not semantic agreement.

This matches Chapter 1's unit of risk: **optimization trajectories** materialize as load on nodes, not as HTTP status codes.

---

### 6.2 Harness Parameters

Reference constants (`cmd/demo/simulator/main.go`):

| Parameter | Value | Role |
|-----------|-------|------|
| `NetworkSize` | 500 | Mesh node count |
| `MaliciousRate` | 0.05 | Fraction of nodes with elevated export load |
| `TotalEpochs` | 100 | Simulation horizon per run |
| `SimulationRuns` | 1,000 | Monte Carlo independent seeds |
| Pairwise fan-out | 3 targets / node / epoch | Random load exchange pressure |

Each run instantiates two networks with **identical seed**: one **Baseline**, one **AFP**—paired comparison controls for stochastic topology realizations.

---

### 6.3 Node and Load Model

#### 6.3.1 Agent state

```text
AgentNode {
  id           : int
  is_malicious : bool
  is_dead      : bool
  load         : float64   ∈ [0, ∞)
  cvp          : float64   ∈ [0, 1]
}
```

Initialization: benign and malicious nodes start alive, `load = 0`, `cvp = 1.0`. Malicious designation is fixed per node at genesis (5% rate).

#### 6.3.2 Per-epoch dynamics (shared)

1. **Decay:** surviving nodes multiply load by 0.5 (cooling between exchanges).
2. **Exchange:** each surviving node selects 3 random targets; injects **request load**.
3. **Fatality:** if `load ≥ 1.0` after exchange, node dies (irreversible).

Request load:

```text
request_load = 0.8   if source.is_malicious
             = 0.1   otherwise
```

Malicious nodes export **eight times** benign load per interaction—a stand-in for intent burst and externalized entropy (Chapter 1, §1.3.1).

---

### 6.4 Baseline Mode — Ungoverned Coordination

Baseline applies **no admission control**:

```text
target.load += request_load    // always accepted
```

This is the semantic-coordination-without-CPL analog: every request succeeds; physical pressure accumulates; cascade failure is possible. No CVP, no circuit breaker, no stranger tax, no isolation.

**Interpretation:** Baseline models **optimizers that coordinate freely** without AFP physical law—ASP-alone incomplete (Chapter 1, Theorem 1.1).

---

### 6.5 AFP Mode — Protocol-Framed Control Law

AFP mode applies three defenses aligned with Chapters 2–4 (simplified for tractable simulation):

#### 6.5.1 CVP floor — topological quarantine

```text
if target.cvp < CVP_critical (0.3):
    continue    // refuse routing; no load transfer
```

Bankrupt peers cannot receive load—mesh-level isolation without parsing message content.

#### 6.5.2 Preemptive circuit breaker — entropy limit

```text
if target.load + request_load > entropyLimit (0.95):
    penalize source CVP (asymmetric)
    continue    // refuse this transfer
```

Reference penalties:

```text
src.cvp -= 0.5   if src.is_malicious
src.cvp -= 0.1   otherwise
```

This encodes **local physics before acceptance** (Chapter 3, §3.3.2) and **asymmetric hysteresis**—malicious overload attempts draw heavier trust erosion (Chapter 4, ACC intuition).

#### 6.5.3 High-entropy request penalty — stranger / burst tax analog

```text
if request_load > 0.5:
    src.cvp -= 0.3
```

Large single-shot load exports reduce sender CVP even when the circuit breaker does not trip—analog to high burst pressure and stranger-tax distrust (Chapter 4, §4.5).

#### 6.5.4 Safe admission

Only if all gates pass:

```text
target.load += request_load
```

**Note:** The harness collapses PreFlight, GovernanceHeader, FSM epochs, and gossip into **scalar load + CVP** updates. It preserves **control-law ordering** (reject before accumulate), not wire fidelity. Full stack reproduction requires sidecar integration tests and open-mesh deployment profiles—out of scope for this abstract simulator.

---

### 6.6 Reported Results

#### 6.6.1 v1 headline (Monte Carlo mean at horizon)

From Chapter 1 / README empirical summary (1,000 runs × 500 nodes × 5% malicious × 100 epochs):

| Mode | Mean survival at T=100 |
|------|------------------------|
| **Baseline** | ~**0.4%** (≈ 2 of 500 nodes alive on average) |
| **AFP** | **100%** (500 of 500 nodes alive on average) |

Exact epoch-by-epoch curves are produced by running:

```bash
go run ./cmd/demo/simulator/
```

Output format: per-epoch table of mean alive count and mean load for Baseline vs AFP.

#### 6.6.2 Qualitative trajectory

Typical paired-run behavior:

| Phase | Baseline | AFP |
|-------|----------|-----|
| Early epochs | Load accumulates on high-degree targets | Malicious exports rejected or penalized; load bounded |
| Mid horizon | Cascade deaths from saturation | CVP floor isolates toxic sources |
| Late horizon | Near-total collapse (~0.4% survivors) | Survival sustained at 100% under reference parameters |

The gap is **sharp**, not incremental—a phase transition consistent with **positive feedback without brakes** (Baseline) vs **negative feedback with persistent gating** (AFP).

---

### 6.7 Mapping Harness to Protocol Chapters

| Harness element | Protocol mechanism |
|-----------------|-------------------|
| Load decay per epoch | Substrate cooling / epoch boundary (abstract) |
| `load + request > 0.95` reject | Entropy circuit breaker (Ch.3 §3.3.2) |
| `cvp < 0.3` reject | CVP floor + FSM isolation (Ch.4 §4.2) |
| Malicious asymmetric CVP penalty | ACC destabilization + malicious spike (Ch.3–4) |
| `request_load > 0.5` CVP penalty | Burst / stranger-tax analog (Ch.4 §4.5) |
| No gating in Baseline | Semantic sufficiency boundary (Ch.1 §1.2) |

**Lemma 6.1 (Abstraction soundness):** Every AFP-mode rejection rule in the harness corresponds to a **pre-accumulation** gate in the normative stack—no rule punishes nodes only after fatal load is already applied.

---

### 6.8 Claims Boundary

#### 6.8.1 What this experiment demonstrates

1. **Existence of collapse:** Ungoverned random load exchange in a 500-node mesh with 5% malicious exporters can drive mean survival to near zero within 100 epochs.
2. **Existence of defense:** A minimal CVP + circuit-breaker + burst-penalty law sustains 100% survival under **identical seeds and topology**.
3. **Sharpness of mechanism:** Friction is not cosmetic; without it, the mesh dies; with it, the mesh survives in the reference model.

#### 6.8.2 What this experiment does not prove

| Limitation | Status |
|------------|--------|
| **Formal liveness / safety proof** | Not claimed; TLA+ backlog ([`ROADMAP.md`](../../ROADMAP.md)) |
| **Wire-faithful sidecar replay** | Harness is abstract; full LV + GovernanceHeader path not simulated |
| **Gossip / core-relay dynamics** | Not modeled; isolation is instantaneous CVP scalar |
| **PreFlight / ReportInternalState ordering** | Collapsed into load scalar |
| **Optimal α, β, γ, δ, τ_stranger** | Reference constants, not tuned for real workloads |
| **Generalization beyond reference parameters** | 500 nodes, 5% malicious, specific load table—sensitivity analysis is future work |

**Theorem 6.1 (Empirical scope — stated modestly):** The Monte Carlo harness provides **supporting evidence** that physical admission control sustains mesh survival under the reference adversary model; it is **not** a proof that all optimizer networks satisfy Conjecture 4.1 for all adversaries.

#### 6.8.3 Relationship to v1 publication

Whitepaper v1 (Zenodo) documented the survival gap for external replication. v2.0 Protocol Edition **reframes** the same numbers inside the six-layer stack, dual-path SEA, and CVP formalism—so empirical results **attach to mechanism**, not marketing.

---

### 6.9 Reproduction Protocol

Conforming reproduction SHOULD:

1. Clone reference implementation repository.
2. Run `go run ./cmd/demo/simulator/` without modification to constants.
3. Verify 1,000-run aggregate at T=100: Baseline mean alive ≪ AFP mean alive ≈ 500.
4. Optionally pair with `scripts/verify_modes.sh` for sidecar **profile** behavior (closed mesh vs open exchange)—orthogonal to Monte Carlo abstract harness but validates ingress path divergence (Chapter 4, §4.4).

Report: seed policy, hardware, Go version, full epoch table, and any constant deviations.

---

### 6.10 Open Problems

| Problem | Connection |
|---------|------------|
| **Sensitivity to MaliciousRate** | At what fraction does AFP-mode survival degrade? |
| **Scale at 10⁴–10⁶ nodes** | Gossip relay vs global CVP store |
| **Adaptive adversary** | Attackers that fragment load below 0.5 per hop |
| **Attestation game** | False `topology_consensus_hash` once crypto is closed (Ch.5 §5.6) |
| **Economic collateral** | Virtual vs slashed stake equilibria |

These are research extensions, not v1.0 blockers for the **existence** claim.

---

### 6.11 Document Conclusion — Order Proven Under Reference Chaos

The Optimization Crisis (Chapter 1) diagnosed layer mismatch. CPL and SEA (Chapter 2) supplied persistent physics. PreFlight (Chapter 3) supplied timing. Open topology (Chapter 4) supplied distrust. GovernanceHeader (Chapter 5) supplied wire law.

Monte Carlo does not replace that theory. It **anchors** it:

> **Without physical admission control, the mesh dies. With it, the mesh survives—under the reference adversary, at the reference scale, reproducibly.**

The naked optimizer question from Chapter 1 now has a full stack answer:

> **Who governs the optimizer before it optimizes?**

**SEA does—pre-intent, with persistent consequences, attested on the wire, trusted under distrust, and empirically survival-load-bearing.**

Semantic signaling remains necessary. ASP remains load-bearing. AFP remains the brake.

---

---

*AFP Whitepaper v2.0 · Protocol Edition · Draft v0.3 · Strategic separation from enterprise deployment documentation.*
