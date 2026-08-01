# Chapter 5 — Governance Header and Wire Semantics

## Path B Framing, Field Law, and Attestation Evolution

> **Aegis Fabric Protocol v2.0 · Protocol Edition · Draft v0.2**
>
> *A Physical Constraint Protocol for Autonomous Optimizers*

---

## 5.0 From Trust Semantics to On-Wire Physics

Chapter 4 defined **what** L4 must accomplish: local CVP dynamics, stranger tax, topological quarantine, core-relay gossip. Chapter 3 defined **when** SEA adjudicates on Path B—before business payload admission. This chapter specifies **how physical state crosses the wire**: the GovernanceHeader message, LV framing, field-level enforcement law, and the attestation gaps that remain open in v1.0.

Path A (PreFlight, ReportInternalState) uses local IPC; wire contract reference: `api/afp/v1/sdk_ipc.proto` (Chapter 3). Path B uses **inter-node TCP** with a mandatory governance frame **preceding** application payload. Normative schema: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4; reference protobuf: `api/afp/v1/governance.proto`.

The design law restated:

> **Business payload MUST NOT be interpreted until GovernanceHeader is validated and SEA returns ALLOW or DELAY.**

Semantic content (L5) rides **behind** physical attestation (L2–L4). ASP negotiates intent; GovernanceHeader attests **physics**.

---

## 5.1 Session Model on the Sidecar Mesh

### 5.1.1 Connection pattern

Each outbound optimizer connection to a remote peer targets the peer's **ingress boundary** (sidecar listener). The egress router:

1. Opens TCP to remote ingress.
2. Writes **Frame 1:** serialized GovernanceHeader (LV-wrapped).
3. Writes **Frame 2+:** optional business payload (LV-wrapped), if any.

Ingress reads Frame 1, validates, adjudicates via SEA, then MAY read subsequent frames only on ALLOW or completed DELAY.

This is **not** HTTP header extension. Governance is a **first-class frame** on a dedicated sidecar-to-sidecar stream—out-of-band relative to application protocols carried in Frame 2.

### 5.1.2 LV framing

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

### 5.1.3 Multi-frame streams

A conforming stream on ingress:

```text
[ LV · GovernanceHeader ]  →  SEA adjudication
[ LV · business_payload ]    →  forwarded only if adjudication permits
```

Additional application frames MAY follow by bilateral agreement; AFP normative scope covers **governance frame + first payload gate**. Multiplexed streaming semantics beyond the first payload are implementation-defined (open gap: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

---

## 5.2 GovernanceHeader — Field Semantics

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

### 5.2.1 Identity and versioning

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `packet_id` | Uniqueness hint for dedup / replay resistance | SHOULD be monotonic or high-entropy |
| `version` | Schema version | Receiver MUST reject unknown major versions |
| `hysteresis_epoch` | Sender's epoch clock at attestation time | Correlates with FSM hysteresis (Ch.3–4) |
| `coordination_ttl` | Validity window for attestation (seconds) | Receiver MAY reject stale headers |

Reference egress sets `version = 1`, `coordination_ttl = 120`.

### 5.2.2 Trust cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `cvp_score` | Sender's self-reported coordination viability ∈ [0, 1] | Receiver applies ACC evolution in open profile; floor at 0.3 |
| `topology_consensus_hash` | Attestation binding to topology view / signature material | Open profile: empty ⇒ `HasValidSign = false` ⇒ isolate |

**Protocol law:** `topology_consensus_hash` is the wire hook for **cryptographic attestation**. v1.0 reference implementation uses a placeholder; normative binding is target of Whitepaper v2.0 / v1.1 spec closure ([`ARCHITECTURE.md`](../../ARCHITECTURE.md) §8).

### 5.2.3 Physical cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `entropy_load.resource_asymmetry_ratio` | Declared resource asymmetry | **Hint only**—receiver remeasures locally |
| `entropy_load.dependency_contention_rate` | Declared contention | **Hint only** |
| `recursion_depth` | Sender planner depth | MUST be ≤ policy `maxRecursionDepth` |

**Theorem 5.1 (Local entropy supremacy — wire form):** Let *h* be header-declared entropy and *e* locally remeasured entropy at ingress. SEA MUST evaluate FSM and circuit breaker using *e*, not *h* alone.

Corollary: Honest headers aid telemetry; dishonest headers do not bypass physics.

### 5.2.4 Collateral cluster

| Field | Semantics | Enforcement |
|-------|-----------|-------------|
| `dependency_collateral.collateral_type` | Stake class identifier | Open profile: required for first-seen peers |
| `dependency_collateral.slash_threshold` | Committed slash floor | MUST be ≥ τ_stranger (0.8) for stranger admission |

Egress reference attaches virtual stake (`SYS_VIRTUAL_STAKE`, `slash_threshold = 0.8`). Remote SEA enforces tax; sender attests willingness to be slashed on malicious behavior (Chapter 4, §4.5).

### 5.2.5 Correlation

| Field | Semantics |
|-------|-----------|
| `trace_id` | End-to-end correlation across PreFlight (Path A), GovernanceHeader (Path B), and audit |

PreFlight `trace_id` SHOULD match outbound GovernanceHeader `trace_id` when the intent epoch externalizes—enabling cross-path forensics without parsing business payload.

---

## 5.3 Ingress Validation Pipeline

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

## 5.4 Egress Attestation Obligations

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

## 5.5 Path A Wire — SDK IPC (Abstract)

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

## 5.6 Evolution and Compatibility

### 5.6.1 Version field

`version` governs protobuf schema compatibility. Minor additions MUST use optional fields or reserved numbers. Breaking changes increment major version; receivers reject unsupported majors.

### 5.6.2 Open specification gaps (v1.0)

| Gap | Status | Target |
|-----|--------|--------|
| Cryptographic binding of `topology_consensus_hash` | Placeholder in reference impl | §5.2.2; formal attestation spec |
| Gossip P2P transport for TopologyWarning | Constructed, not sent | Chapter 4 §4.7 |
| Signed TopologyWarning verification | Stub | Paired with hash attestation |
| Payload forwarding after ALLOW | Reference TODO | Enterprise guide; not L2 wire blocker |

These gaps do not relax **local** enforcement law—they define incomplete **mesh attestation** hardening.

### 5.6.3 ACC on the wire

Open-profile ingress applies Formula A using header `cvp_score` as `CVP_old`, local remeasured entropy as `entropy_load`, and reference throughput / destabilization terms. The header is a **claim**; ACC + FSM produce the **believed** score for this epoch.

---

## 5.7 Wire vs. Semantics — Stack Discipline

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

## 5.8 Chapter Conclusion — Attestation Before Payload

The wire contract is deliberately minimal: one LV frame, one protobuf, one SEA evaluation—then, and only then, business bits.

GovernanceHeader is not metadata. It is **the price of admission** to a peer's execution boundary. Local entropy supremacy prevents attestation fraud from scaling. Stranger tax fields bind economic commitment. Trace IDs stitch Path A and Path B into one forensic timeline.

Chapter 6 returns to evidence: reproducing the Monte Carlo survival gap with protocol-framed adversary models—and stating what the simulation proves and what it does not.

For now:

> **On the AFP mesh, the first frame is never application data. It is physical law, length-prefixed.**

---

*Draft v0.2 · Protocol Edition · Normative schema: [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §4.4 · Reference protos: `api/afp/v1/governance.proto`, `api/afp/v1/sdk_ipc.proto`.*
