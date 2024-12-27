package file

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
)

// File represents singular upload file
type File struct {
	Name         string `json:"name"`
	Mime         string `json:"mime"`
	Size         int64  `json:"size"`
	Error        error  `json:"error"`
	TempFilename string `json:"tmpName"`
	body         io.Reader
	received     bool
}

// Upload moves file content into temporary file
func (f *File) Upload(cfg *Config) error {
	if f.received {
		return errors.Errorf("File '%s' already received", f.Name)
	}

	if !cfg.IsAllowed(f.Name) {
		f.Error = errors.Errorf("File '%s' has unsupported extension", f.Name)
		return f.Error
	}

	tmp, err := os.CreateTemp(cfg.TmpDir(), cfg.FileNamePattern)
	if err != nil {
		f.Error = errors.Wrapf(err, "Unable to upload file '%s'", f.Name)
		return f.Error
	}

	f.TempFilename = strings.TrimPrefix(tmp.Name(), config.StaticPath)
	defer tmp.Close()

	buf := bufio.NewReader(f.body)
	sniff, _ := buf.Peek(256)
	f.Mime = http.DetectContentType(sniff)

	reader := io.MultiReader(buf, f.body)
	lmt := io.LimitReader(reader, cfg.MaxSize)
	written, err := io.Copy(tmp, lmt)
	if err != nil && err != io.EOF {
		f.Error = errors.Wrapf(err, "Unable to upload file '%s'", f.Name)
		return f.Error
	} else if written > cfg.MaxSize {
		os.Remove(tmp.Name())
		f.Error = errors.Errorf("Unable to upload file '%s'. File is too large", f.Name)
		return f.Error
	}

	f.Size = written
	f.received = true
	return nil
}

// NewUpload wraps net/http upload into compatible structure
func NewUpload(name string, r io.Reader) *File {
	return &File{
		Name: name,
		body: r,
	}
}
