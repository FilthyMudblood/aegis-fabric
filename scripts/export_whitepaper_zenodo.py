#!/usr/bin/env python3
"""Export a Zenodo-ready AFP Whitepaper v2 (standalone, no GitHub / repo links).

Source: docs/whitepaper-v2/whitepaper-v2-protocol-edition.md (GitHub edition).
Output: artifacts/zenodo/ (gitignored). Do not commit those files.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "docs/whitepaper-v2/whitepaper-v2-protocol-edition.md"
OUT_DIR = ROOT / "artifacts/zenodo"
OUT_MD = OUT_DIR / "aegis-fabric-protocol-v2-whitepaper.md"
OUT_META = OUT_DIR / "zenodo-metadata.json"
OUT_README = OUT_DIR / "UPLOAD.md"

ZENODO_RECORD = "21330527"
ZENODO_DOI = f"10.5281/zenodo.{ZENODO_RECORD}"
ZENODO_URL = f"https://doi.org/{ZENODO_DOI}"
CC_LICENSE_URL = "https://creativecommons.org/licenses/by-nc-sa/4.0/"


def strip_heading_anchors(text: str) -> str:
    return re.sub(r" \{#[^}]+\}", "", text)


def plain_toc(text: str) -> str:
    toc_block = """## Table of Contents

1. The Optimization Crisis
2. Consequence Persistence Layer
3. Pre-Intent Enforcement
4. Open-Network Topology
5. Governance Header & Wire Semantics
6. Empirical Baseline"""
    text = re.sub(
        r"## Table of Contents\n\n(?:\d+\. \[.*?\]\(#.*?\)\n)+",
        toc_block + "\n",
        text,
        count=1,
    )
    return text


def replace_header(text: str) -> str:
    header = f"""# Aegis Fabric Protocol v2.0 — Protocol Edition

> **Zenodo publication · Draft v0.3** · *A Physical Constraint Protocol for Autonomous Optimizers*
>
> Standalone archival edition. Mechanisms, formulas, wire schemas, and empirical parameters are defined in-line. This document contains no repository hyperlinks.
>
> **Document license:** Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International ([CC BY-NC-SA 4.0]({CC_LICENSE_URL})).
>
> **v1** (empirical archive): Whitepaper v1 empirical baseline — Zenodo record 20674352 · DOI [10.5281/zenodo.20674352](https://doi.org/10.5281/zenodo.20674352)
>
> **This version DOI:** [{ZENODO_DOI}]({ZENODO_URL})

---"""
    text = re.sub(
        r"# Aegis Fabric Protocol v2\.0 — Protocol Edition\n\n>.*?\n\n---",
        header,
        text,
        count=1,
        flags=re.DOTALL,
    )
    return text


def strip_related_repo_docs(text: str) -> str:
    return re.sub(
        r"\n## Related repository docs\n.*?(?=\n---\n\n---\n|\n---\n\n\*Aegis|\Z)",
        "\n",
        text,
        flags=re.DOTALL,
    )


def replace_architecture_links(text: str) -> str:
    def repl(m: re.Match[str]) -> str:
        section = m.group(1) if m.lastindex and m.lastindex >= 1 else ""
        return f"Normative Architecture Specification{section}"

    text = re.sub(
        r"\[`ARCHITECTURE\.md`\]\(\.\./\.\./ARCHITECTURE\.md\)( §[\d.]+(?:–§[\d.]+)?)?",
        repl,
        text,
    )
    text = re.sub(
        r"\[`ARCHITECTURE\.md`\]\(\.\./\.\./ARCHITECTURE\.md\)",
        "Normative Architecture Specification",
        text,
    )
    text = re.sub(
        r"Implementer diagrams and code map:\s*Normative Architecture Specification\.\s*",
        "",
        text,
    )
    text = re.sub(
        r"Hardening backlog:\s*protocol hardening backlog\.",
        "Open hardening items are summarized in §5.7.",
        text,
    )
    return text


def replace_roadmap_links(text: str) -> str:
    text = re.sub(
        r"\[`ROADMAP\.md`\]\(\.\./\.\./ROADMAP\.md\)",
        "the protocol hardening backlog",
        text,
    )
    return text


def replace_zenodo_links(text: str) -> str:
    text = re.sub(
        r"\[Zenodo(?: record)? 20674352\]\(https://zenodo\.org/records/20674352\)",
        "Whitepaper v1 empirical baseline (Zenodo DOI 10.5281/zenodo.20674352)",
        text,
    )
    return text


def replace_repo_paths(text: str) -> str:
    replacements = [
        (r"`internal/control/acc_kernel\.go`", "the reference Adaptive Coordination Calculus kernel"),
        (r"`internal/dataplane/codec\.go`", "the reference wire codec"),
        (r"`api/afp/v1/sdk_ipc\.proto`", "the SDK IPC schema (PreFlight / ReportInternalState)"),
        (r"`api/afp/v1/governance\.proto`", "the GovernanceHeader protobuf schema"),
        (r"`cmd/demo/simulator/main\.go`", "the reference Monte Carlo simulator constants"),
        (r"\[`cmd/demo/simulator/`\]\(\.\./\.\./cmd/demo/simulator/\)", "the reference Monte Carlo harness"),
        (r"`cmd/demo/simulator/`", "the reference Monte Carlo harness"),
        (r"`scripts/verify_modes\.sh`", "the reference deployment-mode verification harness"),
    ]
    for pattern, repl in replacements:
        text = re.sub(pattern, repl, text)
    return text


def replace_go_run_blocks(text: str) -> str:
    text = re.sub(
        r"\n```bash\ngo run \./cmd/demo/simulator/\n```\n",
        "\nReproduce with the reference Monte Carlo harness distributed alongside the Aegis Fabric Protocol reference implementation (constants in §6.2).\n",
        text,
    )
    text = re.sub(r"`go run \./cmd/demo/simulator/`", "the reference Monte Carlo harness", text)
    text = re.sub(
        r"1\. Clone this repository\.\n"
        r"2\. Run `?go run \./cmd/demo/simulator/`? without modifying the constants in §6\.2\.",
        "1. Obtain the reference Monte Carlo harness distributed with the Aegis Fabric Protocol reference implementation.\n"
        "2. Execute the simulator without modifying the reference constants listed in §6.2.",
        text,
    )
    text = re.sub(
        r"1\. Clone this repository\.\n"
        r"2\. Run the reference Monte Carlo harness without modifying the constants in §6\.2\.",
        "1. Obtain the reference Monte Carlo harness distributed with the Aegis Fabric Protocol reference implementation.\n"
        "2. Execute the simulator without modifying the reference constants listed in §6.2.",
        text,
    )
    text = re.sub(
        r"Report seed policy, hardware, Go / runtime version,",
        "Report seed policy, hardware, runtime version,",
        text,
    )
    return text


def neutralize_repo_mentions(text: str) -> str:
    """Remove GitHub / repository phrasing that should not appear on Zenodo."""
    replacements = [
        (
            r"This is the repository whitepaper: narrative first, full names in prose, mechanisms explained in-line\.\s*",
            "",
        ),
        (
            r"Stack diagrams, protobuf schemas, and code map for implementers:.*\n",
            "",
        ),
        (
            r"\*\*GitHub edition · Draft v0\.3\*\*",
            "**Zenodo publication · Draft v0.3**",
        ),
        (
            r"GitHub Protocol Edition",
            "Protocol Edition",
        ),
        (
            r"this repository",
            "the reference implementation distribution",
        ),
        (
            r"Clone this repository",
            "Obtain the reference implementation distribution",
        ),
        (
            r"see roadmap / architecture open gaps",
            "see §5.7 open gaps",
        ),
        (
            r"Implementer diagrams and code map: Normative Architecture Specification\. Hardening backlog: the protocol hardening backlog\.",
            "Open hardening items are listed in §5.7; they do not relax local enforcement law.",
        ),
        (
            r"Implementer diagrams and code map: Normative Architecture Specification\. Open hardening items are summarized in §5\.7\.",
            "Open hardening items are listed in §5.7; they do not relax local enforcement law.",
        ),
    ]
    for pattern, repl in replacements:
        text = re.sub(pattern, repl, text, flags=re.IGNORECASE)
    return text


PUBLICATION_BACKMATTER = f"""
---

## Citation

**APA:**

He, M. (2026). Aegis Fabric Protocol (AFP) v2.0 — Protocol Edition: A Physical Constraint Protocol for Autonomous Optimizers. Zenodo. {ZENODO_URL}

**BibTeX:**

```bibtex
@techreport{{he2026afp,
  author       = {{He, Muchen}},
  title        = {{{{Aegis Fabric Protocol (AFP) v2.0 --- Protocol Edition: A Physical Constraint Protocol for Autonomous Optimizers}}}},
  year         = {{2026}},
  institution  = {{Aegis Fabric Protocol}},
  type         = {{Whitepaper}},
  doi          = {{{ZENODO_DOI}}},
  url          = {{{ZENODO_URL}}}
}}
```

---

## Legal Notice

This whitepaper is provided as-is for architecture evaluation and research. It does not constitute a security certification, legal advice, or a warranty of fitness for any particular purpose. Deployments handling regulated data require independent risk assessment, penetration testing, and operational controls appropriate to your jurisdiction and industry.

**Copyright:** © 2026 He, Muchen / Aegis Fabric Protocol. All rights reserved except as granted under the license below.

**Document license:** Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International (CC BY-NC-SA 4.0).  
License text: {CC_LICENSE_URL}

You may share and adapt this document with attribution for non-commercial purposes under the same license. Commercial republication or redistribution outside that license requires separate permission from the copyright holder.

**Reference software:** Reference code distributed separately is typically offered under the MIT License and is an experimental optimizer-runtime implementation. It has not undergone third-party security audit. Software licensing is independent of this document’s CC BY-NC-SA 4.0 terms.

**Trademarks:** AFP and Aegis Fabric Protocol are project identifiers. Third-party names (for example LangGraph, LangChain, CrewAI, AutoGen, Kafka, Prometheus) are used for comparative description only and remain the property of their respective owners.

**Relationship to academic literature:** This protocol shares the name “Aegis” and a high-level interest in agent safety with certain recent academic publications. The engineering implementation, control-plane structure, wire contracts, and Adaptive Coordination Calculus defined in this document are completely distinct and independent of that literature.

---

*Aegis Fabric Protocol — Consequence Persistence Layer for autonomous optimizers.*  
*Zenodo DOI: {ZENODO_DOI}*
"""


def replace_footer(text: str) -> str:
    text = re.sub(
        r"\n---\n\n---\n\n\*Aegis Fabric Protocol Whitepaper[^\n]*\*\s*$",
        "",
        text,
        flags=re.DOTALL,
    )
    text = re.sub(
        r"\n---\n\n---\n\n\*AFP Whitepaper[^\n]*\*\.?\s*$",
        "",
        text,
        flags=re.DOTALL,
    )
    text = re.sub(
        r"\n---\n\n## Citation\n.*\Z",
        "",
        text,
        flags=re.DOTALL,
    )
    text = re.sub(
        r"\n\*AFP — Aegis Fabric Protocol\.[^\n]*\*\s*$",
        "",
        text,
    )
    text = re.sub(
        r"\n\*Aegis Fabric Protocol Whitepaper[^\n]*\*\s*$",
        "",
        text,
    )
    return text.rstrip()


def scrub_remaining_links(text: str) -> str:
    # Drop any remaining relative markdown links; keep link text if useful.
    text = re.sub(r"\[([^\]]+)\]\(\.\./[^)]+\)", r"\1", text)
    text = re.sub(r"\[([^\]]+)\]\(\./[^)]+\)", r"\1", text)
    # Neutralize absolute GitHub / raw repo URLs if any slipped in.
    text = re.sub(
        r"https?://(?:www\.)?github\.com/[^\s\)\]]+",
        "[repository URL omitted in Zenodo edition]",
        text,
        flags=re.IGNORECASE,
    )
    text = re.sub(
        r"https?://raw\.githubusercontent\.com/[^\s\)\]]+",
        "[repository URL omitted in Zenodo edition]",
        text,
        flags=re.IGNORECASE,
    )
    # Collapse accidental double spaces from link stripping (not inside code fences).
    parts = re.split(r"(```.*?```)", text, flags=re.DOTALL)
    for i, part in enumerate(parts):
        if not part.startswith("```"):
            parts[i] = re.sub(r"[^\S\n]{2,}", " ", part)
    return "".join(parts)


def transform(source: str) -> str:
    text = source
    text = replace_header(text)
    text = plain_toc(text)
    text = strip_heading_anchors(text)
    text = strip_related_repo_docs(text)
    text = replace_architecture_links(text)
    text = replace_roadmap_links(text)
    text = replace_zenodo_links(text)
    text = replace_repo_paths(text)
    text = replace_go_run_blocks(text)
    text = neutralize_repo_mentions(text)
    text = replace_footer(text)
    text = scrub_remaining_links(text)
    return text


def write_metadata() -> None:
    meta = {
        "upload_type": "publication",
        "publication_type": "other",
        "title": "Aegis Fabric Protocol (AFP) v2.0 — Protocol Edition: A Physical Constraint Protocol for Autonomous Optimizers",
        "description": (
            "Standalone Protocol Edition whitepaper for the Aegis Fabric Protocol (AFP) v2.0. "
            "Defines the Consequence Persistence Layer, Single Execution Authority, Adaptive "
            "Coordination Calculus, pre-intent enforcement, open-network topology, GovernanceHeader "
            "wire semantics, and an empirical Monte Carlo baseline. Zenodo edition with no "
            "repository-internal hyperlinks."
        ),
        "creators": [
            {
                "name": "He, Muchen",
                "affiliation": "Aegis Fabric Protocol",
            }
        ],
        "keywords": [
            "AFP",
            "Aegis Fabric Protocol",
            "multi-agent systems",
            "autonomous agents",
            "agent coordination",
            "consequence persistence",
            "pre-intent enforcement",
            "Adaptive Coordination Calculus",
            "optimizer governance",
        ],
        "license": "cc-by-nc-sa-4.0",
        "access_right": "open",
        "language": "eng",
        "version": "2.0.0-draft.v0.3",
        "notes": (
            "Document license CC BY-NC-SA 4.0. Reference software (if distributed separately) "
            "is typically MIT-licensed and experimentally unaudited. This manuscript is "
            "independent of other academic works that share the name Aegis."
        ),
        "related_identifiers": [
            {
                "identifier": ZENODO_DOI,
                "relation": "isIdenticalTo",
                "resource_type": "publication-other",
                "scheme": "doi",
            },
            {
                "identifier": "10.5281/zenodo.20674352",
                "relation": "isNewVersionOf",
                "resource_type": "publication-other",
                "scheme": "doi",
            },
        ],
    }
    OUT_META.write_text(json.dumps(meta, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def write_upload_readme() -> None:
    OUT_README.write_text(
        f"""# Zenodo upload package (local only — not in Git)

Generated by `scripts/export_whitepaper_zenodo.py`. This directory is gitignored.

## Files to upload

| File | Purpose |
|------|---------|
| `aegis-fabric-protocol-v2-whitepaper.md` | **Main manuscript** (upload this) |
| `zenodo-metadata.json` | Copy fields into the Zenodo web form / API |
| `UPLOAD.md` | This checklist |

Optional but recommended: also upload a PDF export of the same Markdown for readers who prefer PDF (Zenodo accepts multiple files under one DOI).

## Pre-upload checklist

1. **No GitHub / repo links** — the Markdown must not contain `github.com`, relative `../` links, or paths like `ARCHITECTURE.md` as hyperlinks.
2. **License on Zenodo form** — select **CC-BY-NC-SA-4.0** (must match the Legal Notice in the file).
3. **Creators** — confirm `He, Muchen` and affiliation before publish.
4. **Related identifiers** — keep v1 DOI `10.5281/zenodo.20674352` as *isNewVersionOf*.
5. **Reserved DOI** — this package assumes DOI `{ZENODO_DOI}` / record `{ZENODO_RECORD}`. If Zenodo assigns a different concept DOI or version DOI on first publish, update the Citation section and regenerate, or edit the Citation block after reserve.
6. **Version field** — use `2.0` (or `2.0.0-draft.v0.3` if you want draft visibility in metadata).
7. **Communities / grants** — add only if you have a real community or funder ID; do not invent them.

## Upload steps (zenodo.org)

1. Sign in → **New upload** (or **New version** of record {ZENODO_RECORD} if updating).
2. Upload `aegis-fabric-protocol-v2-whitepaper.md` (+ optional PDF).
3. Fill metadata from `zenodo-metadata.json`.
4. Set license **CC-BY-NC-SA-4.0**, access **Open**, language **English**.
5. Set publication date and version.
6. Preview → Publish.

Published target: https://zenodo.org/records/{ZENODO_RECORD} · DOI `{ZENODO_DOI}`

## Regenerate

```bash
python3 scripts/export_whitepaper_zenodo.py
```

Then re-check with:

```bash
rg -n 'github\\.com|/\\.\\./|ARCHITECTURE\\.md|ROADMAP\\.md|Clone this repository' artifacts/zenodo/aegis-fabric-protocol-v2-whitepaper.md || echo OK
```
""",
        encoding="utf-8",
    )


def audit(text: str) -> list[str]:
    problems: list[str] = []
    patterns = [
        (r"github\.com", "GitHub URL"),
        (r"raw\.githubusercontent\.com", "raw GitHub URL"),
        (r"\]\(\.\./", "relative parent link"),
        (r"\]\(\./", "relative link"),
        (r"\[`ARCHITECTURE\.md`\]", "ARCHITECTURE.md markdown link"),
        (r"\[`ROADMAP\.md`\]", "ROADMAP.md markdown link"),
        (r"Clone this repository", "clone-repo instruction"),
        (r"go run \./cmd/", "local go run command"),
        (r"GitHub edition", "GitHub edition label"),
        (r"Related repository docs", "related-repo section"),
    ]
    for pattern, label in patterns:
        if re.search(pattern, text, flags=re.IGNORECASE):
            problems.append(label)
    if "## Citation" not in text:
        problems.append("missing Citation section")
    if "## Legal Notice" not in text:
        problems.append("missing Legal Notice section")
    if "CC BY-NC-SA 4.0" not in text and "CC-BY-NC-SA-4.0" not in text:
        problems.append("missing CC BY-NC-SA 4.0 license text")
    if "© 2026" not in text:
        problems.append("missing copyright line")
    return problems


def main() -> None:
    if not SOURCE.is_file():
        raise SystemExit(f"Source not found: {SOURCE}")

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    source_text = SOURCE.read_text(encoding="utf-8")
    out_text = transform(source_text).rstrip()
    # Avoid --- / --- double rules before Citation backmatter.
    out_text = re.sub(r"(?:\n---)+\s*\Z", "", out_text).rstrip()
    out_text = out_text + "\n" + PUBLICATION_BACKMATTER.lstrip("\n")
    OUT_MD.write_text(out_text, encoding="utf-8")
    write_metadata()
    write_upload_readme()

    problems = audit(out_text)
    if problems:
        print("AUDIT WARNINGS:", file=sys.stderr)
        for p in problems:
            print(f"  - {p}", file=sys.stderr)
        raise SystemExit(1)

    print(f"Wrote {OUT_MD} ({OUT_MD.stat().st_size} bytes)")
    print(f"Wrote {OUT_META}")
    print(f"Wrote {OUT_README}")
    print("Audit: OK (no GitHub/repo links; Citation + Legal Notice present)")


if __name__ == "__main__":
    main()
