package http

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/noxyicm/wsf/application/file"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPEUploadsMiddleware is a name of this middleware
	TYPEUploadsMiddleware = "uploads"

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

var braketsPattern = regexp.MustCompile(`\[(.)+\]`)

func init() {
	RegisterMiddleware(TYPEUploadsMiddleware, NewUploadsMiddleware)
}

type UploadsMiddleware struct {
	Options         *MiddlewareConfig
	FTType          string
	Location        *regexp.Regexp
	Directory       string
	FileNamePattern string
	Allowed         []string
	MaxMemory       int64
	MaxSize         int64
	StoreAccess     fs.FileMode
	PassArgs        bool
}

// Init initializes middleware
func (m *UploadsMiddleware) Init(options *MiddlewareConfig) (bool, error) {
	m.Options = options

	if utp, ok := m.Options.Params["file_transfer_type"]; ok {
		if m.FTType, ok = utp.(string); !ok {
			m.FTType = file.TYPEDefaultTransfer
		}
	}

	var location string
	if ulocation, ok := m.Options.Params["location"]; ok {
		if location, ok = ulocation.(string); !ok {
			return false, errors.New("Location pattern must be specified")
		}
	}
	m.Location = regexp.MustCompile(location)

	if ufnp, ok := m.Options.Params["file_name_pattern"]; ok {
		if m.FileNamePattern, ok = ufnp.(string); !ok {
			m.FileNamePattern = ""
		}
	}

	if udr, ok := m.Options.Params["dir"]; ok {
		if m.Directory, ok = udr.(string); !ok || m.Directory == "" {
			return false, errors.New("Directory must be specified")
		}

		dirInfo, err := os.Stat(m.Directory)
		if err != nil {
			return false, errors.Errorf("Error accessing directory %s", m.Directory)
		}
		mode := dirInfo.Mode()
		if mode&os.ModeDir != os.ModeDir {
			return false, errors.Errorf("%s is not a directory", m.Directory)
		} else if mode&OS_USER_W != OS_USER_W || mode&OS_GROUP_W != OS_GROUP_W {
			return false, errors.Errorf("Directory %s is not writable", m.Directory)
		}
	}

	if uallowed, ok := m.Options.Params["allowed"]; ok {
		switch uallowed.(type) {
		case []string:
			m.Allowed = uallowed.([]string)

		case []interface{}:
			for _, uv := range uallowed.([]interface{}) {
				switch v := uv.(type) {
				case string:
					m.Allowed = append(m.Allowed, v)
				}
			}
		}
	}

	if umms, ok := m.Options.Params["max_memory_size"]; ok {
		switch mms := umms.(type) {
		case string:
			var b int64
			var bs string
			n, err := fmt.Sscanf(mms, "%d%s", &b, &bs)
			if n == 0 || err != nil {
				m.MaxMemory = 5 * 1024 * 1024
			} else {
				switch strings.ToLower(bs) {
				case "kb":
					m.MaxMemory = b * 1024

				case "mb":
					m.MaxMemory = b * 1024 * 1024

				case "gb":
					m.MaxMemory = b * 1024 * 1024 * 1024

				case "tb":
					m.MaxMemory = b * 1024 * 1024 * 1024 * 1024
				}
			}

		case int:
			m.MaxMemory = int64(mms)

		case int64:
			m.MaxMemory = mms
		}
	}

	if ums, ok := m.Options.Params["max_file_size"]; ok {
		switch ms := ums.(type) {
		case string:
			var b int64
			var bs string
			n, err := fmt.Sscanf(ms, "%d%s", &b, &bs)
			if n == 0 || err != nil {
				m.MaxSize = 5 * 1024 * 1024
			} else {
				switch strings.ToLower(bs) {
				case "kb":
					m.MaxSize = b * 1024

				case "mb":
					m.MaxSize = b * 1024 * 1024

				case "gb":
					m.MaxSize = b * 1024 * 1024 * 1024

				case "tb":
					m.MaxSize = b * 1024 * 1024 * 1024 * 1024
				}
			}

		case int:
			m.MaxSize = int64(ms)

		case int64:
			m.MaxSize = ms
		}
	}

	var sa string
	pat := regexp.MustCompile(`user:([rwx]+)? group:([rwx]+)? other:([rwx]+)?`)
	if usa, ok := m.Options.Params["store_access"]; ok {
		if sa, ok = usa.(string); ok {
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

				m.StoreAccess = os.FileMode(prm)
			} else {
				m.StoreAccess = os.FileMode(0664)
			}
		} else {
			m.StoreAccess = os.FileMode(0664)
		}
	}

	return true, nil
}

// Handle middleware
func (m *UploadsMiddleware) Handle(s *Service, r request.Interface, w response.Interface) bool {
	if !m.Location.MatchString(r.PathInfo()) {
		return false
	}

	var rqs *request.HTTP
	var ok bool
	if rqs, ok = r.(*request.HTTP); !ok {
		return false
	}

	if !rqs.IsPost() {
		return false
	}

	ft, err := file.NewTransferFromConfig(m.FTType, &file.Config{
		Type:            m.FTType,
		Directory:       m.Directory,
		FileNamePattern: m.FileNamePattern,
		Allowed:         m.Allowed,
		MaxMemory:       m.MaxMemory,
		MaxSize:         m.MaxSize,
		StoreAccess:     m.StoreAccess,
	})
	if err != nil {
		s.Logger.Warning(errors.Wrap(err, "Unable to create file transfer instance"), map[string]string{})
		return false
	}

	r.SetFileTransfer(ft)

	return false
}

// AlwaysForbid must return true if file extension is not allowed for the upload
func (m *UploadsMiddleware) AlwaysForbid(filename string, s []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	return utils.InSSlice(ext[1:], s)
}

// AlwaysServe must indicate that file is expected to be served by static service
func (m *UploadsMiddleware) AlwaysServe(filename string, s []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	return utils.InSSlice(ext[1:], s)
}

// NewUploadsMiddleware creates new uploads middleware
func NewUploadsMiddleware(cfg *MiddlewareConfig) (mi Middleware, err error) {
	c := &UploadsMiddleware{
		Allowed: make([]string, 0),
	}
	c.Options = cfg
	return c, nil
}
