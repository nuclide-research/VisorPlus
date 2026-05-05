# VisorPlus

Unified CLI that orchestrates the NuClide AI/LLM security assessment chain. Single entry point that calls JAXEN (Shodan harvest) → VisorSD (severity-ranked dorks) → aimap (active fingerprinting) → BARE (exploit ranking) → VisorCorpus (adversarial corpus generation).

## Language
Go

## Build & Run
```
go build -o visorplus .
./visorplus install      # auto-clones + builds JAXEN/VisorSD/VisorCorpus/BARE/aimap into ~/Tools/
./visorplus full         # end-to-end Shodan-driven assessment
go test ./...
```

## Claude Code Notes
- Check README for full CLI surface (audit / hunt / enum / assess / corpus / full / install)
- The 7 subcommands map to chain stages — read the cmd/ files (`assess.go`, `audit.go`, `corpus.go`, `enum.go`, `full.go`, `hunt.go`, `install.go`) for stage-specific logic
- Output files land in `visorplus-run/` by default — `hunt/`, `assess/<ip>/`, `corpora/` subdirs are stable contracts that downstream tools (VisorLog) read
- Requires Shodan API key in `SHODAN_API_KEY` env var
- Built with [Claude Code](https://claude.ai/code)
