package main

//go:generate oapi-codegen --package=openhue -generate=client,types -o ./openhue/openhue.gen.go https://github.com/openhue/openhue-api/releases/download/0.6/openhue.yaml

import "openhue-cli/cmd"

func main() {
	cmd.Execute()
}
