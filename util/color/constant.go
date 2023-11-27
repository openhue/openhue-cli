package color

import (
	"fmt"
	"strings"
)

// The Table contains the list of supported colors by name
var Table = map[RGBHex]*XY{
	// basic colors
	"white":   rgb("#FFFFFF").toXY(),
	"red":     rgb("#FF0000").toXY(),
	"lime":    rgb("#00FF00").toXY(),
	"blue":    rgb("#0000FF").toXY(),
	"yellow":  rgb("#FFFF00").toXY(),
	"cyan":    rgb("#00FFFF").toXY(),
	"aqua":    rgb("#00FFFF").toXY(),
	"magenta": rgb("#FF00FF").toXY(),
	"fuchsia": rgb("#FF00FF").toXY(),
	"silver":  rgb("#C0C0C0").toXY(),
	"gray":    rgb("#FF00FF").toXY(),
	"maroon":  rgb("#808080").toXY(),
	"olive":   rgb("#808000").toXY(),
	"green":   rgb("#008000").toXY(),
	"purple":  rgb("#800080").toXY(),
	"teal":    rgb("#008080").toXY(),
	"navy":    rgb("#000080").toXY(),

	// more colors
	"dark_red":                rgb("#8B0000").toXY(),
	"brown":                   rgb("#A52A2A").toXY(),
	"firebrick":               rgb("#B22222").toXY(),
	"crimson":                 rgb("#DC143C").toXY(),
	"tomato":                  rgb("#FF6347").toXY(),
	"coral":                   rgb("#FF7F50").toXY(),
	"indian_red":              rgb("#CD5C5C").toXY(),
	"light_coral":             rgb("#F08080").toXY(),
	"dark_salmon":             rgb("#E9967A").toXY(),
	"salmon":                  rgb("#FA8072").toXY(),
	"light_salmon":            rgb("#FFA07A").toXY(),
	"orange_red":              rgb("#FF4500").toXY(),
	"dark_orange":             rgb("#FF8C00").toXY(),
	"orange":                  rgb("#FFA500").toXY(),
	"gold":                    rgb("#FFD700").toXY(),
	"dark_golden_rod":         rgb("#B8860B").toXY(),
	"golden_rod":              rgb("#DAA520").toXY(),
	"pale_golden_rod":         rgb("#EEE8AA").toXY(),
	"dark_khaki":              rgb("#BDB76B").toXY(),
	"khaki":                   rgb("#F0E68C").toXY(),
	"yellow_green":            rgb("#9ACD32").toXY(),
	"dark_olive_green":        rgb("#556B2F").toXY(),
	"olive_drab":              rgb("#6B8E23").toXY(),
	"lawn_green":              rgb("#7CFC00").toXY(),
	"chartreuse":              rgb("#7FFF00").toXY(),
	"green_yellow":            rgb("#ADFF2F").toXY(),
	"dark_green":              rgb("#006400").toXY(),
	"forest_green":            rgb("#228B22").toXY(),
	"lime_green":              rgb("#32CD32").toXY(),
	"light_green":             rgb("#90EE90").toXY(),
	"pale_green":              rgb("#98FB98").toXY(),
	"dark_sea_green":          rgb("#8FBC8F").toXY(),
	"medium_spring_green":     rgb("#00FA9A").toXY(),
	"spring_green":            rgb("#00FF7F").toXY(),
	"sea_green":               rgb("#2E8B57").toXY(),
	"medium_aqua_marine":      rgb("#66CDAA").toXY(),
	"medium_sea_green":        rgb("#3CB371").toXY(),
	"light_sea_green":         rgb("#20B2AA").toXY(),
	"dark_slate_gray":         rgb("#2F4F4F").toXY(),
	"dark_cyan":               rgb("#008B8B").toXY(),
	"light_cyan":              rgb("#E0FFFF").toXY(),
	"dark_turquoise":          rgb("#00CED1").toXY(),
	"turquoise":               rgb("#40E0D0").toXY(),
	"medium_turquoise":        rgb("#48D1CC").toXY(),
	"pale_turquoise":          rgb("#AFEEEE").toXY(),
	"aqua_marine":             rgb("#7FFFD4").toXY(),
	"powder_blue":             rgb("#B0E0E6").toXY(),
	"cadet_blue":              rgb("#5F9EA0").toXY(),
	"steel_blue":              rgb("#4682B4").toXY(),
	"corn_flower_blue":        rgb("#6495ED").toXY(),
	"deep_sky_blue":           rgb("#00BFFF").toXY(),
	"dodger_blue":             rgb("#1E90FF").toXY(),
	"light_blue":              rgb("#ADD8E6").toXY(),
	"sky_blue":                rgb("#87CEEB").toXY(),
	"light_sky_blue":          rgb("#87CEFA").toXY(),
	"midnight_blue":           rgb("#191970").toXY(),
	"dark_blue":               rgb("#00008B").toXY(),
	"medium_blue":             rgb("#0000CD").toXY(),
	"royal_blue":              rgb("#4169E1").toXY(),
	"blue_violet":             rgb("#8A2BE2").toXY(),
	"indigo":                  rgb("#4B0082").toXY(),
	"dark_slate_blue":         rgb("#483D8B").toXY(),
	"slate_blue":              rgb("#6A5ACD").toXY(),
	"medium_slate_blue":       rgb("#7B68EE").toXY(),
	"medium_purple":           rgb("#9370DB").toXY(),
	"dark_magenta":            rgb("#8B008B").toXY(),
	"dark_violet":             rgb("#9400D3").toXY(),
	"dark_orchid":             rgb("#9932CC").toXY(),
	"medium_orchid":           rgb("#BA55D3").toXY(),
	"thistle":                 rgb("#D8BFD8").toXY(),
	"plum":                    rgb("#DDA0DD").toXY(),
	"violet":                  rgb("#EE82EE").toXY(),
	"orchid":                  rgb("#DA70D6").toXY(),
	"medium_violet_red":       rgb("#C71585").toXY(),
	"pale_violet_red":         rgb("#DB7093").toXY(),
	"deep_pink":               rgb("#FF1493").toXY(),
	"hot_pink":                rgb("#FF69B4").toXY(),
	"light_pink":              rgb("#FFB6C1").toXY(),
	"pink":                    rgb("#FFC0CB").toXY(),
	"antique_white":           rgb("#FAEBD7").toXY(),
	"beige":                   rgb("#F5F5DC").toXY(),
	"bisque":                  rgb("#FFE4C4").toXY(),
	"blanched_almond":         rgb("#FFEBCD").toXY(),
	"wheat":                   rgb("#F5DEB3").toXY(),
	"corn_silk":               rgb("#FFF8DC").toXY(),
	"lemon_chiffon":           rgb("#FFFACD").toXY(),
	"light golden rod yellow": rgb("#FAFAD2").toXY(),
	"light_yellow":            rgb("#FFFFE0").toXY(),
	"saddle_brown":            rgb("#8B4513").toXY(),
	"sienna":                  rgb("#A0522D").toXY(),
	"chocolate":               rgb("#D2691E").toXY(),
	"peru":                    rgb("#CD853F").toXY(),
	"sandy_brown":             rgb("#F4A460").toXY(),
	"burly_wood":              rgb("#DEB887").toXY(),
	"tan":                     rgb("#D2B48C").toXY(),
	"rosy_brown":              rgb("#BC8F8F").toXY(),
	"moccasin":                rgb("#FFE4B5").toXY(),
	"navajo_white":            rgb("#FFDEAD").toXY(),
	"peach_puff":              rgb("#FFDAB9").toXY(),
	"misty_rose":              rgb("#FFE4E1").toXY(),
	"lavender_blush":          rgb("#FFF0F5").toXY(),
	"linen":                   rgb("#FAF0E6").toXY(),
	"old_lace":                rgb("#FDF5E6").toXY(),
	"papaya_whip":             rgb("#FFEFD5").toXY(),
	"sea_shell":               rgb("#FFF5EE").toXY(),
	"mint_cream":              rgb("#F5FFFA").toXY(),
	"slate_gray":              rgb("#708090").toXY(),
	"light_slate_gray":        rgb("#778899").toXY(),
	"light_steel_blue":        rgb("#B0C4DE").toXY(),
	"lavender":                rgb("#E6E6FA").toXY(),
	"floral_white":            rgb("#FFFAF0").toXY(),
	"alice_blue":              rgb("#F0F8FF").toXY(),
	"ghost_white":             rgb("#F8F8FF").toXY(),
	"honeydew":                rgb("#F0FFF0").toXY(),
	"ivory":                   rgb("#FFFFF0").toXY(),
	"azure":                   rgb("#F0FFFF").toXY(),
	"snow":                    rgb("#FFFAFA").toXY(),
	"dim_gray":                rgb("#696969").toXY(),
	"dark_gray":               rgb("#A9A9A9").toXY(),
	"light_gray":              rgb("#D3D3D3").toXY(),
	"gainsboro":               rgb("#DCDCDC").toXY(),
	"white_smoke":             rgb("#F5F5F5").toXY(),
}

// GetSupportColorList returns the list of the supported colors contained in the Table map
func GetSupportColorList() []string {
	colors := make([]string, len(Table))
	i := 0
	for c := range Table {
		colors[i] = string(c)
		i++
	}
	return colors
}

// FindColorByName helps find a XY color by its name
func FindColorByName(name string) (*XY, error) {
	c := Table[RGBHex(name)]

	if c == nil {
		return nil, fmt.Errorf("color with name '%s' not found. Supported colors: \n%s", name, strings.Join(GetSupportColorList(), ", "))
	}

	return c, nil
}

// rgb Local helper function that returns a pointer to a RGBHex value
func rgb(v RGBHex) *RGBHex {
	return &v
}
