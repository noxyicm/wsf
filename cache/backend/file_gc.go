package backend

import (
	"container/list"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
)

// FileGC is a ttl watcher
type FileGC struct {
	Options  *FileConfig
	StopChan chan bool
	ErrChan  chan error
	Removed  []string
	Logger   *log.Log
	List     *list.List
	Table    map[string]*list.Element
	MaxLen   int

	mu sync.Mutex
}

// Init the file gc
func (g *FileGC) Init(options config.Config) (bool, error) {
	cfg := &FileConfig{}
	cfg.Defaults()
	cfg.Populate(options)
	g.Options = cfg

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("Log resource is not configured")
	}

	g.Logger = logResource.(*log.Log)
	g.Start()
	return true, nil
}

// Start the gc
func (g *FileGC) Start() {
	go g.startRoutine()
	return
}

// Check filepath, read file and remove it if needed
func (g *FileGC) Check(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("scanning source '%s' failed: %v", path, err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	fd, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "check failed")
	}
	defer fd.Close()

	if g.Options.TagsHolder != "" && fd.Name() == g.Options.Dir+"/"+g.Options.TagsHolder+g.Options.Suffix {
		return nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, "check failed")
	}

	d := make([]byte, fi.Size())
	n, err := fd.Read(d)
	if err != nil {
		return errors.Wrap(err, "check failed")
	}

	if n == 0 {
		if err := os.Remove(fd.Name()); err != nil {
			return errors.Wrapf(err, "unable to remove file '%s'", fd.Name())
		}

		g.Removed = append(g.Removed, fd.Name())
		return nil
	}

	fdt := FileData{}
	if err := json.Unmarshal(d, &fdt); err != nil {
		return errors.Wrap(err, "unable to deserialize data")
	}

	if fdt.Expires != 0 && time.Now().After(time.Unix(fdt.Expires, 0)) {
		if err := os.Remove(fd.Name()); err != nil {
			return errors.Wrapf(err, "unable to remove file '%s'", fd.Name())
		}

		g.Removed = append(g.Removed, fd.Name())
		return nil
	}

	return nil
}

func (g *FileGC) startRoutine() {
Mainloop:
	for {
		select {
		case <-g.StopChan:
			break Mainloop
		case <-time.After(time.Duration(g.Options.GC) * time.Second):
			g.mu.Lock()
			g.Removed = []string{}

			if _, err := os.Stat(g.Options.Dir); !os.IsNotExist(err) {
				err := filepath.Walk(g.Options.Dir, g.Check)
				if err != nil {
					g.Logger.Warningf("[File] Unable to process file: %v", nil, err.Error())
				}

				if err := g.clearRemoved(); err != nil {
					g.Logger.Warning(err.Error(), nil)
				}
			}

			g.mu.Unlock()
		}
	}

	return
}

func (g *FileGC) clearRemoved() error {
	if g.Options.TagsHolder != "" {
		tagsFilePath := g.Options.Dir + "/" + g.Options.TagsHolder + g.Options.Suffix
		tfd, err := os.OpenFile(tagsFilePath, os.O_RDWR|os.O_CREATE, 0664)
		if err != nil {
			return errors.Wrapf(err, "[File] Unable to open file: %s", tagsFilePath)
		}
		defer tfd.Close()

		fi, err := os.Stat(tagsFilePath)
		if err != nil {
			return errors.Wrapf(err, "[File] Unable to get stats for file: %s", tagsFilePath)
		}

		d := make([]byte, fi.Size())
		n, err := tfd.Read(d)
		if err != nil {
			return errors.Wrapf(err, "[File] Unable to read file: %s", tagsFilePath)
		}

		m := make(map[string][]string)
		if n > 0 {
			if err := json.Unmarshal(d, &m); err != nil {
				return errors.Wrapf(err, "[File] Unable to unmarshal file: %s", tagsFilePath)
			}
		}

		for _, fname := range g.Removed {
			id := strings.Replace(fname, g.Options.Suffix, "", -1)
			for tag, storedIDs := range m {
				key, hasKey := utils.SKey(id, storedIDs)
				if hasKey {
					storedIDs = append(storedIDs[:key], storedIDs[key+1:]...)
					if len(storedIDs) > 0 {
						m[tag] = storedIDs
					}
				}
			}
		}

		encoded, _ := json.Marshal(m)
		tfd.Truncate(0)
		if _, err := tfd.WriteAt(encoded, 0); err != nil {
			return errors.Wrapf(err, "[File] Unable to write file: %s", tagsFilePath)
		}
	}

	return nil
}

// NewFileGC creates a new file gc instance
func NewFileGC(options *FileConfig) (*FileGC, error) {
	g := &FileGC{
		StopChan: make(chan bool, 1),
		ErrChan:  make(chan error, 1),
		Removed:  make([]string, 0),
	}
	g.Options = options
	return g, nil
}
