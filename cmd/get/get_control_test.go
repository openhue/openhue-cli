package get

import (
	"bytes"
	"encoding/json"
	"openhue-cli/openhue"
	"testing"
	"time"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func controlsHome() *openhue.HomeModel {
	room := openhue.Room{
		Resource: openhue.Resource{Id: "room-1", Name: "Living Room", Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom)},
		HueData:  &oh.RoomGet{},
	}
	event := oh.ButtonGetButtonButtonReportEventShortRelease
	at := time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)
	press := &oh.ButtonGet{Id: "b1", Type: "button"}
	press.Button = &struct {
		ButtonReport *struct {
			Event   *oh.ButtonGetButtonButtonReportEvent `json:"event,omitempty"`
			Updated *time.Time                           `json:"updated,omitempty"`
		} `json:"button_report,omitempty"`
		EventValues    *[]oh.ButtonGetButtonEventValues `json:"event_values,omitempty"`
		LastEvent      *oh.ButtonGetButtonLastEvent     `json:"last_event,omitempty"`
		RepeatInterval *int                             `json:"repeat_interval,omitempty"`
	}{ButtonReport: &struct {
		Event   *oh.ButtonGetButtonButtonReportEvent `json:"event,omitempty"`
		Updated *time.Time                           `json:"updated,omitempty"`
	}{Event: &event, Updated: &at}}

	dial := openhue.Resource{Id: "d1", Name: "Dial", Parent: &room.Resource}
	switchRes := openhue.Resource{Id: "d2", Name: "Hall switch"}
	room.Devices = []openhue.Device{{
		Resource: dial,
		Controls: []openhue.Control{
			{Resource: openhue.Resource{Id: "b1", Parent: &dial}, Kind: openhue.ControlKindButton, ControlId: 1, Button: press},
			{Resource: openhue.Resource{Id: "r1", Parent: &dial}, Kind: openhue.ControlKindRotary, ControlId: 1, Rotary: &oh.RelativeRotaryGet{Id: "r1", Type: "relative_rotary"}},
		},
	}}

	return &openhue.HomeModel{
		Rooms: []openhue.Room{room},
		Devices: []openhue.Device{{
			Resource: switchRes,
			Controls: []openhue.Control{
				{Resource: openhue.Resource{Id: "b2", Parent: &switchRes}, Kind: openhue.ControlKindButton, ControlId: 1, Button: &oh.ButtonGet{Id: "b2", Type: "button"}},
			},
		}},
		HueData: &oh.BridgeHomeGet{},
	}
}

func runGetControl(t *testing.T, home *openhue.HomeModel, args ...string) (string, error) {
	t.Helper()
	ctx, out := openhue.NewTestContext(home)
	cmd := NewCmdGetControl(ctx, &CmdGetOptions{})
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func runGetControlJson(t *testing.T, home *openhue.HomeModel, args ...string) []map[string]any {
	t.Helper()
	ctx, out := openhue.NewTestContext(home)
	cmd := NewCmdGetControl(ctx, &CmdGetOptions{Json: true})
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.Execute())
	var got []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	return got
}

func TestGetControlListAll(t *testing.T) {
	out, err := runGetControl(t, controlsHome())
	require.NoError(t, err)
	assert.Contains(t, out, "b1")
	assert.Contains(t, out, "r1")
	assert.Contains(t, out, "b2")
	assert.Contains(t, out, "short_release")
	assert.Contains(t, out, "Living Room")
	assert.Contains(t, out, "Hall switch", "a device in no room is listed with an empty room")
}

func TestGetControlFilters(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"room", []string{"--room", "Living Room"}, []string{"b1", "r1"}},
		{"device", []string{"--device", "Hall switch"}, []string{"b2"}},
		{"device by id", []string{"-d", "d1"}, []string{"b1", "r1"}},
		{"kind", []string{"--kind", "rotary"}, []string{"r1"}},
		{"room and kind", []string{"-r", "room-1", "--kind", "button"}, []string{"b1"}},
		{"id", []string{"b2"}, []string{"b2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runGetControlJson(t, controlsHome(), tt.args...)
			var ids []string
			for _, c := range got {
				ids = append(ids, c["id"].(string))
			}
			assert.ElementsMatch(t, tt.want, ids)
		})
	}
}

func TestGetControlUnknownId(t *testing.T) {
	_, err := runGetControl(t, controlsHome(), "nope")
	assert.ErrorContains(t, err, "not found")
}

func TestGetControlInvalidKind(t *testing.T) {
	_, err := runGetControl(t, controlsHome(), "--kind", "slider")
	assert.ErrorContains(t, err, "button, rotary")
}

func TestGetControlEmptyHome(t *testing.T) {
	home := &openhue.HomeModel{HueData: &oh.BridgeHomeGet{}}

	out, err := runGetControl(t, home)
	require.NoError(t, err)
	assert.Contains(t, out, "ID")

	got := runGetControlJson(t, home)
	assert.Empty(t, got)
	ctx, jsonOut := openhue.NewTestContext(home)
	cmd := NewCmdGetControl(ctx, &CmdGetOptions{Json: true})
	require.NoError(t, cmd.Execute())
	assert.JSONEq(t, "[]", jsonOut.String())
}

func TestGetControlJsonIsRawResource(t *testing.T) {
	got := runGetControlJson(t, controlsHome(), "--kind", "button", "--room", "room-1")
	require.Len(t, got, 1, "a single result is still an array")
	assert.Equal(t, "button", got[0]["type"])
	assert.Contains(t, got[0], "button")
}
