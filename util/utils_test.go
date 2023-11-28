package util

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test"
	"strings"
	"testing"
)

type Light struct {
	id   string
	name string
}

func PrintLightLineProcessor(light Light) string {
	return light.id + "\t" + light.name
}

func TestPrintTable(t *testing.T) {

	lights := []Light{{id: "1234", name: "Light 1"}, {id: "4321", name: "Light 2"}}

	streams, _, out, _ := openhue.NewTestIOStreams()

	PrintTable(streams, lights, PrintLightLineProcessor, "id", "name")

	lines := strings.Split(out.String(), "\n")

	test.AssertThatLineEqualsTo(t, lines, 0, "id     name   ")
	test.AssertThatLineEqualsTo(t, lines, 1, "----   ----   ")
	test.AssertThatLineEqualsTo(t, lines, 2, "1234   Light 1")
	test.AssertThatLineEqualsTo(t, lines, 3, "4321   Light 2")
}
