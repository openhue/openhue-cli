package main

//go:generate oapi-codegen --package=openhue -generate=client,types -o ./openhue/openhue.gen.go https://api.redocly.com/registry/bundle/openhue/openhue/v2/openapi.yaml?branch=main

import "openhue-cli/cmd"

func main() {
	cmd.Execute()
}
