package main

//go:generate oapi-codegen --package=openhue -generate=client,types -o ./openhue/openhue.gen.go /Users/thibault.pensec/local/perso/openhue-api/build/openhue.yaml

import (
	"openhue-cli/cmd"
	"openhue-cli/util"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(util.NewBuildInfo(version, commit, date))
}
