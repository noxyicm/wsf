package file

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
)

// File represents singular upload file
type File struct {
	Name         string `json:"name"`
	Mime         string `json:"mime"`
	Size         int64  `json:"size"`
	Error        int    `json:"error"`
	TempFilename string `json:"tmpName"`
	header       *multipart.FileHeader
	received     bool
}

// Upload moves file content into temporary file
func (f *File) Upload(cfg *Config) error {
	if f.received {
		f.Error = UploadErrorUploded
		return errors.New("File '" + f.Name + "' already received")
	}

	if !cfg.Allowed(f.Name) {
		f.Error = UploadErrorExtension
		return errors.New("File '" + f.Name + "' has unsupported extension")
	}

	file, err := f.header.Open()
	if err != nil {
		f.Error = UploadErrorNoFile
		return fmt.Errorf("Unable to upload file '%s': %v", f.Name, err)
	}
	defer file.Close()

	tmp, err := ioutil.TempFile(cfg.TmpDir(), "")
	if err != nil {
		f.Error = UploadErrorNoTmpDir
		return fmt.Errorf("Unable to upload file '%s': %v", f.Name, err)
	}

	f.TempFilename = tmp.Name()
	defer tmp.Close()

	if f.Size, err = io.Copy(tmp, file); err != nil {
		f.Error = UploadErrorCantWrite
	}

	f.received = true
	return nil
}

// NewUpload wraps net/http upload into compatible structure
func NewUpload(f *multipart.FileHeader) *File {
	return &File{
		Name:   f.Filename,
		Mime:   f.Header.Get("Content-Type"),
		Error:  UploadErrorOK,
		header: f,
	}
}
