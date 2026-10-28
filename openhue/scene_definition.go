package openhue

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"openhue-cli/util/color"
)

// SceneDefinition is the single internal model of a scene authored from flags or a YAML file.
// All the pointer and slice fields are optional: an unset field means "not specified", which
// lets the same type act as a full definition (create) or as a patch (update).
type SceneDefinition struct {
	Name         string             `yaml:"name,omitempty"`
	Room         string             `yaml:"room,omitempty"`
	Speed        *float32           `yaml:"speed,omitempty"` // 0-100
	AutoDynamic  *bool              `yaml:"auto_dynamic,omitempty"`
	Palette      *PaletteDefinition `yaml:"palette,omitempty"`
	Defaults     *LightSettings     `yaml:"defaults,omitempty"`
	Lights       []LightEntry       `yaml:"lights,omitempty"`
	RemoveLights []string           `yaml:"remove_lights,omitempty"`

	// ClearPalette is only used by the update command, it has no YAML representation.
	ClearPalette bool `yaml:"-"`
}

// PaletteDefinition describes the palette used when a scene is played dynamically.
type PaletteDefinition struct {
	Colors       []string  `yaml:"colors,omitempty"` // color name, #RRGGBB or xy@x/y
	Temperatures []int     `yaml:"temperatures,omitempty"`
	Brightness   []float32 `yaml:"brightness,omitempty"`
	Effects      []string  `yaml:"effects,omitempty"`
}

// LightEntry is the settings of one light of the scene.
type LightEntry struct {
	Light         string `yaml:"light"`
	LightSettings `yaml:",inline"`
}

// GradientSettings describes a gradient. Points are color names, #RRGGBB or xy@x/y.
type GradientSettings struct {
	Mode   string   `yaml:"mode,omitempty"`
	Points []string `yaml:"points,omitempty"`
}

// LightSettings is the block of settings applicable to a light (or to all of them as defaults).
type LightSettings struct {
	On          *bool             `yaml:"on,omitempty"`
	Brightness  *float32          `yaml:"brightness,omitempty"`
	Rgb         string            `yaml:"rgb,omitempty"`
	X           *float32          `yaml:"x,omitempty"`
	Y           *float32          `yaml:"y,omitempty"`
	Color       string            `yaml:"color,omitempty"`
	Temperature *int              `yaml:"temperature,omitempty"`
	Transition  string            `yaml:"transition,omitempty"`
	Effect      string            `yaml:"effect,omitempty"`
	Gradient    *GradientSettings `yaml:"gradient,omitempty"`
}

// IsEmpty returns true when no setting is specified.
func (s *LightSettings) IsEmpty() bool {
	return s == nil || (s.On == nil && s.Brightness == nil && !s.hasColorGroup() && s.Transition == "" && s.Effect == "")
}

// hasColorGroup returns true if any of the mutually exclusive color/temperature/gradient settings is set.
func (s *LightSettings) hasColorGroup() bool {
	return s.Rgb != "" || s.X != nil || s.Y != nil || s.Color != "" || s.Temperature != nil || s.Gradient != nil
}

// IsEmpty returns true when the definition requests nothing, ignoring the name and the room.
func (d *SceneDefinition) IsEmpty() bool {
	return d.Speed == nil && d.AutoDynamic == nil && (d.Palette == nil || d.Palette.isEmpty()) &&
		d.Defaults.IsEmpty() && len(d.Lights) == 0 && len(d.RemoveLights) == 0 && !d.ClearPalette
}

func (p *PaletteDefinition) isEmpty() bool {
	return p == nil || (len(p.Colors) == 0 && len(p.Temperatures) == 0 && len(p.Brightness) == 0 && len(p.Effects) == 0)
}

// Validate checks every setting of the definition without needing any home resource.
func (d *SceneDefinition) Validate() error {
	if d.Speed != nil && (*d.Speed < 0 || *d.Speed > 100) {
		return fmt.Errorf("speed must be between 0 and 100, current value is %.2f", *d.Speed)
	}
	if d.Palette != nil {
		if err := d.Palette.Validate(); err != nil {
			return err
		}
	}
	if d.Defaults != nil {
		if err := d.Defaults.Validate(); err != nil {
			return fmt.Errorf("defaults: %w", err)
		}
	}
	seen := map[string]bool{}
	for _, l := range d.Lights {
		if l.Light == "" {
			return fmt.Errorf("a light entry has no 'light' reference")
		}
		if seen[l.Light] {
			return fmt.Errorf("light '%s' is specified more than once", l.Light)
		}
		seen[l.Light] = true
		if err := l.Validate(); err != nil {
			return fmt.Errorf("light '%s': %w", l.Light, err)
		}
	}
	for _, r := range d.RemoveLights {
		if seen[r] {
			return fmt.Errorf("light '%s' cannot be both updated and removed", r)
		}
	}
	return nil
}

// Validate checks the palette values.
func (p *PaletteDefinition) Validate() error {
	for _, c := range p.Colors {
		if _, err := ParseColor(c); err != nil {
			return fmt.Errorf("palette color: %w", err)
		}
	}
	for _, t := range p.Temperatures {
		if t < 153 || t > 500 {
			return fmt.Errorf("palette temperature must be between 153 and 500, current value is %d", t)
		}
	}
	for _, b := range p.Brightness {
		if b < 0 || b > 100 {
			return fmt.Errorf("palette brightness must be between 0 and 100, current value is %.2f", b)
		}
	}
	for _, e := range p.Effects {
		if err := validateEffectName(e); err != nil {
			return fmt.Errorf("palette effect: %w", err)
		}
	}
	return nil
}

// Validate checks the settings values and their consistency.
func (s *LightSettings) Validate() error {
	if s == nil {
		return nil
	}
	if s.Brightness != nil && (*s.Brightness < 0 || *s.Brightness > 100) {
		return fmt.Errorf("brightness must be between 0 and 100, current value is %.2f", *s.Brightness)
	}
	if s.Temperature != nil && (*s.Temperature < 153 || *s.Temperature > 500) {
		return fmt.Errorf("temperature must be between 153 and 500, current value is %d", *s.Temperature)
	}
	if (s.X == nil) != (s.Y == nil) {
		return fmt.Errorf("x and y must be set together")
	}
	if s.X != nil && (*s.X < 0 || *s.X > 1 || *s.Y < 0 || *s.Y > 1) {
		return fmt.Errorf("x and y must be between 0 and 1")
	}
	groups := 0
	for _, set := range []bool{s.Rgb != "", s.X != nil, s.Color != "", s.Temperature != nil, s.Gradient != nil} {
		if set {
			groups++
		}
	}
	if groups > 1 {
		return fmt.Errorf("only one of rgb, x/y, color, temperature or gradient can be set")
	}
	if s.Rgb != "" {
		if _, err := color.NewRGBFomHex(s.Rgb); err != nil {
			return fmt.Errorf("rgb: %w", err)
		}
	}
	if s.Color != "" {
		if _, err := color.FindColorByName(s.Color); err != nil {
			return err
		}
	}
	if s.Transition != "" {
		if _, err := s.TransitionDuration(); err != nil {
			return err
		}
	}
	if s.Effect != "" {
		if err := validateEffectName(s.Effect); err != nil {
			return err
		}
	}
	if s.Gradient != nil {
		if s.Gradient.Mode != "" {
			if err := validateGradientMode(s.Gradient.Mode); err != nil {
				return err
			}
		}
		if len(s.Gradient.Points) == 1 {
			return fmt.Errorf("a gradient needs at least 2 points")
		}
		for _, p := range s.Gradient.Points {
			if _, err := ParseColor(p); err != nil {
				return fmt.Errorf("gradient point: %w", err)
			}
		}
	}
	return nil
}

// TransitionDuration parses the transition setting.
func (s *LightSettings) TransitionDuration() (time.Duration, error) {
	d, err := time.ParseDuration(s.Transition)
	if err != nil {
		return 0, fmt.Errorf("invalid transition '%s': %w", s.Transition, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("transition must not be negative, current value is %s", s.Transition)
	}
	return d, nil
}

// xy returns the explicit color of the settings, if any.
func (s *LightSettings) xy() (*color.XY, error) {
	switch {
	case s.Rgb != "":
		rgb, err := color.NewRGBFomHex(s.Rgb)
		if err != nil {
			return nil, err
		}
		return rgb.ToXY(), nil
	case s.X != nil && s.Y != nil:
		return &color.XY{X: *s.X, Y: *s.Y}, nil
	case s.Color != "":
		return color.FindColorByName(s.Color)
	}
	return nil, nil
}

// ParseColor converts a color name, a #RRGGBB hexadecimal value or a xy@x/y value to CIE coordinates.
func ParseColor(s string) (*color.XY, error) {
	switch {
	case strings.HasPrefix(s, "#"):
		rgb, err := color.NewRGBFomHex(s)
		if err != nil {
			return nil, err
		}
		return rgb.ToXY(), nil
	case strings.HasPrefix(s, "xy@"):
		parts := strings.Split(strings.TrimPrefix(s, "xy@"), "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid color '%s': expected xy@x/y", s)
		}
		x, errX := strconv.ParseFloat(parts[0], 32)
		y, errY := strconv.ParseFloat(parts[1], 32)
		if errX != nil || errY != nil || x < 0 || x > 1 || y < 0 || y > 1 {
			return nil, fmt.Errorf("invalid color '%s': x and y must be numbers between 0 and 1", s)
		}
		return &color.XY{X: float32(x), Y: float32(y)}, nil
	default:
		return color.FindColorByName(s)
	}
}

// FormatXY is the reverse of ParseColor for the xy@x/y notation.
func FormatXY(x, y float32) string {
	return "xy@" + strconv.FormatFloat(float64(x), 'g', -1, 32) + "/" + strconv.FormatFloat(float64(y), 'g', -1, 32)
}

var effectNames = []string{"candle", "cosmos", "enchant", "fire", "glisten", "no_effect", "opal", "prism", "sparkle", "sunbeam", "underwater"}
var gradientModes = []string{"interpolated_palette", "interpolated_palette_mirrored", "random_pixelated"}

func validateEffectName(e string) error {
	for _, n := range effectNames {
		if n == e {
			return nil
		}
	}
	return fmt.Errorf("unknown effect '%s'. Supported effects: %s", e, strings.Join(effectNames, ", "))
}

func validateGradientMode(m string) error {
	for _, n := range gradientModes {
		if n == m {
			return nil
		}
	}
	return fmt.Errorf("unknown gradient mode '%s'. Supported modes: %s", m, strings.Join(gradientModes, ", "))
}

//
// --light parsing
//

// ParseLightArg parses a `--light "<light>:<settings>"` value. The value is split at the last ':' so that
// light names containing ':' work. Settings are comma-separated: bare `on`/`off` or `key=value`. Gradient
// points are separated by ';'.
func ParseLightArg(arg string) (*LightEntry, error) {
	idx := strings.LastIndex(arg, ":")
	if idx <= 0 || idx == len(arg)-1 {
		return nil, fmt.Errorf("invalid --light value '%s': expected \"<light>:<settings>\"", arg)
	}

	entry := &LightEntry{Light: strings.TrimSpace(arg[:idx])}
	if entry.Light == "" {
		return nil, fmt.Errorf("invalid --light value '%s': the light is empty", arg)
	}

	for _, item := range strings.Split(arg[idx+1:], ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("light '%s': empty setting", entry.Light)
		}
		if err := entry.apply(item); err != nil {
			return nil, fmt.Errorf("light '%s': %w", entry.Light, err)
		}
	}

	if err := entry.Validate(); err != nil {
		return nil, fmt.Errorf("light '%s': %w", entry.Light, err)
	}
	return entry, nil
}

func (e *LightEntry) apply(item string) error {
	key, value, hasValue := strings.Cut(item, "=")
	t := true
	f := false

	if !hasValue {
		switch key {
		case "on":
			e.On = &t
		case "off":
			e.On = &f
		default:
			return fmt.Errorf("unknown setting '%s'", item)
		}
		return nil
	}

	switch key {
	case "brightness":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("brightness: '%s' is not a number", value)
		}
		b := float32(v)
		e.Brightness = &b
	case "rgb":
		e.Rgb = value
	case "x", "y":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("%s: '%s' is not a number", key, value)
		}
		n := float32(v)
		if key == "x" {
			e.X = &n
		} else {
			e.Y = &n
		}
	case "color":
		e.Color = value
	case "temperature":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("temperature: '%s' is not an integer", value)
		}
		e.Temperature = &v
	case "transition":
		e.Transition = value
	case "effect":
		e.Effect = value
	case "gradient":
		if e.Gradient == nil {
			e.Gradient = &GradientSettings{}
		}
		e.Gradient.Points = strings.Split(value, ";")
	case "gradient-mode":
		if e.Gradient == nil {
			e.Gradient = &GradientSettings{}
		}
		e.Gradient.Mode = value
	default:
		return fmt.Errorf("unknown setting '%s'", key)
	}
	return nil
}

//
// Merge
//

// MergeSceneDefinitions returns base overridden by over: scalar values of over replace the ones of base,
// settings blocks are merged key by key, light entries are matched on their light reference.
func MergeSceneDefinitions(base, over *SceneDefinition) *SceneDefinition {
	m := *base
	m.Lights = append([]LightEntry(nil), base.Lights...)
	m.RemoveLights = append([]string(nil), base.RemoveLights...)

	if over.Name != "" {
		m.Name = over.Name
	}
	if over.Room != "" {
		m.Room = over.Room
	}
	if over.Speed != nil {
		m.Speed = over.Speed
	}
	if over.AutoDynamic != nil {
		m.AutoDynamic = over.AutoDynamic
	}
	m.Palette = mergePalettes(base.Palette, over.Palette)
	m.Defaults = mergeSettings(base.Defaults, over.Defaults)
	m.ClearPalette = base.ClearPalette || over.ClearPalette

	for _, o := range over.Lights {
		found := false
		for i := range m.Lights {
			if m.Lights[i].Light == o.Light {
				m.Lights[i].LightSettings = *mergeSettings(&m.Lights[i].LightSettings, &o.LightSettings)
				found = true
				break
			}
		}
		if !found {
			m.Lights = append(m.Lights, o)
		}
	}

	for _, r := range over.RemoveLights {
		dup := false
		for _, existing := range m.RemoveLights {
			dup = dup || existing == r
		}
		if !dup {
			m.RemoveLights = append(m.RemoveLights, r)
		}
	}

	return &m
}

func mergePalettes(base, over *PaletteDefinition) *PaletteDefinition {
	if over == nil {
		return base
	}
	if base == nil {
		return over
	}
	m := *base
	if len(over.Colors) > 0 {
		m.Colors = over.Colors
	}
	if len(over.Temperatures) > 0 {
		m.Temperatures = over.Temperatures
	}
	if len(over.Brightness) > 0 {
		m.Brightness = over.Brightness
	}
	if len(over.Effects) > 0 {
		m.Effects = over.Effects
	}
	return &m
}

// mergeSettings never returns nil. A color-group setting in over (rgb, x/y, color, temperature or gradient
// points) replaces the whole color group of base.
func mergeSettings(base, over *LightSettings) *LightSettings {
	m := LightSettings{}
	if base != nil {
		m = *base
	}
	if over == nil {
		return &m
	}
	if over.On != nil {
		m.On = over.On
	}
	if over.Brightness != nil {
		m.Brightness = over.Brightness
	}
	if over.Transition != "" {
		m.Transition = over.Transition
	}
	if over.Effect != "" {
		m.Effect = over.Effect
	}
	if over.hasColorGroup() {
		onlyMode := over.Gradient != nil && len(over.Gradient.Points) == 0 &&
			over.Rgb == "" && over.X == nil && over.Color == "" && over.Temperature == nil
		if onlyMode && m.Gradient != nil {
			g := *m.Gradient
			g.Mode = over.Gradient.Mode
			m.Gradient = &g
		} else {
			m.Rgb, m.X, m.Y, m.Color, m.Temperature, m.Gradient = over.Rgb, over.X, over.Y, over.Color, over.Temperature, over.Gradient
		}
	}
	return &m
}
