package main

import (
	"fmt"
	"os"

	"github.com/Nicholas-Kloster/VisorPlus/cmd"
)

const banner = `
  VisorPlus · AI/LLM Hunt & Assessment
  github.com/Nicholas-Kloster/VisorPlus
`

func usage() {
	fmt.Println(banner)
	fmt.Println(`USAGE:
  visorplus <command> [flags]

COMMANDS:
  install             Install all Nuclide tools (JAXEN, VisorSD, VisorCorpus, BARE, aimap)
  hunt   <dork>       Hunt Shodan with a query, harvest into empire.db
  audit  [flags]      VisorSD severity-ranked scan (beginner AI/LLM stack)
  enum   <ip:port>    Enumerate Ollama /api/tags + model details on a single host
  assess <ip>         Full passive recon on a single IP (nmap, whois, passive intel)
  corpus [flags]      Generate adversarial LLM prompt corpus via VisorCorpus
  full   [dork]       End-to-end chain: audit → hunt → enum → flag → assess

FLAGS (full command):
  -key   string       Shodan API key (or set SHODAN_API_KEY env)
  -out   string       Output directory (default: ./visorplus-output)
  -org   string       Scope to organization name
  -asn   string       Scope to ASN (e.g. AS48090)
  -net   string       Scope to CIDR
  -limit int          Max hosts to process (default 40)

EXAMPLES:
  visorplus install
  visorplus audit -dry-run
  visorplus hunt 'http.html:"Ollama is running" -port:443'
  visorplus enum 93.123.109.107:11434
  visorplus assess 93.123.109.107
  visorplus full 'http.html:"Ollama is running" -port:443'
  visorplus full -org "Acme Corp"
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "install":
		cmd.Install(os.Args[2:])
	case "hunt":
		cmd.Hunt(os.Args[2:])
	case "audit":
		cmd.Audit(os.Args[2:])
	case "enum":
		cmd.Enum(os.Args[2:])
	case "assess":
		cmd.Assess(os.Args[2:])
	case "corpus":
		cmd.Corpus(os.Args[2:])
	case "full":
		cmd.Full(os.Args[2:])
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}
