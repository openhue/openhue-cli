package openhue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const behaviorScenes = `[
	{"id":"s1","type":"scene","metadata":{"name":"Cozy"},"group":{"rid":"room-1","rtype":"room"}},
	{"id":"s2","type":"scene","metadata":{"name":"Soho"},"group":{"rid":"room-1","rtype":"room"}},
	{"id":"s3","type":"scene","metadata":{"name":"Dusk"},"group":{"rid":"room-1","rtype":"room"}},
	{"id":"s4","type":"scene","metadata":{"name":"Cozy"},"group":{"rid":"room-1","rtype":"room"}},
	{"id":"s5","type":"scene","metadata":{"name":"Elsewhere"},"group":{"rid":"room-2","rtype":"room"}}]`

// behaviorInstances configures the dimmer d1: b1 recalls a single scene and dims up on hold, b2 cycles three scenes,
// b3 is time based, b4 recalls a scene that no longer exists. The tap dial d2 has no instance.
const behaviorInstances = `[
	{"id":"bi1","type":"behavior_instance","script_id":"script","configuration":{
		"device":{"rid":"d1","rtype":"device"},
		"model_id":"RWL022",
		"buttons":{
			"b1":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
				"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],"with_off":{"enabled":true}}},
				"on_long_press":{"action":"do_nothing"},
				"on_repeat":{"action":"dim_up"},
				"future_key":{"a":[1,2,{"b":null}]}},
			"b2":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
				"on_short_release":{"scene_cycle_extended":{"slots":[[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}],[{"action":{"recall":{"rid":"s3","rtype":"scene"}}}]],"repeat_timeout":{"seconds":5},"with_off":{"enabled":false}}}},
			"b3":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
				"on_short_release":{"time_based_extended":{"slots":[{"start_time":{"hour":7,"minute":0},"actions":[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}]}],"with_off":{"enabled":false}}}},
			"b4":{"where":[{"group":{"rid":"room-1","rtype":"room"}}],
				"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"gone","rtype":"scene"}}}],"with_off":{"enabled":false}}}}}}}]`

func behaviorHome(t *testing.T) *HomeModel {
	return loadControlsHomeWith(t, false, behaviorScenes, behaviorInstances)
}

func controlById(t *testing.T, home *HomeModel, id string) *Control {
	t.Helper()
	c, err := FindControl(home, id, 0)
	require.NoError(t, err)
	return c
}

func TestLoadHomeLinksBehaviorToControls(t *testing.T) {
	home := behaviorHome(t)

	b1 := controlById(t, home, "b1")
	require.NotNil(t, b1.Behavior)
	assert.Equal(t, "bi1", b1.Behavior.Id)
	assert.NotNil(t, b1.ButtonConfig())
	assert.Equal(t, "scene Cozy", b1.ShortPressSummary())
	assert.Equal(t, "dim-up", b1.HoldSummary())

	// a button the instance does not list has no configuration
	b5 := controlById(t, home, "b5")
	assert.Nil(t, b5.Behavior, "unconfigured accessory")
	assert.Nil(t, b5.ButtonConfig())
	assert.Equal(t, "-", b5.ShortPressSummary())
	assert.Equal(t, "-", b5.HoldSummary())
}

func TestLoadHomeToleratesBehaviorFailure(t *testing.T) {
	home := loadControlsHomeWith(t, false, behaviorScenes, "fail")

	b1 := controlById(t, home, "b1")
	assert.Nil(t, b1.Behavior)
	assert.Equal(t, "-", b1.ShortPressSummary())
}

func TestControlJsonCarriesConfiguration(t *testing.T) {
	home := behaviorHome(t)

	raw, err := json.Marshal(controlById(t, home, "b1"))
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw, &doc))
	assert.Equal(t, "b1", doc["id"])
	cfg := doc["configuration"].(map[string]any)
	assert.Contains(t, cfg, "on_long_press", "parts the table does not summarize are included")
	assert.Contains(t, cfg, "future_key")

	raw, err = json.Marshal(controlById(t, home, "b5"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "configuration")
}

func TestShortPressSummaries(t *testing.T) {
	home := behaviorHome(t)
	tests := map[string]string{
		"b1": "scene Cozy",
		"b2": "cycle: Cozy, Soho, Dusk (3)",
		"b3": "time-based",
		"b4": "scene <unknown scene>",
		"b5": "-",
	}
	for id, want := range tests {
		assert.Equal(t, want, controlById(t, home, id).ShortPressSummary(), id)
	}
}

func TestSummarizeShortPressShapes(t *testing.T) {
	names := func(id string) (string, bool) { return map[string]string{"s1": "Cozy"}[id], id == "s1" }
	parse := func(s string) map[string]any {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(s), &m))
		return m
	}
	tests := []struct {
		name, button, want string
	}{
		{"nothing", `{}`, "-"},
		{"empty release", `{"on_short_release":{}}`, "-"},
		{"smart scene", `{"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"ss1","rtype":"smart_scene"}}}]}}}`, "smart-scene ss1"},
		{"two actions", `{"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s1","rtype":"scene"}}},{"action":{"recall":{"rid":"s1","rtype":"scene"}}}]}}}`, "custom"},
		{"unknown kind", `{"on_short_release":{"something_new":{}}}`, "custom"},
		{"cycle with smart scene", `{"on_short_release":{"scene_cycle_extended":{"slots":[[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],[{"action":{"recall":{"rid":"ss1","rtype":"smart_scene"}}}]]}}}`, "cycle: Cozy, smart-scene ss1 (2)"},
		{"cycle with bad slot", `{"on_short_release":{"scene_cycle_extended":{"slots":[[{"action":{"turn":"on"}}]]}}}`, "custom"},
		{"cycle without slots", `{"on_short_release":{"scene_cycle_extended":{}}}`, "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SummarizeShortPress(parse(tt.button), names))
		})
	}
	assert.Equal(t, "scene <unknown scene>", SummarizeShortPress(parse(`{"on_short_release":{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}]}}}`), nil))
}

func TestSummarizeHold(t *testing.T) {
	for in, want := range map[string]string{
		`{}`:                                       "-",
		`{"on_repeat":{}}`:                         "-",
		`{"on_repeat":{"action":"dim_up"}}`:        "dim-up",
		`{"on_repeat":{"action":"dim_down"}}`:      "dim-down",
		`{"on_repeat":{"action":"dim_alternate"}}`: "dim-alternate",
		`{"on_repeat":{"action":"do_nothing"}}`:    "-",
		`{"on_repeat":{"action":"explode"}}`:       "custom",
	} {
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(in), &m))
		assert.Equal(t, want, SummarizeHold(m), in)
	}
}

// config returns a fresh decoded copy of the instance bi1 configuration.
func config(t *testing.T) map[string]any {
	t.Helper()
	var instances []struct {
		Configuration map[string]any `json:"configuration"`
	}
	require.NoError(t, json.Unmarshal([]byte(behaviorInstances), &instances))
	return instances[0].Configuration
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}

func buttonOf(cfg map[string]any, id string) map[string]any {
	return cfg["buttons"].(map[string]any)[id].(map[string]any)
}

func TestBuildSingleScene(t *testing.T) {
	cfg := config(t)
	before := marshal(t, cfg)

	out, err := BuildButtonConfiguration(cfg, "b2", ButtonPatch{SceneIds: []string{"s3"}})
	require.NoError(t, err)

	assert.Equal(t, before, marshal(t, cfg), "the input is not modified")
	assert.JSONEq(t, `{"recall_single_extended":{"actions":[{"action":{"recall":{"rid":"s3","rtype":"scene"}}}],"with_off":{"enabled":false}}}`,
		marshal(t, buttonOf(out, "b2")["on_short_release"]), "a cycle is replaced by a single scene")
	assert.Equal(t, marshal(t, buttonOf(cfg, "b1")), marshal(t, buttonOf(out, "b1")), "neighbours are unchanged")
	assert.Equal(t, marshal(t, buttonOf(cfg, "b3")), marshal(t, buttonOf(out, "b3")))
}

func TestBuildSingleSceneWithOff(t *testing.T) {
	yes, no := true, false

	out, err := BuildButtonConfiguration(config(t), "b2", ButtonPatch{SceneIds: []string{"s3"}, WithOff: &yes})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b2")["on_short_release"]), `"with_off":{"enabled":true}`)

	// the current with_off value is kept when the flag is absent, and overridden when given
	out, err = BuildButtonConfiguration(config(t), "b1", ButtonPatch{SceneIds: []string{"s2"}})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b1")["on_short_release"]), `"with_off":{"enabled":true}`)
	out, err = BuildButtonConfiguration(config(t), "b1", ButtonPatch{SceneIds: []string{"s2"}, WithOff: &no})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b1")["on_short_release"]), `"with_off":{"enabled":false}`)
}

func TestBuildCycle(t *testing.T) {
	out, err := BuildButtonConfiguration(config(t), "b1", ButtonPatch{SceneIds: []string{"s1", "s2"}})
	require.NoError(t, err)

	assert.JSONEq(t, `{"scene_cycle_extended":{
		"slots":[[{"action":{"recall":{"rid":"s1","rtype":"scene"}}}],[{"action":{"recall":{"rid":"s2","rtype":"scene"}}}]],
		"repeat_timeout":{"seconds":3},"with_off":{"enabled":true}}}`, marshal(t, buttonOf(out, "b1")["on_short_release"]),
		"default timeout, with_off kept")

	out, err = BuildButtonConfiguration(config(t), "b1", ButtonPatch{SceneIds: []string{"s1", "s2"}, RepeatTimeout: 20 * time.Second})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b1")["on_short_release"]), `"repeat_timeout":{"seconds":20}`)

	// re-assigning a cycle keeps its timeout
	out, err = BuildButtonConfiguration(config(t), "b2", ButtonPatch{SceneIds: []string{"s3", "s1"}})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b2")["on_short_release"]), `"repeat_timeout":{"seconds":5}`)
}

func TestBuildTimeoutAndWithOffAlone(t *testing.T) {
	yes := true

	out, err := BuildButtonConfiguration(config(t), "b2", ButtonPatch{RepeatTimeout: 9 * time.Second, WithOff: &yes})
	require.NoError(t, err)
	got := marshal(t, buttonOf(out, "b2")["on_short_release"])
	assert.Contains(t, got, `"repeat_timeout":{"seconds":9}`)
	assert.Contains(t, got, `"with_off":{"enabled":true}`)
	assert.Contains(t, got, `"slots"`, "the slots are kept")

	out, err = BuildButtonConfiguration(config(t), "b1", ButtonPatch{WithOff: new(bool)})
	require.NoError(t, err)
	assert.Contains(t, marshal(t, buttonOf(out, "b1")["on_short_release"]), `"with_off":{"enabled":false}`)

	_, err = BuildButtonConfiguration(config(t), "b1", ButtonPatch{RepeatTimeout: 9 * time.Second})
	assert.ErrorContains(t, err, "cycling")
	_, err = BuildButtonConfiguration(config(t), "b3", ButtonPatch{WithOff: &yes})
	assert.ErrorContains(t, err, "--scene", "time-based assignments are not edited implicitly")
}

func TestBuildHoldAndClear(t *testing.T) {
	cfg := config(t)

	out, err := BuildButtonConfiguration(cfg, "b2", ButtonPatch{OnHold: "dim-alternate"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"action":"dim_alternate"}`, marshal(t, buttonOf(out, "b2")["on_repeat"]))
	assert.Equal(t, marshal(t, buttonOf(cfg, "b2")["on_short_release"]), marshal(t, buttonOf(out, "b2")["on_short_release"]), "short press unchanged")

	out, err = BuildButtonConfiguration(cfg, "b1", ButtonPatch{OnHold: "none"})
	require.NoError(t, err)
	assert.NotContains(t, buttonOf(out, "b1"), "on_repeat")

	out, err = BuildButtonConfiguration(cfg, "b1", ButtonPatch{Clear: true})
	require.NoError(t, err)
	assert.NotContains(t, buttonOf(out, "b1"), "on_short_release")
	assert.Contains(t, buttonOf(out, "b1"), "on_repeat", "hold is kept")

	out, err = BuildButtonConfiguration(cfg, "b3", ButtonPatch{Clear: true})
	require.NoError(t, err)
	assert.NotContains(t, buttonOf(out, "b3"), "on_short_release", "an explicit clear removes a time-based assignment")
}

func TestBuildPreservesUnknownKeys(t *testing.T) {
	cfg := config(t)
	out, err := BuildButtonConfiguration(cfg, "b1", ButtonPatch{SceneIds: []string{"s2"}, OnHold: "dim-down"})
	require.NoError(t, err)

	b1 := buttonOf(out, "b1")
	assert.Equal(t, marshal(t, buttonOf(cfg, "b1")["future_key"]), marshal(t, b1["future_key"]))
	assert.Equal(t, marshal(t, buttonOf(cfg, "b1")["on_long_press"]), marshal(t, b1["on_long_press"]))
	assert.Equal(t, marshal(t, buttonOf(cfg, "b1")["where"]), marshal(t, b1["where"]))
	assert.Equal(t, "RWL022", out["model_id"])

	// an empty patch is a byte-identical round trip
	same, err := BuildButtonConfiguration(cfg, "b1", ButtonPatch{})
	require.NoError(t, err)
	assert.Equal(t, marshal(t, cfg), marshal(t, same))
}

func TestBuildRejects(t *testing.T) {
	tests := []struct {
		name    string
		button  string
		patch   ButtonPatch
		wantErr string
	}{
		{"clear and scene", "b1", ButtonPatch{Clear: true, SceneIds: []string{"s1"}}, "--clear"},
		{"bad hold", "b1", ButtonPatch{OnHold: "dim-sideways"}, "--on-hold"},
		{"short timeout", "b2", ButtonPatch{RepeatTimeout: time.Millisecond}, "at least 1s"},
		{"timeout on single scene", "b1", ButtonPatch{SceneIds: []string{"s1"}, RepeatTimeout: 5 * time.Second}, "several scenes"},
		{"unknown button", "nope", ButtonPatch{Clear: true}, "no configuration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildButtonConfiguration(config(t), tt.button, tt.patch)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestControlScenes(t *testing.T) {
	home := behaviorHome(t)
	b1 := controlById(t, home, "b1")

	ids, err := ControlScenes(home, b1, []string{"Soho", "s3"})
	require.NoError(t, err)
	assert.Equal(t, []string{"s2", "s3"}, ids)

	home.Rooms = append(home.Rooms, Room{
		Resource: Resource{Id: "room-2", Name: "Kitchen"},
		Scenes:   []Scene{{Resource: Resource{Id: "s5", Name: "Elsewhere"}}},
	})
	_, err = ControlScenes(home, b1, []string{"Elsewhere"})
	assert.ErrorContains(t, err, "Kitchen", "wrong room")
	_, err = ControlScenes(home, b1, []string{"Cozy"})
	assert.ErrorContains(t, err, "use a scene ID", "ambiguous name")
	_, err = ControlScenes(home, b1, []string{"Nope"})
	assert.ErrorContains(t, err, "not found")

	name, ok := SceneName(home, "s2")
	assert.True(t, ok)
	assert.Equal(t, "Soho", name)
	_, ok = SceneName(home, "zzz")
	assert.False(t, ok)
}
