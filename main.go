package main

//go:generate make generate

import (
	_ "embed"
	"openhue-cli/cmd"
	"openhue-cli/openhue"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// skill is the root SKILL.md, bundled so that `openhue install-skill` can write it on disk
//
//go:embed SKILL.md
var skill []byte

func main() {
	cmd.Execute(openhue.NewBuildInfo(version, commit, date), skill)
}
