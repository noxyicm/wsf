package file

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
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

	// TYPEDefaultTransfer is the name of transfer
	TYPEDefaultTransfer = "default"
)

var (
	buildHandlers = map[string]func(*Config) (TransferInterface, error){}
)

func init() {
	RegisterTransfer(TYPEDefaultTransfer, NewDefaultTransfer)
}

// Interface is an interface for controllers
type TransferInterface interface {
	UploadFile(file *File) error
	Add(file *File) error
	Has(key string) bool
	Get(name string) *File
	Remove(file *File) error
	Clear() error
	Upload() error
	Uploaded() bool
	MarshalJSON() ([]byte, error)
}

// Transfer manages files transfers
type Transfer struct {
	Options  *Config
	files    map[string]*File
	uploaded bool
	mur      sync.RWMutex
	wg       sync.WaitGroup
}

// MarshalJSON marshal tree into JSON
func (t *Transfer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.files)
}

// UploadFile performs a upload of a file
func (t *Transfer) UploadFile(file *File) (err error) {
	return file.Upload(t.Options)
}

// Upload moves all uploaded files to temp directory
func (t *Transfer) Upload() error {
	var err error
	if t.uploaded {
		return errors.New("Files already uploaded")
	}

	for k, v := range t.files {
		go t.tryUploadFile(v, k)
	}

	t.wg.Wait()
	t.uploaded = true
	return err
}

func (t *Transfer) tryUploadFile(f *File, key string) (err error) {
	t.wg.Add(1)
	defer t.wg.Done()

	if er := f.Upload(t.Options); er != nil {
		err = er
	}

	return err
}

// Clear deletes all temporary files
func (t *Transfer) Clear() error {
	for _, v := range t.files {
		if v.TempFilename != "" && exists(v.TempFilename) {
			if err := os.Remove(v.TempFilename); err != nil {
				return err
			}
		}
	}

	t.files = make(map[string]*File)
	t.uploaded = false
	return nil
}

// Uploaded returns keys of uploaded files
func (t *Transfer) Uploaded() bool {
	return t.uploaded
}

// Has returns true if transfer contains a key
func (t *Transfer) Has(key string) bool {
	if _, ok := t.files[key]; ok {
		return true
	}

	return false
}

// Add adds file to file transfer
func (t *Transfer) Add(file *File) error {
	t.files[file.Name] = file
	t.uploaded = false
	return nil
}

// Get retrives file
func (t *Transfer) Get(key string) *File {
	if v, ok := t.files[key]; ok {
		return v
	}

	return nil
}

func (t *Transfer) Remove(file *File) error {
	for k, v := range t.files {
		if v.TempFilename != "" && exists(v.TempFilename) {
			if err := os.Remove(v.TempFilename); err != nil {
				return err
			}
		}

		delete(t.files, k)
		t.uploaded = false
		return nil
	}

	return errors.New("No such file")
}

// NewDefaultTransfer creates new default file transfer
func NewDefaultTransfer(cfg *Config) (TransferInterface, error) {
	t := &Transfer{
		Options:  cfg,
		files:    make(map[string]*File),
		uploaded: false,
	}

	return t, nil
}

// NewController creates a new controller specified by type
func NewTransfer(transferType string, options config.Config) (TransferInterface, error) {
	cfg := &Config{}
	cfg.Defaults()
	if err := cfg.Populate(options); err != nil {
		return nil, errors.Wrap(err, "Unable to create file.Transfer")
	}

	if transferType == "" {
		transferType = TYPEDefaultTransfer
	}

	if f, ok := buildHandlers[transferType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized file transfer type \"%v\"", transferType)
}

// NewController creates a new controller specified by type
func NewTransferFromConfig(transferType string, cfg *Config) (TransferInterface, error) {
	if f, ok := buildHandlers[transferType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized file transfer type \"%v\"", transferType)
}

// RegisterTransfer registers a handler for file transfer creation
func RegisterTransfer(transferType string, handler func(*Config) (TransferInterface, error)) {
	buildHandlers[transferType] = handler
}

// exists cheks if file exists
func exists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}

	return false
}
