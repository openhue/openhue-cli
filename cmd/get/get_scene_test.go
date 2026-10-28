package get

import (
	"bytes"
	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
	"strings"
	"testing"
)

func TestGetAllScenes(t *testing.T) {
	mockCtx, _ := openhue.NewTestContext(NewFakeHome())
	cmd := NewCmdGetScene(mockCtx, &CmdGetOptions{Json: true})
	err := cmd.Execute()
	assert.Nil(t, err)
}

func NewFakeHome() *openhue.HomeModel {
	return &openhue.HomeModel{
		Resource: openhue.Resource{
			Id:     "home-01",
			Name:   "HomeModel",
			Type:   openhue.HomeResourceType(oh.BridgeHomeGetTypeBridgeHome),
			Parent: nil,
		},
		Rooms: []openhue.Room{
			{
				Resource: openhue.Resource{
					Id:   "room-01",
					Name: "Room 1",
					Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom),
				},
				Devices: nil,
				Scenes: []openhue.Scene{
					{
						Resource: openhue.Resource{
							Id:   "scene-01",
							Name: "Soho",
						},
						HueData: &oh.SceneGet{},
					},
				},
				GroupedLight: nil,
				HueData:      &oh.RoomGet{},
			},
		},
		Devices: []openhue.Device{},
		HueData: &oh.BridgeHomeGet{},
	}
}

func yamlHome() *openhue.HomeModel {
	h := fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Desk"), fake.NewFakeLight("Ceiling")},
		fake.NewFakeScene("s1", "Cozy", fake.StoredAction("Desk", true, 70, 300), fake.StoredAction("Ceiling", false, 20, 400)),
		fake.NewFakeScene("s2", "Cozy", fake.StoredAction("Desk", true, 10, 200)),
		fake.NewFakeScene("s3", "Solo", fake.StoredAction("Desk", true, 40, 250)))
	h.Rooms[0].Scenes[0].HueData.Speed = 0.25
	return h
}

func runGetSceneYaml(t *testing.T, args ...string) (string, error) {
	t.Helper()
	ctx, out := openhue.NewTestContext(yamlHome())
	cmd := NewCmdGetScene(ctx, &CmdGetOptions{})
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func TestGetSceneYaml(t *testing.T) {
	out, err := runGetSceneYaml(t, "Solo", "--yaml")
	require.NoError(t, err)

	def, err := openhue.ParseSceneDefinition(strings.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, "Solo", def.Name)
	assert.Equal(t, "Office", def.Room)
	require.Len(t, def.Lights, 1)
	assert.Equal(t, "Desk", def.Lights[0].Light)
	assert.Equal(t, float32(40), *def.Lights[0].Brightness)
	assert.Equal(t, 250, *def.Lights[0].Temperature)
}

func TestGetSceneYamlParsesBackToTheDefinition(t *testing.T) {
	out, err := runGetSceneYaml(t, "s1", "--yaml")
	require.NoError(t, err)

	def, err := openhue.ParseSceneDefinition(strings.NewReader(out))
	require.NoError(t, err)
	home := yamlHome()
	want := openhue.SceneToDefinition(&home.Rooms[0], &home.Rooms[0].Scenes[0])
	assert.Equal(t, want, def)
	assert.Equal(t, float32(25), *def.Speed)
}

func TestGetSceneYamlErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"no match":      {"Nope", "--yaml"},
		"several":       {"Cozy", "--yaml"},
		"no argument":   {"--yaml"},
		"two arguments": {"Solo", "s1", "--yaml"},
	} {
		t.Run(name, func(t *testing.T) {
			out, err := runGetSceneYaml(t, args...)
			require.Error(t, err)
			assert.Empty(t, out)
		})
	}

	_, err := runGetSceneYaml(t, "Cozy", "-r", "Office", "--yaml")
	assert.Error(t, err, "still several scenes in the room")
}

func TestGetSceneYamlExcludesJson(t *testing.T) {
	ctx, out := openhue.NewTestContext(yamlHome())
	cmd := NewCmdGetScene(ctx, &CmdGetOptions{Json: true})
	cmd.SetArgs([]string{"Solo", "--yaml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	assert.Error(t, cmd.Execute())
	assert.Empty(t, out.String())
}
