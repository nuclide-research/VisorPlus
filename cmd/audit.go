package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type stackTier struct {
	Name        string
	Description string
	Components  string
	Stacks      []string // VisorSD stack names
}

var tiers = []stackTier{
	{
		Name:        "beginner",
		Description: "Local dev stack — no auth, everything default",
		Components:  "Ollama + Open WebUI + ChromaDB + n8n + Cloudflared",
		Stacks:      []string{"beginner"},
	},
	{
		Name:        "intermediate",
		Description: "LangChain-era RAG stack — FastAPI glue, production vector DB",
		Components:  "LangChain/LangGraph + Qdrant/Weaviate + FastAPI + Langfuse",
		Stacks:      []string{"inference", "vector-db", "rag"},
	},
	{
		Name:        "advanced",
		Description: "Inference cluster — GPU serving, MLflow experiment tracking, custom RAG API",
		Components:  "vLLM/TGI + Kubernetes vector DB + MLflow + custom RAG API",
		Stacks:      []string{"inference", "vector-db", "rag", "data"},
	},
	{
		Name:        "enterprise",
		Description: "Full production AI platform — multi-tenant, audited, orchestrated",
		Components:  "OpenSearch + Airflow + Prometheus/Grafana + multi-tenant auth",
		Stacks:      []string{"inference", "vector-db", "data", "observability", "rag"},
	},
}

func listTiers() {
	fmt.Printf("\n%s%sAvailable stack tiers:%s\n\n", bold, cyan, reset)
	for _, t := range tiers {
		fmt.Printf("  %s%-14s%s %s\n", yellow, t.Name, reset, t.Components)
		fmt.Printf("  %s              %s%s\n\n", "", t.Description, reset)
	}
	fmt.Printf("Usage: visorplus audit -tier <name> [-dry-run]\n\n")
}

func Audit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	tier    := fs.String("tier", "", "stack tier: beginner|intermediate|advanced|enterprise (omit to list)")
	org     := fs.String("org", "", "filter by organization name")
	asn     := fs.String("asn", "", "filter by ASN (e.g. AS48090)")
	net     := fs.String("net", "", "filter by CIDR")
	limit   := fs.Int("limit", 10, "max results per query")
	dryRun  := fs.Bool("dry-run", false, "print queries without calling Shodan")
	format  := fs.String("format", "text", "output format: text|json|csv")
	outFile := fs.String("out", "", "write results to file")
	failOn  := fs.String("fail-on", "", "exit non-zero if severity >= threshold (critical|high|medium|low)")
	fs.Parse(args)

	if *tier == "" {
		listTiers()
		return
	}

	var selected *stackTier
	for i := range tiers {
		if tiers[i].Name == strings.ToLower(*tier) {
			selected = &tiers[i]
			break
		}
	}
	if selected == nil {
		fail("unknown tier %q — run `visorplus audit` to list", *tier)
		os.Exit(1)
	}

	requireKey()
	visorsd := requireTool("visorsd")

	fmt.Printf("\n%s%s▶ %s stack%s\n", bold, cyan, selected.Name, reset)
	fmt.Printf("  %s%s%s\n", yellow, selected.Components, reset)
	fmt.Printf("  %s\n\n", selected.Description)

	baseArgs := []string{}
	if *org != ""     { baseArgs = append(baseArgs, "-org", *org) }
	if *asn != ""     { baseArgs = append(baseArgs, "-asn", *asn) }
	if *net != ""     { baseArgs = append(baseArgs, "-net", *net) }
	if *limit != 10   { baseArgs = append(baseArgs, "-limit", itoa(*limit)) }
	if *dryRun        { baseArgs = append(baseArgs, "-dry-run") }
	if *format != ""  { baseArgs = append(baseArgs, "-format", *format) }
	if *outFile != "" { baseArgs = append(baseArgs, "-out", *outFile) }
	if *failOn != ""  { baseArgs = append(baseArgs, "-fail-on", *failOn) }

	for _, s := range selected.Stacks {
		header(s)
		runArgs := append([]string{"-stack", s}, baseArgs...)
		if err := run(visorsd, runArgs...); err != nil {
			if *failOn != "" {
				os.Exit(1)
			}
		}
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
