package util

import "testing"

type Light struct {
	id   string
	name string
}

func PrintLightLineProcessor(light Light) string {
	return light.id + "\t" + light.name
}

func TestPrintTable(t *testing.T) {

	lights := []Light{{id: "1234", name: "Light 1"}, {id: "4321", name: "Light 2"}}

	PrintTable(lights, PrintLightLineProcessor, "id", "name")
}
