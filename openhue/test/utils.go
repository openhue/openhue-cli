package test

import (
	"math"
)

// AlmostEqual32 returns true if a and b are equal within a relative error of e.
// See http://floating-point-gui.de/errors/comparison/ for the details of the applied method.
func AlmostEqual32(a, b, e float32) bool {

	MinNormal32 := math.Float32frombits(0x00800000)

	if a == b {
		return true
	}
	absA := Abs32(a)
	absB := Abs32(b)
	diff := Abs32(a - b)
	if a == 0 || b == 0 || absA+absB < MinNormal32 {
		return diff < e*MinNormal32
	}
	return diff/Min32(absA+absB, math.MaxFloat32) < e
}

// Abs32 works like math.Abs, but for float32.
func Abs32(x float32) float32 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

// Min32 works like math.Min, but for float32.
func Min32(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}
