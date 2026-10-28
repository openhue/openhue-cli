package openhue

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/openhue/openhue-go"
)

// Aliases of the anonymous struct types used by the generated API model.
type (
	actionColorTemperature = struct {
		Mirek *openhue.Mirek `json:"mirek,omitempty"`
	}
	actionEffects = struct {
		Effect *openhue.SupportedEffects `json:"effect,omitempty"`
	}
	paletteColorTemperature = struct {
		Mirek *openhue.Mirek `json:"mirek,omitempty"`
	}
	paletteEffect = struct {
		Effect *openhue.SupportedEffects `json:"effect,omitempty"`
	}
)

// lightRef pairs a light of the room with its identifier as scene action target.
func lightTarget(l *Light) openhue.ResourceIdentifier {
	return openhue.ResourceIdentifier{Rid: l.Id, Rtype: openhue.ResourceIdentifierRtypeLight}
}

// RoomLights returns the lights of the room.
func (r *Room) RoomLights() []*Light {
	var lights []*Light
	for i := range r.Devices {
		if r.Devices[i].Light != nil {
			lights = append(lights, r.Devices[i].Light)
		}
	}
	return lights
}

// ResolveLight finds a light of the room by name or ID. Zero or several matches are errors.
func (r *Room) ResolveLight(ref string) (*Light, error) {
	var found []*Light
	for _, l := range r.RoomLights() {
		if l.Id == ref || l.Name == ref {
			found = append(found, l)
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("light '%s' not found in room '%s'", ref, r.Name)
	case 1:
		return found[0], nil
	default:
		return nil, fmt.Errorf("several lights are named '%s' in room '%s', use the ID instead", ref, r.Name)
	}
}

//
// Capabilities
//

func supportsColor(l *Light) bool { return l.HueData == nil || l.HueData.Color != nil }
func supportsTemperature(l *Light) bool {
	return l.HueData == nil || l.HueData.ColorTemperature != nil
}
func supportsGradient(l *Light) bool { return l.HueData == nil || l.HueData.Gradient != nil }
func supportsEffect(l *Light, effect string) bool {
	if l == nil || l.HueData == nil {
		return true
	}
	if l.HueData.Effects == nil {
		return false
	}
	values := l.HueData.Effects.EffectValues
	if values == nil || effect == string(openhue.NoEffect) {
		return true
	}
	for _, v := range *values {
		if string(v) == effect {
			return true
		}
	}
	return false
}

//
// Baselines
//

// snapshotAction returns an action reproducing the current state of the light.
func snapshotAction(l *Light) openhue.ActionPost {
	a := openhue.ActionPost{Target: lightTarget(l)}
	if l.HueData == nil {
		return a
	}
	on := l.HueData.On
	a.Action.On = &on
	if l.HueData.Dimming != nil && l.HueData.Dimming.Brightness != nil {
		a.Action.Dimming = &openhue.Dimming{Brightness: *l.HueData.Dimming.Brightness}
	}
	ct := l.HueData.ColorTemperature
	if ct != nil && ct.Mirek != nil && (ct.MirekValid == nil || *ct.MirekValid) {
		m := *ct.Mirek
		a.Action.ColorTemperature = &actionColorTemperature{Mirek: &m}
	} else if l.HueData.Color != nil && l.HueData.Color.Xy != nil {
		a.Action.Color = &openhue.Color{Xy: *l.HueData.Color.Xy}
	}
	return a
}

// offAction returns an action turning the light off and nothing else.
func offAction(lightId string) openhue.ActionPost {
	a := openhue.ActionPost{Target: openhue.ResourceIdentifier{Rid: lightId, Rtype: openhue.ResourceIdentifierRtypeLight}}
	a.Action.On = &openhue.On{On: false}
	return a
}

// storedAction converts a stored action to the post model, preserving every field.
func storedAction(a openhue.ActionGet) (openhue.ActionPost, error) {
	var p openhue.ActionPost
	raw, err := json.Marshal(a)
	if err == nil {
		err = json.Unmarshal(raw, &p)
	}
	return p, err
}

//
// Applying settings
//

// applySettings applies the settings on the action. Explicit settings that the light does not support are
// errors, the ones coming from defaults are skipped with a warning.
func applySettings(a *openhue.ActionPost, s *LightSettings, l *Light, explicit bool, warnings *[]string) error {
	if s == nil {
		return nil
	}

	unsupported := func(what string) error {
		if explicit {
			return fmt.Errorf("light '%s' does not support %s", l.Name, what)
		}
		*warnings = append(*warnings, fmt.Sprintf("light '%s' does not support %s, skipped", l.Name, what))
		return nil
	}

	if s.On != nil {
		a.Action.On = &openhue.On{On: *s.On}
	}
	if s.Brightness != nil {
		if l.HueData != nil && l.HueData.Dimming == nil {
			if err := unsupported("brightness"); err != nil {
				return err
			}
		} else {
			a.Action.Dimming = &openhue.Dimming{Brightness: *s.Brightness}
		}
	}

	xy, err := s.xy()
	if err != nil {
		return err
	}
	switch {
	case xy != nil:
		if !supportsColor(l) {
			if err := unsupported("colors"); err != nil {
				return err
			}
		} else {
			a.Action.Color = &openhue.Color{Xy: openhue.GamutPosition{X: xy.X, Y: xy.Y}}
			a.Action.ColorTemperature, a.Action.Gradient, a.Action.Effects = nil, nil, nil
		}
	case s.Temperature != nil:
		if !supportsTemperature(l) {
			if err := unsupported("color temperatures"); err != nil {
				return err
			}
		} else {
			m := *s.Temperature
			a.Action.ColorTemperature = &actionColorTemperature{Mirek: &m}
			a.Action.Color, a.Action.Gradient, a.Action.Effects = nil, nil, nil
		}
	case s.Gradient != nil && len(s.Gradient.Points) > 0:
		if !supportsGradient(l) {
			if err := unsupported("gradients"); err != nil {
				return err
			}
		} else {
			g := &openhue.Gradient{}
			points := make([]openhue.Color, 0, len(s.Gradient.Points))
			for _, p := range s.Gradient.Points {
				c, err := ParseColor(p)
				if err != nil {
					return err
				}
				points = append(points, openhue.Color{Xy: openhue.GamutPosition{X: c.X, Y: c.Y}})
			}
			g.Points = &points
			if s.Gradient.Mode != "" {
				mode := openhue.SupportedGradientMode(s.Gradient.Mode)
				g.Mode = &mode
			}
			a.Action.Gradient = g
			a.Action.Color, a.Action.ColorTemperature, a.Action.Effects = nil, nil, nil
		}
	case s.Gradient != nil && s.Gradient.Mode != "":
		if a.Action.Gradient == nil {
			return fmt.Errorf("light '%s': a gradient mode needs gradient points", l.Name)
		}
		mode := openhue.SupportedGradientMode(s.Gradient.Mode)
		a.Action.Gradient.Mode = &mode
	}

	if s.Effect != "" {
		if !supportsEffect(l, s.Effect) {
			if err := unsupported(fmt.Sprintf("the effect '%s'", s.Effect)); err != nil {
				return err
			}
		} else {
			e := openhue.SupportedEffects(s.Effect)
			a.Action.Effects = &actionEffects{Effect: &e}
			if s.Effect != string(openhue.NoEffect) {
				// the bridge rejects an effect combined with a color, a temperature or a gradient: the effect wins
				a.Action.Color, a.Action.ColorTemperature, a.Action.Gradient = nil, nil, nil
			}
		}
	}

	if s.Transition != "" {
		d, err := s.TransitionDuration()
		if err != nil {
			return err
		}
		ms := int(d / time.Millisecond)
		a.Action.Dynamics = &openhue.Dynamics{Duration: &ms}
	}

	return nil
}

//
// Palette
//

func buildPalette(p *PaletteDefinition) (*openhue.ScenePalette, error) {
	if p.isEmpty() {
		return nil, nil
	}
	out := &openhue.ScenePalette{}
	if len(p.Colors) > 0 {
		list := make([]openhue.ColorPaletteGet, 0, len(p.Colors))
		for _, c := range p.Colors {
			xy, err := ParseColor(c)
			if err != nil {
				return nil, err
			}
			list = append(list, openhue.ColorPaletteGet{
				Color:   &openhue.Color{Xy: openhue.GamutPosition{X: xy.X, Y: xy.Y}},
				Dimming: &openhue.Dimming{Brightness: 100}, // required by the bridge
			})
		}
		out.Color = &list
	}
	if len(p.Temperatures) > 0 {
		list := make([]openhue.ColorTemperaturePalettePost, 0, len(p.Temperatures))
		for _, t := range p.Temperatures {
			m := t
			list = append(list, openhue.ColorTemperaturePalettePost{
				ColorTemperature: &paletteColorTemperature{Mirek: &m},
				Dimming:          &openhue.Dimming{Brightness: 100}, // required by the bridge
			})
		}
		out.ColorTemperature = &list
	}
	if len(p.Brightness) > 0 {
		list := make([]openhue.Dimming, 0, len(p.Brightness))
		for _, b := range p.Brightness {
			list = append(list, openhue.Dimming{Brightness: b})
		}
		out.Dimming = &list
	}
	if len(p.Effects) > 0 {
		list := make([]paletteEffect, 0, len(p.Effects))
		for _, e := range p.Effects {
			ef := openhue.SupportedEffects(e)
			list = append(list, paletteEffect{Effect: &ef})
		}
		out.Effects = &list
	}
	return out, nil
}

func paletteHasEntries(p *openhue.ScenePalette) bool {
	if p == nil {
		return false
	}
	return (p.Color != nil && len(*p.Color) > 0) || (p.ColorTemperature != nil && len(*p.ColorTemperature) > 0) ||
		(p.Dimming != nil && len(*p.Dimming) > 0) || (p.Effects != nil && len(*p.Effects) > 0)
}

const epsilon = 1e-4

func approx(a, b float32) bool { return math.Abs(float64(a-b)) < epsilon }

// mergePalette replaces in the stored palette the lists that the definition names and that really differ
// (stored colors carry extra per-color data that a re-applied export cannot restore, so equal lists are kept).
func mergePalette(stored *openhue.ScenePalette, def *PaletteDefinition) (*openhue.ScenePalette, bool, error) {
	wanted, err := buildPalette(def)
	if err != nil || wanted == nil {
		return stored, false, err
	}
	result := openhue.ScenePalette{}
	if stored != nil {
		result = *stored
	}
	changed := false

	if wanted.Color != nil && !sameColors(result.Color, wanted.Color) {
		result.Color, changed = wanted.Color, true
	}
	if wanted.ColorTemperature != nil && !sameTemperatures(result.ColorTemperature, wanted.ColorTemperature) {
		result.ColorTemperature, changed = wanted.ColorTemperature, true
	}
	if wanted.Dimming != nil && !sameDimmings(result.Dimming, wanted.Dimming) {
		result.Dimming, changed = wanted.Dimming, true
	}
	if wanted.Effects != nil && !sameEffects(result.Effects, wanted.Effects) {
		result.Effects, changed = wanted.Effects, true
	}
	return &result, changed, nil
}

func sameColors(a, b *[]openhue.ColorPaletteGet) bool {
	if a == nil || len(*a) != len(*b) {
		return false
	}
	for i := range *a {
		x, y := (*a)[i].Color, (*b)[i].Color
		if x == nil || y == nil || !approx(x.Xy.X, y.Xy.X) || !approx(x.Xy.Y, y.Xy.Y) {
			return false
		}
	}
	return true
}

func sameTemperatures(a, b *[]openhue.ColorTemperaturePalettePost) bool {
	if a == nil || len(*a) != len(*b) {
		return false
	}
	for i := range *a {
		x, y := (*a)[i].ColorTemperature, (*b)[i].ColorTemperature
		if x == nil || x.Mirek == nil || *x.Mirek != *y.Mirek {
			return false
		}
	}
	return true
}

func sameDimmings(a, b *[]openhue.Dimming) bool {
	if a == nil || len(*a) != len(*b) {
		return false
	}
	for i := range *a {
		if !approx((*a)[i].Brightness, (*b)[i].Brightness) {
			return false
		}
	}
	return true
}

func sameEffects(a, b *[]struct {
	Effect *openhue.SupportedEffects `json:"effect,omitempty"`
}) bool {
	if a == nil || len(*a) != len(*b) {
		return false
	}
	for i := range *a {
		x, y := (*a)[i].Effect, (*b)[i].Effect
		if x == nil || *x != *y {
			return false
		}
	}
	return true
}

func speedToApi(s float32) float32 { return s / 100 }

//
// Create
//

// BuildScenePost builds the request creating a scene in the room from the definition. Nothing is sent: the
// returned warnings are non-fatal remarks (for instance a default color skipped for a white-only light).
func BuildScenePost(room *Room, def *SceneDefinition, onlyListed bool) (*openhue.ScenePost, []string, error) {
	var warnings []string

	if strings.TrimSpace(def.Name) == "" {
		return nil, nil, fmt.Errorf("a scene name is required")
	}
	if err := def.Validate(); err != nil {
		return nil, nil, err
	}
	for _, s := range room.Scenes {
		if s.Name == def.Name {
			return nil, nil, fmt.Errorf("a scene named '%s' already exists in room '%s'", def.Name, room.Name)
		}
	}
	if len(def.RemoveLights) > 0 {
		return nil, nil, fmt.Errorf("remove_lights is only valid when updating a scene")
	}

	roomLights := room.RoomLights()
	if len(roomLights) == 0 {
		return nil, nil, fmt.Errorf("room '%s' has no light", room.Name)
	}

	entries := map[string]*LightEntry{}
	var order []*Light
	for i := range def.Lights {
		l, err := room.ResolveLight(def.Lights[i].Light)
		if err != nil {
			return nil, nil, err
		}
		if _, dup := entries[l.Id]; dup {
			return nil, nil, fmt.Errorf("light '%s' is specified more than once", l.Name)
		}
		entries[l.Id] = &def.Lights[i]
		order = append(order, l)
	}

	if onlyListed && len(order) == 0 {
		return nil, nil, fmt.Errorf("--only-listed requires at least one light")
	}

	post := &openhue.ScenePost{
		Group:    openhue.ResourceIdentifier{Rid: room.Id, Rtype: openhue.ResourceIdentifierRtypeRoom},
		Metadata: openhue.SceneMetadata{Name: &def.Name},
		Actions:  []openhue.ActionPost{},
	}
	t := openhue.ScenePostTypeScene
	post.Type = &t

	for _, l := range roomLights {
		if _, listed := entries[l.Id]; onlyListed && !listed {
			// the bridge requires every light of the room in a scene: the others are turned off
			post.Actions = append(post.Actions, offAction(l.Id))
			continue
		}
		a := snapshotAction(l)
		if err := applySettings(&a, def.Defaults, l, false, &warnings); err != nil {
			return nil, nil, err
		}
		if e, ok := entries[l.Id]; ok {
			if err := applySettings(&a, &e.LightSettings, l, true, &warnings); err != nil {
				return nil, nil, err
			}
		}
		post.Actions = append(post.Actions, a)
	}

	if def.Speed != nil {
		s := speedToApi(*def.Speed)
		post.Speed = &s
	}
	if def.AutoDynamic != nil {
		post.AutoDynamic = def.AutoDynamic
	}
	palette, err := buildPalette(def.Palette)
	if err != nil {
		return nil, nil, err
	}
	post.Palette = palette
	if def.AutoDynamic != nil && *def.AutoDynamic && !paletteHasEntries(palette) {
		return nil, nil, fmt.Errorf("auto-dynamic requires at least one palette entry")
	}

	return post, warnings, nil
}

//
// Update
//

// BuildScenePut builds the request updating the scene from a patch definition. The complete merged action
// list is sent when any light is concerned, because the bridge replaces the whole list.
func BuildScenePut(room *Room, scene *Scene, def *SceneDefinition) (*openhue.ScenePut, []string, error) {
	var warnings []string

	if err := def.Validate(); err != nil {
		return nil, nil, err
	}
	if scene.HueData == nil {
		return nil, nil, fmt.Errorf("scene '%s' has no data", scene.Name)
	}

	put := &openhue.ScenePut{}

	// stored actions, in order
	actions := make([]openhue.ActionPost, 0, len(scene.HueData.Actions))
	for _, stored := range scene.HueData.Actions {
		a, err := storedAction(stored)
		if err != nil {
			return nil, nil, err
		}
		actions = append(actions, a)
	}
	index := func(rid string) int {
		for i := range actions {
			if actions[i].Target.Rid == rid {
				return i
			}
		}
		return -1
	}
	lightOf := func(rid string) *Light {
		for _, l := range room.RoomLights() {
			if l.Id == rid {
				return l
			}
		}
		return &Light{Resource: Resource{Id: rid, Name: rid}}
	}

	// removals
	removed := map[string]bool{}
	for _, ref := range def.RemoveLights {
		l, err := room.ResolveLight(ref)
		if err != nil {
			return nil, nil, err
		}
		if index(l.Id) < 0 {
			return nil, nil, fmt.Errorf("light '%s' is not in scene '%s'", l.Name, scene.Name)
		}
		removed[l.Id] = true
	}
	// the bridge requires every light of the room in a scene: a removed light is turned off instead
	for i := range actions {
		if removed[actions[i].Target.Rid] {
			actions[i] = offAction(actions[i].Target.Rid)
		}
	}

	// entries (resolve everything before changing anything)
	type resolved struct {
		light *Light
		entry *LightEntry
	}
	var entries []resolved
	seen := map[string]bool{}
	for i := range def.Lights {
		l, err := room.ResolveLight(def.Lights[i].Light)
		if err != nil {
			return nil, nil, err
		}
		if seen[l.Id] {
			return nil, nil, fmt.Errorf("light '%s' is specified more than once", l.Name)
		}
		if removed[l.Id] {
			return nil, nil, fmt.Errorf("light '%s' cannot be both updated and removed", l.Name)
		}
		seen[l.Id] = true
		entries = append(entries, resolved{l, &def.Lights[i]})
	}

	// defaults on every light of the scene, then entries (added lights start from their current state)
	for i := range actions {
		if removed[actions[i].Target.Rid] {
			continue
		}
		if err := applySettings(&actions[i], def.Defaults, lightOf(actions[i].Target.Rid), false, &warnings); err != nil {
			return nil, nil, err
		}
	}
	for _, r := range entries {
		i := index(r.light.Id)
		if i < 0 {
			a := snapshotAction(r.light)
			if err := applySettings(&a, def.Defaults, r.light, false, &warnings); err != nil {
				return nil, nil, err
			}
			actions = append(actions, a)
			i = len(actions) - 1
		}
		if err := applySettings(&actions[i], &r.entry.LightSettings, r.light, true, &warnings); err != nil {
			return nil, nil, err
		}
	}
	if len(removed) > 0 || !def.Defaults.IsEmpty() || len(entries) > 0 {
		put.Actions = &actions
	}

	// rename
	if def.Name != "" && def.Name != scene.Name {
		for _, s := range room.Scenes {
			if s.Name == def.Name && s.Id != scene.Id {
				return nil, nil, fmt.Errorf("a scene named '%s' already exists in room '%s'", def.Name, room.Name)
			}
		}
		name := def.Name
		put.Metadata = &openhue.SceneMetadata{Name: &name}
	}

	// dynamics
	if def.Speed != nil && !approx(speedToApi(*def.Speed), scene.HueData.Speed) {
		s := speedToApi(*def.Speed)
		put.Speed = &s
	}
	autoDynamic := scene.HueData.AutoDynamic
	if def.AutoDynamic != nil {
		autoDynamic = *def.AutoDynamic
		if autoDynamic != scene.HueData.AutoDynamic {
			put.AutoDynamic = def.AutoDynamic
		}
	}

	palette := scene.HueData.Palette
	if def.ClearPalette {
		if autoDynamic {
			return nil, nil, fmt.Errorf("the palette cannot be cleared while auto-dynamic is on")
		}
		empty := openhue.ScenePalette{
			Color:            &[]openhue.ColorPaletteGet{},
			ColorTemperature: &[]openhue.ColorTemperaturePalettePost{},
			Dimming:          &[]openhue.Dimming{},
			Effects:          &[]paletteEffect{},
		}
		palette, put.Palette = &empty, &empty
	}
	if !def.Palette.isEmpty() {
		merged, changed, err := mergePalette(palette, def.Palette)
		if err != nil {
			return nil, nil, err
		}
		palette = merged
		if changed {
			put.Palette = merged
		}
	}
	if autoDynamic && !paletteHasEntries(palette) {
		return nil, nil, fmt.Errorf("auto-dynamic requires at least one palette entry")
	}

	return put, warnings, nil
}
