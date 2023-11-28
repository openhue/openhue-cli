package main

//go:generate oapi-codegen --package=gen -generate=client,types -o ./openhue/gen/openhue.gen.go https://api.redocly.com/registry/bundle/openhue/openhue/v2/openapi.yaml?branch=main

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
