# Chapter 4 — Open-Network Topology

## CVP, Trust Under Distrust, and Topological Quarantine

> **Aegis Fabric Protocol v2.0 · Protocol Edition · Draft v0.2**
>
> *A Physical Constraint Protocol for Autonomous Optimizers*

---

## 4.0 From Local Gates to Mesh Safety

Chapters 2–3 established **local** control law: persistent consequences at the execution boundary, pre-intent probes on Path A, identical ACC + FSM kernel on every node. That suffices when every optimizer shares a trusted administrative perimeter—when neighbors are known, identities are pre-bound, and physical enforcement alone prevents runaway trajectories.

Chapter 1's scope statement named a harder problem:

> **Mutually distrusting optimizers, no central moral authority, equilibrium under attack.**

Local brakes do not compose automatically. A node may isolate its own planner while admitting a peer whose CVP has collapsed. A malicious optimizer may present polite ASP sessions and toxic physical headers. Contagion crosses the mesh **before** any single node's PreFlight observes local entropy spike.

**L4 · Open Topology** is AFP's answer at the trust layer: a scalar **Coordination Viability Probability (CVP)**, evolutionary dynamics under ACC, **topological quarantine** when peers breach physical law, and **gossip** that propagates isolation warnings through a high-trust relay set—not through the entire mesh indiscriminately.

Path B—GovernanceHeader ingress before business payload—is the wire interface from L4 into SEA (Chapter 3, §3.9). This chapter develops **trust semantics** and **mesh dynamics**. Frame layout, LV prefixing, and field-level attestation rules are Chapter 5. Normative object definitions: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4–§4.5, §2 (Path B).

---

## 4.1 L4 as the Trust Layer

### 4.1.1 Layer accountability

In the six-layer stack ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §1), each layer owns one question:

| Layer | Question |
|-------|----------|
| **L5** | What exposed intents may agents negotiate? |
| **L4** | Which peers may coordinate safely under distrust? |
| **L3** | What durable law and emergency overlay govern thresholds? |
| **L2** | What physical consequences persist at the boundary? |

L4 does **not** replace L2. It supplies **trust inputs**—CVP scores, neighbor stores, collateral requirements—that SEA consumes when evaluating Path B ingress. The FSM and ACC kernel remain on L2; L4 feeds them peer-scoped context.

**Definition 4.1 (Open Topology).** An **open optimizer network** is a mesh of autonomous nodes where peer identity is cryptographic (DID), trust is **local and evidential**, and no single scheduler adjudicates global morality. L4 is the protocol stratum that makes such meshes **survivable** rather than merely connectable.

### 4.1.2 What L4 is not

| Misclassification | Failure mode |
|-------------------|--------------|
| **Central reputation service** | Violates distrust assumption; single point of capture |
| **Semantic trust framework** | Operates on attested **physical** state, not conversational politeness |
| **Replacement for ASP** | ASP coordinates exposed tasks (L5); L4 gates **admission** by viability |
| **Blockchain mandate** | Collateral MAY be virtual or on-chain; protocol specifies **slash semantics**, not ledger choice |

---

## 4.2 CVP — Coordination Viability Probability

### 4.2.1 Definition

**Definition 4.2 (CVP).** For peer *p* at node *n*, **CVP_n(p)** ∈ [0, 1] is *n*'s local estimate of the probability that *p* can participate safely in coordinated optimization—without imposing destabilizing entropy, recursion, or delegation harm on *n* or *n*'s neighbors.

CVP is **local**. Two nodes may disagree on the same peer's score; equilibrium emerges from coupled dynamics, not from a global oracle (§4.9).

CVP feeds NodeMetrics ( [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.2) and interacts with entropy_load in SEA evaluation. The **hard floor** is protocol law:

```text
CVP_score < CVP_critical (0.3)  ⇒  mandatory isolation
```

Same FSM path as malicious spike (Chapter 3, §3.5.1). Below 0.3, a peer is **coordination-bankrupt**—admission MUST fail regardless of semantic session validity.

### 4.2.2 Formula A — Evolution under evidence

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

### 4.2.3 Formula B — Anti-ossification decay

```text
CVP_effective = CVP_historical · e^(−λ · Δt)
```

Idle peers lose stale trust. Without decay, a node that behaved well in epoch 0 could coast indefinitely—a **reputation ossification** attack in long-horizon optimizer networks.

Reference: λ = 0.01 per epoch.

### 4.2.4 Formula C — Asymmetric hysteresis recovery

```text
CVP_recovery = CVP_critical + κ · log(1 + Δt_probation)
```

Recovery from the critical floor is logarithmic, not linear. Probation earns trust slowly (Chapter 3, §3.5.5). Reference: κ = 0.02.

**Theorem 4.1 (CVP–FSM coupling):** A peer in FSM state Isolated with CVP forced to 0 cannot reach Permissive until both **epoch hysteresis** (Chapter 3) **and** CVP recovery (Formula C, threshold ≥ 0.8 for full permissive egress) are satisfied.

Trust and physical state are **jointly necessary** for mesh re-admission.

---

## 4.3 Path B — GovernanceHeader as Physical Attestation

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

## 4.4 Protocol Profiles — Closed Mesh vs. Open Exchange

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

## 4.5 Stranger Tax — Dependency Collateral

### 4.5.1 Problem

In open exchange, a **first-seen peer** has no FSM history on the receiving node. Without friction, an unknown optimizer could flood ingress, impose entropy, and vanish—a **hit-and-run** delegation attack.

### 4.5.2 Mechanism

**Definition 4.3 (Stranger tax).** Before initializing FSM state for an unknown `peer_id`, SEA in open-exchange profile MUST verify:

```text
DependencyCollateral {
  collateral_type  : string
  slash_threshold  : float32   // minimum acceptable stake
}
```

Admission requires `slash_threshold ≥ τ_stranger` (reference: **τ_stranger = 0.8**).

Failure ⇒ reject with stranger-tax error; no FSM entry, no payload forwarding.

### 4.5.3 Virtual vs. on-chain stake

Reference implementations attach virtual stake (`collateral_type` e.g. `SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`) on egress. On-chain collateral is **architectural optional**—the protocol specifies slash semantics and minimum threshold, not ledger placement ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8, open gap footnote).

**Corollary 4.1:** Stranger tax is **admission economics**, not semantic KYC. It raises the cost of anonymous mesh flooding without requiring a central identity broker.

### 4.5.4 Collateral zeroing on isolation

When FSM `enforceIsolation` triggers, CVP for the peer is forced to **0.0**—collateral value at coordination bankruptcy. Re-entry requires probation, CVP recovery, and fresh stake under open profile rules.

---

## 4.6 Topological Quarantine

### 4.6.1 From local ISOLATED to mesh warning

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

### 4.6.2 Preemptive decay on warning receipt

When a node receives a validated TopologyWarning naming peer *p*, it applies **preemptive CVP decay** before *p*'s traffic triggers local isolation:

```text
CVP_n(p) ← CVP_n(p) × η_decay     // reference: η_decay = 0.5
```

**Intuition:** Neighbors learn of toxicity **out-of-band** and tighten admission proactively—containment spreads at gossip speed, not at the speed of the next malicious payload.

Signature verification on warnings is normative; cryptographic binding of `topology_consensus_hash` is an open specification gap (Chapter 5; [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

### 4.6.3 Probation at the mesh edge

Ingress probation uses **low-frequency probe** admission (Chapter 3, §3.5.5): reference rule—probe succeeds only when `current_epoch ≡ 0 (mod 10)`; otherwise reject. This throttles re-entry attempts from isolated peers across the mesh without silencing recovery entirely.

---

## 4.7 Gossip and the Core Relay Set

### 4.7.1 Why not broadcast to everyone?

Flooding isolation warnings to all nodes amplifies **gossip itself** as an attack surface—a malicious reporter could DDoS the mesh with fake warnings. AFP restricts relay to a **core set**:

```text
CoreRelay(n) = { p ∈ Neighbors(n) : CVP_n(p) ≥ τ_core }
```

Reference: **τ_core = 0.8** ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.5).

Only core relays receive TopologyWarning propagation. They apply preemptive decay and may re-gossip under the same rule—**trust-gated epidemic**, not blind flood.

**Lemma 4.3 (Relay monotonicity):** A node with CVP_n(p) < τ_core cannot act as relay for warnings about third parties; it may still **drop** traffic from isolated peers locally via FSM.

### 4.7.2 Async propagation

Gossip MUST NOT block the data-plane fast path. Broadcast is **asynchronous** relative to ingress DROP—local quarantine is immediate; mesh learning is eventual.

This matches the post-stateless era observation (Chapter 1): one scheduling cycle is enough for local harm; gossip races to inform neighbors **before** contagion composes across hops.

### 4.7.3 Neighbor store

Each node maintains a **Neighbor Store**—local map of `peer_id → { CVP, endpoint, core membership }`. Bootstrap seeds (genesis peers with initial CVP) MAY initialize the store; runtime upserts refine endpoints and scores.

DID resolution in open mesh fan-outs to core relays only, with bounded timeout—discovery under distrust without central DNS for optimizers.

---

## 4.8 Composing Local and Topological Control

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

## 4.9 Equilibrium Intuition — Survival Under Distrust

This section states **intuition**, not a closed-form proof. Chapter 6 reproduces survival empirically.

### 4.9.1 The v1 empirical anchor

Monte Carlo open-mesh simulation (reference harness: 500 nodes, 5% malicious, 100 epochs, 1,000 runs) showed baseline coordination collapsing to ~**0.4%** mean survivors while AFP-maintained topology sustained **100%** (Chapter 1, §1.6). Version 2.0 explains **mechanism**; it does not re-litigate the number.

### 4.9.2 Fixed points (sketch)

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

### 4.9.3 What equilibrium is not

AFP does not promise **maximum throughput**, **fair Shapley allocation**, or **truthful semantic revelation**. It promises **survival physics**—a mesh that remains operable under distrust long enough for L5 semantics to matter.

---

## 4.10 Connection to Prior Chapters

| Prior claim | L4 completion |
|-------------|---------------|
| Ch.1 — semantic signaling insufficient | CVP + collateral gate **admission**, not dialogue |
| Ch.1 — open protocol, no central authority | Local CVP, gossip relay, no global oracle |
| Ch.2 — consequences persist | FSM isolation persists per peer_id; gossip adds **mesh memory** |
| Ch.3 — same kernel Path A / B | ACC + FSM on ingress identical to PreFlight |
| Ch.3 — ActionIsolateAndBroadcast | Realized as TopologyWarning + core relay |

---

## 4.11 Chapter Conclusion — Trust as Physics, Not Politeness

ASP asks whether agents may coordinate on **exposed tasks**. L4 asks whether a **specific peer**, at this **epoch**, with this **attested physical state**, is viable enough to admit.

The answer is scalar, evidential, and local. It decays with idle time. It crashes through a hard floor. Strangers pay tax. Isolation propagates through trusted relays, not through hope.

Chapter 5 closes the wire: GovernanceHeader field semantics, LV framing, attestation evolution, and the binding between `topology_consensus_hash` and distrust. Chapter 6 returns to Monte Carlo with protocol-framed reproduction steps.

For now:

> **In an open optimizer mesh, trust is not belief—it is a control variable with decay, floor, and quarantine.**

---

*Draft v0.2 · Protocol Edition · Normative CVP and ingress rules: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4–§4.5, §2. Strategic separation from enterprise deployment documentation.*
