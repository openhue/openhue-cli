package openhue_test

import (
	"encoding/json"
	"net/http"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
)

func richScene() *openhue.HomeModel {
	strip := fake.NewFakeLight("Strip")
	strip.Gradient, strip.Effects = true, true

	a1 := fake.StoredAction("Desk", true, 70.5, 300)
	e := oh.Candle
	a1.Action.ColorTemperature = nil // the bridge does not accept an effect with a temperature
	a1.Action.Effects = &struct {
		Effect *oh.SupportedEffects `json:"effect,omitempty"`
	}{Effect: &e}

	a2 := oh.ActionGet{Target: oh.ResourceIdentifier{Rid: "light-Strip", Rtype: oh.ResourceIdentifierRtypeLight}}
	a2.Action.On = &oh.On{On: true}
	a2.Action.Dimming = &oh.Dimming{Brightness: 50}
	mode := oh.InterpolatedPalette
	a2.Action.Gradient = &oh.Gradient{Mode: &mode, Points: &[]oh.Color{{Xy: oh.GamutPosition{X: 0.7, Y: 0.3}}, {Xy: oh.GamutPosition{X: 0.15, Y: 0.06}}}}

	a3 := oh.ActionGet{Target: oh.ResourceIdentifier{Rid: "light-Lamp", Rtype: oh.ResourceIdentifierRtypeLight}}
	a3.Action.On = &oh.On{On: false}
	a3.Action.Dimming = &oh.Dimming{Brightness: 50}
	a3.Action.Color = &oh.Color{Xy: oh.GamutPosition{X: 0.3127, Y: 0.329}}

	desk, lamp := fake.NewFakeLight("Desk"), fake.NewFakeLight("Lamp")
	desk.Effects = true
	scene := fake.NewFakeScene("s1", "Aurora", a1, a2, a3)
	scene.HueData.Speed = 0.3
	scene.HueData.AutoDynamic = true
	c := oh.Color{Xy: oh.GamutPosition{X: 0.1, Y: 0.2}}
	scene.HueData.Palette = &oh.ScenePalette{
		Color:   &[]oh.ColorPaletteGet{{Color: &c, Dimming: &oh.Dimming{Brightness: 80}}},
		Dimming: &[]oh.Dimming{{Brightness: 40}},
	}
	return fake.NewFakeHome([]fake.FakeLight{desk, strip, lamp}, scene)
}

func TestExportScene(t *testing.T) {
	home := richScene()
	def := openhue.SceneToDefinition(&home.Rooms[0], &home.Rooms[0].Scenes[0])

	assert.Equal(t, "Aurora", def.Name)
	assert.Equal(t, "Office", def.Room)
	assert.Equal(t, float32(30), *def.Speed)
	assert.True(t, *def.AutoDynamic)
	assert.Equal(t, []string{"xy@0.1/0.2"}, def.Palette.Colors)
	require.Len(t, def.Lights, 3)
	assert.Equal(t, "Desk", def.Lights[0].Light)
	assert.Nil(t, def.Lights[0].Temperature)
	assert.Equal(t, "candle", def.Lights[0].Effect)
	assert.Len(t, def.Lights[1].Gradient.Points, 2)
	assert.Equal(t, float32(0.3127), *def.Lights[2].X)

	out, err := openhue.MarshalSceneDefinition(def)
	require.NoError(t, err)
	assert.Contains(t, string(out), "light: Desk")
}

func TestExportUsesIdWhenNameAmbiguous(t *testing.T) {
	home := fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Lamp"), fake.NewFakeLight("Lamp")},
		fake.NewFakeScene("s1", "S", fake.StoredAction("Lamp", true, 10, 200)))
	def := openhue.SceneToDefinition(&home.Rooms[0], &home.Rooms[0].Scenes[0])
	assert.Equal(t, "light-Lamp", def.Lights[0].Light)
}

func TestExportReapplyIsNoOp(t *testing.T) {
	home := richScene()
	room, scene := &home.Rooms[0], &home.Rooms[0].Scenes[0]

	out, err := openhue.MarshalSceneDefinition(openhue.SceneToDefinition(room, scene))
	require.NoError(t, err)
	def, err := openhue.ParseSceneDefinition(bytesReader(out))
	require.NoError(t, err)

	put, _, err := openhue.BuildScenePut(room, scene, def)
	require.NoError(t, err)

	// actions unchanged
	require.NotNil(t, put.Actions)
	want, err := json.Marshal(storedAsPost(t, scene.HueData.Actions))
	require.NoError(t, err)
	got, err := json.Marshal(*put.Actions)
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(got))
	// no other change
	assert.Nil(t, put.Metadata)
	assert.Nil(t, put.Speed)
	assert.Nil(t, put.AutoDynamic)
	assert.Nil(t, put.Palette)
}

func TestExportCloneIsEquivalent(t *testing.T) {
	home := richScene()
	room, scene := &home.Rooms[0], &home.Rooms[0].Scenes[0]

	out, err := openhue.MarshalSceneDefinition(openhue.SceneToDefinition(room, scene))
	require.NoError(t, err)
	def, err := openhue.ParseSceneDefinition(bytesReader(out))
	require.NoError(t, err)
	def.Name = "Copy"

	post, _, err := openhue.BuildScenePost(room, def, true)
	require.NoError(t, err)

	want, err := json.Marshal(storedAsPost(t, scene.HueData.Actions))
	require.NoError(t, err)
	got, err := json.Marshal(post.Actions)
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(got))
	assert.InDelta(t, 0.3, *post.Speed, 1e-6)
	assert.True(t, *post.AutoDynamic)
	assert.Len(t, *post.Palette.Color, 1)
	assert.Len(t, *post.Palette.Dimming, 1)
}

//
// Services
//

func TestSceneServices(t *testing.T) {
	home := twoLightHome(fake.NewFakeScene("s1", "Cozy", fake.StoredAction("Desk", true, 70, 300)))
	h, m := fake.NewTestHomeWithHandlers()
	room, scene := &home.Rooms[0], &home.Rooms[0].Scenes[0]

	id, err := room.CreateScene(h, oh.ScenePost{})
	require.NoError(t, err)
	assert.Equal(t, "new-scene", id)

	require.NoError(t, scene.Update(h, oh.ScenePut{}))
	require.NoError(t, scene.Delete(h))
	m.AssertCalled(t, "DeleteSceneWithResponse", mock.Anything, "s1", mock.Anything)
}

func TestSceneServiceBridgeError(t *testing.T) {
	h, m := oh.NewTestHome()
	m.On("DeleteSceneWithResponse", mock.Anything, mock.Anything, mock.Anything).
		Return(&oh.DeleteSceneResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError, Status: "500"}}, nil)
	s := openhue.Scene{Resource: openhue.Resource{Id: "x"}}
	assert.Error(t, s.Delete(h))
}

func TestFindRoomAndScene(t *testing.T) {
	home := twoLightHome(fake.NewFakeScene("s1", "Cozy"), fake.NewFakeScene("s2", "Cozy"), fake.NewFakeScene("s3", "Solo"))

	r, err := openhue.FindRoom(home, "Office")
	require.NoError(t, err)
	assert.Equal(t, "room-1", r.Id)
	_, err = openhue.FindRoom(home, "room-1")
	assert.NoError(t, err)
	_, err = openhue.FindRoom(home, "Nope")
	assert.Error(t, err)

	s, room, err := openhue.FindScene(home, "", "Solo")
	require.NoError(t, err)
	assert.Equal(t, "s3", s.Id)
	assert.Equal(t, "room-1", room.Id)
	_, _, err = openhue.FindScene(home, "", "s2")
	assert.NoError(t, err)
	_, _, err = openhue.FindScene(home, "", "Cozy")
	assert.ErrorContains(t, err, "several")
	_, _, err = openhue.FindScene(home, "Office", "Cozy")
	assert.ErrorContains(t, err, "several")
	_, _, err = openhue.FindScene(home, "", "Nope")
	assert.ErrorContains(t, err, "not found")
	_, _, err = openhue.FindScene(home, "Elsewhere", "Solo")
	assert.ErrorContains(t, err, "not found")
}
