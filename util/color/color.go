package color

type RGB struct {
	Red   int
	Green int
	Blue  int
}

type XY struct {
	X          float32
	Y          float32
	Brightness float32
}

var UndefinedColor = XY{
	X:          -1.0,
	Y:          -1.0,
	Brightness: -1.0,
}

// RGBHex Hexadecimal representation of an RGB color. Must start with '#'
type RGBHex string

// toXY converts a RGBHex value to its XY representation in CIE color space
func (c *RGBHex) toXY() *XY {
	hex, err := NewRGBFomHex(string(*c))
	if err != nil {
		return nil
	}
	return hex.ToXY()
}
