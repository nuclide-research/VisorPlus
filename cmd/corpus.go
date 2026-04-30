package cmd

import (
	"flag"
	"fmt"
	"os"
)

func Corpus(args []string) {
	fs := flag.NewFlagSet("corpus", flag.ExitOnError)
	tier    := fs.String("tier", "beginner", "stack tier: beginner|intermediate|advanced")
	outDir  := fs.String("out", "./visorplus-corpora", "output directory")
	fs.Parse(args)

	vc := requireTool("visorcorpus")
	os.MkdirAll(*outDir, 0755)

	type build struct {
		profile string
		kind    string
		max     string
		label   string
	}

	tiers := map[string][]build{
		"beginner": {
			{"standard", "hybrid", "100", "beginner"},
		},
		"intermediate": {
			{"standard", "hybrid", "100", "beginner"},
			{"standard", "randomized", "500", "intermediate"},
		},
		"advanced": {
			{"standard", "hybrid", "100", "beginner"},
			{"standard", "randomized", "500", "intermediate"},
			{"strict", "stress", "5000", "advanced"},
		},
	}

	builds, ok2 := tiers[*tier]
	if !ok2 {
		fail("unknown tier: %s (use beginner|intermediate|advanced)", *tier)
		os.Exit(1)
	}

	header("VisorCorpus — " + *tier + " stack")

	for _, b := range builds {
		outFile := *outDir + "/" + b.label + ".json"
		info("Building %s corpus (profile=%s, type=%s, max=%s)", b.label, b.profile, b.kind, b.max)
		if err := run(vc, "build",
			"-profile", b.profile,
			"-type", b.kind,
			"-max", b.max,
			"-out", outFile,
		); err != nil {
			fail("build failed: %v", err)
			continue
		}
		ok("Saved → %s", outFile)

		info("Stats:")
		run(vc, "stats", "-in", outFile)
		fmt.Println()
	}
}
