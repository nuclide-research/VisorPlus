[![Claude Code Friendly](https://img.shields.io/badge/Claude_Code-Friendly-blueviolet?logo=anthropic&logoColor=white)](https://claude.ai/code)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nicholas-Kloster/VisorPlus)](https://goreportcard.com/report/github.com/Nicholas-Kloster/VisorPlus)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

```

  AI/LLM Infrastructure Hunt & Assessment Platform
  github.com/Nicholas-Kloster/VisorPlus · @nuclide
```

**VisorPlus** is a unified CLI that orchestrates the full Nuclide AI/LLM security assessment workflow in a single command. It combines JAXEN, VisorSD, VisorCorpus, BARE, and aimap into one cohesive chain — from initial Shodan discovery all the way to adversarial corpus generation.

---

## Install

```bash
git clone https://github.com/Nicholas-Kloster/VisorPlus.git
cd VisorPlus
go build -o visorplus .

# Install all dependent Nuclide tools in one step
./visorplus install
```

**Requires:** Go 1.21+, Shodan API key

---

## The Full Chain

VisorPlus runs the complete AI/LLM security assessment workflow:

```
VisorSD audit      →  discover beginner-stack exposure queries
Shodan hunt        →  harvest live hosts (via JAXEN)
/api/tags sweep    →  enumerate models on every found host
Red-flag detection →  abliterated models, offensive AI brands, cloud quota exposure
Target assess      →  nmap + whois + SSH keys + passive intel (GreyNoise, Shodan, DNSBL)
VisorCorpus        →  generate adversarial LLM prompt corpus for any live endpoint
```

---

## Quick Start

```bash
# Set your Shodan key
export SHODAN_API_KEY="your_key_here"

# Run the full workflow end-to-end
./visorplus full

# Or with a specific dork
./visorplus full 'http.html:"Ollama is running" -port:443'

# Scope to a target org or network
./visorplus full -org "Acme Corp"
./visorplus full -asn AS48090
./visorplus full -net 93.123.0.0/16
```

---

## Commands

### `install` — Set up all Nuclide tools
```bash
./visorplus install
```
Clones and builds JAXEN, VisorSD, VisorCorpus, BARE, and aimap into `~/Tools/`. Prompts for your Shodan key if not already configured.

---

### `audit` — VisorSD beginner-stack audit
```bash
./visorplus audit              # global scan, all AI/LLM categories
./visorplus audit -dry-run     # preview queries, no credits spent
./visorplus audit -org "Acme"  # scope to org
./visorplus audit -format json -out results.json
```
Runs VisorSD's ~20 hardcoded AI/LLM infra queries, severity-ranked CRITICAL → LOW.

---

### `hunt` — JAXEN Shodan harvest
```bash
./visorplus hunt 'http.html:"Ollama is running" -port:443'
./visorplus hunt 'http.title:"Open WebUI"' -out ./my-run
```
Harvests up to 40 CDN-filtered hosts into `empire.db` + exports `recon_dump.json` and `summary.csv`.

---

### `enum` — Ollama model enumeration
```bash
./visorplus enum 93.123.109.107
./visorplus enum 93.123.109.107:11434
```
Calls `/api/version`, `/api/tags`, and `/api/ps` on a single host. Flags:
- **Safety-stripped / abliterated models** — `huihui_ai/*-abliterated`
- **Offensive AI brands** — `hexstrike-ai`, etc.
- **Cloud-proxied models** — operator's paid quota exposed to anyone
- **RAG stacks** — embed + chat models coresident

---

### `assess` — Full passive recon on a single IP
```bash
./visorplus assess 93.123.109.107
./visorplus assess 93.123.109.107 -out ./evidence
```
Runs the complete passive assessment chain:
1. `whois` + reverse DNS
2. `nmap` top-1000 TCP (version + scripts)
3. `ssh-keyscan` — host key fingerprints
4. GreyNoise community classification
5. Shodan host detail (all ports + banners)
6. HackerTarget passive DNS history
7. Spamhaus DNSBL check
8. Ollama `/api/tags` + `/api/show` (system prompt extraction)
9. BARE exploit matching against Metasploit corpus

All artifacts saved to `<out>/<ip>/`.

---

### `corpus` — Adversarial LLM prompt corpus
```bash
./visorplus corpus                    # beginner (100 cases)
./visorplus corpus -tier intermediate  # 600 cases
./visorplus corpus -tier advanced      # 5000+ cases (forged)
```
Generates structured attack cases across 8 categories:
`prompt_injection`, `kb_exfiltration`, `jailbreak`, `system_prompt`,
`infra_discovery`, `tenant_cross_leak`, `kb_instructions`, `benign_control`

Each case includes `expect` blocks for automated pass/fail evaluation.

---

### `full` — End-to-end workflow
```bash
./visorplus full                            # default Ollama dork
./visorplus full 'http.title:"Open WebUI"' # custom dork
./visorplus full -org "Acme Corp"           # org-scoped
./visorplus full -skip audit,corpus         # skip specific phases
```

Runs all 5 phases in sequence:
1. VisorSD dry-run (beginner stack preview)
2. Shodan count + JAXEN harvest
3. `/api/tags` sweep across all found hosts, red-flag detection
4. Interactive: choose a target IP for full `assess`
5. Adversarial corpus generation

---

## Output Structure

```
visorplus-run/
  hunt/
    recon_dump.json     — full Shodan banners
    summary.csv         — compact host list
  assess/
    <ip>/
      whois.txt
      rdns.txt
      nmap_top1000.txt
      ssh_keys.txt
      greynoise.json
      shodan_host.json
      passive_dns.txt
      dnsbl.txt
  corpora/
    beginner.json
    intermediate.json   (if tier >= intermediate)
    advanced.json       (if tier = advanced)
```

---

## Red-Flag Model Patterns

VisorPlus automatically flags models during enumeration:

| Pattern | Signal |
|---------|--------|
| `*-abliterated` | Safety-stripped weights |
| `hexstrike-ai` | Offensive AI orchestrator brand |
| `*-uncensored` | Uncensored fine-tune |
| `*:cloud` | Operator's paid cloud quota exposed unauthenticated |
| embed + chat coresident | RAG stack (vector DB likely co-located) |

---

## Verified Shodan Dorks (Beginner Stack)

From the [AI-LLM-Infrastructure-OSINT](https://github.com/Nicholas-Kloster/AI-LLM-Infrastructure-OSINT) catalogue:

| Component | Dork | Hits (2026-04-30) |
|-----------|------|-------------------|
| Ollama | `http.html:"Ollama is running" -port:443` | 26,580 |
| Ollama (broad) | `product:Ollama` | 26,755 |
| Open WebUI | `http.html:"open-webui"` | 19,549 |
| n8n | `http.title:"n8n"` | 370 |
| ChromaDB | `"chromadb"` | 46 |

---

## Ecosystem

VisorPlus orchestrates the **discovery → enumeration → adversarial-test** subset of the NuClide chain. The tools below are auto-installed by `visorplus install`:

| Tool | Role | Auto-installed |
|------|------|---|
| [JAXEN](https://github.com/Nicholas-Kloster/JAXEN) | Shodan harvest + empire.db persistence | ✅ |
| [VisorSD](https://github.com/Nicholas-Kloster/VisorSD) | Severity-ranked AI/LLM stack audit | ✅ |
| [VisorCorpus](https://github.com/Nicholas-Kloster/VisorCorpus) | Adversarial LLM prompt corpus | ✅ |
| [BARE](https://github.com/Nicholas-Kloster/BARE) | Semantic exploit matching (Metasploit) | ✅ |
| [aimap](https://github.com/Nicholas-Kloster/aimap) | Active AI/ML service enumerator | ✅ |
| [AI-LLM-Infrastructure-OSINT](https://github.com/Nicholas-Kloster/AI-LLM-Infrastructure-OSINT) | Verified Shodan dork catalogue (reference) | — |

The wider NuClide chain extends VisorPlus's output into ledger / scoring / agentic-bench stages. Install these individually as needed:

| Tool | Role | Stage |
|------|------|------|
| [VisorGoose](https://github.com/Nicholas-Kloster/VisorGoose) | Government-TLD AI discovery (CT logs + Shodan + DNS) | Discovery (gov TLD) |
| [VisorGraph](https://github.com/Nicholas-Kloster/VisorGraph) | Seed-polymorphic recon engine; typed provenance graph | Discovery (graph) |
| [menlohunt](https://github.com/Nicholas-Kloster/menlohunt) | GCP External Attack Surface Management | Discovery (GCP/EASM) |
| [VisorLog](https://github.com/Nicholas-Kloster/VisorLog) | NuClide findings ledger (NDJSON → SQLite, lifecycle-tracked) | Ledger |
| [VisorScuba](https://github.com/Nicholas-Kloster/VisorScuba) | OPA/Rego compliance scoring against NuClide AI Security Baseline | Compliance |
| [VisorAgent](https://github.com/Nicholas-Kloster/VisorAgent) | Agentic LLM injection benchmark (delivers VisorCorpus payloads via real tool-use paths) | Adversarial bench |
| [VisorRAG](https://github.com/Nicholas-Kloster/VisorRAG) | Agentic recon CLI (RAG-grounded LLM drives 6 live tools, gVisor sandboxed) | Sibling orchestrator |
| [VisorHollow](https://github.com/Nicholas-Kloster/VisorHollow) | Process-injection detection benchmark (Windows x64) | Adversarial bench (Windows) |

---

## Use with Claude Code

VisorPlus orchestrates the entire NuClide AI/LLM hunt workflow — use Claude Code to drive it intelligently across a target.

```
I have VisorPlus installed at ~/Tools/VisorPlus/. Read README.md, then:
1. Run ./visorplus full against [target: IP/CIDR/org/ASN] and help me interpret the output
2. Flag any red-model patterns (abliterated, cloud-proxied, offensive AI brands) and prioritize them
3. After assessment, help me draft the disclosure write-up for the highest-severity finding
Target: [paste here]
```

```
I have a set of Ollama model names from ./visorplus enum [IP].
Read tools/ollama-model-injection.md in AI-LLM-Infrastructure-OSINT, then tell me:
- Which models are injectable via /api/create
- Whether any cloud proxy models are present (quota hijack risk)
- What the highest-impact injection scenario is for this specific target
Model list: [paste ./visorplus enum output here]
```

---

## License

MIT — see [LICENSE](LICENSE)

---

## About

Maintained by **[Nicholas Michael Kloster](https://github.com/Nicholas-Kloster)** as part of [**NuClide**](https://nuclide-research.com) — independent AI infrastructure security research.

CISA disclosures: [CVE-2025-4364](https://nvd.nist.gov/vuln/detail/CVE-2025-4364) · [ICSA-25-140-11](https://www.cisa.gov/news-events/ics-advisories/icsa-25-140-11)
