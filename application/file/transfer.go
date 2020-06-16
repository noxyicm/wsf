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

	// UploadErrorUploded - file already received
	UploadErrorUploded = 8
)

// Transfer manages files transfers
type Transfer struct {
	cfg  *Config
	tree Tree
}

// MarshalJSON marshal tree into JSON
func (t *Transfer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.tree)
}

// Upload moves all uploaded files to temp directory
func (t *Transfer) Upload() error {
	var err error
	var wg sync.WaitGroup
	for _, v := range t.tree {
		if v, ok := v.(Tree); ok {
			for _, f := range v {
				if f, ok := f.(*File); ok {
					wg.Add(1)
					go func(f *File) {
						defer wg.Done()
						if er := f.Upload(t.cfg); er != nil {
							err = er
						}
					}(f)
				}
			}
		}
	}

	wg.Wait()
	return err
}

// Clear deletes all temporary files
func (t *Transfer) Clear() error {
	for _, v := range t.tree {
		if v, ok := v.(Tree); ok {
			for _, f := range v {
				if f, ok := f.(*File); ok && f.TempFilename != "" && exists(f.TempFilename) {
					if err := os.Remove(f.TempFilename); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Push pushes provided slice of files into tree recursively
func (t *Transfer) Push(key string, files *File) {
	t.tree.Push(key, files)
}

// Get retrives file from tree
func (t *Transfer) Get(key string) *File {
	return t.tree.Get(key)
}

// NewTransfer creates new file transfer
func NewTransfer(cfg *Config) (*Transfer, error) {
	t := &Transfer{
		cfg:  cfg,
		tree: make(Tree),
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
