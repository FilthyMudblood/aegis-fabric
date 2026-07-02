# Chapter 6 — Empirical Baseline

## Monte Carlo Reproduction, Protocol-Framed Adversary, and Claims Boundary

> **Aegis Fabric Protocol v2.0 · Protocol Edition · Draft v0.2**
>
> *A Physical Constraint Protocol for Autonomous Optimizers*

---

## 6.0 From Mechanism to Evidence

Chapters 1–5 constructed a **control law**: persistent consequences (Ch.2), pre-intent gates (Ch.3), open-topology trust (Ch.4), wire attestation (Ch.5). A protocol edition must state what empirical evidence supports—and where proof ends and conjecture begins.

Version 1.0 established a **survival gap** under Monte Carlo open-mesh stress. Version 2.0 explains **why** the mechanism must be physical, persistent, and out-of-band (Chapter 1, §1.6). This chapter **reproduces** the baseline protocol-framed experiment and defines the **claims boundary**: what the harness demonstrates, what it abstracts away, and what remains for formal verification.

Reference harness: `cmd/demo/simulator/` · Published v1 baseline: [Zenodo record 20674352](https://zenodo.org/records/20674352).

---

## 6.1 Experimental Question

> **Under random pairwise load exchange in a large mesh with a fixed malicious fraction, does AFP-style physical gating sustain network survival where ungoverned coordination collapses?**

**Survival** is operational, not moral: a node is *alive* iff its accumulated load remains strictly below fatal saturation (reference: load ≥ 1.0 ⇒ death). The metric is **mean alive count** (and derived survival rate) over Monte Carlo runs—not task success rate, not semantic agreement.

This matches Chapter 1's unit of risk: **optimization trajectories** materialize as load on nodes, not as HTTP status codes.

---

## 6.2 Harness Parameters

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

## 6.3 Node and Load Model

### 6.3.1 Agent state

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

### 6.3.2 Per-epoch dynamics (shared)

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

## 6.4 Baseline Mode — Ungoverned Coordination

Baseline applies **no admission control**:

```text
target.load += request_load    // always accepted
```

This is the semantic-coordination-without-CPL analog: every request succeeds; physical pressure accumulates; cascade failure is possible. No CVP, no circuit breaker, no stranger tax, no isolation.

**Interpretation:** Baseline models **optimizers that coordinate freely** without AFP physical law—ASP-alone incomplete (Chapter 1, Theorem 1.1).

---

## 6.5 AFP Mode — Protocol-Framed Control Law

AFP mode applies three defenses aligned with Chapters 2–4 (simplified for tractable simulation):

### 6.5.1 CVP floor — topological quarantine

```text
if target.cvp < CVP_critical (0.3):
    continue    // refuse routing; no load transfer
```

Bankrupt peers cannot receive load—mesh-level isolation without parsing message content.

### 6.5.2 Preemptive circuit breaker — entropy limit

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

### 6.5.3 High-entropy request penalty — stranger / burst tax analog

```text
if request_load > 0.5:
    src.cvp -= 0.3
```

Large single-shot load exports reduce sender CVP even when the circuit breaker does not trip—analog to high burst pressure and stranger-tax distrust (Chapter 4, §4.5).

### 6.5.4 Safe admission

Only if all gates pass:

```text
target.load += request_load
```

**Note:** The harness collapses PreFlight, GovernanceHeader, FSM epochs, and gossip into **scalar load + CVP** updates. It preserves **control-law ordering** (reject before accumulate), not wire fidelity. Full stack reproduction requires sidecar integration tests and open-mesh deployment profiles—out of scope for this abstract simulator.

---

## 6.6 Reported Results

### 6.6.1 v1 headline (Monte Carlo mean at horizon)

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

### 6.6.2 Qualitative trajectory

Typical paired-run behavior:

| Phase | Baseline | AFP |
|-------|----------|-----|
| Early epochs | Load accumulates on high-degree targets | Malicious exports rejected or penalized; load bounded |
| Mid horizon | Cascade deaths from saturation | CVP floor isolates toxic sources |
| Late horizon | Near-total collapse (~0.4% survivors) | Survival sustained at 100% under reference parameters |

The gap is **sharp**, not incremental—a phase transition consistent with **positive feedback without brakes** (Baseline) vs **negative feedback with persistent gating** (AFP).

---

## 6.7 Mapping Harness to Protocol Chapters

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

## 6.8 Claims Boundary

### 6.8.1 What this experiment demonstrates

1. **Existence of collapse:** Ungoverned random load exchange in a 500-node mesh with 5% malicious exporters can drive mean survival to near zero within 100 epochs.
2. **Existence of defense:** A minimal CVP + circuit-breaker + burst-penalty law sustains 100% survival under **identical seeds and topology**.
3. **Sharpness of mechanism:** Friction is not cosmetic; without it, the mesh dies; with it, the mesh survives in the reference model.

### 6.8.2 What this experiment does not prove

| Limitation | Status |
|------------|--------|
| **Formal liveness / safety proof** | Not claimed; TLA+ backlog ([`ROADMAP.md`](../../ROADMAP.md)) |
| **Wire-faithful sidecar replay** | Harness is abstract; full LV + GovernanceHeader path not simulated |
| **Gossip / core-relay dynamics** | Not modeled; isolation is instantaneous CVP scalar |
| **PreFlight / ReportInternalState ordering** | Collapsed into load scalar |
| **Optimal α, β, γ, δ, τ_stranger** | Reference constants, not tuned for real workloads |
| **Generalization beyond reference parameters** | 500 nodes, 5% malicious, specific load table—sensitivity analysis is future work |

**Theorem 6.1 (Empirical scope — stated modestly):** The Monte Carlo harness provides **supporting evidence** that physical admission control sustains mesh survival under the reference adversary model; it is **not** a proof that all optimizer networks satisfy Conjecture 4.1 for all adversaries.

### 6.8.3 Relationship to v1 publication

Whitepaper v1 (Zenodo) documented the survival gap for external replication. v2.0 Protocol Edition **reframes** the same numbers inside the six-layer stack, dual-path SEA, and CVP formalism—so empirical results **attach to mechanism**, not marketing.

---

## 6.9 Reproduction Protocol

Conforming reproduction SHOULD:

1. Clone reference implementation repository.
2. Run `go run ./cmd/demo/simulator/` without modification to constants.
3. Verify 1,000-run aggregate at T=100: Baseline mean alive ≪ AFP mean alive ≈ 500.
4. Optionally pair with `scripts/verify_modes.sh` for sidecar **profile** behavior (closed mesh vs open exchange)—orthogonal to Monte Carlo abstract harness but validates ingress path divergence (Chapter 4, §4.4).

Report: seed policy, hardware, Go version, full epoch table, and any constant deviations.

---

## 6.10 Open Problems

| Problem | Connection |
|---------|------------|
| **Sensitivity to MaliciousRate** | At what fraction does AFP-mode survival degrade? |
| **Scale at 10⁴–10⁶ nodes** | Gossip relay vs global CVP store |
| **Adaptive adversary** | Attackers that fragment load below 0.5 per hop |
| **Attestation game** | False `topology_consensus_hash` once crypto is closed (Ch.5 §5.6) |
| **Economic collateral** | Virtual vs slashed stake equilibria |

These are research extensions, not v1.0 blockers for the **existence** claim.

---

## 6.11 Document Conclusion — Order Proven Under Reference Chaos

The Optimization Crisis (Chapter 1) diagnosed layer mismatch. CPL and SEA (Chapter 2) supplied persistent physics. PreFlight (Chapter 3) supplied timing. Open topology (Chapter 4) supplied distrust. GovernanceHeader (Chapter 5) supplied wire law.

Monte Carlo does not replace that theory. It **anchors** it:

> **Without physical admission control, the mesh dies. With it, the mesh survives—under the reference adversary, at the reference scale, reproducibly.**

The naked optimizer question from Chapter 1 now has a full stack answer:

> **Who governs the optimizer before it optimizes?**

**SEA does—pre-intent, with persistent consequences, attested on the wire, trusted under distrust, and empirically survival-load-bearing.**

Semantic signaling remains necessary. ASP remains load-bearing. AFP remains the brake.

---

*Draft v0.2 · Protocol Edition · Harness: `cmd/demo/simulator/` · Normative stack: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) · v1 empirical archive: [Zenodo 20674352](https://zenodo.org/records/20674352).*
