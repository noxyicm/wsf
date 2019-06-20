package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"wsf/config"
	"wsf/errors"
	"wsf/utils"
)

const (
	// TYPEFile represents file backend cache
	TYPEFile = "file"
)

func init() {
	Register(TYPEFile, NewFileBackendCache)
}

// File chache handler
type File struct {
	Backend
	Options *FileConfig
	mu      sync.Mutex
}

// Load stored data
func (b *File) Load(id string, testCacheValidity bool) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	filePath := b.Options.Dir + "/" + id + b.Options.Suffix
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "Load failed")
	}
	defer fd.Close()

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "Load failed")
	}

	d := make([]byte, fi.Size())
	n, err := fd.Read(d)
	if err != nil {
		return nil, errors.Wrap(err, "Load failed")
	}

	if n == 0 {
		return nil, errors.New("Load failed. File is empty")
	}

	return d, nil
}

// Test if key exists
func (b *File) Test(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	filePath := b.Options.Dir + "/" + id + b.Options.Suffix
	if _, err := os.Stat(filePath); err != nil {
		return false
	}

	return true
}

// Save data by key
func (b *File) Save(data []byte, id string, tags []string, specificLifetime int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	filePath := b.Options.Dir + "/" + id + b.Options.Suffix
	fd, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return errors.Wrap(err, "Save failed")
	}
	defer fd.Close()

	fd.Truncate(0)
	if _, err := fd.WriteAt(data, 0); err != nil {
		return errors.Wrap(err, "Save failed")
	}

	if len(tags) > 0 {
		tagsFilePath := b.Options.Dir + "/" + b.Options.TagsHolder + b.Options.Suffix
		tfd, err := os.OpenFile(tagsFilePath, os.O_RDWR|os.O_CREATE, 0664)
		if err != nil {
			return errors.Wrap(err, "Save failed")
		}
		defer tfd.Close()

		fi, err := os.Stat(tagsFilePath)
		if err != nil {
			return errors.Wrap(err, "Save failed")
		}

		d := make([]byte, fi.Size())
		n, err := tfd.Read(d)
		if err != nil {
			return errors.Wrap(err, "Save failed")
		}

		m := make(map[string][]string)
		if n > 0 {
			if err := json.Unmarshal(d, &m); err != nil {
				return errors.Wrap(err, "Save failed")
			}
		}

		for _, tag := range tags {
			if storedIDs, ok := m[tag]; ok {
				if !utils.InSSlice(id, storedIDs) {
					storedIDs = append(storedIDs, id)
					m[tag] = storedIDs
				}
			} else {
				m[tag] = []string{id}
			}
		}

		encoded, _ := json.Marshal(m)
		tfd.Truncate(0)
		if _, err := tfd.WriteAt(encoded, 0); err != nil {
			return errors.Wrap(err, "Save failed")
		}
	}

	return nil
}

// Remove data by key
func (b *File) Remove(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	filePath := b.Options.Dir + "/" + id + b.Options.Suffix
	if err := os.Remove(filePath); err != nil {
		return errors.Wrap(err, "Remove failed")
	}

	tagsFilePath := b.Options.Dir + "/" + b.Options.TagsHolder + b.Options.Suffix
	tfd, err := os.OpenFile(tagsFilePath, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return errors.Wrap(err, "Remove failed")
	}
	defer tfd.Close()

	fi, err := os.Stat(tagsFilePath)
	if err != nil {
		return errors.Wrap(err, "Remove failed")
	}

	d := make([]byte, fi.Size())
	n, err := tfd.Read(d)
	if err != nil {
		return err
	}

	m := make(map[string][]string)
	if n > 0 {
		if err := json.Unmarshal(d, &m); err != nil {
			return errors.Wrap(err, "Remove failed")
		}
	}

	for tag, storedIDs := range m {
		key, hasKey := utils.SKey(id, storedIDs)
		if hasKey {
			storedIDs = append(storedIDs[:key], storedIDs[key+1:]...)
			if len(storedIDs) > 0 {
				m[tag] = storedIDs
			}
		}
	}

	encoded, _ := json.Marshal(m)
	tfd.Truncate(0)
	if _, err := tfd.WriteAt(encoded, 0); err != nil {
		return errors.Wrap(err, "Remove failed")
	}

	return nil
}

// Clean stored data by tags
func (b *File) Clean(mode int, tags []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return nil
}

// NewFileBackendCache creates new file backend cache
func NewFileBackendCache(options config.Config) (bi Interface, err error) {
	b := &File{}

	cfg := &FileConfig{}
	cfg.Defaults()
	cfg.Populate(options)
	b.Options = cfg

	if b.Options.Dir, err = filepath.Abs(b.Options.Dir); err != nil {
		return nil, errors.Wrap(err, "Failed to create file backend cache object")
	}

	if _, err := os.Stat(b.Options.Dir); err != nil {
		if err := os.MkdirAll(b.Options.Dir, 0775); err != nil {
			return nil, errors.Wrap(err, "Failed to create file backend cache object")
		}
	}

	return b, nil
}
