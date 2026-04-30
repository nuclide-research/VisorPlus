package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Assess(args []string) {
	fs := flag.NewFlagSet("assess", flag.ExitOnError)
	out := fs.String("out", ".", "output directory")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fail("usage: visorplus assess <ip>")
		os.Exit(1)
	}

	ip := fs.Arg(0)
	outDir := *out + "/" + strings.ReplaceAll(ip, ".", "_")
	os.MkdirAll(outDir, 0755)

	header("Full Assessment — " + ip)

	// ── 1. WHOIS ──────────────────────────────────────────────────────────
	header("1/6 Network Identity")
	runToFile("whois", outDir+"/whois.txt", ip)
	runToFile("dig", outDir+"/rdns.txt", "-x", ip)
	ok("whois + rDNS saved")

	// ── 2. NMAP ───────────────────────────────────────────────────────────
	header("2/6 Port Scan (nmap)")
	info("TCP top-1000")
	runToFile("nmap", outDir+"/nmap_top1000.txt",
		"-Pn", "-sV", "--top-ports", "1000", "--min-rate", "2000", ip)

	// ── 3. SSH KEYSCAN ────────────────────────────────────────────────────
	header("3/6 SSH Fingerprint")
	runToFile("ssh-keyscan", outDir+"/ssh_keys.txt",
		"-t", "rsa,ecdsa,ed25519", ip)
	ok("Host keys saved → %s/ssh_keys.txt", outDir)

	// ── 4. PASSIVE INTEL ──────────────────────────────────────────────────
	header("4/6 Passive Intel")

	client := &http.Client{Timeout: 8 * time.Second}

	// GreyNoise
	info("GreyNoise")
	gnURL := "https://api.greynoise.io/v3/community/" + ip
	if resp, err := client.Get(gnURL); err == nil {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		os.WriteFile(outDir+"/greynoise.json", b, 0644)
		var gn map[string]any
		json.Unmarshal(b, &gn)
		if cls, ok2 := gn["classification"]; ok2 {
			ok("GreyNoise: %v (noise=%v, riot=%v)", cls, gn["noise"], gn["riot"])
		} else {
			ok("GreyNoise: no data")
		}
	}

	// Shodan host detail
	info("Shodan host detail")
	key := shodanKey()
	if key != "" {
		if resp, err := client.Get(fmt.Sprintf("https://api.shodan.io/shodan/host/%s?key=%s", ip, key)); err == nil {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			os.WriteFile(outDir+"/shodan_host.json", b, 0644)
			var sh map[string]any
			json.Unmarshal(b, &sh)
			ok("Shodan: org=%v, country=%v, ports=%v", sh["org"], sh["country_name"], sh["ports"])
		}
	}

	// Passive DNS
	info("Passive DNS (HackerTarget)")
	if resp, err := client.Get("https://api.hackertarget.com/reverseiplookup/?q=" + ip); err == nil {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		os.WriteFile(outDir+"/passive_dns.txt", b, 0644)
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			ok("Passive DNS: %d hostname(s) found", len(lines))
			for _, l := range lines {
				fmt.Printf("    %s\n", l)
			}
		} else {
			ok("Passive DNS: no hostnames")
		}
	}

	// DNSBL
	info("Spamhaus DNSBL")
	parts := strings.Split(ip, ".")
	reversed := parts[3] + "." + parts[2] + "." + parts[1] + "." + parts[0]
	runToFile("dig", outDir+"/dnsbl.txt", "+short", reversed+".zen.spamhaus.org")

	// ── 5. OLLAMA ENUM ────────────────────────────────────────────────────
	header("5/6 Ollama Enumeration")
	Enum([]string{ip + ":11434"})

	// ── 6. BARE ───────────────────────────────────────────────────────────
	header("6/6 Exploit Match (BARE)")
	bare := toolPath("bare")
	if bare == "" {
		warn("BARE not found — skipping exploit match (run `visorplus install`)")
	} else {
		payload := `{"findings":[{"product":"Ollama","version":"unknown","port":11434},{"product":"OpenSSH","version":"unknown","port":22}]}`
		c := exec.Command(bare, "ingest", "--json")
		c.Stdin = strings.NewReader(payload)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Run()
	}

	// ── Report ────────────────────────────────────────────────────────────
	header("Assessment Complete")
	ok("All artifacts saved to %s/", outDir)
	fmt.Printf("\n  Files: whois.txt, rdns.txt, nmap_top1000.txt, ssh_keys.txt,\n")
	fmt.Printf("         greynoise.json, shodan_host.json, passive_dns.txt, dnsbl.txt\n\n")
}

func runToFile(name, dest string, args ...string) {
	c := exec.Command(name, args...)
	c.Env = os.Environ()
	f, err := os.Create(dest)
	if err != nil {
		fail("create %s: %v", dest, err)
		return
	}
	defer f.Close()
	c.Stdout = f
	c.Stderr = f
	if err := c.Run(); err != nil {
		warn("%s: %v", name, err)
	}
}
