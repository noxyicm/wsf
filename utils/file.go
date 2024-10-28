package utils

import (
	"os"
	"path/filepath"
)

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// WalkDirectoryDeep walks throught directory tree
func WalkDirectoryDeep(bPath string, rPath string, walkFn filepath.WalkFunc) error {
	wFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var name string
		name, err = filepath.Rel(bPath, path)
		if err != nil {
			return err
		}

		path = filepath.Join(rPath, name)
		if err == nil && info.Mode() == os.ModeSymlink {
			var symlinkPath string
			symlinkPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}

			info, err = os.Lstat(symlinkPath)
			if err != nil {
				return walkFn(path, info, err)
			}

			if info.IsDir() {
				return WalkDirectoryDeep(symlinkPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}

	err := filepath.Walk(bPath, wFunc)
	return err
}
