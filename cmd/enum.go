package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func Enum(args []string) {
	fs := flag.NewFlagSet("enum", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fail("usage: visorplus enum <ip:port>")
		os.Exit(1)
	}

	host := fs.Arg(0)
	if !strings.Contains(host, ":") {
		host = host + ":11434"
	}

	header("Ollama Enumeration — " + host)
	client := &http.Client{Timeout: 10 * time.Second}

	// /api/version
	info("Fetching /api/version")
	if resp, err := client.Get("http://" + host + "/api/version"); err == nil {
		var v map[string]any
		json.NewDecoder(resp.Body).Decode(&v)
		resp.Body.Close()
		ok("Version: %v", v["version"])
	} else {
		fail("Unreachable: %v", err)
		os.Exit(1)
	}

	// /api/tags
	info("Fetching /api/tags")
	models, err := ollamaTags(host)
	if err != nil {
		fail("Failed: %v", err)
		os.Exit(1)
	}

	ok("Models loaded: %d | Total: %.1f GB", len(models), totalGB(models))
	fmt.Println()

	flags := []string{}
	for _, m := range models {
		sizeGB := float64(m.Size) / 1e9
		flag := redFlag(m.Name)
		flagStr := ""
		if flag != "" {
			flagStr = red + "  ← " + flag + reset
			flags = append(flags, m.Name+" ("+flag+")")
		}
		fmt.Printf("  %-55s  %.1f GB%s\n", m.Name, sizeGB, flagStr)
	}

	// RAG signal
	if ragSignal(models) {
		fmt.Printf("\n" + yellow + "[!] RAG signal: embed + chat models coresident" + reset + "\n")
	}

	// /api/ps
	fmt.Println()
	info("Fetching /api/ps (running models)")
	if resp, err := client.Get("http://" + host + "/api/ps"); err == nil {
		var ps struct {
			Models []struct {
				Name    string `json:"name"`
				Size    int64  `json:"size"`
				Details struct {
					ParameterSize string `json:"parameter_size"`
					QuantLevel    string `json:"quantization_level"`
				} `json:"details"`
				ExpiresAt string `json:"expires_at"`
			} `json:"models"`
		}
		json.NewDecoder(resp.Body).Decode(&ps)
		resp.Body.Close()
		if len(ps.Models) == 0 {
			info("No models currently loaded in memory")
		}
		for _, m := range ps.Models {
			ok("In memory: %s  %s %s  (%.1f GB)  expires %s",
				m.Name, m.Details.ParameterSize, m.Details.QuantLevel,
				float64(m.Size)/1e9, m.ExpiresAt[:10])
		}
	}

	// Summary
	if len(flags) > 0 {
		fmt.Println()
		warn("Red flags detected:")
		for _, f := range flags {
			fmt.Printf("  • %s\n", f)
		}
		fmt.Printf("\n  Run: visorplus assess %s\n\n", strings.Split(host, ":")[0])
	} else {
		fmt.Printf("\n  No red flags. Use `visorplus assess %s` for full passive recon.\n\n", strings.Split(host, ":")[0])
	}
}
