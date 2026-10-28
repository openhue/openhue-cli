package set

import (
	"bytes"
	"errors"
	"net/http"
	"openhue-cli/openhue"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func controlsHome() *openhue.HomeModel {
	dimmer := openhue.Resource{Id: "d1", Name: "Dimmer"}
	dial := openhue.Resource{Id: "d2", Name: "Dial"}
	button := func(id string, parent *openhue.Resource, n int) openhue.Control {
		return openhue.Control{Resource: openhue.Resource{Id: id, Name: parent.Name + " button", Parent: parent}, Kind: openhue.ControlKindButton, ControlId: n}
	}
	return &openhue.HomeModel{
		Devices: []openhue.Device{
			{Resource: dimmer, Controls: []openhue.Control{button("b1", &dimmer, 1), button("b2", &dimmer, 2)}},
			{Resource: dial, Controls: []openhue.Control{
				{Resource: openhue.Resource{Id: "r1", Name: "Dial rotary", Parent: &dial}, Kind: openhue.ControlKindRotary, ControlId: 1},
			}},
		},
		HueData: &oh.BridgeHomeGet{},
	}
}

func runSetControl(t *testing.T, m *oh.ClientWithResponsesMock, h *oh.Home, args ...string) (string, error) {
	t.Helper()
	ctx, out := openhue.NewTestContext(controlsHome())
	ctx.H = h
	cmd := NewCmdSetControl(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func expectUpdate(m *oh.ClientWithResponsesMock, id string, status int) {
	m.On("UpdateButtonWithResponse", mock.Anything, id, mock.Anything, mock.Anything).
		Return(&oh.UpdateButtonResponse{HTTPResponse: &http.Response{StatusCode: status}}, nil)
}

func sentInterval(t *testing.T, m *oh.ClientWithResponsesMock) int {
	t.Helper()
	require.Len(t, m.Calls, 1)
	body, ok := m.Calls[0].Arguments.Get(2).(oh.ButtonPut)
	require.True(t, ok)
	require.NotNil(t, body.Button)
	require.NotNil(t, body.Button.RepeatInterval)
	return *body.Button.RepeatInterval
}

func TestSetControlById(t *testing.T) {
	h, m := oh.NewTestHome()
	expectUpdate(m, "b1", 200)

	out, err := runSetControl(t, m, h, "b1", "--repeat-interval", "800ms")
	require.NoError(t, err)
	assert.Equal(t, 800, sentInterval(t, m))
	assert.Contains(t, out, "b1")
	assert.Contains(t, out, "800")
}

func TestSetControlByDeviceAndNumber(t *testing.T) {
	h, m := oh.NewTestHome()
	expectUpdate(m, "b2", 200)

	_, err := runSetControl(t, m, h, "Dimmer", "--control", "2", "--repeat-interval", "1s")
	require.NoError(t, err)
	assert.Equal(t, 1000, sentInterval(t, m))
}

func TestSetControlPlainMilliseconds(t *testing.T) {
	h, m := oh.NewTestHome()
	expectUpdate(m, "b1", 200)

	_, err := runSetControl(t, m, h, "b1", "--repeat-interval", "250")
	require.NoError(t, err)
	assert.Equal(t, 250, sentInterval(t, m))
}

func TestSetControlRejectsWithoutRequest(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"ambiguous device", []string{"Dimmer", "--repeat-interval", "1s"}, "b1"},
		{"unknown control", []string{"nope", "--repeat-interval", "1s"}, "no control found"},
		{"unknown control number", []string{"Dimmer", "--control", "7", "--repeat-interval", "1s"}, "no control found"},
		{"zero interval", []string{"b1", "--repeat-interval", "0"}, "greater than 0"},
		{"negative interval", []string{"b1", "--repeat-interval", "-5s"}, "greater than 0"},
		{"sub-millisecond interval", []string{"b1", "--repeat-interval", "100us"}, "greater than 0"},
		{"not a duration", []string{"b1", "--repeat-interval", "soon"}, "invalid --repeat-interval"},
		{"rotary", []string{"r1", "--repeat-interval", "1s"}, "no configurable settings"},
		{"rotary by device", []string{"Dial", "--repeat-interval", "1s"}, "no configurable settings"},
		{"no flag", []string{"b1"}, "--repeat-interval"},
		{"bad control number", []string{"Dimmer", "--control", "0", "--repeat-interval", "1s"}, "--control"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m := oh.NewTestHome()
			_, err := runSetControl(t, m, h, tt.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, m.Calls, "no request must be sent")
		})
	}
}

func TestSetControlBridgeError(t *testing.T) {
	h, m := oh.NewTestHome()
	m.On("UpdateButtonWithResponse", mock.Anything, "b1", mock.Anything, mock.Anything).
		Return(&oh.UpdateButtonResponse{}, errors.New("bridge unreachable"))

	_, err := runSetControl(t, m, h, "b1", "--repeat-interval", "1s")
	assert.ErrorContains(t, err, "bridge unreachable")
}

func TestSetControlBridgeRejects(t *testing.T) {
	h, m := oh.NewTestHome()
	expectUpdate(m, "b1", 400)

	_, err := runSetControl(t, m, h, "b1", "--repeat-interval", "1s")
	assert.Error(t, err)
}
