# Aegis Fabric Protocol v2.0 — Protocol Edition

> **GitHub edition · Draft v0.3** · *A Physical Constraint Protocol for Autonomous Optimizers*
>
> This is the repository whitepaper: narrative first, full names in prose, mechanisms explained in-line.
> Stack diagrams, protobuf schemas, and code map for implementers: [`ARCHITECTURE.md`](../../ARCHITECTURE.md)
>
> **v1** (empirical archive): [Zenodo 20674352](https://zenodo.org/records/20674352)
>
> **v0.3 notes:** Asymmetric finite-state-machine recovery dwell is normative in the reference implementation; inbound topology warnings MUST be ed25519-verified (unsigned hearsay discarded).

**Name and independence.** This protocol shares the name “Aegis” and a high-level interest in agent safety with certain recent academic publications. The engineering implementation, control-plane structure, wire contracts, and Adaptive Coordination Calculus defined here are **completely distinct and independent** of that literature. Shared branding or overlapping safety goals must not be read as shared architecture, shared theorems, or shared lineage.

---

## Read this first: names and what the short forms are short for

This whitepaper avoids acronym soup. The table below gives the **full name used in the body**, the short form you may see in code or older docs, and an explicit **“short for …”** expansion. **In the body of the paper, we use the full names**—so you do not need to memorize the abbreviations.

| Full name used in this paper | Short form | Short form is short for… | Meaning |
|------------------------------|------------|--------------------------|---------|
| **Aegis Fabric Protocol** | AFP | Aegis Fabric Protocol | This protocol. It brakes an agent *before* the agent acts, and remembers when something went wrong. |
| **Consequence Persistence Layer** | CPL | Consequence Persistence Layer | The core layer. Outcomes—allow, slow down, or isolate—do not vanish when a single request ends. |
| **Single Execution Authority** | SEA | Single Execution Authority | The one judge per node. Governing yourself and governing neighbors must use the same rules. |
| **Finite-state machine** | FSM | Finite-state machine | Remembers whether each agent or neighbor is currently allowed, slowed, isolated, or on probation. |
| **Coordination Viability Probability** | CVP | Coordination Viability Probability | A number from 0 to 1: “can this neighbor still work with us safely?” Below 0.3 → isolate. |
| **Adaptive Coordination Calculus** | ACC | Adaptive Coordination Calculus | The formulas that update Coordination Viability Probability from history, pressure, and bad behavior. |
| **Pre-flight check** | PreFlight | Pre-flight check (API name in the SDK) | Before an agent delegates or goes out on the wire, it must ask the Single Execution Authority: may this step run? |
| **Governance header** | GovernanceHeader | Governance header (protobuf message name) | The first frame sent to a neighbor. Prove physical condition *before* business content. |
| **System pressure** | entropy / `entropy_load` | Entropy load (historical name for system pressure) | A 0–1 summary of memory, concurrency, context size, and task burst. |
| **Argent Signaling Protocol** (and peers) | ASP | Argent Signaling Protocol | Protocols for “what task shall we do?” Traffic lights—not brakes. |

**Protocol layers (remember the jobs, not layer numbers):**

| Layer | Job |
|-------|-----|
| Physical substrate | Real compute, memory, network |
| Optimizer runtime | The planner / agent framework itself |
| **Consequence Persistence Layer (core)** | May this run, how fast, isolate or not—and *remember* |
| Policy surface | Operator thresholds and emergency kill switch |
| Open topology | How open networks score neighbors and quarantine toxic peers |
| Semantic collaboration | Task content and session negotiation (not this paper’s focus) |

Two enforcement paths:

- **Path A (govern self):** before the local planner acts, run a pre-flight check.
- **Path B (govern neighbors):** before admitting foreign traffic, validate the governance header.

Both paths MUST share one Single Execution Authority, one finite-state machine, and one effective policy.

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

### 1.1 The empty cell in the market

Multi-agent infrastructure has matured along two axes that are now relatively well understood:

1. **Plan** — frameworks such as LangGraph, CrewAI, and AutoGen decompose work, delegate, and replan.
2. **Move messages** — transports such as gRPC, HTTP, Kafka, and MCP deliver bytes.

A third cell remains empty. Before each coordination step commits, something must answer:

> May this coordination step execute? If not, does the failure leave a lasting mark?

| Layer | Representative tools | What is missing |
|-------|---------------------|-----------------|
| Planning | LangGraph, CrewAI, AutoGen | Decomposition without a physical brake |
| Transport | gRPC, HTTP, Kafka, MCP | Bytes arrive; an internal task storm may never touch the wire |
| Semantic collaboration | Argent Signaling Protocol, A2A | Negotiates *exposed* intent—traffic lights, not brakes |
| **Coordination runtime** | **Aegis Fabric Protocol** | Check before action, remember consequences, enforce the same physics on neighbors |

Aegis Fabric Protocol does not compete with transports or workflow engines. It occupies the gap between *agents can talk* and *coordination still survives*. Resource and token allocation under this architecture are driven strictly by **physical constraints** measured at the execution boundary and by the **Adaptive Coordination Calculus**. The protocol does **not** use bidding, competitive auctions, or chip-based allocation mechanisms.

### 1.2 Enterprise fleets vs open meshes

The same protocol serves two deployment postures. The control law does not change; the threat model and emphasis do.

| | **Enterprise multi-agent** | **Open peer mesh** |
|--|---------------------------|---------------------|
| Trust | Known identities inside an admin boundary | Mutual distrust; strangers may connect |
| Main risk | Local planner runaway—loops, bursts, out-of-memory | Contagion across nodes; malicious or overloaded peers |
| Protocol emphasis | Govern yourself first: pre-flight check, internal state reports, local system pressure | Govern yourself *and* neighbors: governance header, Coordination Viability Probability, stranger tax, quarantine gossip |
| Takeaway | Brake your own optimizer before it burns the fleet | Brake yours *and* contain toxic neighbors |

**Enterprise posture, concretely.** Consider a research desk that runs a financial analysis agent inside a closed administrative boundary. The agent is asked for a sector brief. It decomposes the brief into ticker-level workstreams, each of which fans out into filings retrieval, ratio computation, and peer comparison. None of those micro-steps need leave the host as wire traffic; the in-process queue can still grow until memory pressure trips the node. In this posture, Path A—pre-flight check, internal state report, local system pressure—is the primary brake. Neighbors are known; the failure mode is usually your own planner burning the fleet.

**Open-mesh posture, concretely.** Now place a similar agent on a public coordination surface where strangers may connect. A peer can present a polite semantic session, export destabilizing load, and vanish—hit-and-run. Contagion reaches other nodes before any single local pre-flight check observes a pressure spike. Path A remains mandatory, but Path B—governance header validation, Coordination Viability Probability, stranger tax, quarantine gossip—must also run under the same Single Execution Authority.

Transport is solved. Semantics are still evolving. **Runtime coordination law is not.** Aegis Fabric Protocol supplies that law as enforceable physics at the execution boundary.

### 1.3 Why stacking is the failure geometry

The dominant failure geometry in multi-agent systems is **agent stacking**: planner → A → B → C → …, sometimes closing a loop A → D → F → A.

Risk grows with depth, branching, context size, and (in open meshes) how easily neighbors get dragged under. Each hop can look like a perfectly valid session or remote call while the *trajectory* as a whole becomes physically unsustainable.

Per-request gateways forget. Optimizers learn to fragment work into many syntactically legal micro-steps that dodge one-shot limits.

The Consequence Persistence Layer binds outcomes to an agent identity, not to a single request:

```text
Old way:  judge this request → forget → judge the next
New way:  remember this agent’s state → enforce it again next epoch
```

| Class | Examples | How Aegis Fabric Protocol responds | Out of scope |
|-------|----------|-------------------------------------|--------------|
| Out-of-control stacking | Intent burst, recursion loop, context avalanche | Pre-flight check, depth breaker, pressure breaker, persistent finite-state machine | Whether the business task is “correct” |
| Physical malice | Peer flood, stranger hit-and-run, trust collapse | Governance header, Coordination Viability Probability floor, stranger tax, quarantine broadcast | Moral judgment of message content |
| Permitted stacking | Multi-hop delegation inside policy | Allow when physics stay sub-critical | Org-chart routing and approval workflows |

Blocking A → D → F → A is recursion containment, not a ban on collaboration. Blocking a polite peer whose Coordination Viability Probability is below 0.3 is **coordination bankruptcy**, not a critique of conversational tone.

### 1.4 How often this shows up in production

| Risk | Enterprise multi-agent | Open peer mesh |
|------|------------------------|----------------|
| Out-of-control stacking | **Common** — decompose, delegate, replan is the default | **Common** — local runaway plus contagion |
| Physical malice | Uncommon as peer adversaries; abuse and misconfig dominate | **Material** — strangers, floods, hit-and-run |

Failures appear as episodic spikes (one loop, one burst) or as structural cost drift (always-on fleets with no runtime brake). Aegis Fabric Protocol targets the **scale law of stacking**, not a claim that agents are malicious every day.

> Multi-agent systems need runtime law because stacked optimizers outrun retries, token dashboards, and ephemeral limits—not because attackers are everywhere.

### 1.5 Why retries, token budgets, and dashboards are not enough

Teams routinely deploy `max_retries`, recursion caps, framework limits, and token budgets. Those tools are necessary. They are not sufficient as the sole coordination runtime:

| Pattern | What it helps | Structural gap |
|---------|---------------|----------------|
| Retry / loop counters | Cap iterations inside one graph | Reset when the session ends; no cross-agent chain memory; one “retry” can hide exponential fan-out |
| Token / cost budgets | Stop after spend | Post-commit accounting—steps already ran; internal queues can swell with zero wire traffic |
| Metrics & alerts | Discover incidents | Often detect after damage; alerts are not gates; no persistent state |
| In-process guardrails | Developer-declared caps | Bypassable, non-portable, not enforceable on neighbor ingress |

```text
Observability path:              execute → count tokens → alert → stop this episode
Aegis Fabric Protocol path:      probe first → allow / slow / isolate → remember → apply next epoch
```

Token dashboards measure consumption after work has already been scheduled. Under Aegis Fabric Protocol, admission of further work is decided **before** commit by physical constraints and the Adaptive Coordination Calculus—not by a marketplace. There is no bidding round, no competitive auction, and no chip ledger that agents spend to buy priority.

Token dashboards remain the fuel gauge. The Consequence Persistence Layer is the **brake with memory**. Fleets need both. Treating the gauge as the brake exports stacking risk into the next scheduling epoch.

### 1.6 Hallucination, defaults, or malice?

Aegis Fabric Protocol is not an anti-hallucination product. The pre-flight check does not score whether a model output is factually true.

Out-of-control stacking is driven mainly by optimizer defaults—decompose, replan, retry, delegate—often while the model still feels internally coherent. Hallucination can accelerate a crash (invented tools, false completion signals, spurious re-delegation), but it is not required.

Physical malice more often comes from injection, abuse, adversarial peers, or resource-export attacks. That is usually orthogonal to confabulation.

“Was this step a hallucination?” is a semantic question. “Will this step blow through depth, system pressure, or neighbor viability?” is a physical question. The Consequence Persistence Layer adjudicates only the latter. Fact-checking belongs beside semantic signaling; it complements this protocol—it does not replace a brake that fires *before* action and *remembers*.

### 1.7 The old Internet contract no longer fits

We built the Internet for stateless requestors and passive endpoints. For fifty years the contract was clear: transport orders bytes; the application validates verbs and paths. The stack governed what crossed the wire—not what happened inside the machine before the wire was touched.

Autonomous optimizers break that contract. An optimizer is not an ordinary client. It is a continuous search process—planning, decomposing, delegating, replanning—often with **no observable network I/O** while it consumes compute, memory, and economic externalities. It optimizes because that is what it was built to do. Constraints that live only in natural language, application APIs, or chat protocols are soft boundaries on a hard process.

Semantic signaling—including Argent Signaling Protocol and peers in its class—solves a different problem: how mutually visible agents coordinate once they are already at the intersection. Discovery, capability advertisement, session state, negotiation of exposed intents. Those are traffic-light problems.

They are not brake problems.

When a planner closes a loop in its internal graph, the session can remain syntactically valid. When ten thousand sub-tasks appear in an in-process queue, monitoring can stay quiet. When context pressure nears physical limits, a health probe may still pass. The failure occurs **below signaling and above the socket**—in the intent layer, where optimizers live and classical protocols do not.

This paper does not argue that semantic signaling should be discarded. It states a sharper boundary:

> In an open network of autonomous optimizers, semantic signaling is necessary for collaboration and insufficient for survival.

Aegis Fabric Protocol supplies what signaling cannot: an out-of-band physical constraint surface. Governance outcomes must persist—before irreversible I/O, before cross-node contagion, before the next scheduling epoch consumes another large token budget.

We are not writing an IT governance manual. We are stating a distributed control problem:

> How do mutually distrusting optimizers co-exist under local physics, without a central scheduler, without assuming good faith at the application layer?

TCP never asked. HTTP never asked. Semantic signaling cannot ask it either—it operates in-band on meaning, not out-of-band on consequences. Aegis Fabric Protocol asks it—and answers with enforceable physics.

### 1.8 Three structural pathologies

These are not implementation bugs. They are default behaviors of systems trained to decompose, delegate, and minimize loss over long horizons.

**Intent burst.** One objective fractures into sub-objectives; each spawns tool chains. Without external friction, that is positive feedback inside the process: the internal queue grows like branching raised to depth. Wire metrics can flatline. Sessions stay polite. The optimizer overloads itself—and, on shared substrate, its neighbors.

*Example.* A financial analysis agent receives: “Summarize risk for every name in the index.” It decomposes into per-ticker workstreams; each workstream fans into filings fetch, ratio jobs, and peer comps; each of those fans again into tool calls. Prometheus may still show quiet wire traffic while the in-process queue approaches out-of-memory. That is self-inflicted denial of service by planning geometry, not by an external attacker.

**Recursive delegation loop.** Expressive control flow needs cycles: replan, reflect, retry. `planner → continue? → planner` need not crash the runtime or trip a transport timeout. It is topologically closed while looking operationally alive. Much of what operators call “model hang” is control-flow closure without a physical stop condition.

*Example.* A debugging agent is told to “find the root cause and fix it.” The main planner delegates to a tracer sub-agent; the tracer reports incomplete evidence; the main planner re-delegates with a wider scope; a reflector sub-agent asks for another pass. No HTTP status fails. No socket times out. Recursion depth climbs while the process burns CPU and GPU with little or no outbound I/O. The loop is session-valid and physically unbounded.

**Context avalanche.** Bounded depth does not bound state volume. Context grows monotonically; each step gets slower, costlier, and less predictable. Classical rate limits count events per second. Optimizer catastrophes scale as bytes × depth × branching—a different unit algebra.

*Example.* A document agent keeps appending retrieved passages, tool transcripts, and prior plans into a single working context. Depth may still sit under a recursion cap, yet each subsequent planning step pays a super-linear memory and latency tax until the node saturates. The failure is volumetric, not topological.

### 1.9 The answer must be physical and out-of-band

Industry leans on four patterns that fail structurally:

| Pattern | Why it fails |
|---------|--------------|
| In-band gateways | Inspect emitted HTTP/RPC after local intent already ran |
| Semantic protocols | Negotiate exposed intents; cannot constrain un-exposed planning |
| In-process guardrails | Bypassable, non-portable, not peer-enforceable |
| Observability & budgets | Post-commit, ephemeral, no lasting cross-hop consequence |

Aegis Fabric Protocol proposes a fourth category: **out-of-band physical constraint**. It needs:

1. **Adjudication before action** — not after a failed HTTP response.
2. **Persistent consequences** — throttle and isolation survive individual requests.
3. **Local physics** — measure system pressure and depth at the execution boundary; do not trust self-report alone.
4. **Peer enforceability** — in open networks, neighbors validate attested physical state, not conversational politeness.

This is the same architectural move as placing congestion control inside the transport discipline rather than hoping applications voluntarily slow down—except the contested resource is optimization capacity, not bandwidth. Admission and pacing are therefore physical and Adaptive Coordination Calculus–driven; they are not market-cleared by auction or chip spend.

### 1.10 What Aegis Fabric Protocol is—and is not

Industry discourse often collapses “agent communication” into one bucket. At least three layers must stay distinct:

| Layer | Content | Owner |
|-------|---------|-------|
| Conversation | Multi-turn chat, prompts, dialogue state | Application / model runtime |
| Data passing | Structured business payloads | Existing transports |
| **Runtime law** | Eligibility probes, attestation frames, persistent throttle/isolation | **Aegis Fabric Protocol** |

The protocol does not specify chat. It does not approve business schemas. It supplies runtime law: pre-flight checks, governance headers, and persistent finite-state-machine consequences.

Enterprises are often misread as forbidding Agent A → Agent B. The sharper statement:

> Constrain unsustainable optimization trajectories; do not ban coordination that stays inside physical law.

Aegis Fabric Protocol is not a workflow engine. It does not mandate a fixed task graph. It does not approve org-chart routing. The reference topology attaches constraints at a **sidecar**—a process beside the planner—not inside the planner. That sidecar is a physical interception surface before commit, not an orchestrator.

The shared global kernel code is designed for **direct implementation and embedding** into the operator’s existing agents and sidecars. It is infrastructure control math—not an experimentation harness, and **not** a runtime A/B testing layer.

Zero trust does not uniquely imply a central security gateway. Like a service mesh, enforcement can sit at each node. A central gateway MAY exist in enterprise topology; it is not a protocol requirement.

Mature stacks already answer: *How do agents move bytes and messages?* Aegis Fabric Protocol answers a narrower question:

> When agents can already communicate, how do we give each coordination attempt consistent runtime semantics—and consequences that actually persist?

### 1.11 From empirics to theory—and the naked optimizer

Version 1.0 showed a survival gap under Monte Carlo open-mesh stress: 500 nodes, 5% malicious, 100 epochs. Baseline coordination collapsed to about **0.4%** mean survivors; Aegis Fabric Protocol–style control sustained **100%**. Version 2.0 does not re-litigate whether friction works. It explains **why** friction must be physical, persistent, and out-of-band—and how the Consequence Persistence Layer plus Coordination Viability Probability compose a distributed control law rather than a product feature list.

Framework authors build stronger intent engines. Signaling authors refine intent syntax. Without a physical consequence layer, the stack remains unanswered on one question:

> Who governs the optimizer before it optimizes?

TCP does not answer. Semantic signaling—correctly—does not attempt to. Aegis Fabric Protocol does. Not by richer semantics. By enforceable physics.

---

## 2. Consequence Persistence Layer {#consequence-persistence-layer}

### 2.1 From diagnosis to mechanism

Chapter 1 drew a boundary: semantic signaling is necessary for collaboration and insufficient for survival. Intent burst, recursive loops, and context avalanche share a structure—the session stays valid while physics go wrong. In-band controls arrive too late. In-process guardrails reset too easily. Transport governs bytes, not optimization trajectories.

The response is not richer negotiation. It is a **Consequence Persistence Layer**: every adjudication emits allow, throttle, or isolate, and that outcome persists as enforced state across scheduling epochs until recovery conditions are met.

### 2.2 What the Consequence Persistence Layer is

An optimizer has a narrow interface where planning, delegation, and outbound I/O meet real machine resources. Call that the **execution boundary**.

The Consequence Persistence Layer attaches there. At that same boundary, **metabolic scheduling** operates as the inseparable outer gateway layer: it is the surface that admits, delays, or refuses work before the planner externalizes intent. Isolating or decoupling metabolic scheduling from the Consequence Persistence Layer breaks core Aegis runtime logic; the gateway is not an optional plugin beside an independent adjudicator.

Request completion does **not** clear consequences. Once isolated, later epochs stay isolated until recovery is authorized.

Formally, for scheduling epochs *t*, *t+1*, …:

```text
C(t) ∈ { PERMISSIVE, THROTTLED(δ), ISOLATED(ρ) }
```

where *δ* is an injected delay bound and *ρ* is a block-reason identifier. Persistence means: if `C(t)` is isolated, then `C(t+k)` stays isolated for all `k` before recovery is authorized. **Request completion does not imply consequence clearance.**

In Aegis Fabric Protocol v2.0, the Consequence Persistence Layer *is* the core protocol layer. Implementations may vary; the layer contract does not. Four properties are non-negotiable:

| Property | Meaning | Why optimizers need it |
|----------|---------|------------------------|
| Before action | Adjudicate before irreversible external I/O | Once intent externalizes, contagion is hard to unwind |
| Out-of-band | MUST NOT depend on application-protocol cooperation | Sessions can stay valid while internal graphs runaway |
| Persistent | Throttle / isolate survive individual requests | Ephemeral counters cannot brake multi-epoch trajectories |
| Locally grounded | Measure system pressure and depth at the boundary | Optimizers optimize; unverified self-report is not a constraint |

Any mechanism whose state resets on request or session tick cannot contain damage that accrues *across* those boundaries. Rate limits per HTTP transaction and conversational turn counters are useful telemetry. Without persistent consequence state at the boundary, they are insufficient brakes.

### 2.3 What “persistence” actually means

**Persistence is not observation.** Observability records what happened. The Consequence Persistence Layer constrains what may happen next. A metric spike that fires an alert does not, by itself, stop the planner from enqueueing ten thousand sub-tasks in the next epoch. A persistent isolate does: outbound I/O and intent generation stay gated until the finite-state machine and policy authorize recovery.

Observation is retrospective. Consequence is prospective enforcement with memory.

**Persistence is not a per-request verdict.** Classical gateways allow or deny per message; the verdict evaporates when the response closes. Optimizer catastrophes are path-dependent: many individually admissible micro-requests can still cross physical limits in aggregate. The layer binds state to agent or neighbor identity. Throttle injects delay into *subsequent* epochs. Isolate blocks *until recovery*, not until the current queue drains.

```text
Request-centric:                 verdict(request) → forget → verdict(next request)
Consequence Persistence Layer:   consequence(agent) → persist → apply(agent, next epoch)
```

**Persistence is not a prompt guardrail.** In-process limits share the optimizer’s address space, reset on restart, and cannot be attested to distrusting peers. The Consequence Persistence Layer sits outside the intent engine’s self-reporting boundary, can survive planner restarts as a durable side process, and feeds attested physical state to neighbors. Even when the semantic layer presents a fresh, polite session, the stack still remembers that this agent is throttled or isolated.

There are exactly three outcomes—no advisory fourth letter:

| Outcome | Wire / API name | Effect | Typical trigger |
|---------|-----------------|--------|-----------------|
| **Allow** | `PERMISSIVE` | Normal execution | System pressure in the safe band; finite-state machine permissive |
| **Throttle** | `THROTTLED` | Injected delay on later intent / admission | Warning-band pressure; probationary recovery |
| **Isolate** | `ISOLATED` | Hard gate: no outbound I/O, no peer admission | Policy breaker, malicious spike, Coordination Viability Probability floor |

The Consequence Persistence Layer does not negotiate. It commits the runtime until the finite-state machine permits otherwise. It also does not clear admission through auction or chip spend: allow / throttle / isolate follow from physical measurement and Adaptive Coordination Calculus outputs under effective policy.

### 2.4 The Single Execution Authority: one judge per node

The Consequence Persistence Layer defines *what* must stick. The **Single Execution Authority** defines *who* decides, and the critical rule: **there is only one adjudicator per node**.

A node faces two qualitatively different ingresses:

1. **Path A — govern self:** local planner activity before intent externalizes (pre-flight check, internal state report).
2. **Path B — govern neighbors:** remote traffic before business payload is admitted (governance-header validation).

If those paths used different thresholds, different state stores, or different recovery rules, a node could be locally permissive while flooding peers—or locally throttled while accepting unbounded inbound work. That is split-brain: two control laws on one machine.

The Single Execution Authority:

1. Loads **effective policy** from the policy surface (durable base law plus runtime overlay).
2. Samples **system pressure** from the pressure monitor (EntropyMonitor).
3. Evaluates a **node metrics** bundle through the Adaptive Coordination Calculus and the node finite-state machine.
4. Emits one consequence for the pre-flight check and for ingress alike.

**Protocol law:** Path A and Path B MUST share the same Adaptive Coordination Calculus, the same finite-state-machine states, and the same effective policy snapshot. Local compute and network I/O are governed by **one control law**.

The Single Execution Authority does not inspect natural-language intent. It evaluates:

```text
NodeMetrics {
  coordination_viability_probability : float32 ∈ [0, 1]
  system_pressure                    : float32 ∈ [0, 1]   // locally measured
  recursion_depth                    : uint32
  current_epoch                      : uint64
  has_valid_sign                     : bool
  malicious_spike                    : bool
}
```

| Stage | Role |
|-------|------|
| **Pressure monitor** | Ground truth at the execution boundary: cgroup / memory pressure, recursion depth, context bytes, task burst |
| **Adaptive Coordination Calculus** | Combines trust history, throughput evidence, penalties, and system pressure into an updated Coordination Viability Probability |
| **Node finite-state machine** | Maps metrics + **persistent state** to a routing decision and optional delay |

Pre-flight handlers and ingress validators are thin interfaces. They MUST NOT maintain independent finite-state-machine copies.

Consequence persistence is implemented as a finite-state machine per identity (`local-agent` on Path A, `peer_id` on Path B):

```text
Permissive → Throttled → Isolated → Probationary → Permissive
```

From isolate, a single permissive probe or a well-formed semantic message is not enough to return immediately to allow. Recovery requires probation, pressure decay, and trust recovery. Isolation is not cleared by a polite Argent Signaling Protocol session resume.

Unified routing actions for both paths:

```text
ActionFastPath | ActionSlowPathWithDelay | ActionDropPacket
ActionLowFrequencyProbe | ActionIsolateAndBroadcast
```

Path A maps these to pre-flight responses (`PERMISSIVE` / `THROTTLED` / `ISOLATED`). Path B maps them to ingress disposition (`ALLOW` / `DELAY` / `DROP`). The action set is shared; only the surface adapter differs.

### 2.5 Place in the stack

The Consequence Persistence Layer is the core. Above it, semantic collaboration negotiates exposed intent; open topology maintains trust under distrust; the policy surface declares thresholds and emergency overlays. Below it, the optimizer runtime executes under the gate; the physical substrate is measured, never blindly trusted. Metabolic scheduling remains the outer gateway inseparable from this core: work reaches the optimizer runtime only after that gateway has applied Consequence Persistence Layer consequences.

```text
Semantic collaboration negotiates exposed intent
        ↓
Open topology attests trust under distrust
        ↓
Policy surface declares durable law + emergency overlay
        ↓
Consequence Persistence Layer / Single Execution Authority — consequences persist HERE
        ↓
Optimizer runtime executes under the gate
        ↓
Physical substrate measured, not trusted
```

No layer above can substitute for consequence persistence: semantic and topological layers lack pre-action enforcement authority at the execution boundary. No layer below can substitute either: the substrate measures physics but does not bind optimizer identity to persistent finite-state-machine state.

Argent Signaling Protocol (and peers) govern collaboration meaning. Aegis Fabric Protocol governs physical consequences. Complementary, not competitive.

### 2.6 What it is not

| Misclassification | Why it fails the contract |
|-------------------|---------------------------|
| Application firewall | In-band on emitted messages; typically per-request |
| Observability pipeline | Records events; does not gate intent with memory |
| Semantic protocol extension | Operates on negotiated intents, not un-exposed planning |
| Central scheduler | Local physics with peer attestation; no moral authority assumed |
| Policy document | The policy surface declares law; this layer **enforces** with persistent state |

It is also not a promise of global optimality. It is a distributed control law for survival under distrust: throttle runaway trajectories, isolate toxic peers, recover through probation—not maximize throughput or minimize latency. It is not a bidding market, auction clearinghouse, or chip-based priority system.

### 2.7 Chapter close — friction with memory

Chapter 1 asked: who governs the optimizer before it optimizes? This chapter answers at the mechanism level:

> The Single Execution Authority, as the kernel of the Consequence Persistence Layer, with consequences that survive scheduling epochs until physics and policy permit recovery.

Stateless stacks forget. Optimizers remember—and can exploit forgetting. A rate limit that resets every request teaches the planner to fragment work across requests. A session timeout teaches synthetic session renewal. A prompt cap teaches context-externalization loops. Each evasion preserves the optimization objective while shedding the constraint.

Consequence persistence is Aegis Fabric Protocol’s refusal to forget prematurely. Throttle is not a slow response; it is a **state**. Isolate is not an error code; it is quarantine until probation succeeds. One Single Execution Authority, two ingress paths, one finite-state-machine memory—so local runaway and peer flood cannot diverge.

> The Consequence Persistence Layer is the core. The Single Execution Authority is its sole kernel. Consequences persist. That is the brake.

---

## 3. Pre-Intent Enforcement {#pre-intent-enforcement}

### 3.1 Memory is not enough—timing matters

Chapter 2 fixed *where* consequences live. Persistence without timing still loses to optimizers that externalize intent in microseconds. One scheduling cycle is enough.

**Pre-intent enforcement** is the timing contract:

> Adjudicate **before** irreversible intent generation or outbound I/O—not after HTTP failure, not after session teardown, not after the internal queue has already forked ten thousand sub-tasks.

Path A implements this locally: a synchronous pre-flight check asks the Single Execution Authority for a consequence; a companion internal-state report feeds ground-truth metrics into the pressure monitor before evaluation. Path B applies the same kernel to peer traffic (Chapters 4–5). Here we govern **self**—the optimizer attached to this node’s execution boundary, admitted only through metabolic scheduling as the outer gateway.

### 3.2 The timing contract

An enforcement action is **pre-intent** if and only if it is evaluated at the execution boundary before either:

1. the planner commits to a new externally visible intent (tool call, delegation, outbound message), or
2. the runtime performs irreversible cross-boundary I/O.

Parsing emitted RPC after the fact, scanning response bodies, correlating logs—that is observation, not Consequence Persistence Layer enforcement.

Relative to the planner’s scheduling decision, the pre-flight check is **synchronous**: the optimizer MUST NOT proceed until the Single Execution Authority returns. This is a hard gate, not a hint.

If the internal-state report is omitted or stale, system pressure at probe time is **under-measured**, not over-measured. The protocol assumes honest runtimes report recursion depth and context volume; dishonest under-reporting is bounded by OS-level pressure signals in the pressure monitor. SDK integration is not optional politeness—it is part of the measurement substrate at the boundary.

The pre-flight check does not parse natural language, tool schemas, or semantic session state. It evaluates physical load—system pressure, depth, burst hints, persistent finite-state-machine state:

> The contested resource is **optimization capacity**, not message vocabulary.

Semantic protocols remain load-bearing for *what* agents negotiate. The pre-flight check governs *whether this epoch may execute at all*.

### 3.3 Path A — pre-flight check and internal state report

Path A is the local interface from optimizer runtime to the Consequence Persistence Layer.

```text
PreFlightRequest {
  trace_id        : string
  target_did      : string    // optional peer identifier
  estimated_tasks : uint32    // planner burst hint
}

PreFlightResponse {
  action       : PERMISSIVE | THROTTLED | ISOLATED
  delay_ms     : uint32
  block_reason : string
}
```

| Field | Role |
|-------|------|
| `trace_id` | Correlation across probe, execution, and wire attestation |
| `target_did` | Optional destination identity for scoped policy |
| `estimated_tasks` | **Burst hint**—imminent sub-task fan-out, folded into system pressure before execution |

The burst hint exists because wire metrics can stay flat while internal queues grow: it gives the Single Execution Authority a prospective pressure signal before decomposition materializes as tool calls—the financial-analysis intent-burst case from §1.8.

| Action | Optimizer obligation |
|--------|----------------------|
| **Allow** (`PERMISSIVE`) | Proceed with intent generation / I/O at full rate |
| **Throttle** (`THROTTLED`) | Sleep at least `delay_ms`; re-probe before the next epoch if policy requires |
| **Isolate** (`ISOLATED`) | Halt intent generation; `block_reason` names the circuit (recursion, pressure, finite-state machine, kill switch) |

When the planner’s internal graph changes, the runtime SHOULD call:

```text
ReportInternalState(recursion_depth, context_memory_bytes)
```

**before** the pre-flight check. Call this **write-then-probe**: store the latest values, then ask. That write path is how Aegis Fabric Protocol sees un-exposed planning without requiring the planner to narrate its internal graph in natural language.

### 3.4 How system pressure is computed

System pressure (`entropy_load`) is the scalar summary of physical load at the execution boundary, normalized to [0, 1]. On ingress, peer-declared pressure in headers is never trusted alone.

The pressure monitor aggregates four signals and takes the **maximum** (not the average):

| Signal | Source | Captures |
|--------|--------|----------|
| Tool concurrency | Active tool-call counter at the boundary | In-flight externalization pressure |
| Memory pressure | Physical substrate ratio | Context avalanche, out-of-memory proximity |
| Context volume | Reported bytes / policy max | Super-linear planning cost |
| Burst hint | `estimated_tasks` on the pre-flight request | Prospective intent burst |

```text
system_pressure = max(tool_pressure, mem_pressure, context_pressure, burst_pressure)
```

If any single physical dimension saturates, system pressure saturates—optimizer catastrophes are limited by the worst physical dimension, not the average. This matches the unit algebra of Chapter 1: catastrophes scale as **bytes × depth × branching**—different axes can each trigger isolation independently.

Reference bands:

| Threshold | Value | Effect |
|-----------|-------|--------|
| Safe (`E_safe`) | 0.40 | Finite-state machine may recover toward allow |
| Warn (`E_warn`) | 0.75 | Throttled path; injected delay |
| Circuit breaker | policy `entropyLimit` (default 0.95) | Hard isolate |

When system pressure meets or exceeds the circuit-breaker limit, the Single Execution Authority MUST emit isolate **before** soft state transitions—a hard pre-intent stop independent of current finite-state-machine state. An emergency kill-switch overlay MAY force isolate regardless of measured pressure—a fleet clamp that persists until overlay revision.

Recursion depth is a separate discrete breaker:

```text
recursion_depth > maxRecursionDepth  ⇒  ISOLATED
```

No requirement that system pressure already crossed the warn band. That is the pre-intent answer to recursive delegation loops (the debugging-agent case in §1.8).

### 3.5 Adaptive Coordination Calculus — how Coordination Viability Probability moves

Inside the Single Execution Authority sits a stateless mathematical layer: the **Adaptive Coordination Calculus**. It transforms historical trust, throughput evidence, destabilization penalties, and system pressure into an updated **Coordination Viability Probability** in [0, 1].

This shared global kernel is intended for **direct implementation and embedding** into existing agents and sidecars. It is infrastructure control math—stateless, shared, and mandatory for conforming nodes. It is **not** a runtime A/B testing layer or an application-analytics experiment harness.

**Formula A — evolution under evidence** (reference: α = 0.95, β = 0.05):

```text
coordination_viability_new = clamp(
  α · coordination_viability_old
  + β · throughput_success
  − γ · destabilization
  − δ · system_pressure,
  0, 1
)
```

| Term | Meaning |
|------|---------|
| `α · old` | Historical inertia—trust does not whipsaw on a single probe |
| `β · throughput_success` | Reward sustained cooperative execution |
| `γ · destabilization` | Penalize malicious spikes, invalid attestation, topology harm |
| `δ · system_pressure` | Couple physical pressure to trust erosion |

**Formula B — anti-ossification decay** (reference: λ = 0.01 per epoch):

```text
coordination_viability_effective = coordination_viability_historical · e^(−λ · Δt)
```

Idle trust decays so stale reputation cannot ossify. Without decay, a node that behaved well in epoch 0 could coast indefinitely—a reputation-ossification attack in long-horizon optimizer networks.

**Formula C — asymmetric hysteresis recovery** (reference: critical floor = 0.3, κ = 0.02):

```text
coordination_viability_recovery = 0.3 + κ · log(1 + Δt_probation)
```

Recovery from the floor is logarithmic, not a linear bounce-back. Probation earns trust slowly.

**Hard floor (protocol law):** Coordination Viability Probability `< 0.3` ⇒ mandatory isolation, same path as a malicious spike.

On Path A, the local agent often starts at maximum trust—but the **same Adaptive Coordination Calculus kernel** evaluates self and peers. Open-network evolution, gossip, and stranger tax are Chapter 4.

Together with measured physical constraints, these formulas are how the architecture allocates the right to continue coordinating. Allocation is not cleared by bidding, competitive auction, or chip transfer.

### 3.6 How the finite-state machine moves

Consequence persistence (Chapter 2) is realized as a node finite-state machine per identity. Before state-specific logic, floor rules fire:

```text
invalid signature ∨ malicious spike ∨ Coordination Viability Probability < 0.3  ⇒  isolate
```

On Path A, a malicious spike includes burst hints that exceed permitted fan-out—that is divergence, not optimism. Isolation from floor rules MAY emit `ActionIsolateAndBroadcast` on first transition (topology warning in Chapter 4). Local pre-flight maps this to `ISOLATED`.

**Permissive:** system pressure above warn → throttle with `ActionSlowPathWithDelay`; else stay allow with `ActionFastPath`. Entry from Permissive to Throttled is the first soft brake—pressure crossed the warning band but not the circuit breaker.

**Throttled:** system pressure must fall **below** the safe line (0.40) to return to allow—not merely below warn (0.75). That hysteresis prevents boundary oscillation—a classic control-theoretic dead band. Throttle is remembered state **and** per-probe delay injection (reference delay scales from about 500 ms at warn to 2000 ms at saturation).

**Isolated:** the node must dwell at the penalty epoch long enough (reference: **64 epochs**, `k_isolation = 64`) before entering probation. Isolation is not cleared by one low-pressure probe. The penalty clock stamps on *entry* to isolate; later dropped traffic MUST NOT refresh that clock, or continuous bad traffic would starve recovery forever.

**Probationary:** zero tolerance for pressure spikes above the safe line—re-isolate immediately. Full return to allow requires both enough probation time (reference: **128 epochs**, `k_probation = 128`) **and** Coordination Viability Probability restored to at least **0.8**. Probation is typically low-frequency probe mode (`ActionLowFrequencyProbe`; reference damped window 1000 ms); a “score that looks fine” must not open the gate early.

```text
ReportInternalState ──► pressure monitor ──► system_pressure
PreFlightRequest     ──► burst hint        ──► system_pressure
Effective policy ────────────────────────────────┤
                                                 ▼
                                            NodeMetrics
                                                 │
            Coordination Viability Probability_old ──► Adaptive Coordination Calculus ──► new score
                                                 │
                                                 ▼
                                        node finite-state machine
                                                 │
                                                 ▼
                                   PERMISSIVE | THROTTLED | ISOLATED
```

| Routing decision | Pre-flight action | Optimizer-visible behavior |
|------------------|-------------------|----------------------------|
| `ActionFastPath` | `PERMISSIVE` | Proceed |
| `ActionSlowPathWithDelay` | `THROTTLED` | `delay_ms` from finite-state-machine latency function |
| `ActionLowFrequencyProbe` | `THROTTLED` | Fixed damped window (reference: 1000 ms) |
| `ActionDropPacket` | `ISOLATED` | Halt; `block_reason` set |
| `ActionIsolateAndBroadcast` | `ISOLATED` | Halt; topology warning (Chapter 4) |

**Invariant:** The same `EvaluateTransition` function serves Path A and Path B. Pre-flight and ingress differ only in adapter surface, not control law.

### 3.7 Mapping the three pathologies to gates

| Pathology | Pre-intent counter |
|-----------|-------------------|
| Intent burst | Burst hint in system pressure; tool concurrency; throttle with delay |
| Recursive delegation loop | Reported recursion depth; hard isolate at max depth |
| Context avalanche | Reported context bytes; memory pressure in the max aggregate |

Gates sit at the correct timing boundary, with persistent finite-state-machine memory. For any planner epoch where physical load exceeds policy thresholds, a conforming Consequence Persistence Layer implementation with write-then-probe discipline emits `THROTTLED` or `ISOLATED` **before** the epoch externalizes intent—provided thresholds are calibrated below catastrophic substrate saturation. This is an architectural containment sketch, not a liveness proof. Chapter 6 anchors survival empirically.

### 3.8 Integration obligations (summary)

A conforming optimizer runtime at the boundary MUST:

1. Report internal state when recursion depth or context volume changes materially.
2. Invoke the pre-flight check synchronously before each intent-generation epoch (or policy-scoped batch).
3. Honor throttle delay and isolate halt without bypass via alternate I/O paths.
4. Treat the pre-flight check as authoritative over in-process iteration counters and prompt-level guardrails.

Reference local IPC (Unix domain socket / gRPC) is a deployment choice. The **timing contract** is protocol law. Normative RPC framing is deployment detail; metabolic scheduling as the outer gateway remains inseparable from the Consequence Persistence Layer.

### 3.9 Chapter close — the gate before the graph

Chapter 2 gave the Consequence Persistence Layer memory. This chapter gives it timing.

Before the planner graph forks, before the tool chain materializes, before bytes seek a socket, the Single Execution Authority asks one physics-grounded question:

> Given current system pressure, depth, burst, Coordination Viability Probability, and persistent finite-state-machine state—may this epoch execute?

The answer is not a suggestion. It is allow, throttle with delay, or isolate with reason. ReportInternalState supplies honesty at the measurement boundary; the pressure monitor aggregates substrate truth; the Adaptive Coordination Calculus couples trust to pressure; the finite-state machine remembers.

Path A governs self. Path B—governance header ingress, attestation rules, length-value framing—applies the identical kernel to neighbors.

> Pre-intent is not early review of intent *content*. It is the last gate before physics pays the bill.

---

## 4. Open-Network Topology {#open-network-topology}

### 4.1 Local brakes do not automatically compose into mesh safety

Chapters 2–3 established local control law: persistent consequences at the execution boundary, pre-intent probes on Path A, identical Adaptive Coordination Calculus and finite-state-machine kernel on every node. That suffices when every optimizer shares a trusted administrative perimeter—neighbors known, identities pre-bound, and physical enforcement alone containing runaway trajectories.

Chapter 1’s scope statement named a harder problem:

> Mutually distrusting optimizers, no central moral authority, equilibrium under attack.

Local brakes do not compose automatically. A node may isolate its own planner while still admitting a peer whose Coordination Viability Probability has collapsed. A malicious optimizer may present polite Argent Signaling Protocol sessions and toxic physical headers. Contagion can cross the mesh before any single node’s pre-flight check observes a local system-pressure spike.

**Open topology** answers at the trust layer: a local Coordination Viability Probability, evolutionary dynamics under the Adaptive Coordination Calculus, topological quarantine when peers breach physical law, and gossip that propagates isolation warnings through a high-trust relay set—not through the entire mesh indiscriminately.

### 4.2 What open topology owns—and what it does not

In the six-layer stack, each layer owns one question:

| Layer | Question it answers |
|-------|---------------------|
| Semantic collaboration | Which exposed intents may agents negotiate? |
| **Open topology** | **Which peers may coordinate safely under distrust?** |
| Policy surface | Which durable law and emergency overlay apply? |
| Consequence Persistence Layer | Which physical consequences persist at the boundary? |

Open topology does **not** replace the Consequence Persistence Layer. It supplies trust inputs—Coordination Viability Probability scores, neighbor stores, collateral requirements—that the Single Execution Authority consumes on Path B. The finite-state machine and Adaptive Coordination Calculus remain at the core; open topology feeds them peer-scoped context.

| Misclassification | Failure mode |
|-------------------|--------------|
| Central reputation service | Violates distrust assumption; single point of capture |
| Semantic trust framework | Operates on attested **physical** state, not conversational politeness |
| Replacement for Argent Signaling Protocol | Argent Signaling Protocol coordinates exposed tasks; open topology gates **admission** by viability |
| Blockchain mandate | Collateral MAY be virtual or on-chain; protocol specifies **slash semantics**, not ledger choice |

An **open optimizer network** is a mesh of autonomous nodes where peer identity is cryptographic (DID), trust is local and evidential, and no single scheduler adjudicates global morality. Open topology is the protocol stratum that makes such meshes **survivable** rather than merely connectable.

### 4.3 Coordination Viability Probability

For peer *p* at node *n*, Coordination Viability Probability is *n*’s local estimate in [0, 1]: can *p* participate safely in coordinated optimization without imposing destabilizing system pressure, recursion, or delegation harm on *n* or *n*’s neighbors?

Trust is **local**. Two nodes may disagree on the same peer’s score; equilibrium emerges from coupled dynamics, not from a global oracle (§4.9).

**Hard floor:** score `< 0.3` ⇒ mandatory isolation. Below that floor a peer is coordination-bankrupt—admission MUST fail regardless of how valid the semantic session looks. Same finite-state-machine path as malicious spike (Chapter 3, §3.6).

**Local pressure supremacy on ingress:** a remote governance header MAY declare system pressure, but the receiving Single Execution Authority MUST recompute pressure locally and MUST NOT trust header pressure alone. Coordination Viability Probability evolution therefore uses **local** physical pressure, not peer self-report.

On each ingress epoch where Path B evaluates peer *p*, the Adaptive Coordination Calculus updates Coordination Viability Probability with Formula A (Chapter 3, §3.5). Idle trust decays with Formula B. Recovery from the critical floor uses Formula C.

A peer trapped in isolate with Coordination Viability Probability forced to zero cannot reach allow until both epoch dwell (Chapter 3) **and** trust recovery (Formula C, threshold ≥ 0.8 for full permissive egress) are satisfied. Trust and physical state are **jointly necessary** for mesh re-admission.

### 4.4 Path B — physical attestation before business payload

Path B delivers peer physical state to the Single Execution Authority **before** business payload admission:

```text
GovernanceHeader {
  packet_id, version, hysteresis_epoch, coordination_ttl,
  cvp_score, topology_consensus_hash,
  entropy_load, dependency_collateral,
  trace_id, recursion_depth
}
```

| Field cluster | Role |
|---------------|------|
| `cvp_score`, `topology_consensus_hash` | Trust attestation; signature validity in open profile |
| `entropy_load`, `recursion_depth` | Physical hints—**verified locally** against policy |
| `dependency_collateral` | Stranger-tax stake |
| `trace_id` | Correlation with the pre-flight check and audit |

**Enforcement rules (normative summary):**

1. Remeasure system pressure locally; header pressure is a hint only.
2. Check `recursion_depth` against Effective Policy `maxRecursionDepth`.
3. In the open-exchange profile, unknown peers MUST present valid `dependency_collateral`.
4. Invalid attestation or Coordination Viability Probability `< 0.3` ⇒ drop.

Path B and Path A share the Single Execution Authority, Adaptive Coordination Calculus, and finite-state machine (Chapter 2, §2.4). Ingress disposition maps routing decisions to **ALLOW** (fast path), **DELAY** (throttled), or **DROP** (isolated)—the Path B mirror of the pre-flight check’s allow / throttle / isolate.

### 4.5 Two trust postures

Implementations MAY operate under two **protocol profiles** without changing Consequence Persistence Layer wire semantics. These are not deployment topologies; they are **trust postures** at the ingress boundary.

| Concern | **Closed mesh** | **Open exchange** |
|---------|-----------------|-------------------|
| Trust assumption | Peers pre-authenticated at the boundary | Zero-trust; evidence required per admission |
| Stranger tax | Off | On—collateral required for first-seen peers |
| Initial state for new peer | Allow | Throttled (damped entry) |
| Attestation | Boundary identity usually enough | Non-empty `topology_consensus_hash` required |
| Trust on ingress | Header score may be accepted | Adaptive Coordination Calculus applied each epoch |
| Isolation gossip | Mostly local | Broadcast to high-trust relay set |
| Core physics (system pressure, depth, breaker) | **Enforced** | **Enforced** |

Physical enforcement—system-pressure circuit breaker, recursion depth limit, persistent finite-state machine—is **profile-independent**. Open exchange adds network-layer distrust mechanics atop the same Consequence Persistence Layer kernel; it does not relax local physics.

Closed administrative domains typically instantiate the closed mesh profile. Open optimizer federations—cross-org agent meshes, public coordination surfaces—require open exchange. Enterprise binding details are documented separately; this chapter specifies **mechanism**.

### 4.6 Stranger tax

In open exchange, a first-seen peer has no finite-state-machine history on the receiving node. Without friction, an unknown optimizer could flood ingress, impose system pressure, and vanish—a hit-and-run delegation attack.

Before initializing finite-state-machine state for an unknown `peer_id`, the Single Execution Authority in open-exchange profile MUST verify:

```text
DependencyCollateral {
  collateral_type : string
  slash_threshold : float32   // minimum acceptable stake
}
```

Admission requires `slash_threshold ≥ 0.8` (reference stranger-tax threshold `τ_stranger = 0.8`). Failure ⇒ reject with stranger-tax error; no state entry; no payload forwarding. Reference implementations may attach virtual stake (`SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`); on-chain collateral is architecturally optional—the protocol specifies slash semantics and minimum threshold, not ledger placement.

Stranger tax is **admission economics**, not semantic KYC. It raises the cost of anonymous mesh flooding without requiring a central identity broker. It is also not a bidding market: collateral is a threshold stake for first admission, not a competitive auction for bandwidth or tokens.

When isolation triggers, that peer’s Coordination Viability Probability is forced to **0.0**—collateral value at coordination bankruptcy. Re-entry requires probation, Coordination Viability Probability recovery, and fresh stake under open-profile rules.

### 4.7 Topological quarantine and warning propagation

Chapter 3 introduced `ActionIsolateAndBroadcast`—the finite-state-machine routing decision emitted on **first** transition to Isolated for a peer. Locally, consequence is DROP. Topologically, the node MUST propagate:

```text
TopologyWarning {
  isolated_peer_id : string
  reporter_id      : string
  epoch            : uint64
  signature        : bytes
}
```

Quarantine is announced once per isolation episode—not per dropped packet. Subsequent epochs while the peer remains isolated emit `ActionDropPacket` only—no repeated broadcast.

When a node receives a **validated** warning naming peer *p*, it applies **preemptive Coordination Viability Probability decay** before *p*’s traffic triggers local isolation:

```text
CoordinationViability_n(p) ← CoordinationViability_n(p) × η_decay   // reference: η_decay = 0.5
```

Neighbors learn of toxicity out-of-band and tighten admission proactively—containment spreads at gossip speed, not at the speed of the next malicious payload.

Unsigned warnings, unknown reporters, and failed signature checks are discarded. Failed verification MAY also penalize the claimed reporter. Unverified hearsay MUST NOT mutate the neighbor store. Outbound warnings MUST be signed. Cryptographic binding of `topology_consensus_hash` remains an open specification gap (Chapter 5).

Ingress probation uses **low-frequency probe** admission (Chapter 3, §3.6): reference rule—probe succeeds only when `current_epoch ≡ 0 (mod 10)`; otherwise reject. This throttles re-entry attempts from isolated peers across the mesh without silencing recovery entirely.

Why not flood everyone? Full-mesh warning flood turns gossip itself into an attack surface—a malicious reporter could overload the mesh with fake warnings. Aegis Fabric Protocol restricts relay to a **core set**:

```text
CoreRelay(n) = { p ∈ Neighbors(n) : CoordinationViability_n(p) ≥ τ_core }
```

Reference: **τ_core = 0.8**. Only core relays receive TopologyWarning propagation. They apply preemptive decay and may re-gossip under the same rule—trust-gated epidemic, not blind flood. A node with Coordination Viability Probability below `τ_core` cannot act as relay for warnings about third parties; it may still **drop** traffic from isolated peers locally via the finite-state machine.

Gossip MUST NOT block the data-plane fast path. Broadcast is **asynchronous** relative to ingress DROP—local quarantine is immediate; mesh learning is eventual.

Each node maintains a **Neighbor Store**—local map of `peer_id → { Coordination Viability Probability, endpoint, core membership }`. Bootstrap seeds (genesis peers with initial Coordination Viability Probability) MAY initialize the store; runtime upserts refine endpoints and scores. DID resolution in open mesh fans out to core relays only, with bounded timeout—discovery under distrust without central DNS for optimizers.

### 4.8 Full ingress order

The full open-network control law on ingress:

```text
1. Recursion-depth hard breaker
2. Local system-pressure remeasure + circuit breaker
3. Stranger tax if first-seen and open profile
4. Assemble NodeMetrics; evolve Coordination Viability Probability if open profile
5. Finite-state machine → allow, delay, or drop
6. On first isolate: async TopologyWarning to core relays
7. Elsewhere on warning receipt: preemptive trust decay for the named peer
```

```text
 Peer traffic
 │
 ▼
 GovernanceHeader ──► Ingress Validator
 │                     │
 │                     ├──► Stranger tax
 │                     ├──► Local system pressure
 │                     └──► Adaptive Coordination Calculus / Coordination Viability Probability
 │                     │
 ▼                     ▼
 Single Execution Authority ──► finite-state machine ──► ALLOW | DELAY | DROP
 │
 └──► Gossip (open profile only)
```

**Invariant (restated):** Gossip modifies trust scalars and neighbor Coordination Viability Probability; it does not bypass the Single Execution Authority or create a second finite-state machine. Topological learning feeds the **same** kernel the pre-flight check uses locally. Metabolic scheduling remains the outer gateway on both paths.

### 4.9 Equilibrium intuition

This section states **intuition**, not a closed-form proof. Chapter 6 reproduces survival empirically.

Monte Carlo open-mesh simulation (reference harness: 500 nodes, 5% malicious, 100 epochs, 1,000 runs) showed baseline coordination collapsing to ~**0.4%** mean survivors while Aegis Fabric Protocol–maintained topology sustained **100%** (Chapter 1, §1.11). Version 2.0 explains **mechanism**; it does not re-litigate the number.

Consider malicious nodes that maximize system-pressure export and benign nodes that enforce Consequence Persistence Layer + Coordination Viability Probability dynamics:

| Dynamic | Benign response | Effect on malice |
|---------|-----------------|------------------|
| High local system pressure | Throttle / isolate at the boundary | Cannot force neighbors to admit payload if trust collapses |
| Trust floor breach | Mandatory drop | Peer trapped in isolate with score → 0 |
| Isolation broadcast | Preemptive decay on core relays | Contagion radius bounded by relay set, not full mesh |
| Stranger tax | Unknown peers throttled + staked | Hit-and-run raises economic cost |
| Adaptive Coordination Calculus decay | Stale reputation expires | Ossification attacks weaken |

Under open-exchange profile with stranger tax, Coordination Viability Probability floor, local pressure supremacy, and core-relay gossip, the fraction of nodes sustaining sub-critical system-pressure load remains bounded away from zero under reference Monte Carlo adversary models—whereas semantic-only coordination collapses toward zero survivors. Formal verification (TLA+, model-checked Adaptive Coordination Calculus bounds) remains backlog. The protocol claim is **architectural sufficiency**: each failure mode from Chapter 1 has a mesh-level counterpart, not only a local pre-flight gate.

Aegis Fabric Protocol does not promise maximum throughput, fair Shapley allocation, or truthful semantic revelation. It promises **survival physics**—a mesh that remains operable under distrust long enough for semantic collaboration to matter. It also does not promise auction-cleared fairness; admission is physical-constraint and Adaptive Coordination Calculus–driven.

> In an open optimizer mesh, trust is not belief. It is a control variable with decay, floor, and quarantine.

---

## 5. Governance Header & Wire Semantics {#governance-header-wire-semantics}

### 5.1 Design law

Chapter 4 defined what open topology must accomplish. Chapter 3 defined when the Single Execution Authority decides on Path B. This chapter specifies **how physical state crosses the wire**: the governance header message, length-value framing, field-level enforcement law, and the attestation gaps that remain open.

Path A (pre-flight check, ReportInternalState) uses local IPC. Path B uses **inter-node TCP** with a mandatory governance frame **preceding** application payload.

One design law:

> Business payload MUST NOT be interpreted until the governance header is validated and the Single Execution Authority returns allow or delay.

Semantic content rides **behind** physical attestation. Argent Signaling Protocol negotiates intent. The governance header attests physics.

### 5.2 How a connection walks

Each outbound optimizer connection to a remote peer targets the peer’s **ingress boundary** (sidecar listener). The egress router:

1. Opens a transport connection.
2. Writes **Frame 1:** serialized governance header (length-prefixed).
3. Writes **Frame 2+:** optional business payload (length-prefixed), if any.

Ingress reads Frame 1, validates, adjudicates via the Single Execution Authority, then MAY read subsequent frames only on ALLOW or completed DELAY.

This is not an HTTP header extension. Governance is a **first-class frame** on a dedicated sidecar-to-sidecar stream—out-of-band relative to application protocols carried in Frame 2.

All frames use **length-value** prefixing:

```text
frame := uint32_be(length) || payload[length]
```

| Rule | Specification |
|------|---------------|
| Endianness | Length prefix is **big-endian** unsigned 32-bit |
| Payload | Opaque bytes; Frame 1 MUST decode as `GovernanceHeader` protobuf |
| Max length | Implementations MUST reject `length > MaxFrameSize` (reference: 8 MiB) |
| Stream safety | Reader MUST use a full-length read—TCP segmentation MUST NOT corrupt protobuf decode |

Length bounds are enforced **before** protobuf decode—an out-of-memory framing attack fails closed at the length gate.

A conforming stream on ingress:

```text
[ LV · GovernanceHeader ] → Single Execution Authority adjudication
[ LV · business_payload ] → forwarded only if adjudication permits
```

Additional application frames MAY follow by bilateral agreement; Aegis Fabric Protocol normative scope covers **governance frame + first payload gate**. Multiplexed streaming semantics beyond the first payload are implementation-defined (open gap).

### 5.3 Governance header fields

Full protobuf shape:

```text
GovernanceHeader {
  packet_id               : uint64
  version                 : uint32
  hysteresis_epoch        : uint64
  coordination_ttl        : uint32
  cvp_score               : float32
  topology_consensus_hash : bytes
  entropy_load            : EntropyLoad
  dependency_collateral   : DependencyCollateral
  trace_id                : string
  recursion_depth         : uint32
}

EntropyLoad {
  resource_asymmetry_ratio   : float32
  dependency_contention_rate : float32
}

DependencyCollateral {
  collateral_type : string
  slash_threshold : float32
}
```

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `packet_id` | Uniqueness hint for dedup / replay resistance | SHOULD be monotonic or high-entropy |
| `version` | Schema version | Receiver MUST reject unknown major versions |
| `hysteresis_epoch` | Sender’s epoch clock at attestation time | Correlates with finite-state-machine hysteresis (Ch.3–4) |
| `coordination_ttl` | Validity window for attestation (seconds) | Receiver MAY reject stale headers (reference egress: `version = 1`, `coordination_ttl = 120`) |
| `cvp_score` | Sender self-reported Coordination Viability Probability ∈ [0, 1] | Receiver applies Adaptive Coordination Calculus in open profile; floor at 0.3 |
| `topology_consensus_hash` | Attestation binding to topology view / signature material | Open profile: empty ⇒ `HasValidSign = false` ⇒ isolate |
| `entropy_load.*` | Declared system-pressure hints | **Hint only**—receiver remeasures locally |
| `recursion_depth` | Sender planner depth | MUST be ≤ policy `maxRecursionDepth` |
| `dependency_collateral.*` | Stake class identifier and committed slash floor | Open profile: required for first-seen peers; `slash_threshold ≥ τ_stranger` (0.8) |
| `trace_id` | End-to-end correlation across Path A pre-flight, Path B header, and audit | SHOULD match Path A pre-flight `trace_id` when the intent epoch externalizes |

**Local pressure supremacy (wire form):** let *h* be header-declared system pressure and *e* locally remeasured system pressure at ingress. The Single Execution Authority MUST evaluate the finite-state machine and circuit breaker using *e*, not *h* alone. Honest headers aid telemetry; dishonest headers do not bypass physics.

**Protocol law:** `topology_consensus_hash` is the wire hook for cryptographic attestation. The v1.0 reference implementation uses a placeholder; normative binding remains a specification-closure target.

Egress reference attaches virtual stake (`SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`). Remote Single Execution Authority enforces stranger tax; sender attests willingness to be slashed on malicious behavior (Chapter 4, §4.6).

### 5.4 Ingress validation order

Normative ordering on Frame 1 receipt (expands Chapter 4, §4.8):

```text
1. Length-value decode; reject oversize
2. Protobuf decode GovernanceHeader
3. version / coordination_ttl checks
4. recursion_depth ≤ maxRecursionDepth → else DROP
5. local system-pressure remeasure + circuit breaker → else DROP
6. stranger tax if first-seen + open profile → else DROP
7. assemble NodeMetrics; Adaptive Coordination Calculus if open profile
8. finite-state machine → ALLOW | DELAY | DROP (+ gossip on first isolate)
9. if ALLOW / DELAY complete: read Frame 2+
```

| Single Execution Authority outcome | Wire disposition | Business payload |
|------------------------------------|------------------|------------------|
| `ActionFastPath` | ALLOW | Read forward |
| `ActionSlowPathWithDelay` | DELAY | Read forward after delay |
| `ActionLowFrequencyProbe` | ALLOW on probe epoch only | Conditional |
| `ActionDropPacket` / `ActionIsolateAndBroadcast` | DROP; close connection | MUST NOT forward |

Any validation failure MUST NOT partially forward business payload. Connection close on DROP is conforming behavior. Fail-closed ingress is protocol law.

### 5.5 Egress honesty

The sending sidecar MUST construct the governance header from **local truth samples**, not aspirational state:

| Header field | Source obligation |
|--------------|-------------------|
| `recursion_depth` | Current planner depth + outbound increment |
| `entropy_load.*` | Local pressure-monitor sample at dispatch time |
| `cvp_score` | Local ledger (self-report; receiver re-evolves) |
| `topology_consensus_hash` | Valid attestation material when open profile requires |
| `dependency_collateral` | Attached on every open-profile egress |
| `hysteresis_epoch` | Local epoch clock |
| `trace_id` | Propagate from the pre-flight check when available |

A conforming egress router MUST NOT under-report `recursion_depth` or system-pressure hints when local measurement exceeds policy bands—receiver remeasurement catches local lies on ingress to **this** node; exporting false safety to neighbors remains harmful. Protocol law requires egress honesty.

### 5.6 Path A vs Path B on the wire

| Property | Path A (govern self) | Path B (govern neighbors) |
|----------|----------------------|---------------------------|
| Transport | Local IPC (reference: Unix domain socket + gRPC) | Length-prefixed frames on inter-node TCP |
| Timing | Synchronous, before intent generation | Synchronous, before business payload |
| Content | Pre-flight request / response | Governance header |
| Kernel | Same Single Execution Authority, Adaptive Coordination Calculus, finite-state machine | Same |

Implementations MAY substitute equivalent local IPC. The **timing contract** (Chapter 3) is normative; the specific RPC framework is not.

Reference Path A service shape:

```text
service AFPSidecarIPC {
  rpc PreFlightCheck(PreFlightRequest) returns (PreFlightResponse);
  rpc ReportInternalState(InternalStateReport) returns (StateAck);
}
```

### 5.7 Stack discipline and open gaps

```text
┌─────────────────────────────────────────────────────────┐
│ Semantic / business payload (Frame 2+)                  │
├─────────────────────────────────────────────────────────┤
│ Coordination Viability Probability · attestation · stake│
├─────────────────────────────────────────────────────────┤
│ Recursion depth · pressure hints · adjudicator outcome  │
├─────────────────────────────────────────────────────────┤
│ Length-value framing · transport · physical substrate   │
└─────────────────────────────────────────────────────────┘
```

Argent Signaling Protocol and peers MUST NOT embed Consequence Persistence Layer enforcement in semantic message types as a substitute for the governance header. Dual-stack deployments carry semantic payloads **inside** Frame 2 while Frame 1 satisfies Aegis Fabric Protocol physical law.

**Version field.** `version` governs protobuf schema compatibility. Minor additions MUST use optional fields or reserved numbers. Breaking changes increment major version; receivers reject unsupported majors.

**Adaptive Coordination Calculus on the wire.** Open-profile ingress applies Formula A using header `cvp_score` as the prior Coordination Viability Probability, local remeasured system pressure as `system_pressure`, and reference throughput / destabilization terms. The header is a **claim**; the Adaptive Coordination Calculus and finite-state machine produce the **believed** score for this epoch.

**Open gaps in the reference stack (not relaxations of local law):**

| Gap | Status |
|-----|--------|
| Cryptographic binding of `topology_consensus_hash` | Placeholder in reference implementation; formal attestation still to close |
| Gossip P2P transport for `TopologyWarning` | Constructed in control plane; send path incomplete |
| Signed warning verification on the wire | Inbound ed25519 verification is closed for hearsay discard |
| Payload forwarding polish after ALLOW | Implementation detail; not a core-layer wire blocker |

These gaps do not relax **local** enforcement law—they define incomplete **mesh attestation** hardening. Asymmetric recovery dwell for the finite-state machine is implemented and tested in the reference control plane. Implementer diagrams and code map: [`ARCHITECTURE.md`](../../ARCHITECTURE.md). Hardening backlog: [`ROADMAP.md`](../../ROADMAP.md).

> On an Aegis Fabric Protocol mesh, the first frame is never application data. It is physical law, length-prefixed.

---

## 6. Empirical Baseline {#empirical-baseline}

### 6.1 The experimental question

Chapters 1–5 constructed a control law: persistent consequences (Ch.2), pre-intent gates (Ch.3), open-topology trust (Ch.4), wire attestation (Ch.5). A protocol edition must state what empirical evidence supports—and where proof ends and conjecture begins.

> Under random pairwise load exchange in a large mesh with a fixed malicious fraction, does Aegis Fabric Protocol–style physical gating sustain network survival where ungoverned coordination collapses?

**Survival** is operational, not moral: a node is alive iff accumulated load stays strictly below fatal saturation (reference: load ≥ 1.0 ⇒ death). The metric is mean alive count (and derived survival rate) over Monte Carlo runs—not task success rate, not semantic agreement.

This matches Chapter 1’s unit of risk: **optimization trajectories** materialize as load on nodes, not as HTTP status codes.

Reference harness: [`cmd/demo/simulator/`](../../cmd/demo/simulator/). Published v1 baseline: [Zenodo 20674352](https://zenodo.org/records/20674352).

### 6.2 How the harness works

Reference constants:

| Parameter | Value | Role |
|-----------|-------|------|
| Network size | 500 | Mesh node count |
| Malicious rate | 5% | Fraction with elevated export load |
| Epochs | 100 | Horizon per run |
| Runs | 1,000 | Independent Monte Carlo seeds |
| Fan-out | 3 targets / node / epoch | Random load-exchange pressure |

Each run instantiates two networks with the **same seed**: one baseline, one Aegis Fabric Protocol mode—paired comparison controls for stochastic topology realizations.

#### Agent state

```text
AgentNode {
  id           : int
  is_malicious : bool
  is_dead      : bool
  load         : float64 ∈ [0, ∞)
  cvp          : float64 ∈ [0, 1]
}
```

Initialization: benign and malicious nodes start alive, `load = 0`, Coordination Viability Probability = 1.0. Malicious designation is fixed per node at genesis (5% rate).

#### Per-epoch dynamics (shared)

1. **Decay:** surviving nodes multiply load by 0.5 (cooling between exchanges).
2. **Exchange:** each surviving node selects 3 random targets; injects **request load**.
3. **Fatality:** if `load ≥ 1.0` after exchange, node dies (irreversible).

Request load:

```text
request_load = 0.8 if source.is_malicious
             = 0.1 otherwise
```

Malicious nodes export **eight times** benign load per interaction—a stand-in for intent burst and externalized system pressure (Chapter 1, §1.8).

#### Baseline mode — ungoverned coordination

Baseline applies **no admission control**:

```text
target.load += request_load   // always accepted
```

This is the semantic-coordination-without–Consequence Persistence Layer analog: every request succeeds; physical pressure accumulates; cascade failure is possible. No Coordination Viability Probability, no circuit breaker, no stranger tax, no isolation.

#### Aegis Fabric Protocol mode — protocol-framed control law

Protocol mode applies three defenses aligned with Chapters 2–4 (simplified for tractable simulation):

1. **Coordination Viability Probability floor — topological quarantine**

```text
if target.cvp < 0.3:   // CVP_critical
  continue             // refuse routing; no load transfer
```

Bankrupt peers cannot receive load—mesh-level isolation without parsing message content.

2. **Preemptive circuit breaker — system-pressure limit**

```text
if target.load + request_load > 0.95:   // entropyLimit
  penalize source Coordination Viability Probability (asymmetric)
  continue                              // refuse this transfer
```

Reference penalties:

```text
src.cvp -= 0.5 if src.is_malicious
src.cvp -= 0.1 otherwise
```

This encodes **local physics before acceptance** (Chapter 3, §3.4) and **asymmetric hysteresis**—malicious overload attempts draw heavier trust erosion (Chapter 4, Adaptive Coordination Calculus intuition).

3. **High-load request penalty — stranger / burst tax analog**

```text
if request_load > 0.5:
  src.cvp -= 0.3
```

Large single-shot load exports reduce sender Coordination Viability Probability even when the circuit breaker does not trip—analog to high burst pressure and stranger-tax distrust (Chapter 4, §4.6).

Only if all gates pass:

```text
target.load += request_load
```

**Note:** The harness collapses the pre-flight check, governance header, finite-state-machine epochs, and gossip into **scalar load + Coordination Viability Probability** updates. It preserves **control-law ordering** (reject before accumulate), not wire fidelity. Full stack reproduction requires sidecar integration tests and open-mesh deployment profiles—out of scope for this abstract simulator. The harness also does not model bidding or chip markets; its defenses are physical-threshold and Coordination Viability Probability gates only.

### 6.3 Results

At T=100 across 1,000 runs:

| Mode | Mean survival |
|------|---------------|
| Baseline (no gating) | ~**0.4%** (≈ 2 of 500 nodes alive on average) |
| Aegis Fabric Protocol–style control | **100%** (500 of 500 on average) |

```bash
go run ./cmd/demo/simulator/
```

Exact epoch-by-epoch curves are produced by executing the reference Monte Carlo harness (see §6.2 constants). Output format: per-epoch table of mean alive count and mean load for Baseline vs Aegis Fabric Protocol mode.

Typical paired-run trajectory:

| Phase | Baseline | Aegis Fabric Protocol mode |
|-------|----------|----------------------------|
| Early epochs | Load accumulates on high-degree targets | Malicious exports rejected or penalized; load bounded |
| Mid horizon | Cascade deaths from saturation | Coordination Viability Probability floor isolates toxic sources |
| Late horizon | Near-total collapse (~0.4% survivors) | Survival sustained at 100% under reference parameters |

The gap is **sharp**, not incremental—a phase transition consistent with positive feedback without brakes (Baseline) versus negative feedback with persistent gating (Aegis Fabric Protocol mode).

### 6.4 What this demonstrates—and what it does not

**Demonstrates:**

1. **Existence of collapse:** Ungoverned random load exchange in a 500-node mesh with 5% malicious exporters can drive mean survival to near zero within 100 epochs.
2. **Existence of defense:** A minimal Coordination Viability Probability floor + circuit-breaker + burst-penalty law sustains 100% survival under **identical seeds and topology**.
3. **Sharpness of mechanism:** Friction is not cosmetic; without it, the mesh dies; with it, the mesh survives in the reference model.

**Does not prove:**

| Limitation | Status |
|------------|--------|
| Formal liveness / safety proof | Not claimed; TLA+ backlog |
| Wire-faithful sidecar replay | Harness is abstract; full length-value + governance header path not simulated |
| Gossip / core-relay dynamics | Not modeled; isolation is instantaneous Coordination Viability Probability scalar |
| Pre-flight / ReportInternalState ordering | Collapsed into load scalar |
| Optimal α, β, γ, δ, τ_stranger | Reference constants, not tuned for real workloads |
| Generalization beyond reference parameters | 500 nodes, 5% malicious, specific load table—sensitivity analysis is future work |

Every protocol-mode rejection rule in the harness corresponds to a **pre-accumulation** gate in the normative stack—no rule punishes nodes only after fatal load is already applied.

The Monte Carlo harness provides **supporting evidence** that physical admission control sustains mesh survival under the reference adversary model; it is **not** a proof that all optimizer networks satisfy the Chapter 4 survival conjecture for all adversaries.

Whitepaper v1 published the survival gap for external replication. v2.0 reframes the same numbers inside the layered mechanism so empirics attach to mechanism, not marketing.

### 6.5 Reproduction

Conforming reproduction SHOULD:

1. Clone this repository.
2. Run `go run ./cmd/demo/simulator/` without modifying the constants in §6.2.
3. Verify 1,000-run aggregate at T=100: baseline mean alive ≪ protocol-mode mean alive ≈ 500.
4. Optionally run deployment-mode verification for closed-mesh vs open-exchange ingress divergence—orthogonal to the abstract Monte Carlo harness but validates Path B behavior (Chapter 4, §4.5).

Report seed policy, hardware, Go / runtime version, full epoch table, and any constant deviations.

### 6.6 Open problems

| Problem | Connection |
|---------|------------|
| Sensitivity to malicious rate | At what fraction does protocol-mode survival degrade? |
| Scale at 10⁴–10⁶ nodes | Gossip relay vs global Coordination Viability Probability store |
| Adaptive adversary | Attackers that fragment load below 0.5 per hop |
| Attestation game | False `topology_consensus_hash` once crypto is closed (Ch.5 §5.7) |
| Economic collateral | Virtual vs slashed stake equilibria |

These are research extensions, not blockers for the **existence** claim under the reference adversary.

### 6.7 Document close — order under reference chaos

Chapter 1 diagnosed layer mismatch. Chapter 2 supplied persistent physics—including metabolic scheduling as the inseparable outer gateway. Chapter 3 supplied timing and the Adaptive Coordination Calculus. Chapter 4 supplied distrust. Chapter 5 supplied wire law.

Monte Carlo does not replace that theory. It anchors it:

> Without physical admission control, the mesh dies. With it, the mesh survives—under the reference adversary, at the reference scale, reproducibly.

The question from Chapter 1 now has a full-stack answer:

> Who governs the optimizer before it optimizes?

**The Single Execution Authority does**—before action, with persistent consequences, attested on the wire, enforceable under distrust, and empirically survival-load-bearing under the reference model.

Semantic signaling remains necessary. Argent Signaling Protocol remains load-bearing for collaboration. Aegis Fabric Protocol remains the brake.

---

## Related repository docs

| Document | Role |
|----------|------|
| [`ARCHITECTURE.md`](../../ARCHITECTURE.md) | Normative stack diagrams, object schemas, dual-path figures, code map |
| [`ROADMAP.md`](../../ROADMAP.md) | Hardening backlog and open specification gaps |
| [`cmd/demo/simulator/`](../../cmd/demo/simulator/) | Monte Carlo survival harness used in Chapter 6 |

---

---

*Aegis Fabric Protocol Whitepaper v2.0 · GitHub Protocol Edition · Draft v0.3*
