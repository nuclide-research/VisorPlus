package cmd

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Full runs the complete session workflow:
//
//	VisorSD audit → Shodan count → JAXEN hunt → /api/tags sweep →
//	flag red hosts → user picks target → assess → corpus
func Full(args []string) {
	fs := flag.NewFlagSet("full", flag.ExitOnError)
	dork   := fs.String("dork", `http.html:"Ollama is running" -port:443`, "Shodan dork to hunt")
	org    := fs.String("org", "", "scope to organization (VisorSD + Shodan filter)")
	asn    := fs.String("asn", "", "scope to ASN")
	netStr := fs.String("net", "", "scope to CIDR")
	outDir := fs.String("out", "./visorplus-run", "output directory")
	skip   := fs.String("skip", "", "comma-separated phases to skip: audit,hunt,enum,assess,corpus")
	fs.Parse(args)

	if fs.NArg() > 0 {
		*dork = fs.Arg(0)
	}

	skipped := map[string]bool{}
	for _, s := range strings.Split(*skip, ",") {
		skipped[strings.TrimSpace(s)] = true
	}

	os.MkdirAll(*outDir, 0755)
	requireKey()

	fmt.Printf("\n" + bold + cyan +
		"════════════════════════════════════════\n" +
		"  VisorPlus — Full Hunt & Assessment\n" +
		"════════════════════════════════════════\n" + reset + "\n")

	// ── Phase 1: VisorSD audit ────────────────────────────────────────────
	if !skipped["audit"] {
		header("Phase 1/5 — VisorSD Beginner Stack Audit")
		auditArgs := []string{"-dry-run"}
		if *org != ""    { auditArgs = append(auditArgs, "-org", *org) }
		if *asn != ""    { auditArgs = append(auditArgs, "-asn", *asn) }
		if *netStr != "" { auditArgs = append(auditArgs, "-net", *netStr) }
		Audit(auditArgs)
	}

	// ── Phase 2: Shodan count + JAXEN hunt ───────────────────────────────
	if !skipped["hunt"] {
		header("Phase 2/5 — Shodan Hunt")
		count, err := shodanCount(shodanKey(), *dork)
		if err != nil {
			warn("Count failed: %v", err)
		} else {
			info("Dork: %s", *dork)
			ok("Total available: %d", count)
		}
		huntDir := *outDir + "/hunt"
		os.MkdirAll(huntDir, 0755)
		Hunt([]string{"-out", huntDir, *dork})
	}

	// ── Phase 3: /api/tags sweep ──────────────────────────────────────────
	if !skipped["enum"] {
		header("Phase 3/5 — Ollama Model Enumeration")
		huntDump := *outDir + "/hunt/recon_dump.json"
		hosts := extractHosts(huntDump)

		if len(hosts) == 0 {
			warn("No hosts in %s — skipping enum", huntDump)
		} else {
			info("Enumerating %d hosts ...", len(hosts))
			var redHosts []string

			for _, h := range hosts {
				models, err := ollamaTags(h + ":11434")
				if err != nil {
					continue
				}
				var flags []string
				for _, m := range models {
					if f := redFlag(m.Name); f != "" {
						flags = append(flags, m.Name+" ("+f+")")
					}
				}
				rag := ragSignal(models)
				gb := totalGB(models)

				indicator := ""
				if len(flags) > 0 {
					indicator = red + "  RED FLAG" + reset
					redHosts = append(redHosts, h)
				} else if rag {
					indicator = yellow + "  RAG stack" + reset
				}
				fmt.Printf("  %-20s  %d models  %.1f GB%s\n", h, len(models), gb, indicator)
				for _, f := range flags {
					fmt.Printf("    → %s\n", f)
				}
			}

			if len(redHosts) > 0 {
				fmt.Printf("\n" + red + "[!] Red-flag hosts:\n" + reset)
				for _, h := range redHosts {
					fmt.Printf("    %s\n", h)
				}
			}
		}
	}

	// ── Phase 4: Assess ───────────────────────────────────────────────────
	if !skipped["assess"] {
		header("Phase 4/5 — Target Assessment")
		fmt.Printf("\n  Enter target IP to assess (or press Enter to skip): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		target := strings.TrimSpace(scanner.Text())
		if target != "" {
			assessDir := *outDir + "/assess"
			Assess([]string{"-out", assessDir, target})
		} else {
			info("Skipped — run `visorplus assess <ip>` manually")
		}
	}

	// ── Phase 5: Corpus ───────────────────────────────────────────────────
	if !skipped["corpus"] {
		header("Phase 5/5 — Adversarial Corpus")
		Corpus([]string{"-tier", "beginner", "-out", *outDir + "/corpora"})
	}

	header("Run Complete")
	ok("All output in %s/", *outDir)
}

func extractHosts(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []map[string]any
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil
	}
	seen := map[string]bool{}
	var hosts []string
	for _, e := range entries {
		if ip, ok := e["ip_str"].(string); ok && !seen[ip] {
			seen[ip] = true
			hosts = append(hosts, ip)
		}
	}
	return hosts
}
