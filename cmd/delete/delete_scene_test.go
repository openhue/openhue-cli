package delete

import (
	"bytes"
	"net/http"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/fake"
)

func home() *openhue.HomeModel {
	h := fake.NewFakeHome([]fake.FakeLight{fake.NewFakeLight("Desk")},
		fake.NewFakeScene("s1", "Cozy"), fake.NewFakeScene("s2", "Solo"))
	h.Rooms = append(h.Rooms, openhue.Room{
		Resource: openhue.Resource{Id: "room-2", Name: "Bedroom"},
		Scenes:   []openhue.Scene{fake.NewFakeScene("s3", "Cozy")},
	})
	return h
}

func run(t *testing.T, h *oh.Home, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	streams, _, out, _ := openhue.NewTestIOStreams()
	ctx := openhue.NewContext(streams, openhue.NewTestBuildInfo(), h, nil)
	ctx.Home = home()
	cmd := NewCmdDeleteScene(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	return out, cmd.Execute()
}

func TestDeleteByNameAndId(t *testing.T) {
	h, m := fake.NewTestHomeWithHandlers()
	out, err := run(t, h, "Solo")
	require.NoError(t, err)
	m.AssertCalled(t, "DeleteSceneWithResponse", mock.Anything, "s2", mock.Anything)
	assert.Contains(t, out.String(), "Scene 'Solo' deleted from room 'Office' (s2)")

	h, m = fake.NewTestHomeWithHandlers()
	_, err = run(t, h, "s1")
	require.NoError(t, err)
	m.AssertCalled(t, "DeleteSceneWithResponse", mock.Anything, "s1", mock.Anything)
}

func TestDeleteWithRoom(t *testing.T) {
	h, m := fake.NewTestHomeWithHandlers()
	_, err := run(t, h, "Cozy", "-r", "Bedroom")
	require.NoError(t, err)
	m.AssertCalled(t, "DeleteSceneWithResponse", mock.Anything, "s3", mock.Anything)
}

func TestDeleteNotFoundAndAmbiguous(t *testing.T) {
	h, m := fake.NewTestHomeWithHandlers()

	_, err := run(t, h, "Nope")
	assert.ErrorContains(t, err, "not found")

	_, err = run(t, h, "Cozy")
	assert.ErrorContains(t, err, "--room")

	_, err = run(t, h)
	assert.Error(t, err)

	m.AssertNotCalled(t, "DeleteSceneWithResponse", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteBridgeError(t *testing.T) {
	h, m := oh.NewTestHome()
	m.On("DeleteSceneWithResponse", mock.Anything, mock.Anything, mock.Anything).
		Return(&oh.DeleteSceneResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError, Status: "500 Internal Server Error"}}, nil)
	out, err := run(t, h, "Solo")
	require.Error(t, err)
	assert.ErrorContains(t, err, "500")
	assert.Empty(t, out.String())
}
