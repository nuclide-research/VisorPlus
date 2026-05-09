package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type tool struct {
	name   string
	repo   string
	build  []string // relative to cloned dir
	binary string
}

var tools = []tool{
	{
		name:   "JAXEN",
		repo:   "https://github.com/Nicholas-Kloster/JAXEN.git",
		build:  []string{"go", "install", "."},
		binary: "jaxen",
	},
	{
		name:   "VisorSD",
		repo:   "https://github.com/Nicholas-Kloster/VisorSD.git",
		build:  []string{"go", "install", "./cmd/shodan-audit"},
		binary: "visorsd",
	},
	{
		name:   "VisorCorpus",
		repo:   "https://github.com/Nicholas-Kloster/VisorCorpus.git",
		build:  []string{"go", "install", "./cmd/visorcorpus"},
		binary: "visorcorpus",
	},
	{
		name:   "BARE",
		repo:   "https://github.com/Nicholas-Kloster/BARE.git",
		build:  []string{"cargo", "install", "--path", "."},
		binary: "bare",
	},
	{
		name:   "aimap",
		repo:   "https://github.com/Nicholas-Kloster/aimap.git",
		build:  []string{"go", "install", "."},
		binary: "aimap",
	},
}

func Install(_ []string) {
	header("VisorPlus — Install All Tools")

	home, _ := os.UserHomeDir()
	base := filepath.Join(home, "Tools")
	os.MkdirAll(base, 0755)

	// Ensure Shodan key dir exists
	keyDir := filepath.Join(home, ".config", "nuclide")
	os.MkdirAll(keyDir, 0700)
	keyFile := filepath.Join(keyDir, "shodan.key")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		warn("Shodan key not found at %s", keyFile)
		fmt.Print("  Enter your Shodan API key (or press Enter to skip): ")
		var key string
		fmt.Scanln(&key)
		if key != "" {
			os.WriteFile(keyFile, []byte(key), 0600)
			ok("Key saved to %s", keyFile)
		}
	} else {
		ok("Shodan key already configured")
	}

	for _, t := range tools {
		header(t.name)
		dest := filepath.Join(base, t.name)

		if _, err := os.Stat(dest); os.IsNotExist(err) {
			info("Cloning %s → %s", t.repo, dest)
			c := exec.Command("git", "clone", t.repo, dest)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				fail("clone failed: %v", err)
				continue
			}
		} else {
			info("Already cloned — pulling latest")
			c := exec.Command("git", "-C", dest, "pull", "--ff-only")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
		}

		info("Building %s", t.name)
		c := exec.Command(t.build[0], t.build[1:]...)
		c.Dir = dest
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fail("build failed: %v", err)
			continue
		}

		gobin := os.Getenv("GOPATH")
		if gobin == "" {
			gobin = filepath.Join(os.Getenv("HOME"), "go")
		}
		binPath := filepath.Join(gobin, "bin", t.binary)
		if _, err := os.Stat(binPath); err == nil {
			ok("%s installed → %s", t.name, binPath)
		} else {
			warn("binary not found at %s — check $GOPATH", binPath)
		}
	}

	header("Done")
	ok("All tools installed under %s", base)
	fmt.Printf("\n  Add to PATH: export PATH=\"$PATH:%s\"\n\n", base)
}
