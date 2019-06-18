package reader

import (
	"fmt"
	"wsf/file"
)

// Reader Reads files
type Reader struct {
}

// NewReader Creates new file reader
func NewReader() (*Reader, error) {
	return &Reader{}, nil
}

// ReadAsByteArray reads files into byte array
func (r *Reader) ReadAsByteArray(file file.File, len int64) ([]byte, error) {
	fmt.Println("Reading file")
	data := make([]byte, len)
	_, err := file.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
