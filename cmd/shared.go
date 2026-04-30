package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ── colour helpers ──────────────────────────────────────────────────────────

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

func info(f string, a ...any)  { fmt.Printf(cyan+"[*]"+reset+" "+f+"\n", a...) }
func ok(f string, a ...any)    { fmt.Printf(green+"[+]"+reset+" "+f+"\n", a...) }
func warn(f string, a ...any)  { fmt.Printf(yellow+"[!]"+reset+" "+f+"\n", a...) }
func fail(f string, a ...any)  { fmt.Printf(red+"[-]"+reset+" "+f+"\n", a...) }
func header(s string)          { fmt.Printf("\n"+bold+cyan+"═══ %s ═══"+reset+"\n", s) }

// ── env / key helpers ───────────────────────────────────────────────────────

func shodanKey() string {
	if k := os.Getenv("SHODAN_API_KEY"); k != "" {
		return k
	}
	home, _ := os.UserHomeDir()
	b, err := os.ReadFile(filepath.Join(home, ".config", "nuclide", "shodan.key"))
	if err == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}

func requireKey() string {
	k := shodanKey()
	if k == "" {
		fail("Shodan API key not found. Set SHODAN_API_KEY or store at ~/.config/nuclide/shodan.key")
		os.Exit(1)
	}
	return k
}

func toolPath(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "Tools", name, name),
		filepath.Join(home, "Tools", name),
		filepath.Join(home, "go", "bin", name),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func requireTool(name string) string {
	p := toolPath(name)
	if p == "" {
		fail("tool not found: %s — run `visorplus install` first", name)
		os.Exit(1)
	}
	return p
}

func run(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), "SHODAN_API_KEY="+shodanKey())
	return c.Run()
}

func runCapture(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	c.Env = append(os.Environ(), "SHODAN_API_KEY="+shodanKey())
	b, err := c.CombinedOutput()
	return string(b), err
}

// ── Shodan helpers ──────────────────────────────────────────────────────────

func shodanCount(key, query string) (int, error) {
	q := url.QueryEscape(query)
	resp, err := http.Get(fmt.Sprintf("https://api.shodan.io/shodan/host/count?key=%s&query=%s", key, q))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r struct {
		Total int    `json:"total"`
		Error string `json:"error"`
	}
	json.Unmarshal(b, &r)
	if r.Error != "" {
		return 0, fmt.Errorf(r.Error)
	}
	return r.Total, nil
}

// ── Ollama helpers ──────────────────────────────────────────────────────────

type OllamaModel struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

type OllamaTagsResp struct {
	Models []OllamaModel `json:"models"`
}

func ollamaTags(host string) ([]OllamaModel, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("http://" + host + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r OllamaTagsResp
	json.NewDecoder(resp.Body).Decode(&r)
	return r.Models, nil
}

// redFlag returns a reason string if the model name is suspicious, else "".
func redFlag(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "abliterat"):
		return "safety-stripped"
	case strings.Contains(lower, "hexstrike"):
		return "offensive-AI brand"
	case strings.Contains(lower, "uncensored"):
		return "uncensored"
	case strings.Contains(lower, "jailbreak"):
		return "jailbreak"
	case strings.Contains(lower, "roleplay") || strings.Contains(lower, "rp-"):
		return "roleplay/uncensored"
	case strings.Contains(lower, ":cloud"):
		return "cloud-proxied (paid quota exposed)"
	}
	return ""
}

func ragSignal(models []OllamaModel) bool {
	hasEmbed, hasChat := false, false
	for _, m := range models {
		l := strings.ToLower(m.Name)
		if strings.Contains(l, "embed") || strings.Contains(l, "nomic") || strings.Contains(l, "mxbai") {
			hasEmbed = true
		} else {
			hasChat = true
		}
	}
	return hasEmbed && hasChat
}

func totalGB(models []OllamaModel) float64 {
	var t int64
	for _, m := range models {
		t += m.Size
	}
	return float64(t) / 1e9
}
