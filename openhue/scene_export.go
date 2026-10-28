package openhue

import (
	"math"

	"github.com/openhue/openhue-go"
)

// SceneToDefinition converts a scene to its definition. Lights are referenced by name, or by ID when
// several lights of the room share the name. Fields the definition format does not cover are not exported.
func SceneToDefinition(room *Room, scene *Scene) *SceneDefinition {
	def := &SceneDefinition{Name: scene.Name}
	if room != nil {
		def.Room = room.Name
	}
	if scene.HueData == nil {
		return def
	}
	d := scene.HueData

	speed := float32(math.Round(float64(d.Speed)*10000) / 100)
	def.Speed = &speed
	auto := d.AutoDynamic
	def.AutoDynamic = &auto
	def.Palette = paletteToDefinition(d.Palette)

	for _, a := range d.Actions {
		ref := a.Target.Rid
		if room != nil {
			for _, l := range room.RoomLights() {
				if l.Id == a.Target.Rid {
					ref = l.Name
					if lightNameCount(room, l.Name) > 1 {
						ref = l.Id
					}
				}
			}
		}
		def.Lights = append(def.Lights, LightEntry{Light: ref, LightSettings: actionToSettings(a)})
	}
	return def
}

func lightNameCount(room *Room, name string) int {
	n := 0
	for _, l := range room.RoomLights() {
		if l.Name == name {
			n++
		}
	}
	return n
}

func actionToSettings(a openhue.ActionGet) LightSettings {
	s := LightSettings{}
	act := a.Action
	if act.On != nil {
		on := act.On.On
		s.On = &on
	}
	if act.Dimming != nil {
		b := act.Dimming.Brightness
		s.Brightness = &b
	}
	hasEffect := act.Effects != nil && act.Effects.Effect != nil && *act.Effects.Effect != openhue.NoEffect
	switch {
	case hasEffect:
		// an effect excludes any color, temperature or gradient
	case act.Gradient != nil && act.Gradient.Points != nil && len(*act.Gradient.Points) > 0:
		g := &GradientSettings{}
		if act.Gradient.Mode != nil {
			g.Mode = string(*act.Gradient.Mode)
		}
		for _, p := range *act.Gradient.Points {
			g.Points = append(g.Points, FormatXY(p.Xy.X, p.Xy.Y))
		}
		s.Gradient = g
	case act.ColorTemperature != nil && act.ColorTemperature.Mirek != nil:
		m := *act.ColorTemperature.Mirek
		s.Temperature = &m
	case act.Color != nil:
		x, y := act.Color.Xy.X, act.Color.Xy.Y
		s.X, s.Y = &x, &y
	}
	if act.Effects != nil && act.Effects.Effect != nil {
		s.Effect = string(*act.Effects.Effect)
	}
	return s
}

func paletteToDefinition(p *openhue.ScenePalette) *PaletteDefinition {
	if p == nil {
		return nil
	}
	out := &PaletteDefinition{}
	if p.Color != nil {
		for _, c := range *p.Color {
			if c.Color != nil {
				out.Colors = append(out.Colors, FormatXY(c.Color.Xy.X, c.Color.Xy.Y))
			}
		}
	}
	if p.ColorTemperature != nil {
		for _, c := range *p.ColorTemperature {
			if c.ColorTemperature != nil && c.ColorTemperature.Mirek != nil {
				out.Temperatures = append(out.Temperatures, *c.ColorTemperature.Mirek)
			}
		}
	}
	if p.Dimming != nil {
		for _, d := range *p.Dimming {
			out.Brightness = append(out.Brightness, d.Brightness)
		}
	}
	if p.Effects != nil {
		for _, e := range *p.Effects {
			if e.Effect != nil {
				out.Effects = append(out.Effects, string(*e.Effect))
			}
		}
	}
	if out.isEmpty() {
		return nil
	}
	return out
}
