package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

// Bishop runs VisorBishop, the meta-fingerprinter for the AI observability tier.
// VisorBishop is a standalone Go tool; this subcommand is a thin wrapper that
// delegates to the installed `visorbishop` binary.
//
// Source: github.com/nuclide-research/VisorBishop
func Bishop(args []string) {
	fs := flag.NewFlagSet("bishop", flag.ExitOnError)
	input := fs.String("i", "", "Input file with one URL per line (or - for stdin)")
	target := fs.String("t", "", "Single target URL")
	ipShadow := fs.Bool("ip-shadow", false, "Run IP-direct-shadow port sweep on confirmed platform IPs")
	ipShadowAll := fs.Bool("ip-shadow-all", false, "Run IP-shadow on every target, even non-platform")
	conc := fs.Int("c", 16, "Concurrent probes")
	timeout := fs.String("timeout", "8s", "Per-probe timeout")
	jsonOut := fs.String("json", "", "JSON output file")
	csvOut := fs.String("csv", "", "CSV output file")
	quiet := fs.Bool("q", false, "Quiet mode")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `visorplus bishop — VisorBishop meta-fingerprinter for AI observability platforms

USAGE:
  visorplus bishop -t <url>                       Single-target probe
  visorplus bishop -i <file>                      Batch probe
  visorplus bishop -i <file> -ip-shadow           Add IP-direct-shadow sweep

FLAGS:`)
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
EXAMPLES:
  visorplus bishop -t http://190.210.105.193:6006
  visorplus bishop -i unauth-hosts.txt -ip-shadow -json out.json
  visorplus bishop -i hosts.txt -c 32 -timeout 6s

Detects:
  Phoenix (Arize AI), Langfuse, Helicone, LangSmith, OpenLIT, Lunary, Pezzo

IP-direct-shadow (-ip-shadow) probes 15 ports per host for co-located
unauth services: NFS, MailHog, MailCatcher, Postgres, ClickHouse, Redis,
Kibana, Prometheus, AlertManager, node_exporter, Elasticsearch, etc.

Read-only. No credential testing, no payload fuzzing.

Source: https://github.com/nuclide-research/VisorBishop`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *input == "" && *target == "" {
		fs.Usage()
		os.Exit(2)
	}

	// Locate visorbishop binary
	bin, err := exec.LookPath("visorbishop")
	if err != nil {
		fmt.Fprintln(os.Stderr, "visorplus bishop: visorbishop binary not found in PATH")
		fmt.Fprintln(os.Stderr, "  install with:")
		fmt.Fprintln(os.Stderr, "    go install github.com/nuclide-research/VisorBishop/cmd/visorbishop@latest")
		os.Exit(1)
	}

	// Build forwarded argv
	fwd := []string{}
	if *input != "" {
		fwd = append(fwd, "-i", *input)
	}
	if *target != "" {
		fwd = append(fwd, "-t", *target)
	}
	if *ipShadow {
		fwd = append(fwd, "-ip-shadow")
	}
	if *ipShadowAll {
		fwd = append(fwd, "-ip-shadow-all")
	}
	fwd = append(fwd, "-c", fmt.Sprintf("%d", *conc))
	fwd = append(fwd, "-timeout", *timeout)
	if *jsonOut != "" {
		fwd = append(fwd, "-json", *jsonOut)
	}
	if *csvOut != "" {
		fwd = append(fwd, "-csv", *csvOut)
	}
	if *quiet {
		fwd = append(fwd, "-q")
	}

	c := exec.Command(bin, fwd...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "visorplus bishop:", err)
		os.Exit(1)
	}
}
