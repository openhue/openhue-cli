package color

import "testing"

func TestConvertHexToXY(t *testing.T) {

	var c RGBHex = "#123456"
	xy := c.toXY()
	if xy == nil {
		t.Fatalf("unable to convert %s to XY", c)
	}
}

func TestFailToConvertHexToXY(t *testing.T) {

	var c RGBHex = "#foo"
	xy := c.toXY()
	if xy != nil {
		t.Fatalf("it should not be possible to convert %s to XY", c)
	}
}
