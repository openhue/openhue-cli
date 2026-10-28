package update

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
)

type env struct {
	out  *bytes.Buffer
	mock *oh.ClientWithResponsesMock
}

func run(t *testing.T, home *openhue.HomeModel, args ...string) (*env, error) {
	t.Helper()
	streams, _, out, _ := openhue.NewTestIOStreams()
	h, m := fake.NewTestHomeWithHandlers()
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home

	cmd := NewCmdUpdateScene(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	return &env{out: out, mock: m}, cmd.Execute()
}

func (e *env) put(t *testing.T) (string, oh.ScenePut) {
	t.Helper()
	for _, c := range e.mock.Calls {
		if c.Method == "UpdateSceneWithResponse" {
			return c.Arguments.String(1), c.Arguments.Get(2).(oh.ScenePut)
		}
	}
	t.Fatal("scene update was not requested")
	return "", oh.ScenePut{}
}

func (e *env) notUpdated(t *testing.T) {
	t.Helper()
	e.mock.AssertNotCalled(t, "UpdateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func home() *openhue.HomeModel {
	deskAction := fake.StoredAction("Desk", true, 70, 300)
	effect := oh.Fire
	deskAction.Action.Effects = &struct {
		Effect *oh.SupportedEffects `json:"effect,omitempty"`
	}{Effect: &effect}
	return fake.NewFakeHome(
		[]fake.FakeLight{fake.NewFakeLight("Desk"), fake.NewFakeLight("Ceiling"), fake.NewFakeLight("Shelf")},
		fake.NewFakeScene("s1", "Cozy", deskAction, fake.StoredAction("Ceiling", false, 20, 400)),
		fake.NewFakeScene("s2", "Other", fake.StoredAction("Desk", true, 10, 200)),
	)
}

func actionOf(put oh.ScenePut, light string) oh.ActionPost {
	for _, a := range *put.Actions {
		if a.Target.Rid == "light-"+light {
			return a
		}
	}
	return oh.ActionPost{}
}

func TestUpdateByNameAndById(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--light", "Desk:brightness=20")
	require.NoError(t, err)
	id, put := e.put(t)
	assert.Equal(t, "s1", id)
	assert.Equal(t, float32(20), actionOf(put, "Desk").Action.Dimming.Brightness)
	assert.Contains(t, e.out.String(), "Scene 'Cozy' updated in room 'Office' (s1)")

	e, err = run(t, home(), "s2", "--on")
	require.NoError(t, err)
	id, _ = e.put(t)
	assert.Equal(t, "s2", id)
}

func TestUpdatePreservesUntouchedLights(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--light", "Ceiling:brightness=5")
	require.NoError(t, err)
	_, put := e.put(t)
	require.Len(t, *put.Actions, 2)
	desk := actionOf(put, "Desk")
	assert.Equal(t, float32(70), desk.Action.Dimming.Brightness)
	assert.Equal(t, oh.Fire, *desk.Action.Effects.Effect)
	assert.Equal(t, 300, *desk.Action.ColorTemperature.Mirek)
}

func TestUpdateSceneLookup(t *testing.T) {
	h := home()
	h.Rooms = append(h.Rooms, openhue.Room{
		Resource: openhue.Resource{Id: "room-2", Name: "Bedroom"},
		Scenes:   []openhue.Scene{fake.NewFakeScene("s9", "Cozy")},
	})

	e, err := run(t, h, "Cozy", "--on")
	require.Error(t, err)
	assert.ErrorContains(t, err, "--room")
	e.notUpdated(t)

	e, err = run(t, h, "Cozy", "-r", "Office", "--on")
	require.NoError(t, err)
	id, _ := e.put(t)
	assert.Equal(t, "s1", id)

	e, err = run(t, h, "Nope", "--on")
	require.Error(t, err)
	assert.ErrorContains(t, err, "Nope")
	e.notUpdated(t)
}

func TestUpdateDefaultsAddAndRemove(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--brightness", "10", "--light", "Shelf:on,brightness=30")
	require.NoError(t, err)
	_, put := e.put(t)
	require.Len(t, *put.Actions, 3)
	assert.Equal(t, float32(10), actionOf(put, "Desk").Action.Dimming.Brightness)
	assert.Equal(t, float32(10), actionOf(put, "Ceiling").Action.Dimming.Brightness)
	assert.Equal(t, float32(30), actionOf(put, "Shelf").Action.Dimming.Brightness)

	e, err = run(t, home(), "Cozy", "--remove-light", "Ceiling")
	require.NoError(t, err)
	_, put = e.put(t)
	require.Len(t, *put.Actions, 2)
	assert.True(t, actionOf(put, "Desk").Action.On.On)
	assert.False(t, actionOf(put, "Ceiling").Action.On.On)
	assert.Nil(t, actionOf(put, "Ceiling").Action.Dimming)
}

func TestUpdateRenameDynamicsPalette(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--name", "Relax", "--speed", "60", "--auto-dynamic", "--palette-color", "red")
	require.NoError(t, err)
	_, put := e.put(t)
	assert.Equal(t, "Relax", *put.Metadata.Name)
	assert.InDelta(t, 0.6, *put.Speed, 1e-6)
	assert.True(t, *put.AutoDynamic)
	assert.Len(t, *put.Palette.Color, 1)
	assert.Nil(t, put.Actions, "no light change, no action list")
	assert.Contains(t, e.out.String(), "Scene 'Relax' updated")

	e, err = run(t, home(), "Cozy", "--clear-palette")
	require.NoError(t, err)
	_, put = e.put(t)
	assert.Empty(t, *put.Palette.Color)
}

func TestUpdateFailures(t *testing.T) {
	for name, args := range map[string][]string{
		"no change":       {"Cozy"},
		"name collision":  {"Cozy", "--name", "Other"},
		"unknown light":   {"Cozy", "--light", "Nope:on"},
		"remove absent":   {"Cozy", "--remove-light", "Shelf"},
		"remove and edit": {"Cozy", "--remove-light", "Desk", "--light", "Desk:on"},
		"auto no palette": {"Cozy", "--auto-dynamic"},
		"bad speed":       {"Cozy", "--speed", "200"},
		"no scene arg":    {},
	} {
		t.Run(name, func(t *testing.T) {
			e, err := run(t, home(), args...)
			require.Error(t, err)
			e.notUpdated(t)
		})
	}

	_, err := run(t, home(), "Cozy")
	assert.ErrorContains(t, err, "at least one change")
}

func TestUpdateDryRun(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--brightness", "10", "--dry-run")
	require.NoError(t, err)
	e.notUpdated(t)
	assert.Contains(t, e.out.String(), `"actions"`)
	assert.Contains(t, e.out.String(), `"brightness": 10`)
}

func TestUpdateFromFilePatch(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "patch.yaml")
	require.NoError(t, os.WriteFile(file, []byte("lights:\n  - light: Desk\n    brightness: 20\nremove_lights: [Ceiling]\n"), 0o600))

	e, err := run(t, home(), "Cozy", "--from-file", file)
	require.NoError(t, err)
	_, put := e.put(t)
	require.Len(t, *put.Actions, 2)
	assert.Equal(t, float32(20), actionOf(put, "Desk").Action.Dimming.Brightness)
	assert.False(t, actionOf(put, "Ceiling").Action.On.On)

	// room checks
	match := filepath.Join(dir, "match.yaml")
	require.NoError(t, os.WriteFile(match, []byte("room: Office\nname: Relax\n"), 0o600))
	_, err = run(t, home(), "Cozy", "--from-file", match)
	require.NoError(t, err)

	mismatch := filepath.Join(dir, "mismatch.yaml")
	require.NoError(t, os.WriteFile(mismatch, []byte("room: Elsewhere\nname: Relax\n"), 0o600))
	e, err = run(t, home(), "Cozy", "--from-file", mismatch)
	require.Error(t, err)
	assert.ErrorContains(t, err, "does not match")
	e.notUpdated(t)

	// flags override the file
	e, err = run(t, home(), "Cozy", "--from-file", file, "--light", "Desk:brightness=99")
	require.NoError(t, err)
	_, put = e.put(t)
	assert.Equal(t, float32(99), actionOf(put, "Desk").Action.Dimming.Brightness)
}

func TestUpdateJson(t *testing.T) {
	streams, _, out, _ := openhue.NewTestIOStreams()
	h, m := fake.NewTestHomeWithHandlers()
	name := "Cozy"
	m.On("GetSceneWithResponse", mock.Anything, "s1", mock.Anything).Return(&oh.GetSceneResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200: &struct {
			Data   []oh.SceneGet `json:"data"`
			Errors []oh.Error    `json:"errors"`
		}{Data: []oh.SceneGet{{Id: "s1", Metadata: oh.SceneMetadata{Name: &name}}}},
	}, nil)
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()

	cmd := NewCmdUpdateScene(ctx)
	cmd.SetArgs([]string{"Cozy", "--on", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"id": "s1"`)
}

func TestUpdateBridgeError(t *testing.T) {
	streams, _, out, _ := openhue.NewTestIOStreams()
	h, m := oh.NewTestHome()
	m.On("UpdateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&oh.UpdateSceneResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"}}, nil)
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()

	cmd := NewCmdUpdateScene(ctx)
	cmd.SetArgs([]string{"Cozy", "--on"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Empty(t, out.String())
}
