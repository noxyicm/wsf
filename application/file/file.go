package file

import (
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
}

// Upload moves file content into temporary file
func (f *File) Upload(cfg *Config) error {
	if !cfg.Allowed(f.Name) {
		f.Error = UploadErrorExtension
		return nil
	}

	file, err := f.header.Open()
	if err != nil {
		f.Error = UploadErrorNoFile
		return err
	}
	defer file.Close()

	tmp, err := ioutil.TempFile(cfg.TmpDir(), "upload")
	if err != nil {
		f.Error = UploadErrorNoTmpDir
		return err
	}

	f.TempFilename = tmp.Name()
	defer tmp.Close()

	if f.Size, err = io.Copy(tmp, file); err != nil {
		f.Error = UploadErrorCantWrite
	}

	return err
}

// NewUpload wraps net/http upload into PRS-7 compatible structure
func NewUpload(f *multipart.FileHeader) *File {
	return &File{
		Name:   f.Filename,
		Mime:   f.Header.Get("Content-Type"),
		Error:  UploadErrorOK,
		header: f,
	}
}
