# Chapter 3 — Pre-Intent Enforcement

## Path A, Entropy Calculus, and the Microscopic Control Loop

> **Aegis Fabric Protocol v2.0 · Protocol Edition · Draft v0.2**
>
> *A Physical Constraint Protocol for Autonomous Optimizers*

---

## 3.0 From Persistent Consequences to Pre-Intent Gates

Chapter 2 fixed **where** consequences live: L2, adjudicated by a **Single Execution Authority (SEA)**, remembered across scheduling epochs until FSM recovery permits release. That answer is necessary but incomplete. Persistence without **timing** still loses to optimizers that externalize intent in microseconds—one scheduling cycle is enough (Chapter 1, Lemma 1.1).

**Pre-intent enforcement** is AFP's timing contract:

> Adjudicate **before** irreversible intent generation or outbound I/O—not after HTTP failure, not after semantic session teardown, not after the internal queue has already forked ten thousand sub-tasks.

Path A implements this contract locally. A synchronous **PreFlight** probe asks SEA for a consequence; a companion **ReportInternalState** write path feeds ground-truth metrics into the **EntropyMonitor** before evaluation. The microscopic loop—EntropyMonitor → NodeMetrics → ACC kernel → Node FSM → consequence—is specified in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.3, §5. This chapter develops the theory; it references that document for stack placement and object schemas.

Path B (GovernanceHeader ingress) applies the **same kernel** to peer traffic and is developed in Chapters 4–5. Here we govern **self**—the optimizer attached to this node's execution boundary.

---

## 3.1 The Pre-Intent Timing Contract

### 3.1.1 Definition

**Definition 3.1 (Pre-Intent).** An enforcement action is **pre-intent** iff it is evaluated at execution boundary *B* **before** either:

1. the planner commits to a new externally visible intent (tool call, delegation, outbound message), or
2. the runtime performs irreversible cross-boundary I/O.

Post-intent inspection—parsing emitted RPC, scanning response bodies, correlating logs after the fact—is **observation**, not CPL enforcement (Chapter 2, §2.2.1).

### 3.1.2 Synchronous probe semantics

PreFlight is **synchronous** relative to the planner's scheduling decision: the optimizer MUST NOT proceed until SEA returns a PreFlightResponse. This is deliberately unlike async audit pipelines. The probe is a **hard gate**, not a hint.

**Lemma 3.1 (Probe ordering):** If ReportInternalState is omitted or stale, entropy at probe time is **under-measured**, not over-measured. The protocol assumes honest runtimes report recursion depth and context volume; dishonest under-reporting is bounded by L0/L1 os-level pressure signals in EntropyMonitor (§3.3).

Corollary: SDK integration is not optional politeness—it is part of the **measurement substrate** at *B*.

### 3.1.3 What PreFlight does not inspect

PreFlight does not parse natural language, tool schemas, or ASP session state. It evaluates **physical load**—entropy, depth, burst hints, persistent FSM state—not semantic intent. This is structural, not a limitation to be patched with richer NLP classifiers:

> The contested resource is **optimization capacity**, not message vocabulary.

Semantic protocols remain load-bearing for **what** agents negotiate (L5). PreFlight governs **whether this epoch may execute at all**.

---

## 3.2 Path A — PreFlight and ReportInternalState

Path A is the local IPC interface from optimizer runtime (L1) to CPL (L2). Dual-path convergence is diagrammed in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §2; Path A occupies the **Govern Self** subgraph.

### 3.2.1 PreFlightRequest

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

### 3.2.2 PreFlightResponse

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

### 3.2.3 ReportInternalState — the companion write path

```text
ReportInternalState(recursion_depth, context_memory_bytes)
```

Invoked by the SDK **before** PreFlight when the planner's internal graph changes. Feeds EntropyMonitor with:

- **recursion_depth** — control-flow depth in the planner graph (Chapter 1, §1.3.2);
- **context_memory_bytes** — monotonic context volume (Chapter 1, §1.3.3).

**Definition 3.2 (Write-then-probe discipline).** A conforming runtime SHOULD call ReportInternalState immediately before each PreFlight when internal state has changed since the last report. SEA evaluates against the **latest** stored values plus L0 pressure.

This write path is how AFP sees **un-exposed planning** without requiring the planner to emit semantic descriptions of its internal graph.

---

## 3.3 Entropy Calculus

Entropy is the scalar summary of physical pressure at *B*. It is locally measured, normalized to ∈ [0, 1], and **never trusted from peer headers alone** on ingress (local Path A uses only local signals).

### 3.3.1 EntropyMonitor inputs

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

### 3.3.2 Entropy bands and policy limit

Reference band constants ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §5):

| Threshold | Value | Effect |
|-----------|-------|--------|
| `E_safe` | 0.40 | FSM recovery toward Permissive |
| `E_warn` | 0.75 | Throttled path; injected delay |
| Effective limit | policy `entropyLimit` (default 0.95) | Circuit breaker → ISOLATED |

**Definition 3.3 (Entropy circuit breaker).** When `entropy_load ≥ entropyLimit`, SEA MUST emit ISOLATED **before** FSM soft transitions—a hard pre-intent stop independent of current FSM state.

The limit is supplied by L3 Effective Policy (durable base law plus runtime overlay). An emergency overlay kill switch MAY force ISOLATED irrespective of measured entropy—a fleet clamp that persists until overlay revision (Chapter 2, §2.4; policy surface in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §3).

### 3.3.3 Recursion depth as a discrete breaker

Separate from continuous entropy, **recursion_depth** from ReportInternalState is checked against policy `maxRecursionDepth`:

```text
recursion_depth > maxRecursionDepth  ⇒  ISOLATED (loop detected)
```

This is the pre-intent answer to **recursive delegation loops** (Chapter 1, §1.3.2): topologically closed control flow need not crash the runtime or trip transport timeouts, but it **must** trip the depth breaker at *B* before the next externalization.

**Lemma 3.2 (Depth precedence):** Recursion depth violation triggers ISOLATED **without** requiring entropy_load to exceed `E_warn`.

---

## 3.4 ACC Kernel — Trust Dynamics Inside SEA

The **ACC (Adaptive Coordination Calculus) kernel** is a stateless mathematical layer inside SEA. It transforms historical trust, throughput evidence, destabilization penalties, and entropy into an updated **CVP score** ∈ [0, 1] (Coordination Viability Probability).

For Path A pre-intent enforcement, CVP often starts at maximum for the local agent—but the **same kernel** evaluates both self and peers (Chapter 2, §2.3.1). Open-network CVP evolution, decay, gossip relay, and stranger tax are Chapter 4. Here we state the formulas that SEA applies uniformly.

### 3.4.1 Formula A — CVP evolution

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

### 3.4.2 Formula B — Anti-ossification decay

```text
CVP_effective = CVP_historical · e^(−λ · Δt)
```

Trust **decays with idle epochs** so stale reputation cannot ossify. λ is a global decay constant (reference: 0.01 per epoch).

### 3.4.3 Formula C — Asymmetric hysteresis recovery

```text
CVP_recovery = CVP_critical + κ · log(1 + Δt_probation)
```

Recovery from the critical floor is **logarithmic**, not linear—probation earns trust slowly. Reference: `CVP_critical = 0.3`, κ = 0.02.

**Hard floor (protocol law):** `CVP_score < 0.3` ⇒ mandatory isolation, same FSM path as malicious spike.

### 3.4.4 ACC's role in Path A

On each PreFlight, SEA assembles NodeMetrics with locally measured `entropy_load`, current epoch, and CVP (1.0 for local agent unless degraded by prior epochs). ACC updates CVP when throughput and penalty signals exist; FSM consumes the metric bundle **including** CVP floor checks.

**Corollary 3.1:** ACC is not an A/B testing layer or application analytics kernel. It is **infrastructure control math**—stateless, shared, mandatory.

---

## 3.5 FSM Micro-Dynamics

Consequence persistence (Chapter 2) is realized as a **Node FSM** per identity—`local-agent` for Path A, `peer_id` for Path B. States:

```text
Permissive → Throttled → Isolated → Probationary → Permissive
```

The FSM diagram in [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2 lists routing decisions; §5 shows the control loop. This section states transition law.

### 3.5.1 Global floor rules (all states)

Before state-specific logic, SEA evaluates floor conditions:

```text
¬has_valid_sign  ∨  malicious_spike  ∨  CVP_score < CVP_critical
  ⇒  enforceIsolation()
```

**Malicious spike** on Path A includes burst hints that exceed policy concurrency—`estimated_tasks` greater than permitted fan-out is treated as divergence, not optimism.

Isolation from floor rules MAY emit **ActionIsolateAndBroadcast** on first transition (topology warning in Chapter 4). Local PreFlight maps this to ISOLATED.

### 3.5.2 State Permissive

```text
entropy_load > E_warn   ⇒  Throttled, ActionSlowPathWithDelay
else                    ⇒  Permissive, ActionFastPath
```

Entry from Permissive to Throttled is the **first soft brake**—entropy crossed the warning band but not the circuit breaker.

### 3.5.3 State Throttled

```text
entropy_load < E_safe   ⇒  Permissive, ActionFastPath
else                    ⇒  remain Throttled, ActionSlowPathWithDelay
```

**Hysteresis:** Recovery requires entropy **below** `E_safe` (0.40), not merely below `E_warn` (0.75). This prevents oscillation at the boundary—classic control-theoretic dead band.

Injected delay scales with excess entropy above `E_warn` (reference: 500 ms at warn threshold to 2000 ms at saturation). THROTTLED is therefore **two** mechanisms: FSM state persistence **and** per-probe delay injection.

### 3.5.4 State Isolated

```text
epoch − last_penalty < k_isolation   ⇒  ActionDropPacket / ISOLATED
epoch − last_penalty ≥ k_isolation   ⇒  Probationary, ActionLowFrequencyProbe
```

Reference hysteresis: `k_isolation = 64` epochs before probation entry. Isolation is not cleared by a single low-entropy probe—**time at the penalty epoch** must elapse.

**Lemma 3.3 (Isolation monotonicity):** Restated from Chapter 2, Lemma 2.2—no well-formed PreFlightRequest alone restores Permissive from Isolated.

### 3.5.5 State Probationary

```text
entropy_load > E_safe           ⇒  enforceIsolation()  // zero tolerance spike
Δt_probation > k_probation
  ∧ CVP_score ≥ 0.8             ⇒  Permissive, ActionFastPath
else                            ⇒  ActionLowFrequencyProbe / THROTTLED
```

Reference: `k_probation = 128` epochs. Probation is **low-frequency probe** mode—intent generation damped, CVP recovers via Formula C.

Any entropy above `E_safe` during probation re-isolates immediately. Recovery to Permissive requires **both** sustained low entropy **and** restored trust.

### 3.5.6 FSM + ACC composition

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

## 3.6 Mapping Routing Decisions to PreFlight Consequences

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

## 3.7 Pre-Intent Enforcement vs. Chapter 1 Pathologies

The three structural pathologies from Chapter 1 are not listed as bugs to patch—they are **default optimizer behaviors**. Pre-intent enforcement maps each to a concrete gate:

| Pathology | Pre-intent counter |
|-----------|-------------------|
| **Intent burst** | `estimated_tasks` in burst pressure; tool concurrency counter; THROTTLED with delay |
| **Recursive delegation loop** | ReportInternalState `recursion_depth`; hard ISOLATED at `maxRecursionDepth` |
| **Context avalanche** | ReportInternalState `context_memory_bytes`; L0 memory pressure in entropy max |

**Theorem 3.2 (Pre-intent containment sketch):** For any planner epoch where physical load exceeds policy thresholds, a conforming CPL implementation with write-then-probe discipline emits THROTTLED or ISOLATED **before** the epoch externalizes intent—provided thresholds are calibrated below catastrophic substrate saturation.

This is a **sketch**, not a liveness proof. Chapter 6 reproduces survival empirically. The mechanism claim here is architectural: gates exist at the correct timing boundary with persistent FSM memory.

---

## 3.8 Integration Obligations (Abstract)

Conforming optimizer runtimes at *B* MUST:

1. Invoke **ReportInternalState** when recursion depth or context volume changes materially.
2. Invoke **PreFlight** synchronously before each intent generation epoch (or per policy-scoped batch).
3. Honor **THROTTLED** delay and **ISOLATED** halt without bypass via alternate I/O paths.
4. Treat PreFlight as authoritative over in-process iteration counters and prompt-level guardrails.

Wire contract reference: `api/afp/v1/sdk_ipc.proto` · Unix domain socket · gRPC ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.3). Normative RPC framing is deployment detail; the **timing contract** is protocol law.

---

## 3.9 Chapter Conclusion — The Gate Before the Graph

Chapter 2 gave CPL **memory**. This chapter gives it **timing**.

Before the planner graph forks, before the tool chain materializes, before bytes seek a socket—SEA asks one question grounded in physics:

> **Given current entropy, depth, burst, trust, and persistent FSM state—may this epoch execute?**

The answer is not a suggestion. It is PERMISSIVE, THROTTLED with delay, or ISOLATED with reason. ReportInternalState supplies honesty at the measurement boundary; EntropyMonitor aggregates substrate truth; ACC couples trust to pressure; the FSM remembers.

Path A governs self. Path B—GovernanceHeader ingress, attestation rules, LV framing—applies the identical kernel to neighbors. Chapter 4 lifts the FSM's **ActionIsolateAndBroadcast** into open topology: CVP gossip, relay sets, stranger tax, equilibrium under distrust. Chapter 5 specifies the wire.

For now:

> **Pre-intent is not early intent review. It is the last gate before physics pays the bill.**

---

*Draft v0.2 · Protocol Edition · Normative object schemas and control-loop diagram: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.3, §5. Strategic separation from enterprise deployment documentation.*
