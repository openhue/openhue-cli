package set

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"openhue-cli/openhue"
	"testing"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const behaviorConfig = `{
	"device":{"rid":"d1","rtype":"device"},
	"buttons":{
		"b1":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
			"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],"with_off":{"enabled":false}}},
			"on_long_press":{"action":"do_nothing"},"on_repeat":{"action":"dim_up"}},
		"b2":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
			"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}],"with_off":{"enabled":false}}}}}}`

// behaviorHome has a dimmer d1 (buttons b1, b2, configured), a switch d3 with no configuration (button b9), and the
// rotary r1 of a dial. Room "Living" has the scenes Cozy, Soho and two scenes named Twin; room "Kitchen" has Bright.
func behaviorHome(t *testing.T) *openhue.HomeModel {
	t.Helper()
	var cfg map[string]any
	require.NoError(t, json.Unmarshal([]byte(behaviorConfig), &cfg))
	instance := &openhue.BehaviorInstance{Id: "bi1", Configuration: cfg}

	living := openhue.Resource{Id: "room-1", Name: "Living", Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom)}
	scene := func(id, name string) openhue.Scene {
		return openhue.Scene{Resource: openhue.Resource{Id: id, Name: name, Parent: &living}}
	}
	dimmer := openhue.Resource{Id: "d1", Name: "Dimmer", Parent: &living}
	plain := openhue.Resource{Id: "d3", Name: "Plain", Parent: &living}
	button := func(id string, parent *openhue.Resource, n int, b *openhue.BehaviorInstance) openhue.Control {
		return openhue.Control{Resource: openhue.Resource{Id: id, Name: parent.Name + " button", Parent: parent}, Kind: openhue.ControlKindButton, ControlId: n, Behavior: b}
	}
	rotary := openhue.Control{Resource: openhue.Resource{Id: "r1", Name: "Dimmer rotary", Parent: &dimmer}, Kind: openhue.ControlKindRotary, ControlId: 3, Behavior: instance}

	return &openhue.HomeModel{
		Rooms: []openhue.Room{
			{
				Resource: living,
				Scenes:   []openhue.Scene{scene("s1", "Cozy"), scene("s2", "Soho"), scene("s3", "Twin"), scene("s4", "Twin")},
				Devices: []openhue.Device{
					{Resource: dimmer, Controls: []openhue.Control{button("b1", &dimmer, 1, instance), button("b2", &dimmer, 2, instance), rotary}},
					{Resource: plain, Controls: []openhue.Control{button("b9", &plain, 1, nil)}},
				},
			},
			{
				Resource: openhue.Resource{Id: "room-2", Name: "Kitchen", Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom)},
				Scenes:   []openhue.Scene{{Resource: openhue.Resource{Id: "s5", Name: "Bright"}}},
			},
		},
		HueData: &oh.BridgeHomeGet{},
	}
}

func runSetBehavior(t *testing.T, m *oh.ClientWithResponsesMock, h *oh.Home, args ...string) (string, error) {
	t.Helper()
	ctx, out := openhue.NewTestContext(behaviorHome(t))
	ctx.H = h
	cmd := NewCmdSetControl(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func expectBehaviorUpdate(m *oh.ClientWithResponsesMock, status int) {
	m.On("UpdateBehaviorInstanceWithResponse", mock.Anything, "bi1", mock.Anything, mock.Anything).
		Return(&oh.UpdateBehaviorInstanceResponse{HTTPResponse: &http.Response{StatusCode: status}}, nil)
}

// sentButton returns the configuration of button id in the single behavior update the mock received.
func sentButton(t *testing.T, m *oh.ClientWithResponsesMock, id string) map[string]any {
	t.Helper()
	require.Len(t, m.Calls, 1)
	body, ok := m.Calls[0].Arguments.Get(2).(oh.BehaviorInstancePut)
	require.True(t, ok)
	require.NotNil(t, body.Configuration)
	return (*body.Configuration)["buttons"].(map[string]any)[id].(map[string]any)
}

func asJson(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}

func TestSetControlSingleScene(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 200)

	out, err := runSetBehavior(t, m, h, "Dimmer", "--control", "1", "--scene", "Soho", "--with-off")
	require.NoError(t, err)

	b := sentButton(t, m, "b1")
	assert.JSONEq(t, `{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}],"with_off":{"enabled":true}}}`, asJson(t, b["on_short_release"]))
	assert.Contains(t, b, "on_long_press", "long press untouched")
	assert.JSONEq(t, `{"action":"dim_up"}`, asJson(t, b["on_repeat"]), "hold untouched")
	assert.Contains(t, out, "scene Soho")
}

func TestSetControlCycle(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 200)

	out, err := runSetBehavior(t, m, h, "b2", "--scene", "Cozy", "--scene", "Soho", "--scene", "s1", "--repeat-timeout", "10s")
	require.NoError(t, err)

	b := sentButton(t, m, "b2")
	assert.JSONEq(t, `{"scene_cycle_extended":{"slots":[
		[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],
		[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}],
		[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}]],
		"repeat_timeout":{"seconds":10},"with_off":{"enabled":false}}}`, asJson(t, b["on_short_release"]))
	assert.Contains(t, out, "cycle: Cozy, Soho, Cozy (3)")
}

func TestSetControlNeighbourUntouched(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 200)

	_, err := runSetBehavior(t, m, h, "b1", "--scene", "Soho")
	require.NoError(t, err)

	body := m.Calls[0].Arguments.Get(2).(oh.BehaviorInstancePut)
	var original map[string]any
	require.NoError(t, json.Unmarshal([]byte(behaviorConfig), &original))
	assert.Equal(t, asJson(t, original["buttons"].(map[string]any)["b2"]), asJson(t, (*body.Configuration)["buttons"].(map[string]any)["b2"]))
	assert.Equal(t, asJson(t, original["device"]), asJson(t, (*body.Configuration)["device"]))
}

func TestSetControlHoldAndClear(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 200)
	_, err := runSetBehavior(t, m, h, "b1", "--on-hold", "dim-alternate")
	require.NoError(t, err)
	b := sentButton(t, m, "b1")
	assert.JSONEq(t, `{"action":"dim_alternate"}`, asJson(t, b["on_repeat"]))
	assert.Contains(t, b, "on_short_release", "short press unchanged")

	h, m = oh.NewTestHome()
	expectBehaviorUpdate(m, 200)
	_, err = runSetBehavior(t, m, h, "b1", "--clear", "--on-hold", "none")
	require.NoError(t, err)
	b = sentButton(t, m, "b1")
	assert.NotContains(t, b, "on_short_release")
	assert.NotContains(t, b, "on_repeat")
}

func TestSetControlDryRun(t *testing.T) {
	h, m := oh.NewTestHome()

	out, err := runSetBehavior(t, m, h, "b1", "--scene", "Soho", "--dry-run")
	require.NoError(t, err)
	assert.Empty(t, m.Calls, "nothing is sent")

	var printed struct {
		Configuration map[string]any `json:"configuration"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &printed))
	assert.Contains(t, asJson(t, printed.Configuration), `"rid":"s2"`)
}

func TestSetControlBehaviorRejectsWithoutRequest(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"scene of another room", []string{"b1", "--scene", "Bright"}, "Kitchen"},
		{"unknown scene", []string{"b1", "--scene", "Nope"}, "not found"},
		{"ambiguous scene", []string{"b1", "--scene", "Twin"}, "use a scene ID"},
		{"second scene unknown", []string{"b1", "--scene", "Cozy", "--scene", "Nope"}, "not found"},
		{"clear and scene", []string{"b1", "--clear", "--scene", "Cozy"}, "--clear"},
		{"bad hold", []string{"b1", "--on-hold", "sideways"}, "--on-hold"},
		{"empty hold", []string{"b1", "--on-hold", ""}, "--on-hold"},
		{"bad timeout", []string{"b1", "--scene", "Cozy", "--scene", "Soho", "--repeat-timeout", "soon"}, "--repeat-timeout"},
		{"timeout on single scene", []string{"b1", "--scene", "Cozy", "--repeat-timeout", "5s"}, "several scenes"},
		{"rotary", []string{"r1", "--scene", "Cozy"}, "no configurable settings"},
		{"unconfigured accessory", []string{"b9", "--scene", "Cozy"}, "no behavior configuration"},
		{"unconfigured accessory dry run", []string{"b9", "--on-hold", "dim-up", "--dry-run"}, "no behavior configuration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m := oh.NewTestHome()
			_, err := runSetBehavior(t, m, h, tt.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, m.Calls, "no request must be sent")
		})
	}
}

func TestSetControlBehaviorBridgeError(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 400)
	_, err := runSetBehavior(t, m, h, "b1", "--scene", "Soho")
	assert.Error(t, err)

	h, m = oh.NewTestHome()
	m.On("UpdateBehaviorInstanceWithResponse", mock.Anything, "bi1", mock.Anything, mock.Anything).
		Return(&oh.UpdateBehaviorInstanceResponse{}, errors.New("bridge unreachable"))
	_, err = runSetBehavior(t, m, h, "b1", "--scene", "Soho")
	assert.ErrorContains(t, err, "bridge unreachable")
}

func TestSetControlIntervalAndBehaviorTogether(t *testing.T) {
	h, m := oh.NewTestHome()
	expectBehaviorUpdate(m, 200)
	expectUpdate(m, "b1", 200)

	_, err := runSetBehavior(t, m, h, "b1", "--repeat-interval", "1s", "--on-hold", "dim-down")
	require.NoError(t, err)
	assert.Len(t, m.Calls, 2)
}
