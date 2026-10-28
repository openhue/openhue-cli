package create

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
)

type env struct {
	cmd  interface{ Execute() error }
	out  *bytes.Buffer
	err  *bytes.Buffer
	in   *bytes.Buffer
	mock *oh.ClientWithResponsesMock
	home *openhue.HomeModel
}

func run(t *testing.T, home *openhue.HomeModel, args ...string) (*env, error) {
	t.Helper()
	streams, in, out, errOut := openhue.NewTestIOStreams()
	h, m := fake.NewTestHomeWithHandlers()
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home

	cmd := NewCmdCreateScene(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	e := &env{cmd: cmd, out: out, err: errOut, in: in, mock: m, home: home}
	return e, cmd.Execute()
}

func (e *env) created(t *testing.T) oh.ScenePost {
	t.Helper()
	for _, c := range e.mock.Calls {
		if c.Method == "CreateSceneWithResponse" {
			return c.Arguments.Get(1).(oh.ScenePost)
		}
	}
	t.Fatal("scene creation was not requested")
	return oh.ScenePost{}
}

func (e *env) notCreated(t *testing.T) {
	t.Helper()
	e.mock.AssertNotCalled(t, "CreateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything)
}

func home() *openhue.HomeModel {
	ceiling := fake.NewFakeLight("Ceiling")
	ceiling.On = false
	return fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Desk"), ceiling}, fake.NewFakeScene("s1", "Existing"))
}

func action(post oh.ScenePost, light string) oh.ActionPost {
	for _, a := range post.Actions {
		if a.Target.Rid == "light-"+light {
			return a
		}
	}
	return oh.ActionPost{}
}

func TestCreateSnapshot(t *testing.T) {
	e, err := run(t, home(), "Cozy", "--room", "Office")
	require.NoError(t, err)
	post := e.created(t)
	assert.Equal(t, "Cozy", *post.Metadata.Name)
	assert.Equal(t, "room-1", post.Group.Rid)
	assert.Len(t, post.Actions, 2)
	assert.False(t, action(post, "Ceiling").Action.On.On)
	assert.Contains(t, e.out.String(), "Scene 'Cozy' created in room 'Office' (new-scene)")
}

func TestCreateByRoomId(t *testing.T) {
	e, err := run(t, home(), "Cozy", "-r", "room-1")
	require.NoError(t, err)
	assert.Equal(t, "room-1", e.created(t).Group.Rid)
}

func TestCreatePerLight(t *testing.T) {
	e, err := run(t, home(), "Work", "-r", "Office", "--light", "Desk:on,brightness=100,temperature=250", "--light", "Ceiling:off,transition=3s")
	require.NoError(t, err)
	post := e.created(t)
	desk := action(post, "Desk")
	assert.Equal(t, float32(100), desk.Action.Dimming.Brightness)
	assert.Equal(t, 250, *desk.Action.ColorTemperature.Mirek)
	ceiling := action(post, "Ceiling")
	assert.False(t, ceiling.Action.On.On)
	assert.Equal(t, 3000, *ceiling.Action.Dynamics.Duration)
}

func TestCreateDefaultsWithOverride(t *testing.T) {
	e, err := run(t, home(), "Dim", "-r", "Office", "--brightness", "10", "--color", "red", "--transition-time", "2s", "--light", "Desk:brightness=40")
	require.NoError(t, err)
	post := e.created(t)
	assert.Equal(t, float32(40), action(post, "Desk").Action.Dimming.Brightness)
	assert.Equal(t, float32(10), action(post, "Ceiling").Action.Dimming.Brightness)
	assert.NotNil(t, action(post, "Ceiling").Action.Color)
	assert.Equal(t, 2000, *action(post, "Ceiling").Action.Dynamics.Duration)
}

func TestCreateDefaultWarningOnWhiteLight(t *testing.T) {
	white := fake.NewFakeLight("Lamp")
	white.Color = false
	h := fake.NewFakeHome([]fake.FakeLight{white, fake.NewFakeLight("Desk")})
	e, err := run(t, h, "S", "-r", "Office", "--color", "red")
	require.NoError(t, err)
	assert.Contains(t, e.err.String(), "Lamp")
	assert.Nil(t, action(e.created(t), "Lamp").Action.Color)
}

func TestCreateOnlyListed(t *testing.T) {
	e, err := run(t, home(), "Desk", "-r", "Office", "--only-listed", "--light", "Desk:on")
	require.NoError(t, err)
	post := e.created(t)
	assert.Len(t, post.Actions, 2, "the bridge needs every light of the room")
	assert.True(t, action(post, "Desk").Action.On.On)
	assert.False(t, action(post, "Ceiling").Action.On.On)

	e, err = run(t, home(), "Desk", "-r", "Office", "--only-listed")
	require.Error(t, err)
	e.notCreated(t)
}

func TestCreateDynamics(t *testing.T) {
	e, err := run(t, home(), "Aurora", "-r", "Office", "--auto-dynamic", "--speed", "30",
		"--palette-color", "blue", "--palette-color", "#00FF88", "--palette-temperature", "200", "--palette-brightness", "60", "--palette-effect", "fire")
	require.NoError(t, err)
	post := e.created(t)
	assert.InDelta(t, 0.3, *post.Speed, 1e-6)
	assert.True(t, *post.AutoDynamic)
	assert.Len(t, *post.Palette.Color, 2)
	assert.Len(t, *post.Palette.ColorTemperature, 1)
	assert.Len(t, *post.Palette.Dimming, 1)
	assert.Len(t, *post.Palette.Effects, 1)
}

func TestCreateFromFileAndFlagsOverride(t *testing.T) {
	file := filepath.Join(t.TempDir(), "s.yaml")
	require.NoError(t, os.WriteFile(file, []byte(`
name: Sunset
room: Elsewhere
defaults: {brightness: 20}
lights:
  - {light: Desk, temperature: 300}
`), 0o600))

	e, err := run(t, home(), "--from-file", file, "--room", "Office")
	require.NoError(t, err)
	post := e.created(t)
	assert.Equal(t, "Sunset", *post.Metadata.Name)
	assert.Equal(t, 300, *action(post, "Desk").Action.ColorTemperature.Mirek)
	assert.Equal(t, float32(20), action(post, "Ceiling").Action.Dimming.Brightness)

	e, err = run(t, home(), "Renamed", "--from-file", file, "--room", "Office", "--brightness", "5", "--light", "Desk:brightness=90")
	require.NoError(t, err)
	post = e.created(t)
	assert.Equal(t, "Renamed", *post.Metadata.Name)
	assert.Equal(t, float32(5), action(post, "Ceiling").Action.Dimming.Brightness)
	assert.Equal(t, float32(90), action(post, "Desk").Action.Dimming.Brightness)
	assert.Equal(t, 300, *action(post, "Desk").Action.ColorTemperature.Mirek, "file entry merged with the flag entry")
}

func TestCreateFromStdin(t *testing.T) {
	streams, in, out, _ := openhue.NewTestIOStreams()
	h, m := fake.NewTestHomeWithHandlers()
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()
	in.WriteString("name: Piped\nroom: Office\n")

	cmd := NewCmdCreateScene(ctx)
	cmd.SetArgs([]string{"--from-file", "-"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Piped")
	m.AssertCalled(t, "CreateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateFileErrors(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(bad, []byte("lights:\n  - light: Desk\n    brigthness: 50\n"), 0o600))

	e, err := run(t, home(), "S", "-r", "Office", "--from-file", bad)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "brigthness")
	e.notCreated(t)

	e, err = run(t, home(), "S", "-r", "Office", "--from-file", filepath.Join(dir, "missing.yaml"))
	require.Error(t, err)
	e.notCreated(t)
}

func TestCreateDryRun(t *testing.T) {
	e, err := run(t, home(), "Cozy", "-r", "Office", "--dry-run")
	require.NoError(t, err)
	e.notCreated(t)
	assert.Contains(t, e.out.String(), `"name": "Cozy"`)

	// validation still applies
	e, err = run(t, home(), "Existing", "-r", "Office", "--dry-run")
	require.Error(t, err)
	assert.Empty(t, e.out.String())
}

func TestCreateJson(t *testing.T) {
	streams, _, out, _ := openhue.NewTestIOStreams()
	h, m := fake.NewTestHomeWithHandlers()
	name := "Cozy"
	m.On("GetSceneWithResponse", mock.Anything, "new-scene", mock.Anything).Return(&oh.GetSceneResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		JSON200: &struct {
			Data   []oh.SceneGet `json:"data"`
			Errors []oh.Error    `json:"errors"`
		}{Data: []oh.SceneGet{{Id: "new-scene", Metadata: oh.SceneMetadata{Name: &name}}}},
	}, nil)
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()

	cmd := NewCmdCreateScene(ctx)
	cmd.SetArgs([]string{"Cozy", "-r", "Office", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"id": "new-scene"`)
	assert.NotContains(t, out.String(), "created in room")
}

func TestCreateArgumentErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"missing name":    {"-r", "Office"},
		"missing room":    {"Cozy"},
		"unknown room":    {"Cozy", "-r", "Nope"},
		"duplicate name":  {"Existing", "-r", "Office"},
		"unknown light":   {"Cozy", "-r", "Office", "--light", "Nope:on"},
		"bad setting":     {"Cozy", "-r", "Office", "--light", "Desk:foo=1"},
		"bad speed":       {"Cozy", "-r", "Office", "--speed", "101"},
		"auto no palette": {"Cozy", "-r", "Office", "--auto-dynamic"},
		"two args":        {"A", "B", "-r", "Office"},
	} {
		t.Run(name, func(t *testing.T) {
			e, err := run(t, home(), args...)
			require.Error(t, err)
			e.notCreated(t)
			assert.Empty(t, strings.TrimSpace(e.out.String()))
		})
	}

	t.Run("missing room message", func(t *testing.T) {
		_, err := run(t, home(), "Cozy")
		assert.ErrorContains(t, err, "room is required")
	})
	t.Run("unknown room message", func(t *testing.T) {
		_, err := run(t, home(), "Cozy", "-r", "Nope")
		assert.ErrorContains(t, err, "Nope")
	})
}

func TestCreateRoomWithoutLights(t *testing.T) {
	e, err := run(t, fake.NewFakeHome(nil), "Cozy", "-r", "Office")
	require.Error(t, err)
	e.notCreated(t)
}

func TestCreateBridgeError(t *testing.T) {
	streams, _, out, _ := openhue.NewTestIOStreams()
	h, m := oh.NewTestHome()
	m.On("CreateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything).
		Return(&oh.CreateSceneResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict, Status: "409 Conflict"}}, nil)
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()

	cmd := NewCmdCreateScene(ctx)
	cmd.SetArgs([]string{"Cozy", "-r", "Office"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409")
	assert.Empty(t, out.String())
}
