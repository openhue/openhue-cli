package main

//go:generate make generate

import (
	"openhue-cli/cmd"
	"openhue-cli/openhue"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(openhue.NewBuildInfo(version, commit, date))
}
