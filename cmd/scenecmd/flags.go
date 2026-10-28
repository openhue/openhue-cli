// Package scenecmd holds what the 'create scene' and 'update scene' commands share: the flags describing a
// scene, their conversion to an openhue.SceneDefinition, and the request output helpers.
package scenecmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"openhue-cli/openhue"
	"openhue-cli/util/color"
)

// Flags are the flags common to 'create scene' and 'update scene'.
type Flags struct {
	Room     string
	FromFile string
	Lights   []string

	// defaults for all the lights
	On             bool
	Off            bool
	Brightness     float32
	Rgb            string
	X              float32
	Y              float32
	ColorName      string
	Temperature    int
	TransitionTime string

	// dynamics
	Speed             float32
	AutoDynamic       bool
	PaletteColors     []string
	PaletteTemps      []int
	PaletteBrightness []float32
	PaletteEffects    []string

	DryRun bool
	Json   bool

	// update only
	Name         string
	RemoveLights []string
	ClearPalette bool
}

// Register adds the flags to the command. The update flags are only registered when update is true.
func (f *Flags) Register(cmd *cobra.Command, update bool) {
	fl := cmd.Flags()

	if update {
		fl.StringVarP(&f.Room, "room", "r", "", "Room of the scene (name or ID), to disambiguate scenes sharing a name")
		fl.StringVar(&f.Name, "name", "", "Rename the scene")
		fl.StringArrayVar(&f.RemoveLights, "remove-light", nil, "Remove a light (name or ID) from the scene (repeatable)")
		fl.BoolVar(&f.ClearPalette, "clear-palette", false, "Remove the dynamic palette of the scene")
		fl.StringVar(&f.FromFile, "from-file", "", "YAML definition applied as a patch, '-' to read from stdin")
	} else {
		fl.StringVarP(&f.Room, "room", "r", "", "Room of the scene (name or ID)")
		fl.StringVar(&f.FromFile, "from-file", "", "YAML definition of the scene, '-' to read from stdin")
	}

	fl.StringArrayVar(&f.Lights, "light", nil, `Per-light settings "<light>:<settings>" (repeatable). Settings are comma-separated: on, off, brightness=, rgb=, x=, y=, color=, temperature=, transition=, effect=, gradient=<color>;<color>, gradient-mode=`)

	// defaults for all the lights
	fl.BoolVar(&f.On, "on", false, "Turn all the lights on")
	fl.BoolVar(&f.Off, "off", false, "Turn all the lights off")
	cmd.MarkFlagsMutuallyExclusive("on", "off")
	fl.Float32VarP(&f.Brightness, "brightness", "b", -1, "Brightness of all the lights [min=0, max=100]")
	fl.StringVar(&f.Rgb, "rgb", "", "RGB hexadecimal color of all the lights (example #CCE5FF)")
	fl.Float32VarP(&f.X, "cie-x", "x", -1, "X coordinate in the CIE color space (example 0.675)")
	fl.Float32VarP(&f.Y, "cie-y", "y", -1, "Y coordinate in the CIE color space (example 0.322)")
	cmd.MarkFlagsRequiredTogether("cie-x", "cie-y")
	fl.StringVarP(&f.ColorName, "color", "c", "", "Color name of all the lights. Allowed: white, lime, green, blue, etc.")
	_ = cmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return color.GetSupportColorList(), cobra.ShellCompDirectiveDefault
	})
	fl.IntVarP(&f.Temperature, "temperature", "t", -1, "Color temperature of all the lights in Mirek [min=153, max=500]")
	fl.StringVar(&f.TransitionTime, "transition-time", "0s", "Duration of the transition to the scene for all the lights")
	cmd.MarkFlagsMutuallyExclusive("color", "rgb", "cie-x", "temperature")

	// dynamics
	fl.Float32Var(&f.Speed, "speed", 0, "Speed of the dynamic palette [min=0, max=100]")
	fl.BoolVar(&f.AutoDynamic, "auto-dynamic", false, "Play the scene dynamically when it is recalled")
	fl.StringArrayVar(&f.PaletteColors, "palette-color", nil, "Palette color: name, #RRGGBB (repeatable)")
	fl.IntSliceVar(&f.PaletteTemps, "palette-temperature", nil, "Palette color temperature in Mirek (repeatable)")
	fl.Float32SliceVar(&f.PaletteBrightness, "palette-brightness", nil, "Palette brightness [0-100] (repeatable)")
	fl.StringArrayVar(&f.PaletteEffects, "palette-effect", nil, "Palette effect (repeatable)")

	fl.BoolVar(&f.DryRun, "dry-run", false, "Print the request instead of sending it")
	fl.BoolVarP(&f.Json, "json", "j", false, "Print the resulting scene as JSON")
}

// Definition converts the flags that were set to a scene definition. withRoom is false for the update
// command where --room only selects the scene.
func (f *Flags) Definition(cmd *cobra.Command, withRoom bool) (*openhue.SceneDefinition, error) {
	changed := cmd.Flags().Changed
	def := &openhue.SceneDefinition{ClearPalette: f.ClearPalette, RemoveLights: f.RemoveLights}

	if withRoom {
		def.Room = f.Room
	}
	def.Name = f.Name

	if changed("speed") {
		s := f.Speed
		def.Speed = &s
	}
	if changed("auto-dynamic") {
		a := f.AutoDynamic
		def.AutoDynamic = &a
	}
	if len(f.PaletteColors)+len(f.PaletteTemps)+len(f.PaletteBrightness)+len(f.PaletteEffects) > 0 {
		def.Palette = &openhue.PaletteDefinition{
			Colors: f.PaletteColors, Temperatures: f.PaletteTemps, Brightness: f.PaletteBrightness, Effects: f.PaletteEffects,
		}
	}

	d := &openhue.LightSettings{}
	if changed("on") && f.On {
		t := true
		d.On = &t
	}
	if changed("off") && f.Off {
		o := false
		d.On = &o
	}
	if changed("brightness") {
		b := f.Brightness
		d.Brightness = &b
	}
	if changed("rgb") {
		d.Rgb = f.Rgb
	}
	if changed("cie-x") && changed("cie-y") {
		x, y := f.X, f.Y
		d.X, d.Y = &x, &y
	}
	if changed("color") {
		d.Color = f.ColorName
	}
	if changed("temperature") {
		t := f.Temperature
		d.Temperature = &t
	}
	if changed("transition-time") {
		d.Transition = f.TransitionTime
	}
	if !d.IsEmpty() {
		def.Defaults = d
	}

	for _, l := range f.Lights {
		entry, err := openhue.ParseLightArg(l)
		if err != nil {
			return nil, err
		}
		def.Lights = append(def.Lights, *entry)
	}
	return def, nil
}

// Resolve loads the file (if any) and merges the flags over it. A non-empty name overrides the file's.
func (f *Flags) Resolve(cmd *cobra.Command, stdin io.Reader, name string, withRoom bool) (*openhue.SceneDefinition, error) {
	flagDef, err := f.Definition(cmd, withRoom)
	if err != nil {
		return nil, err
	}
	if name != "" {
		flagDef.Name = name
	}

	def := flagDef
	if f.FromFile != "" {
		fileDef, err := openhue.LoadSceneDefinition(f.FromFile, stdin)
		if err != nil {
			return nil, err
		}
		def = openhue.MergeSceneDefinitions(fileDef, flagDef)
	}
	if err := def.Validate(); err != nil {
		return nil, err
	}
	return def, nil
}

// PrintRequest prints the body as indented JSON.
func PrintRequest(io openhue.IOStreams, body any) error {
	out, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to format the request: %w", err)
	}
	io.Println(string(out))
	return nil
}

// PrintWarnings prints the warnings on the error output.
func PrintWarnings(io openhue.IOStreams, warnings []string) {
	for _, w := range warnings {
		io.ErrPrintf("Warning: %s\n", w)
	}
}
