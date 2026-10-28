package openhue

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// BehaviorInstance is the configuration of an accessory ("Hue Accessories" script) as stored on the bridge. The
// configuration is kept as an untyped tree so that whatever the CLI does not understand survives a write.
type BehaviorInstance struct {
	Id            string
	Configuration map[string]any
}

// findBehavior returns the behavior instance configuring the given device, nil if there is none.
func findBehavior(ctx *hueHomeCtx, deviceId string) *BehaviorInstance {
	for _, b := range ctx.behaviors {
		if b.Configuration == nil {
			continue
		}
		if rid, _ := dig(*b.Configuration, "device", "rid").(string); rid == deviceId {
			if _, ok := (*b.Configuration)["buttons"]; !ok {
				continue
			}
			return &BehaviorInstance{Id: b.Id, Configuration: *b.Configuration}
		}
	}
	return nil
}

// ButtonConfig returns the stored configuration of the button, nil if the accessory has none for it.
func (c *Control) ButtonConfig() map[string]any {
	if c.Behavior == nil {
		return nil
	}
	cfg, _ := dig(c.Behavior.Configuration, "buttons", c.Id).(map[string]any)
	return cfg
}

// ShortPressSummary describes what a short press of the button does.
func (c *Control) ShortPressSummary() string {
	var names func(string) (string, bool)
	if c.ctx != nil {
		names = func(id string) (string, bool) {
			s, ok := c.ctx.scenes[id]
			if !ok || s.Metadata.Name == nil {
				return "", false
			}
			return *s.Metadata.Name, true
		}
	}
	return SummarizeShortPress(c.ButtonConfig(), names)
}

// HoldSummary describes what holding the button does.
func (c *Control) HoldSummary() string {
	return SummarizeHold(c.ButtonConfig())
}

// SummarizeShortPress turns the on_short_release part of a button configuration into text. sceneName resolves a scene ID
// to its name. Unrecognized shapes are reported as "custom".
func SummarizeShortPress(button map[string]any, sceneName func(id string) (string, bool)) string {
	release, _ := button["on_short_release"].(map[string]any)
	if len(release) == 0 {
		return "-"
	}

	describe := func(rid, rtype string) string {
		if rtype == "smart_scene" {
			return "smart-scene " + rid
		}
		if sceneName != nil {
			if n, ok := sceneName(rid); ok {
				return n
			}
		}
		return "<unknown scene>"
	}

	if single, ok := release["recall_single_extended"].(map[string]any); ok {
		actions, _ := single["actions"].([]any)
		if len(actions) == 1 {
			if rid, rtype, ok := recallOf(actions[0]); ok {
				if rtype == "smart_scene" {
					return describe(rid, rtype)
				}
				return "scene " + describe(rid, rtype)
			}
		}
	}

	if cycle, ok := release["scene_cycle_extended"].(map[string]any); ok {
		slots, _ := cycle["slots"].([]any)
		var names []string
		valid := len(slots) > 0
		for _, slot := range slots {
			actions, _ := slot.([]any)
			if len(actions) != 1 {
				valid = false
				break
			}
			rid, rtype, ok := recallOf(actions[0])
			if !ok {
				valid = false
				break
			}
			names = append(names, describe(rid, rtype))
		}
		if valid {
			return fmt.Sprintf("cycle: %s (%d)", strings.Join(names, ", "), len(names))
		}
	}

	if _, ok := release["time_based_extended"]; ok && len(release) == 1 {
		return "time-based"
	}

	return "custom"
}

// SummarizeHold turns the on_repeat part of a button configuration into text.
func SummarizeHold(button map[string]any) string {
	repeat, _ := button["on_repeat"].(map[string]any)
	if len(repeat) == 0 {
		return "-"
	}
	switch action, _ := repeat["action"].(string); action {
	case "dim_up", "dim_down", "dim_alternate":
		return strings.ReplaceAll(action, "_", "-")
	case "", "do_nothing":
		return "-"
	}
	return "custom"
}

// recallOf extracts the scene recalled by an {action: {recall: {rid, rtype}}} node.
func recallOf(node any) (rid, rtype string, ok bool) {
	recall, _ := dig(node, "action", "recall").(map[string]any)
	rid, _ = recall["rid"].(string)
	rtype, _ = recall["rtype"].(string)
	return rid, rtype, rid != ""
}

// dig walks an untyped tree along the given keys, nil if the path does not exist.
func dig(node any, keys ...string) any {
	for _, k := range keys {
		m, ok := node.(map[string]any)
		if !ok {
			return nil
		}
		node = m[k]
	}
	return node
}

// ButtonPatch lists the changes to apply to a button. Fields left at their zero value are not touched.
type ButtonPatch struct {
	// SceneIds are the scenes a short press recalls: one for a single scene, several for a cycle.
	SceneIds []string
	// WithOff makes a second press turn the lights off. Nil keeps the current value (false if there is none).
	WithOff *bool
	// RepeatTimeout is the idle time after which a cycle restarts. Zero keeps the current value (3s if there is none).
	RepeatTimeout time.Duration
	// OnHold is one of dim-up, dim-down, dim-alternate, none, or empty to leave the hold behavior alone.
	OnHold string
	// Clear removes the short-press action.
	Clear bool
}

const defaultRepeatTimeout = 3 * time.Second

// BuildButtonConfiguration returns a copy of the behavior configuration where the given button is patched. Everything
// the patch does not cover, other buttons and unknown keys included, is copied as is. config is not modified.
func BuildButtonConfiguration(config map[string]any, buttonId string, p ButtonPatch) (map[string]any, error) {
	if p.Clear && len(p.SceneIds) > 0 {
		return nil, fmt.Errorf("--clear cannot be combined with --scene")
	}
	if p.OnHold != "" && !slices.Contains([]string{"dim-up", "dim-down", "dim-alternate", "none"}, p.OnHold) {
		return nil, fmt.Errorf("invalid --on-hold '%s', allowed values: dim-up, dim-down, dim-alternate, none", p.OnHold)
	}
	if p.RepeatTimeout < 0 || (p.RepeatTimeout > 0 && p.RepeatTimeout < time.Second) {
		return nil, fmt.Errorf("invalid --repeat-timeout %s: must be at least 1s", p.RepeatTimeout)
	}
	if len(p.SceneIds) == 1 && p.RepeatTimeout > 0 {
		return nil, fmt.Errorf("--repeat-timeout only applies to a cycle of several scenes")
	}

	out := deepCopy(config).(map[string]any)
	button, ok := dig(out, "buttons", buttonId).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("button %s has no configuration on the bridge", buttonId)
	}

	release, _ := button["on_short_release"].(map[string]any)
	current := currentWithOff(release)

	switch {
	case p.Clear:
		delete(button, "on_short_release")

	case len(p.SceneIds) > 0:
		withOff := current
		if p.WithOff != nil {
			withOff = *p.WithOff
		}
		button["on_short_release"] = buildRelease(p.SceneIds, withOff, p.RepeatTimeout, release)

	case p.WithOff != nil || p.RepeatTimeout > 0:
		kind, body := releaseBody(release)
		switch {
		case kind == "scene_cycle_extended":
		case kind == "recall_single_extended" && p.RepeatTimeout == 0:
		case p.RepeatTimeout > 0:
			return nil, fmt.Errorf("--repeat-timeout needs a button cycling through scenes: add several --scene")
		default:
			return nil, fmt.Errorf("--with-off needs a button recalling scenes: add --scene")
		}
		if p.WithOff != nil {
			body["with_off"] = map[string]any{"enabled": *p.WithOff}
		}
		if p.RepeatTimeout > 0 {
			body["repeat_timeout"] = map[string]any{"seconds": float64(int(p.RepeatTimeout.Seconds()))}
		}
	}

	switch p.OnHold {
	case "none":
		delete(button, "on_repeat")
	case "":
	default:
		repeat, _ := button["on_repeat"].(map[string]any)
		if repeat == nil {
			repeat = map[string]any{}
		}
		repeat["action"] = strings.ReplaceAll(p.OnHold, "-", "_")
		button["on_repeat"] = repeat
	}

	return out, nil
}

// buildRelease builds an on_short_release node for one scene, or a cycle for several.
func buildRelease(sceneIds []string, withOff bool, timeout time.Duration, previous map[string]any) map[string]any {
	recall := func(id string) map[string]any {
		return map[string]any{"action": map[string]any{"recall": map[string]any{"rid": id, "rtype": "scene"}}}
	}
	off := map[string]any{"enabled": withOff}

	if len(sceneIds) == 1 {
		return map[string]any{"recall_single_extended": map[string]any{
			"actions":  []any{recall(sceneIds[0])},
			"with_off": off,
		}}
	}

	if timeout == 0 {
		timeout = defaultRepeatTimeout
		if _, body := releaseBody(previous); body != nil {
			if s, ok := dig(body, "repeat_timeout", "seconds").(float64); ok && s >= 1 {
				timeout = time.Duration(s) * time.Second
			}
		}
	}
	slots := make([]any, 0, len(sceneIds))
	for _, id := range sceneIds {
		slots = append(slots, []any{recall(id)})
	}
	return map[string]any{"scene_cycle_extended": map[string]any{
		"slots":          slots,
		"repeat_timeout": map[string]any{"seconds": float64(int(timeout.Seconds()))},
		"with_off":       off,
	}}
}

// releaseBody returns the kind and body of a single or cycle short-press node, empty if it is neither.
func releaseBody(release map[string]any) (string, map[string]any) {
	for _, kind := range []string{"recall_single_extended", "scene_cycle_extended"} {
		if body, ok := release[kind].(map[string]any); ok {
			return kind, body
		}
	}
	return "", nil
}

func currentWithOff(release map[string]any) bool {
	_, body := releaseBody(release)
	enabled, _ := dig(body, "with_off", "enabled").(bool)
	return enabled
}

func deepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, e := range t {
			m[k] = deepCopy(e)
		}
		return m
	case []any:
		s := make([]any, len(t))
		for i, e := range t {
			s[i] = deepCopy(e)
		}
		return s
	default:
		return v
	}
}

// ControlScenes resolves scene names or IDs to scene IDs among the scenes of the room the button is bound to: the room
// of its "where" in the configuration, else the room of its device.
func ControlScenes(home *HomeModel, c *Control, nameOrIds []string) ([]string, error) {
	roomId := ""
	if where, ok := dig(c.ButtonConfig(), "where").([]any); ok {
		for _, w := range where {
			if rid, _ := dig(w, "group", "rid").(string); rid != "" && dig(w, "group", "rtype") == "room" {
				roomId = rid
				break
			}
		}
	}
	if roomId == "" {
		if r := c.Room(); r != nil {
			roomId = r.Id
		}
	}
	if roomId == "" {
		return nil, fmt.Errorf("control %s is not bound to a room: cannot pick scenes", c.Id)
	}

	var scenes []Scene
	roomName := roomId
	for _, r := range home.Rooms {
		if r.Id == roomId {
			scenes, roomName = r.Scenes, r.Name
		}
	}

	var ids []string
	for _, ref := range nameOrIds {
		var matches []Scene
		for _, s := range scenes {
			if s.matchesNameOrId(ref) {
				matches = append(matches, s)
			}
		}
		switch len(matches) {
		case 1:
			ids = append(ids, matches[0].Id)
		case 0:
			msg := fmt.Sprintf("scene '%s' not found in room '%s'", ref, roomName)
			for _, r := range home.Rooms {
				for _, s := range r.Scenes {
					if r.Id != roomId && s.matchesNameOrId(ref) {
						msg += fmt.Sprintf(" (it exists in room '%s', the button is bound to '%s')", r.Name, roomName)
						return nil, fmt.Errorf("%s", msg)
					}
				}
			}
			return nil, fmt.Errorf("%s", msg)
		default:
			return nil, fmt.Errorf("%d scenes match '%s' in room '%s', use a scene ID", len(matches), ref, roomName)
		}
	}
	return ids, nil
}

// SceneName returns the name of the scene with the given ID, false if the home has no such scene.
func SceneName(home *HomeModel, id string) (string, bool) {
	for _, r := range home.Rooms {
		for _, s := range r.Scenes {
			if s.Id == id {
				return s.Name, true
			}
		}
	}
	return "", false
}
