package util

import (
	"openhue-cli/openhue"
	"openhue-cli/openhue/test/assert"
	"strings"
	"testing"
)

type Light struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func PrintLightLineProcessor(light Light) string {
	return light.Id + "\t" + light.Name
}

func TestPrintTable(t *testing.T) {

	lights := []Light{{Id: "1234", Name: "Light 1"}, {Id: "4321", Name: "Light 2"}}

	streams, _, out, _ := openhue.NewTestIOStreams()

	PrintTable(streams, lights, PrintLightLineProcessor, "Id", "Name")

	lines := strings.Split(out.String(), "\n")

	assert.ThatLineEqualsTo(t, lines, 0, "Id     Name   ")
	assert.ThatLineEqualsTo(t, lines, 1, "----   ----   ")
	assert.ThatLineEqualsTo(t, lines, 2, "1234   Light 1")
	assert.ThatLineEqualsTo(t, lines, 3, "4321   Light 2")
}

func TestPrintJsonArrayWithOneElement(t *testing.T) {

	expected := `{
  "id": "1234",
  "name": "Light 1"
}
`

	lights := []Light{
		{Id: "1234", Name: "Light 1"},
	}

	streams, _, out, _ := openhue.NewTestIOStreams()

	PrintJsonArray(streams, lights)

	if out.String() != expected {
		t.Fatalf("Output is: \n%s\nExpected:\n%s", out.String(), expected)
	}
}

func TestPrintJsonArrayWithTwoElements(t *testing.T) {

	expected := `[
  {
    "id": "1234",
    "name": "Light 1"
  },
  {
    "id": "4321",
    "name": "Light 2"
  }
]
`

	lights := []Light{
		{Id: "1234", Name: "Light 1"},
		{Id: "4321", Name: "Light 2"},
	}

	streams, _, out, _ := openhue.NewTestIOStreams()

	PrintJsonArray(streams, lights)

	if out.String() != expected {
		t.Fatalf("Output is: \n%s\nExpected:\n%s", out.String(), expected)
	}
}
