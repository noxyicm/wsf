package loader

import (
	"fmt"
	"os"
	"time"
	"wsf/file"

	"github.com/pkg/errors"
)

// OsLoader structure
type OsLoader struct {
}

// NewOsLoader creates an os loader
func NewOsLoader() (FileLoader, error) {
	return &OsLoader{}, nil
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (fs *OsLoader) Create(name string) (file.File, error) {
	return os.Create(name)
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (fs *OsLoader) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (fs *OsLoader) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Open opens a file, returning it or an error, if any happens.
func (fs *OsLoader) Open(name string) (file.File, error) {
	fmt.Printf("Opening file %s \n", name)
	file, err := os.Open(name)
	if err != nil {
		err = errors.Wrap(err, "file.loader.OsLoader.Open()")
		return nil, err
	}

	fmt.Println("File opened")
	return file, nil
}

// OpenFile opens a file using the given flags and the given mode.
func (fs *OsLoader) OpenFile(name string, flag int, perm os.FileMode) (file.File, error) {
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		err = errors.Wrap(err, "file.loader.OsLoader.OpenFile()")
		return nil, err
	}

	return file, nil
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (fs *OsLoader) Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (fs *OsLoader) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Rename renames a file.
func (fs *OsLoader) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (fs *OsLoader) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Name The name of this FileSystem
func (fs *OsLoader) Name() string {
	return "osLoader"
}

//Chmod changes the mode of the named file to mode.
func (fs *OsLoader) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

//Chtimes changes the access and modification times of the named file
func (fs *OsLoader) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
