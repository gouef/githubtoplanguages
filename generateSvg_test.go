package main

import (
	"flag"
	"os"
	"testing"
)

func TestGenerateSvg(t *testing.T) {

	t.Run("Generate main with Ignore", func(t *testing.T) {
		resetFlags()
		token := os.Getenv("GITHUB_TOKEN")
		os.Args = []string{"cmd", "-user=JanGalek", "-limit=5", "-ignore-orgs=wowmua", "-ignore-repos=wowmua/Maps", "-gh-token=" + token}
		main()
	})

	t.Run("Generate main", func(t *testing.T) {
		resetFlags()
		token := os.Getenv("GITHUB_TOKEN")
		os.Args = []string{"cmd", "-user=JanGalek", "-limit=6", "-ignore-orgs=wowmua", "-ignore-repos=wowmua/Maps", "-gh-token=" + token}
		main()
	})

	t.Run("Generate main with Forks", func(t *testing.T) {
		resetFlags()
		token := os.Getenv("GITHUB_TOKEN")
		os.Args = []string{"cmd", "-user=JanGalek", "-limit=6", "-ignore-orgs=wowmua", "-ignore-repos=wowmua/Maps", "-gh-token=" + token, "-with-forks=true"}
		main()
	})

	t.Run("Generate main with ignore langs", func(t *testing.T) {
		resetFlags()
		token := os.Getenv("GITHUB_TOKEN")
		os.Args = []string{"cmd", "-user=JanGalek", "-limit=6", "-ignore-orgs=wowmua", "-ignore-repos=wowmua/Maps", "-gh-token=" + token, "-ignore-langs=Go,Python"}
		main()
	})
}
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
