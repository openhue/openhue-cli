package openhue_test

import (
	"encoding/json"
	"strings"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
)

func f32(v float32) *float32 { return &v }
func bl(v bool) *bool        { return &v }
func in(v int) *int          { return &v }

func twoLightHome(scenes ...openhue.Scene) *openhue.HomeModel {
	desk := fake.NewFakeLight("Desk")
	ceiling := fake.NewFakeLight("Ceiling")
	ceiling.On = false
	ceiling.Bri = 80
	return fake.NewFakeHome([]fake.FakeLight{desk, ceiling}, scenes...)
}

func action(t *testing.T, post *oh.ScenePost, light string) oh.ActionPost {
	t.Helper()
	for _, a := range post.Actions {
		if a.Target.Rid == "light-"+light {
			return a
		}
	}
	t.Fatalf("no action for %s", light)
	return oh.ActionPost{}
}

func TestBuildScenePostSnapshot(t *testing.T) {
	home := twoLightHome()
	post, warns, err := openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "Cozy"}, false)
	require.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, "Cozy", *post.Metadata.Name)
	assert.Equal(t, "room-1", post.Group.Rid)
	assert.Equal(t, oh.ResourceIdentifierRtypeRoom, post.Group.Rtype)
	require.Len(t, post.Actions, 2)

	desk := action(t, post, "Desk")
	assert.True(t, desk.Action.On.On)
	assert.Equal(t, float32(50), desk.Action.Dimming.Brightness)
	assert.Equal(t, 366, *desk.Action.ColorTemperature.Mirek)
	assert.Nil(t, desk.Action.Color)

	ceiling := action(t, post, "Ceiling")
	assert.False(t, ceiling.Action.On.On)
	assert.Equal(t, float32(80), ceiling.Action.Dimming.Brightness)
}

func TestBuildScenePostSnapshotUsesColorWhenNoTemperature(t *testing.T) {
	l := fake.NewFakeLight("Desk")
	l.Mirek = nil
	l.XY = &[2]float32{0.1, 0.2}
	home := fake.NewFakeHome([]fake.FakeLight{l})
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "S"}, false)
	require.NoError(t, err)
	a := action(t, post, "Desk")
	assert.Equal(t, float32(0.1), a.Action.Color.Xy.X)
}

func TestBuildScenePostPrecedence(t *testing.T) {
	home := twoLightHome()
	def := &openhue.SceneDefinition{
		Name:     "Dim",
		Defaults: &openhue.LightSettings{Brightness: f32(10), Transition: "1s"},
		Lights: []openhue.LightEntry{
			{Light: "Desk", LightSettings: openhue.LightSettings{Brightness: f32(40), Transition: "3s"}},
		},
	}
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)

	desk := action(t, post, "Desk")
	assert.Equal(t, float32(40), desk.Action.Dimming.Brightness)
	assert.Equal(t, 3000, *desk.Action.Dynamics.Duration)
	// other attributes still from the snapshot
	assert.Equal(t, 366, *desk.Action.ColorTemperature.Mirek)

	ceiling := action(t, post, "Ceiling")
	assert.Equal(t, float32(10), ceiling.Action.Dimming.Brightness)
	assert.Equal(t, 1000, *ceiling.Action.Dynamics.Duration)
}

func TestBuildScenePostColorGroupsAreExclusive(t *testing.T) {
	home := twoLightHome()
	def := &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{
		{Light: "Desk", LightSettings: openhue.LightSettings{Color: "red"}},
	}}
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	a := action(t, post, "Desk")
	assert.NotNil(t, a.Action.Color)
	assert.Nil(t, a.Action.ColorTemperature, "the snapshot temperature is replaced by the color")

	def.Lights[0].LightSettings = openhue.LightSettings{Temperature: in(200)}
	post, _, err = openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	a = action(t, post, "Desk")
	assert.Equal(t, 200, *a.Action.ColorTemperature.Mirek)
	assert.Nil(t, a.Action.Color)
}

func TestBuildScenePostOnlyListed(t *testing.T) {
	home := twoLightHome()
	def := &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Desk", LightSettings: openhue.LightSettings{On: bl(true)}}}}
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], def, true)
	require.NoError(t, err)
	require.Len(t, post.Actions, 2, "the bridge needs every light of the room")
	assert.True(t, action(t, post, "Desk").Action.On.On)
	ceiling := action(t, post, "Ceiling")
	assert.False(t, ceiling.Action.On.On)
	assert.Nil(t, ceiling.Action.Dimming)
	assert.Nil(t, ceiling.Action.Color)

	_, _, err = openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "S"}, true)
	require.Error(t, err)
}

func TestBuildScenePostWhiteOnlyLight(t *testing.T) {
	white := fake.NewFakeLight("Lamp")
	white.Color = false
	home := fake.NewFakeHome([]fake.FakeLight{white, fake.NewFakeLight("Desk")})

	// default color: skipped with a warning
	def := &openhue.SceneDefinition{Name: "S", Defaults: &openhue.LightSettings{Color: "red"}}
	post, warns, err := openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	require.Len(t, warns, 1)
	assert.Contains(t, warns[0], "Lamp")
	assert.Nil(t, action(t, post, "Lamp").Action.Color)
	assert.NotNil(t, action(t, post, "Desk").Action.Color)

	// explicit color: error naming the light
	def = &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Lamp", LightSettings: openhue.LightSettings{Color: "red"}}}}
	_, _, err = openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Lamp")
	assert.Contains(t, err.Error(), "colors")
}

func TestBuildScenePostEffectAndGradient(t *testing.T) {
	strip := fake.NewFakeLight("Strip")
	strip.Gradient, strip.Effects = true, true
	home := fake.NewFakeHome([]fake.FakeLight{strip, fake.NewFakeLight("Desk")})

	def := &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Strip", LightSettings: openhue.LightSettings{
		Gradient: &openhue.GradientSettings{Mode: "interpolated_palette", Points: []string{"#FF0000", "#0000FF"}},
	}}}}
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	a := action(t, post, "Strip")
	assert.Len(t, *a.Action.Gradient.Points, 2)
	assert.Equal(t, oh.InterpolatedPalette, *a.Action.Gradient.Mode)
	assert.Nil(t, a.Action.ColorTemperature)
	assert.Nil(t, a.Action.Color)

	// an effect replaces the snapshot color/temperature, the bridge rejects both together
	def.Lights[0].LightSettings = openhue.LightSettings{Effect: "candle"}
	post, _, err = openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	a = action(t, post, "Strip")
	assert.Equal(t, oh.Candle, *a.Action.Effects.Effect)
	assert.Nil(t, a.Action.Color)
	assert.Nil(t, a.Action.ColorTemperature)
	assert.Nil(t, a.Action.Gradient)

	for _, s := range []openhue.LightSettings{{Effect: "candle"}, {Gradient: &openhue.GradientSettings{Points: []string{"red", "blue"}}}} {
		def = &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Desk", LightSettings: s}}}
		_, _, err = openhue.BuildScenePost(&home.Rooms[0], def, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Desk")
	}
}

func TestBuildScenePostDynamics(t *testing.T) {
	home := twoLightHome()
	def := &openhue.SceneDefinition{
		Name: "Aurora", Speed: f32(30), AutoDynamic: bl(true),
		Palette: &openhue.PaletteDefinition{Colors: []string{"blue", "#00FF88"}, Temperatures: []int{200}, Brightness: []float32{50}, Effects: []string{"fire"}},
	}
	post, _, err := openhue.BuildScenePost(&home.Rooms[0], def, false)
	require.NoError(t, err)
	assert.InDelta(t, 0.3, *post.Speed, 1e-6)
	assert.True(t, *post.AutoDynamic)
	assert.Len(t, *post.Palette.Color, 2)
	assert.Equal(t, 200, *(*post.Palette.ColorTemperature)[0].ColorTemperature.Mirek)
	assert.Len(t, *post.Palette.Dimming, 1)
	assert.Len(t, *post.Palette.Effects, 1)
}

func TestBuildScenePostValidation(t *testing.T) {
	home := twoLightHome(fake.NewFakeScene("s1", "Exists"))
	room := &home.Rooms[0]
	empty := fake.NewFakeHome(nil)

	cases := map[string]struct {
		room *openhue.Room
		def  *openhue.SceneDefinition
		msg  string
	}{
		"no name":        {room, &openhue.SceneDefinition{}, "name is required"},
		"duplicate":      {room, &openhue.SceneDefinition{Name: "Exists"}, "already exists"},
		"no lights":      {&empty.Rooms[0], &openhue.SceneDefinition{Name: "S"}, "no light"},
		"unknown light":  {room, &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Nope"}}}, "not found"},
		"twice":          {room, &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Desk"}, {Light: "light-Desk"}}}, "more than once"},
		"bad speed":      {room, &openhue.SceneDefinition{Name: "S", Speed: f32(101)}, "speed"},
		"bad palette":    {room, &openhue.SceneDefinition{Name: "S", Palette: &openhue.PaletteDefinition{Temperatures: []int{10}}}, "palette temperature"},
		"auto no palete": {room, &openhue.SceneDefinition{Name: "S", AutoDynamic: bl(true)}, "palette entry"},
		"remove":         {room, &openhue.SceneDefinition{Name: "S", RemoveLights: []string{"Desk"}}, "updating"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := openhue.BuildScenePost(c.room, c.def, false)
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.msg)
		})
	}
}

func TestBuildScenePostAmbiguousLight(t *testing.T) {
	home := fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Lamp"), fake.NewFakeLight("Lamp")})
	_, _, err := openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Lamp"}}}, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "several")
}

//
// Update
//

func updateHome() (*openhue.HomeModel, *openhue.Scene) {
	stored := fake.StoredAction("Desk", true, 70, 300)
	effect := oh.Fire
	stored.Action.Effects = &struct {
		Effect *oh.SupportedEffects `json:"effect,omitempty"`
	}{Effect: &effect}
	other := fake.StoredAction("Ceiling", false, 20, 400)
	home := twoLightHome(fake.NewFakeScene("s1", "Cozy", stored, other), fake.NewFakeScene("s2", "Other", fake.StoredAction("Desk", true, 10, 200)))
	return home, &home.Rooms[0].Scenes[0]
}

func putAction(t *testing.T, put *oh.ScenePut, light string) oh.ActionPost {
	t.Helper()
	require.NotNil(t, put.Actions)
	for _, a := range *put.Actions {
		if a.Target.Rid == "light-"+light {
			return a
		}
	}
	t.Fatalf("no action for %s", light)
	return oh.ActionPost{}
}

func TestBuildScenePutKeepsUntouchedLights(t *testing.T) {
	home, scene := updateHome()
	def := &openhue.SceneDefinition{Lights: []openhue.LightEntry{{Light: "Ceiling", LightSettings: openhue.LightSettings{Brightness: f32(5)}}}}
	put, _, err := openhue.BuildScenePut(&home.Rooms[0], scene, def)
	require.NoError(t, err)
	require.Len(t, *put.Actions, 2)

	desk := putAction(t, put, "Desk")
	assert.Equal(t, float32(70), desk.Action.Dimming.Brightness, "stored value, not the current light state")
	assert.Equal(t, 300, *desk.Action.ColorTemperature.Mirek)
	assert.Equal(t, oh.Fire, *desk.Action.Effects.Effect, "stored effect preserved")
	assert.Equal(t, float32(5), putAction(t, put, "Ceiling").Action.Dimming.Brightness)
	assert.Nil(t, put.Metadata)
	assert.Nil(t, put.Speed)
}

func TestBuildScenePutDefaultsAndAdd(t *testing.T) {
	home := fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Desk"), fake.NewFakeLight("Shelf")},
		fake.NewFakeScene("s1", "Cozy", fake.StoredAction("Desk", true, 70, 300)))
	scene := &home.Rooms[0].Scenes[0]

	def := &openhue.SceneDefinition{
		Defaults: &openhue.LightSettings{Brightness: f32(10)},
		Lights:   []openhue.LightEntry{{Light: "Shelf", LightSettings: openhue.LightSettings{On: bl(true), Brightness: f32(30)}}},
	}
	put, _, err := openhue.BuildScenePut(&home.Rooms[0], scene, def)
	require.NoError(t, err)
	require.Len(t, *put.Actions, 2)
	assert.Equal(t, float32(10), putAction(t, put, "Desk").Action.Dimming.Brightness)
	shelf := putAction(t, put, "Shelf")
	assert.Equal(t, float32(30), shelf.Action.Dimming.Brightness)
	assert.Equal(t, 366, *shelf.Action.ColorTemperature.Mirek, "added light starts from its current state")
}

func TestBuildScenePutRemove(t *testing.T) {
	home, scene := updateHome()
	room := &home.Rooms[0]

	put, _, err := openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{RemoveLights: []string{"Ceiling"}})
	require.NoError(t, err)
	require.Len(t, *put.Actions, 2, "a removed light is turned off, not dropped")
	assert.Equal(t, float32(70), putAction(t, put, "Desk").Action.Dimming.Brightness)
	ceiling := putAction(t, put, "Ceiling")
	assert.False(t, ceiling.Action.On.On)
	assert.Nil(t, ceiling.Action.Dimming)
	assert.Nil(t, ceiling.Action.ColorTemperature)

	// removing every light turns them all off; defaults do not touch the removed ones
	put, _, err = openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{RemoveLights: []string{"Desk"}, Defaults: &openhue.LightSettings{Brightness: f32(10)}})
	require.NoError(t, err)
	assert.Nil(t, putAction(t, put, "Desk").Action.Dimming)
	assert.Equal(t, float32(10), putAction(t, put, "Ceiling").Action.Dimming.Brightness)

	for name, def := range map[string]*openhue.SceneDefinition{
		"not there": {RemoveLights: []string{"Nope"}},
		"both":      {RemoveLights: []string{"Desk"}, Lights: []openhue.LightEntry{{Light: "Desk", LightSettings: openhue.LightSettings{On: bl(true)}}}},
	} {
		_, _, err = openhue.BuildScenePut(room, scene, def)
		assert.Error(t, err, name)
	}

	// light of the room but not of the scene
	_, _, err = openhue.BuildScenePut(room, &home.Rooms[0].Scenes[1], &openhue.SceneDefinition{RemoveLights: []string{"Ceiling"}})
	assert.Error(t, err)
}

func TestBuildScenePutRename(t *testing.T) {
	home, scene := updateHome()
	put, _, err := openhue.BuildScenePut(&home.Rooms[0], scene, &openhue.SceneDefinition{Name: "Relax"})
	require.NoError(t, err)
	assert.Equal(t, "Relax", *put.Metadata.Name)
	assert.Nil(t, put.Actions)

	_, _, err = openhue.BuildScenePut(&home.Rooms[0], scene, &openhue.SceneDefinition{Name: "Other"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	put, _, err = openhue.BuildScenePut(&home.Rooms[0], scene, &openhue.SceneDefinition{Name: "Cozy"})
	require.NoError(t, err)
	assert.Nil(t, put.Metadata)
}

func TestBuildScenePutDynamicsAndPalette(t *testing.T) {
	home, scene := updateHome()
	room := &home.Rooms[0]

	put, _, err := openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{Speed: f32(60)})
	require.NoError(t, err)
	assert.InDelta(t, 0.6, *put.Speed, 1e-6)
	assert.Nil(t, put.Actions)
	assert.Nil(t, put.Palette)

	// auto dynamic needs a palette
	_, _, err = openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{AutoDynamic: bl(true)})
	require.Error(t, err)

	put, _, err = openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{AutoDynamic: bl(true), Palette: &openhue.PaletteDefinition{Colors: []string{"red"}}})
	require.NoError(t, err)
	assert.True(t, *put.AutoDynamic)
	assert.Len(t, *put.Palette.Color, 1)

	// clear
	put, _, err = openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{ClearPalette: true})
	require.NoError(t, err)
	require.NotNil(t, put.Palette)
	assert.Empty(t, *put.Palette.Color)

	scene.HueData.AutoDynamic = true
	_, _, err = openhue.BuildScenePut(room, scene, &openhue.SceneDefinition{ClearPalette: true})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "auto-dynamic")
}

func TestBuildScenePutPaletteListsReplaceOnlyNamedLists(t *testing.T) {
	home, scene := updateHome()
	c := oh.Color{Xy: oh.GamutPosition{X: 0.1, Y: 0.2}}
	stored := []oh.ColorPaletteGet{{Color: &c, Dimming: &oh.Dimming{Brightness: 40}}}
	scene.HueData.Palette = &oh.ScenePalette{Color: &stored, Dimming: &[]oh.Dimming{{Brightness: 10}}}

	put, _, err := openhue.BuildScenePut(&home.Rooms[0], scene, &openhue.SceneDefinition{Palette: &openhue.PaletteDefinition{Temperatures: []int{250}}})
	require.NoError(t, err)
	assert.Len(t, *put.Palette.Color, 1, "unnamed lists are kept")
	assert.Len(t, *put.Palette.Dimming, 1)
	assert.Len(t, *put.Palette.ColorTemperature, 1)

	// same colors as stored: nothing to change
	put, _, err = openhue.BuildScenePut(&home.Rooms[0], scene, &openhue.SceneDefinition{Palette: &openhue.PaletteDefinition{Colors: []string{"xy@0.1/0.2"}}})
	require.NoError(t, err)
	assert.Nil(t, put.Palette)
}

func bytesReader(b []byte) *strings.Reader { return strings.NewReader(string(b)) }

func storedAsPost(t *testing.T, actions []oh.ActionGet) []oh.ActionPost {
	t.Helper()
	var out []oh.ActionPost
	raw, err := json.Marshal(actions)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestBuildScenePostLightWithoutDimming(t *testing.T) {
	l := fake.NewFakeLight("Cloud")
	l.NoDimming = true
	home := fake.NewFakeHome([]fake.FakeLight{l, fake.NewFakeLight("Desk")})

	post, warns, err := openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "S", Defaults: &openhue.LightSettings{Brightness: f32(20)}}, false)
	require.NoError(t, err)
	assert.Len(t, warns, 1)
	assert.Nil(t, action(t, post, "Cloud").Action.Dimming)
	assert.Equal(t, float32(20), action(t, post, "Desk").Action.Dimming.Brightness)

	_, _, err = openhue.BuildScenePost(&home.Rooms[0], &openhue.SceneDefinition{Name: "S", Lights: []openhue.LightEntry{{Light: "Cloud", LightSettings: openhue.LightSettings{Brightness: f32(20)}}}}, false)
	assert.ErrorContains(t, err, "brightness")
}
