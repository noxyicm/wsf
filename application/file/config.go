package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/utils"
)

const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)

// Config represents file configuration
type Config struct {
	Type            string
	Directory       string
	FileNamePattern string
	Allowed         []string
	MaxMemory       int64
	MaxSize         int64
	StoreAccess     fs.FileMode
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	c.Directory = cfg.GetString("dir")
	if c.Directory != "" {
		dirInfo, err := os.Stat(filepath.Join(config.StaticPath, c.Directory))
		if err != nil {
			c.Directory = ""
			log.Warning(fmt.Sprintf("Error accessing directory %s", filepath.Join(config.StaticPath, c.Directory)), map[string]string{})
		} else {
			mode := dirInfo.Mode()
			if mode&os.ModeDir != os.ModeDir {
				log.Warning(fmt.Sprintf("%s is not a directory", filepath.Join(config.StaticPath, c.Directory)), map[string]string{})
			} else if mode&OS_USER_W != OS_USER_W || mode&OS_GROUP_W != OS_GROUP_W {
				log.Warning(fmt.Sprintf("Directory %s is not writable", filepath.Join(config.StaticPath, c.Directory)), map[string]string{})
			}
		}
	}

	c.Allowed = cfg.GetStringSlice("allowed")

	mms := cfg.GetString("max_memory_size")
	if mms != "" {
		var b int64
		var bs string
		n, err := fmt.Sscanf(mms, "%d%s", &b, &bs)
		if n == 0 || err != nil {
			c.MaxMemory = 5 * 1024 * 1024
		} else {
			switch strings.ToLower(bs) {
			case "kb":
				c.MaxMemory = b * 1024

			case "mb":
				c.MaxMemory = b * 1024 * 1024

			case "gb":
				c.MaxMemory = b * 1024 * 1024 * 1024

			case "tb":
				c.MaxMemory = b * 1024 * 1024 * 1024 * 1024
			}
		}
	}

	ms := cfg.GetString("max_file_size")
	if ms != "" {
		var b int64
		var bs string
		n, err := fmt.Sscanf(ms, "%d%s", &b, &bs)
		if n == 0 || err != nil {
			c.MaxSize = 5 * 1024 * 1024
		} else {
			switch strings.ToLower(bs) {
			case "kb":
				c.MaxSize = b * 1024

			case "mb":
				c.MaxSize = b * 1024 * 1024

			case "gb":
				c.MaxSize = b * 1024 * 1024 * 1024

			case "tb":
				c.MaxSize = b * 1024 * 1024 * 1024 * 1024
			}
		}
	}

	pat := regexp.MustCompile(`user:([rwx]+)? group:([rwx]+)? other:([rwx]+)?`)
	sa := cfg.GetString("store_access")
	if sa != "" {
		matches := pat.FindAllStringSubmatch(sa, -1)
		if len(matches) == 1 {
			perms := make([]string, 10)
			for i := range perms {
				perms[i] = "-"
			}

			var prm uint
			for n := 1; n <= 3; n++ {
				var shift uint
				if n == 1 {
					shift = OS_USER_SHIFT
				} else if n == 2 {
					shift = OS_GROUP_SHIFT
				} else {
					shift = OS_OTH_SHIFT
				}

				for i := range matches[0][n] {
					switch matches[0][n][i] {
					case 'r':
						perms[(n-1)*3+1] = "r"
						prm = prm | OS_READ<<shift

					case 'w':
						perms[(n-1)*3+2] = "w"
						prm = prm | OS_WRITE<<shift

					case 'x':
						perms[(n-1)*3+3] = "x"
						prm = prm | OS_EX<<shift
					}
				}
			}

			c.StoreAccess = os.FileMode(prm)
		} else {
			c.StoreAccess = os.FileMode(0664)
		}
	} else {
		c.StoreAccess = os.FileMode(0664)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = TYPEDefaultTransfer
	c.Directory = ""
	c.FileNamePattern = ""
	c.Allowed = []string{}
	c.MaxMemory = 1 << 26
	c.MaxSize = 1 << 26 // 64Mb
	c.StoreAccess = os.FileMode(0664)
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}

// TmpDir returns temporary directory
func (c *Config) TmpDir() string {
	if c.Directory != "" {
		return filepath.Join(config.StaticPath, c.Directory)
	}

	return os.TempDir()
}

// Allowed returns true if file allowed to be transfered
func (c *Config) IsAllowed(filename string) bool {
	if len(c.Allowed) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return utils.InSSlice(ext[1:], c.Allowed)
}
