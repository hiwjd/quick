package util

import "math"

func F322d(m float32) float32 {
	n := math.Round(float64(m)*100) / 100
	return float32(n)
}
