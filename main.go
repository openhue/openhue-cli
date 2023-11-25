package main

//go:generate oapi-codegen --package=openhue -generate=client,types -o ./openhue/openhue.gen.go https://api.redocly.com/registry/bundle/openhue/openhue/v2/openapi.yaml?branch=main

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
