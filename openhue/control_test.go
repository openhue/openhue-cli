package openhue

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// loadControlsHome loads a home through the openhue-go mock:
//   - room "Living" holding a dimmer switch (d1, 4 buttons, last one unreported) and a tap dial (d2, 1 button, 1 rotary)
//   - the switch "Hall switch" (d3) in no room, with one button
//
// When controlsFail is set, the buttons endpoint returns an error.
func loadControlsHome(t *testing.T, controlsFail bool) *HomeModel {
	t.Helper()
	h, m := openhue.NewTestHome()
	st := &http.Response{StatusCode: 200}

	set := func(method string, resp any, jsonData string) {
		// the generated response structs have no JSON tags, so the field name drives the decoding
		require.NoError(t, json.Unmarshal([]byte(`{"JSON200":{"data":`+jsonData+`}}`), resp))
		m.On(method, mock.Anything, mock.Anything).Return(resp, nil).Maybe()
	}

	home := &openhue.GetBridgeHomesResponse{HTTPResponse: st}
	set("GetBridgeHomesWithResponse", home, `[{"id":"home","type":"bridge_home","children":[{"rid":"room-1","rtype":"room"},{"rid":"d3","rtype":"device"}]}]`)
	rooms := &openhue.GetRoomsResponse{HTTPResponse: st}
	set("GetRoomsWithResponse", rooms, `[{"id":"room-1","type":"room","metadata":{"name":"Living","archetype":"living_room"},"children":[{"rid":"d1","rtype":"device"},{"rid":"d2","rtype":"device"}],"services":[{"rid":"gl1","rtype":"grouped_light"}]}]`)
	devices := &openhue.GetDevicesResponse{HTTPResponse: st}
	set("GetDevicesWithResponse", devices, `[
		{"id":"d1","type":"device","metadata":{"name":"Dimmer","archetype":"unknown_archetype"},"services":[]},
		{"id":"d2","type":"device","metadata":{"name":"Dial","archetype":"unknown_archetype"},"services":[]},
		{"id":"d3","type":"device","metadata":{"name":"Hall switch","archetype":"unknown_archetype"},"services":[]},
		{"id":"d4","type":"device","metadata":{"name":"Plain","archetype":"unknown_archetype"},"services":[]}]`)
	set("GetLightsWithResponse", &openhue.GetLightsResponse{HTTPResponse: st}, `[]`)
	set("GetGroupedLightsWithResponse", &openhue.GetGroupedLightsResponse{HTTPResponse: st}, `[{"id":"gl1","type":"grouped_light"}]`)
	set("GetScenesWithResponse", &openhue.GetScenesResponse{HTTPResponse: st}, `[]`)

	if controlsFail {
		m.On("GetButtonsWithResponse", mock.Anything, mock.Anything).Return(&openhue.GetButtonsResponse{HTTPResponse: &http.Response{StatusCode: 500}}, nil)
		m.On("GetRelativeRotariesWithResponse", mock.Anything, mock.Anything).Return((*openhue.GetRelativeRotariesResponse)(nil), errors.New("boom"))
	} else {
		set("GetButtonsWithResponse", &openhue.GetButtonsResponse{HTTPResponse: st}, `[
			{"id":"b1","type":"button","owner":{"rid":"d1","rtype":"device"},"metadata":{"control_id":1},"button":{"button_report":{"event":"short_release","updated":"2024-05-01T10:00:00Z"}}},
			{"id":"b2","type":"button","owner":{"rid":"d1","rtype":"device"},"metadata":{"control_id":2},"button":{"last_event":"long_press"}},
			{"id":"b3","type":"button","owner":{"rid":"d1","rtype":"device"},"metadata":{"control_id":3}},
			{"id":"b4","type":"button","owner":{"rid":"d1","rtype":"device"},"metadata":{"control_id":4}},
			{"id":"b5","type":"button","owner":{"rid":"d2","rtype":"device"},"metadata":{"control_id":1}},
			{"id":"b6","type":"button","owner":{"rid":"d3","rtype":"device"},"metadata":{"control_id":1}}]`)
		set("GetRelativeRotariesWithResponse", &openhue.GetRelativeRotariesResponse{HTTPResponse: st}, `[
			{"id":"r1","type":"relative_rotary","owner":{"rid":"d2","rtype":"device"},"metadata":{"control_id":1},"relative_rotary":{"rotary_report":{"action":"start","rotation":{"direction":"clock_wise","steps":3},"updated":"2024-05-01T11:00:00Z"}}}]`)
	}

	model, err := LoadHome(h)
	require.NoError(t, err)
	return model
}

func controlIds(controls []Control) []string {
	var ids []string
	for _, c := range controls {
		ids = append(ids, c.Id)
	}
	return ids
}

func TestLoadHomeAttachesControls(t *testing.T) {
	home := loadControlsHome(t, false)

	dimmer := home.Rooms[0].Devices[0]
	require.Len(t, dimmer.Controls, 4)
	assert.Equal(t, ControlKindButton, dimmer.Controls[0].Kind)
	assert.Equal(t, 1, dimmer.Controls[0].ControlId)
	assert.Equal(t, "Living", dimmer.Controls[0].Room().Name)
	assert.Equal(t, "Dimmer", dimmer.Controls[0].Device().Name)

	dial := home.Rooms[0].Devices[1]
	require.Len(t, dial.Controls, 2)
	assert.Equal(t, ControlKindButton, dial.Controls[0].Kind)
	assert.Equal(t, ControlKindRotary, dial.Controls[1].Kind)

	// a device in no room has controls but no room
	require.Len(t, home.Devices, 1)
	require.Len(t, home.Devices[0].Controls, 1)
	assert.Nil(t, home.Devices[0].Controls[0].Room())
}

func TestControlLastEvent(t *testing.T) {
	home := loadControlsHome(t, false)
	dimmer := home.Rooms[0].Devices[0].Controls

	ev, at := dimmer[0].LastEvent()
	assert.Equal(t, "short_release", ev)
	require.NotNil(t, at)

	// deprecated last_event is the fallback
	ev, at = dimmer[1].LastEvent()
	assert.Equal(t, "long_press", ev)
	assert.Nil(t, at)

	// never reported
	ev, at = dimmer[2].LastEvent()
	assert.Empty(t, ev)
	assert.Nil(t, at)

	ev, at = home.Rooms[0].Devices[1].Controls[1].LastEvent()
	assert.Equal(t, "start clock_wise 3", ev)
	assert.NotNil(t, at)
}

func TestLoadHomeToleratesControlsFailure(t *testing.T) {
	home := loadControlsHome(t, true)
	assert.Len(t, home.Rooms, 1)
	assert.Empty(t, home.Rooms[0].Devices[0].Controls)
	assert.Empty(t, SearchControls(home, "", "", "", nil))
}

func TestDeviceWithoutControls(t *testing.T) {
	home := loadControlsHome(t, false)
	assert.Empty(t, SearchControls(home, "", "Plain", "", nil))
}

func TestSearchControls(t *testing.T) {
	home := loadControlsHome(t, false)

	tests := []struct {
		name         string
		room, device string
		kind         ControlKind
		ids          []string
		want         []string
	}{
		{"all", "", "", "", nil, []string{"b1", "b2", "b3", "b4", "b5", "r1", "b6"}},
		{"room", "Living", "", "", nil, []string{"b1", "b2", "b3", "b4", "b5", "r1"}},
		{"room by id", "room-1", "", ControlKindRotary, nil, []string{"r1"}},
		{"device", "", "Dial", "", nil, []string{"b5", "r1"}},
		{"device by id", "", "d3", "", nil, []string{"b6"}},
		{"kind", "", "", ControlKindRotary, nil, []string{"r1"}},
		{"ids", "", "", "", []string{"b2", "r1"}, []string{"b2", "r1"}},
		{"unknown id", "", "", "", []string{"nope"}, nil},
		{"unknown room", "Kitchen", "", "", nil, nil},
		{"room and device combined", "Living", "Hall switch", "", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SearchControls(home, tt.room, tt.device, tt.kind, tt.ids)
			assert.NotNil(t, got)
			assert.ElementsMatch(t, tt.want, controlIds(got))
		})
	}
}

func TestFindControl(t *testing.T) {
	home := loadControlsHome(t, false)

	c, err := FindControl(home, "b3", 0)
	require.NoError(t, err)
	assert.Equal(t, "b3", c.Id)

	c, err = FindControl(home, "Dimmer", 2)
	require.NoError(t, err)
	assert.Equal(t, "b2", c.Id)

	c, err = FindControl(home, "Hall switch", 0)
	require.NoError(t, err, "a device with a single control needs no control number")
	assert.Equal(t, "b6", c.Id)

	// ambiguous: four buttons, no control number
	_, err = FindControl(home, "Dimmer", 0)
	require.Error(t, err)
	for _, id := range []string{"b1", "b2", "b3", "b4"} {
		assert.Contains(t, err.Error(), id)
	}

	// control number that does not exist, unknown device
	_, err = FindControl(home, "Dimmer", 9)
	assert.ErrorContains(t, err, "no control found")
	_, err = FindControl(home, "Nope", 0)
	assert.ErrorContains(t, err, "no control found")
}

func TestControlJson(t *testing.T) {
	home := loadControlsHome(t, false)
	out, err := json.Marshal(SearchControls(home, "", "", "", []string{"b1", "r1"}))
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{"id":"b1","type":"button","owner":{"rid":"d1","rtype":"device"},"metadata":{"control_id":1},"button":{"button_report":{"event":"short_release","updated":"2024-05-01T10:00:00Z"}}},
		{"id":"r1","type":"relative_rotary","owner":{"rid":"d2","rtype":"device"},"metadata":{"control_id":1},"relative_rotary":{"rotary_report":{"action":"start","rotation":{"direction":"clock_wise","steps":3},"updated":"2024-05-01T11:00:00Z"}}}]`, string(out))
}
