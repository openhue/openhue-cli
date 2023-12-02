package color

import (
	"errors"
	"math"
	"strconv"
)

// ToXY RGB to XY conversion
// Based on https://developers.meethue.com/develop/application-design-guidance/color-conversion-formulas-rgb-to-xy-and-back/
func (c *RGB) ToXY() *XY {

	red := 1.0 - (255-float32(c.Red))/255
	green := 1.0 - (255-float32(c.Green))/255
	blue := 1.0 - (255-float32(c.Blue))/255

	// Gamma correction
	gammaCorrection(&red)
	gammaCorrection(&green)
	gammaCorrection(&blue)

	// Convert the RGB values to XYZ using the Wide RGB D65 conversion formula
	X := red*0.4124 + green*0.3576 + blue*0.1805
	Y := red*0.2126 + green*0.7152 + blue*0.0722
	Z := red*0.0193 + green*0.1192 + blue*0.9505

	// Calculate the xy values from the XYZ values
	return &XY{
		X:          X / (X + Y + Z),
		Y:          Y / (X + Y + Z),
		Brightness: Y,
	}
}

// NewRGBFomHex Converts a hexadecimal string representation of a color to its RGB value.
// Example: #FF0000 will return RGB{255, 0, 0}
func NewRGBFomHex(hex string) (*RGB, error) {

	if !(hex[0:1] == "#") || len(hex) != 7 {
		return nil, errors.New("wrong format for the input hexadecimal string")
	}

	r, _ := strconv.ParseInt(hex[1:3], 16, 32)
	g, _ := strconv.ParseInt(hex[3:5], 16, 32)
	b, _ := strconv.ParseInt(hex[5:7], 16, 32)

	return &RGB{
		int(r),
		int(g),
		int(b),
	}, nil
}

func gammaCorrection(color *float32) {
	if *color > 0.04045 {
		*color = float32(math.Pow((float64(*color)+0.055)/(1.0+0.055), 2.4))
	} else {
		*color = *color / 12.92
	}
}
