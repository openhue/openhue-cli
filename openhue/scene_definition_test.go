package openhue_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
)

func TestParseLightArg(t *testing.T) {
	e, err := openhue.ParseLightArg("Desk:on,brightness=100,temperature=250,transition=3s")
	require.NoError(t, err)
	assert.Equal(t, "Desk", e.Light)
	assert.True(t, *e.On)
	assert.Equal(t, float32(100), *e.Brightness)
	assert.Equal(t, 250, *e.Temperature)
	assert.Equal(t, "3s", e.Transition)
}

func TestParseLightArgNameWithColon(t *testing.T) {
	e, err := openhue.ParseLightArg("Desk: left:off")
	require.NoError(t, err)
	assert.Equal(t, "Desk: left", e.Light)
	assert.False(t, *e.On)
}

func TestParseLightArgGradient(t *testing.T) {
	e, err := openhue.ParseLightArg("Strip:gradient=#FF0000;#0000FF,gradient-mode=interpolated_palette_mirrored")
	require.NoError(t, err)
	assert.Equal(t, []string{"#FF0000", "#0000FF"}, e.Gradient.Points)
	assert.Equal(t, "interpolated_palette_mirrored", e.Gradient.Mode)
}

func TestParseLightArgErrors(t *testing.T) {
	for arg, msg := range map[string]string{
		"Desk":                           "expected",
		"Desk:":                          "expected",
		":on":                            "expected",
		"Desk:foo=1":                     "unknown setting",
		"Desk:foo":                       "unknown setting",
		"Desk:brightness=101":            "between 0 and 100",
		"Desk:brightness=abc":            "not a number",
		"Desk:temperature=100":           "between 153 and 500",
		"Desk:temperature=x":             "not an integer",
		"Desk:x=0.5":                     "together",
		"Desk:x=2,y=0.1":                 "between 0 and 1",
		"Desk:rgb=#ZZZZZZ":               "hex",
		"Desk:color=nope":                "not found",
		"Desk:transition=fast":           "invalid transition",
		"Desk:effect=disco":              "unknown effect",
		"Desk:gradient=#FF0000":          "at least 2 points",
		"Desk:gradient-mode=x":           "unknown gradient mode",
		"Desk:color=red,temperature=200": "only one of",
		"Desk:on,,off":                   "empty setting",
	} {
		_, err := openhue.ParseLightArg(arg)
		if assert.Error(t, err, arg) {
			assert.Contains(t, err.Error(), msg, arg)
		}
	}
}

func TestParseColor(t *testing.T) {
	xy, err := openhue.ParseColor("xy@0.5/0.25")
	require.NoError(t, err)
	assert.Equal(t, float32(0.5), xy.X)
	assert.Equal(t, float32(0.25), xy.Y)
	_, err = openhue.ParseColor("#FF0000")
	assert.NoError(t, err)
	_, err = openhue.ParseColor("xy@2/0")
	assert.Error(t, err)
	assert.Equal(t, "xy@0.5/0.25", openhue.FormatXY(0.5, 0.25))
}

func TestParseSceneDefinition(t *testing.T) {
	def, err := openhue.ParseSceneDefinition(strings.NewReader(`
name: Sunset
room: Office
speed: 30
auto_dynamic: true
palette:
  colors: [blue, "#00FF88"]
  temperatures: [200]
defaults:
  brightness: 40
lights:
  - light: Desk
    on: true
    temperature: 250
    transition: 2s
  - light: Strip
    gradient:
      mode: interpolated_palette
      points: ["#FF0000", "#0000FF"]
remove_lights: [Old]
`))
	require.NoError(t, err)
	assert.Equal(t, "Sunset", def.Name)
	assert.Equal(t, float32(30), *def.Speed)
	assert.Len(t, def.Lights, 2)
	assert.Equal(t, 250, *def.Lights[0].Temperature)
	assert.Len(t, def.Lights[1].Gradient.Points, 2)
	assert.Equal(t, []string{"Old"}, def.RemoveLights)
}

func TestParseSceneDefinitionErrors(t *testing.T) {
	_, err := openhue.ParseSceneDefinition(strings.NewReader("lights:\n  - light: Desk\n    brigthness: 50\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3")
	assert.Contains(t, err.Error(), "brigthness")

	_, err = openhue.ParseSceneDefinition(strings.NewReader("speed: fast\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 1")

	_, err = openhue.ParseSceneDefinition(strings.NewReader("speed: 300\n"))
	require.Error(t, err)

	_, err = openhue.ParseSceneDefinition(strings.NewReader("a: [\n"))
	require.Error(t, err)

	_, err = openhue.LoadSceneDefinition("/does/not/exist.yaml", nil)
	require.Error(t, err)
}

func TestLoadSceneDefinitionFromStdin(t *testing.T) {
	def, err := openhue.LoadSceneDefinition("-", strings.NewReader("name: Piped\n"))
	require.NoError(t, err)
	assert.Equal(t, "Piped", def.Name)
}

func TestMergeSceneDefinitions(t *testing.T) {
	base, err := openhue.ParseSceneDefinition(strings.NewReader(`
name: File
room: Office
speed: 10
palette: {colors: [red], effects: [fire]}
defaults: {brightness: 20, temperature: 300}
lights:
  - {light: Desk, brightness: 30, effect: candle}
  - {light: Lamp, on: true}
`))
	require.NoError(t, err)

	speed := float32(60)
	b := float32(80)
	deskOn := true
	over := &openhue.SceneDefinition{
		Name:     "Flag",
		Speed:    &speed,
		Palette:  &openhue.PaletteDefinition{Colors: []string{"blue"}},
		Defaults: &openhue.LightSettings{Color: "green"},
		Lights: []openhue.LightEntry{
			{Light: "Desk", LightSettings: openhue.LightSettings{Brightness: &b, On: &deskOn}},
			{Light: "Shelf", LightSettings: openhue.LightSettings{On: &deskOn}},
		},
	}
	m := openhue.MergeSceneDefinitions(base, over)

	assert.Equal(t, "Flag", m.Name)
	assert.Equal(t, "Office", m.Room)
	assert.Equal(t, float32(60), *m.Speed)
	assert.Equal(t, []string{"blue"}, m.Palette.Colors)
	assert.Equal(t, []string{"fire"}, m.Palette.Effects)
	// color group replaced, brightness kept
	assert.Equal(t, "green", m.Defaults.Color)
	assert.Nil(t, m.Defaults.Temperature)
	assert.Equal(t, float32(20), *m.Defaults.Brightness)
	// same light merged key by key
	require.Len(t, m.Lights, 3)
	assert.Equal(t, float32(80), *m.Lights[0].Brightness)
	assert.Equal(t, "candle", m.Lights[0].Effect)
	assert.Equal(t, "Shelf", m.Lights[2].Light)
	// base untouched
	assert.Equal(t, float32(30), *base.Lights[0].Brightness)
}

func TestSampleDefinitionFileIsValid(t *testing.T) {
	f, err := os.Open("../docs/examples/scene.yaml")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	def, err := openhue.ParseSceneDefinition(f)
	require.NoError(t, err)
	assert.Equal(t, "Sunset", def.Name)
	assert.Len(t, def.Lights, 4)
	assert.Len(t, def.Palette.Colors, 3)
}
