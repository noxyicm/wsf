package utils

// Clamp is a clamp
func Clamp(value, min, max float64) float64 {
	if value > max {
		return max
	}

	if value < min {
		return min
	}

	return value
}
