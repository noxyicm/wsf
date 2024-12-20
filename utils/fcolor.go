package utils

// RGBAF64 represents an RGBA color using the range 0.0 to 1.0 with a float64 for each channel.
type RGBAF64 struct {
	R, G, B, A float64
}

// NewRGBAF64 returns a new RGBAF64 color based on the provided uint8 values.
// uint8 value 0 maps to 0, 128 to 0.5 and 255 to 1.0.
func NewRGBAF64(r, g, b, a uint8) RGBAF64 {
	return RGBAF64{float64(r) / 255, float64(g) / 255, float64(b) / 255, float64(a) / 255}
}

// Clamp limits the channel values of the RGBAF64 color to the range 0.0 to 1.0.
func (c *RGBAF64) Clamp() {
	c.R = Clamp(c.R, 0, 1)
	c.G = Clamp(c.G, 0, 1)
	c.B = Clamp(c.B, 0, 1)
	c.A = Clamp(c.A, 0, 1)
}
