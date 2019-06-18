package file

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	// UploadErrorOK - no error, the file uploaded with success
	UploadErrorOK = 0

	// UploadErrorNoFile - no file was uploaded
	UploadErrorNoFile = 4

	// UploadErrorNoTmpDir - missing a temporary folder
	UploadErrorNoTmpDir = 5

	// UploadErrorCantWrite - failed to write file to disk
	UploadErrorCantWrite = 6

	// UploadErrorExtension - forbidden file extension
	UploadErrorExtension = 7
)

// Transfer manages files transfers
type Transfer struct {
	cfg  *Config
	tree Tree
	list []*File
}

// MarshalJSON marshal tree tree into JSON
func (t *Transfer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.tree)
}

// Upload moves all uploaded files to temp directory
func (t *Transfer) Upload() {
	var wg sync.WaitGroup
	for _, f := range t.list {
		wg.Add(1)
		go func(f *File) {
			defer wg.Done()
			f.Upload(t.cfg)
		}(f)
	}

	wg.Wait()
}

// Clear deletes all temporary files
func (t *Transfer) Clear() error {
	for _, f := range t.list {
		if f.TempFilename != "" && exists(f.TempFilename) {
			if err := os.Remove(f.TempFilename); err != nil {
				return err
			}
		}
	}

	return nil
}

// Append appends provided slice of files into transfer
func (t *Transfer) Append(files []*File) {
	t.list = append(t.list, files...)
}

// Push pushes provided slice of files into tree recursively
func (t *Transfer) Push(key string, files []*File) {
	t.tree.push(key, files)
}

// NewTransfer creates new file transfer
func NewTransfer(cfg *Config) (*Transfer, error) {
	t := &Transfer{
		cfg:  cfg,
		tree: make(Tree),
		list: make([]*File, 0),
	}

	return t, nil
}

// exists cheks if file exists
func exists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}

	return false
}
