package locale

import "wsf/errors"

// Encoding specifies encoding of the input data.
type Encoding uint

const (
	// UTF8 interprets the input data as UTF-8.
	UTF8 = iota

	// ISO88591 interprets the input data as ISO-8859-1.
	ISO88591
)

// Convert Interprets a byte buffer either as an ISO-8859-1 or UTF-8 encoded string.
// For ISO-8859-1 we can convert each byte straight into a rune since the
// first 256 unicode code points cover ISO-8859-1.
func Convert(buf []byte, enc Encoding) (string, error) {
	switch enc {
	case UTF8:
		return string(buf), nil

	case ISO88591:
		runes := make([]rune, len(buf))
		for i, b := range buf {
			runes[i] = rune(b)
		}
		return string(runes), nil

	default:
		return "", errors.Errorf("Unsupported encoding %v", enc)
	}
}
