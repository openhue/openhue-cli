package color

import (
	"fmt"
	"testing"
)

func TestRGBToXY(t *testing.T) {

	red := RGB{
		255,
		0,
		0,
	}

	xy := red.ToXY()

	fmt.Println(xy)
}

func TestNewRGBFromHex(t *testing.T) {
	assertRGBFromHexString(t, "#330033", 51, 0, 51)
	assertRGBFromHexString(t, "#FF0000", 255, 0, 0)
	assertRGBFromHexString(t, "#00FF00", 0, 255, 0)
	assertRGBFromHexString(t, "#0000FF", 0, 0, 255)
	assertRGBFromHexString(t, "#FFFFFF", 255, 255, 255)
	assertRGBFromHexString(t, "#000000", 0, 0, 0)
}

func assertRGBFromHexString(t *testing.T, hexStr string, r int, g int, b int) {
	rgb, _ := NewRGBFomHex(hexStr)

	if rgb.Red != r {
		t.Fatalf("Red should be %d, obtained value is %d", r, rgb.Red)
	}
	if rgb.Green != g {
		t.Fatalf("Green should be %d, obtained value is %d", g, rgb.Green)
	}
	if rgb.Blue != b {
		t.Fatalf("Blue should be %d, obtained value is %d", b, rgb.Blue)
	}
}
