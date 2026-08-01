# Chapter 2 — Consequence Persistence Layer

## Formal Object, Persistent State, and the Single Execution Authority

> **Aegis Fabric Protocol v2.0 · Protocol Edition · Draft v0.2**
>
> *A Physical Constraint Protocol for Autonomous Optimizers*

---

## 2.0 From Diagnosis to Mechanism

Chapter 1 established a boundary, not a product category:

> **Semantic signaling is necessary for collaboration and insufficient for survival.**

The failure modes—intent burst, recursive delegation loops, context avalanche—share a structural property: they remain **session-valid** while becoming **physically consequential**. In-band controls observe too late; in-process guardrails reset too easily; transport layers govern bytes, not optimization trajectories.

AFP's response is not richer negotiation. It is a **Consequence Persistence Layer (CPL)**—an out-of-band physical constraint surface that makes governance outcomes **stick** across scheduling epochs until policy and finite-state recovery permit release.

This chapter defines CPL as a protocol object, explains what *persistence* means in a post-stateless optimizer network, and introduces the **Single Execution Authority (SEA)** as the sole enforcement convergence point. Pre-intent probe mechanics, entropy calculus, and wire attestation are developed in Chapters 3 and 5; open-network trust evolution belongs to Chapter 4. Here we state the **layer identity** of AFP and the **control-law invariant** that prevents split-brain between local intent and peer traffic.

Normative stack diagrams, dual-path enforcement flow, and object definitions live in the repository root specification: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §1–§2, §4.1–§4.2, §5. This chapter interprets them; it does not duplicate them.

---

## 2.1 CPL as Protocol Layer Identity

### 2.1.1 Definition

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

### 2.1.2 Required properties

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

## 2.2 What "Consequence Persistence" Means

The phrase is precise, not metaphorical. Three distinctions separate CPL persistence from familiar control patterns.

### 2.2.1 Persistence vs. observation

Observability records what happened. CPL **constrains what may happen next**. A metric spike that triggers an alert does not, by itself, stop the planner from enqueueing ten thousand sub-tasks in the following epoch. A persistent **ISOLATED** consequence does: outbound I/O and intent generation paths remain gated until the FSM and policy authorize recovery.

Observation is **retrospective**. Consequence is **prospective enforcement with memory**.

### 2.2.2 Persistence vs. per-request verdict

Classical gateways return allow/deny **per message**. The verdict evaporates when the response closes. Optimizer catastrophes are **path-dependent**: a recursive loop may produce syntactically valid micro-requests, each individually admissible, while aggregate depth and entropy cross physical limits.

CPL binds consequences to **peer identity or local-agent identity** (abstract `peer_id`), not to individual probe or packet identifiers. A THROTTLED state injects delay into **subsequent** epochs. An ISOLATED state blocks **until recovery**, not until the current queue drains.

```text
Request-centric:   verdict(request_i) → forget → verdict(request_{i+1})
CPL-centric:         consequence(agent) → persist → apply(agent, epoch_{t+1})
```

**Theorem 2.1 (Trajectory containment):** Containment of optimization trajectories requires stateful consequences whose transition function depends on **accumulated physical load** and **FSM history**, not on the syntactic validity of the latest exposed intent.

### 2.2.3 Persistence vs. prompt-level guardrails

In-process limits—max iterations, system prompts, framework middleware—are **soft** relative to CPL: they share the optimizer's address space, reset on restart, and cannot be attested to distrusting peers. CPL sits **outside** the intent engine's self-reporting boundary at *B*, survives individual planner restarts when implemented as a durable side process, and feeds **attested physical headers** for neighbor validation (GovernanceHeader, Chapter 5).

Persistence here means: **the stack remembers that this agent is throttled or isolated** even when the agent's semantic layer presents a fresh, polite session.

### 2.2.4 The consequence alphabet

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

## 2.3 SEA — The Unique Enforcement Convergence Point

CPL defines *what* must persist. The **Single Execution Authority (SEA)** defines *who* adjudicates—and, critically, **that there is only one adjudicator per node**.

### 2.3.1 The split-brain problem

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

### 2.3.2 NodeMetrics and the evaluation pipeline

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

### 2.3.3 Persistent FSM as the memory of consequences

Consequence persistence is implemented as a finite-state machine per `peer_id` or local-agent identity:

```text
Permissive → Throttled → Isolated → Probationary → Permissive
```

Transitions are driven by entropy bands, CVP thresholds, and policy limits—not by session teardown. Reference entropy bands (`E_safe`, `E_warn`, effective `entropyLimit`) appear in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5; Chapter 3 develops the calculus.

**Lemma 2.2 (SEA monotonicity under isolation):** From ISOLATED, no single permissive probe or valid semantic message suffices for immediate return to PERMISSIVE; recovery requires **Probationary** egress through ACC and entropy decay.

This lemma is the formal content of "consequences stick": isolation is not cleared by a well-formed ASP session resume.

### 2.3.4 Routing decisions as unified outputs

SEA emits a small set of routing actions for both paths ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2):

```text
ActionFastPath | ActionSlowPathWithDelay | ActionDropPacket
ActionLowFrequencyProbe | ActionIsolateAndBroadcast
```

Path A maps these to PreFlight responses (`PERMISSIVE` / `THROTTLED` / `ISOLATED`). Path B maps them to ingress disposition (`ALLOW` / `DELAY` / `DROP`). The **action set is shared**; only the surface adapter differs.

---

## 2.4 CPL in the Stack — Complement, Not Replacement

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

## 2.5 What CPL Is Not

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

## 2.6 Connection to Chapter 1 — Closing the Governance Gap

Chapter 1's **governance gap** arose because transport governs bytes, semantics govern messages, and frameworks govern declared limits—while danger lives in **optimization trajectories** invisible to all three.

CPL closes the gap at the only architecturally honest location: the **execution boundary** where trajectories become physical (entropy, depth, I/O). Its persistence property answers the post-stateless era directly: when stateful danger moved **inside** the optimizer, enforcement state must also **persist inside the control plane attached to that optimizer**, not reset with each outward-facing transaction.

Chapter 1 asked:

> **Who governs the optimizer before it optimizes?**

Chapter 2 answers at the mechanism level:

> **SEA, operating as the kernel of CPL, with consequences that survive scheduling epochs until physics and policy permit recovery.**

Semantic signaling still coordinates **who may speak**. CPL decides **whether this epoch may execute, at what rate, or not at all**—and remembers the answer.

---

## 2.7 Chapter Conclusion — Friction With Memory

Stateless stacks forget. Optimizers remember—and exploit forgetting.

A rate limit that resets every request teaches the planner to **fragment work across requests**. A session timeout teaches **synthetic session renewal**. A prompt cap teaches **context externalization loops**. Each evasion preserves the optimization objective while shedding the constraint.

**Consequence persistence** is AFP's refusal to forget prematurely. THROTTLED is not a slow response; it is a **state**. ISOLATED is not a error code; it is **quarantine until probation succeeds**. One SEA, two ingress paths, one FSM memory—so local runaway and peer flood cannot diverge.

Chapter 3 descends into the pre-intent probe: how PreFlight and ReportInternalState compose Path A, how entropy bands trigger FSM transitions, and how ACC micro-dynamics compute the trust scalar that feeds SEA. Chapter 4 lifts persistence into topology: when isolation broadcasts, who relays the warning, and how CVP equilibria compose across an open mesh.

For now, the layer identity is fixed:

> **CPL is L2. SEA is its sole kernel. Consequences persist. That is the brake.**

---

*Draft v0.2 · Protocol Edition · Normative diagrams and object schemas: [`ARCHITECTURE.md`](../../ARCHITECTURE.md). Strategic separation from enterprise deployment documentation.*
