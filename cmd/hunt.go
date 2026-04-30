package cmd

import (
	"flag"
	"fmt"
	"os"
)

func Hunt(args []string) {
	fs := flag.NewFlagSet("hunt", flag.ExitOnError)
	out := fs.String("out", ".", "output directory")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fail("usage: visorplus hunt <shodan-dork>")
		os.Exit(1)
	}

	dork := fs.Arg(0)
	key := requireKey()
	jaxen := requireTool("jaxen")

	header("JAXEN Hunt")
	info("Dork: %s", dork)

	count, err := shodanCount(key, dork)
	if err != nil {
		fail("Shodan count: %v", err)
	} else {
		info("Total available: %d", count)
	}

	os.MkdirAll(*out, 0755)

	if err := run(jaxen, "hunt", "--clean", "--export", dork); err != nil {
		fail("JAXEN hunt failed: %v", err)
		os.Exit(1)
	}

	for _, f := range []string{"recon_dump.json", "summary.csv"} {
		if _, err := os.Stat(f); err == nil {
			dest := *out + "/" + f
			os.Rename(f, dest)
			ok("Saved → %s", dest)
		}
	}

	fmt.Printf("\n  empire.db updated. Run `visorplus enum` on any host to enumerate models.\n\n")
}
