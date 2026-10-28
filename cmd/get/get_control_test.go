package get

import (
	"bytes"
	"encoding/json"
	"openhue-cli/openhue"
	"strings"
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
	cmd.SetArgs(append([]string{}, args...))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func runGetControlJson(t *testing.T, home *openhue.HomeModel, args ...string) []map[string]any {
	t.Helper()
	ctx, out := openhue.NewTestContext(home)
	cmd := NewCmdGetControl(ctx, &CmdGetOptions{Json: true})
	cmd.SetArgs(append([]string{}, args...))
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
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())
	assert.JSONEq(t, "[]", jsonOut.String())
}

func TestGetControlJsonIsRawResource(t *testing.T) {
	got := runGetControlJson(t, controlsHome(), "--kind", "button", "--room", "room-1")
	require.Len(t, got, 1, "a single result is still an array")
	assert.Equal(t, "button", got[0]["type"])
	assert.Contains(t, got[0], "button")
}

func TestGetControlShowsButtonActions(t *testing.T) {
	var cfg map[string]any
	require.NoError(t, json.Unmarshal([]byte(`{
		"device":{"rid":"d1","rtype":"device"},
		"buttons":{
			"b1":{"on_short_release":{"scene_cycle_extended":{"slots":[[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}]]}},
				"on_long_press":{"action":"do_nothing"},"on_repeat":{"action":"dim_up"}}}}`), &cfg))

	home := controlsHome()
	home.Rooms[0].Devices[0].Controls[0].Behavior = &openhue.BehaviorInstance{Id: "bi1", Configuration: cfg}

	out, err := runGetControl(t, home)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Contains(t, lines[0], "Short press")
	assert.Contains(t, lines[0], "Hold")
	assert.Regexp(t, `b1\s.*cycle: <unknown scene>, <unknown scene> \(2\)\s+dim-up`, out)
	assert.Regexp(t, `b2\s.*\s-\s+-\s`, out, "unconfigured accessory shows dashes")

	got := runGetControlJson(t, home, "b1")
	conf, ok := got[0]["configuration"].(map[string]any)
	require.True(t, ok, "JSON carries the configuration")
	assert.Contains(t, conf, "on_long_press")
	got = runGetControlJson(t, home, "b2")
	assert.NotContains(t, got[0], "configuration")
}
