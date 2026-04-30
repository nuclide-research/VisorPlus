package cmd

import (
	"flag"
	"fmt"
	"os"
)

func Audit(args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	org    := fs.String("org", "", "filter by organization name")
	asn    := fs.String("asn", "", "filter by ASN (e.g. AS48090)")
	net    := fs.String("net", "", "filter by CIDR")
	limit  := fs.Int("limit", 10, "max results per query")
	dryRun := fs.Bool("dry-run", false, "print queries without calling Shodan")
	format := fs.String("format", "text", "output format: text|json|csv")
	outFile := fs.String("out", "", "write results to file")
	failOn := fs.String("fail-on", "", "exit non-zero if severity >= threshold (critical|high|medium|low)")
	fs.Parse(args)

	requireKey()
	visorsd := requireTool("visorsd")

	header("VisorSD — Beginner AI/LLM Stack Audit")

	var runArgs []string
	if *org != ""     { runArgs = append(runArgs, "-org", *org) }
	if *asn != ""     { runArgs = append(runArgs, "-asn", *asn) }
	if *net != ""     { runArgs = append(runArgs, "-net", *net) }
	if *limit != 10   { runArgs = append(runArgs, "-limit", itoa(*limit)) }
	if *dryRun        { runArgs = append(runArgs, "-dry-run") }
	if *format != ""  { runArgs = append(runArgs, "-format", *format) }
	if *outFile != "" { runArgs = append(runArgs, "-out", *outFile) }
	if *failOn != ""  { runArgs = append(runArgs, "-fail-on", *failOn) }

	if err := run(visorsd, runArgs...); err != nil {
		if *failOn != "" {
			os.Exit(1)
		}
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
