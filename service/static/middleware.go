package static

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	wsfhttp "github.com/noxyicm/wsf/service/http"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPEStaticMiddleware is a name of this middleware
	TYPEStaticMiddleware = "static"
)

func init() {
	wsfhttp.RegisterMiddleware(TYPEStaticMiddleware, NewStaticMiddleware)
}

type StaticMiddleware struct {
	Options   *wsfhttp.MiddlewareConfig
	Directory string
	Forbiden  []string
	Always    []string
}

// Init initializes middleware
func (m *StaticMiddleware) Init(options *wsfhttp.MiddlewareConfig) (bool, error) {
	m.Options = options

	if udr, ok := m.Options.Params["dir"]; ok {
		if m.Directory, ok = udr.(string); !ok || m.Directory == "" {
			return false, errors.New("Directory must be specified")
		}
	}

	if uforbid, ok := m.Options.Params["forbid"]; ok {
		switch uforbid.(type) {
		case []string:
			m.Forbiden = uforbid.([]string)

		case []interface{}:
			for _, uv := range uforbid.([]interface{}) {
				switch v := uv.(type) {
				case string:
					m.Forbiden = append(m.Forbiden, v)
				}
			}
		}
	}

	if ualways, ok := m.Options.Params["always"]; ok {
		switch ualways.(type) {
		case []string:
			m.Always = ualways.([]string)

		case []interface{}:
			for _, uv := range ualways.([]interface{}) {
				switch v := uv.(type) {
				case string:
					m.Always = append(m.Always, v)
				}
			}
		}
	}
	return true, nil
}

// Handle middleware
func (m *StaticMiddleware) Handle(s *wsfhttp.Service, r request.Interface, w response.Interface) bool {
	fPath := path.Clean(r.PathInfo())
	_, err := filepath.Rel(m.Directory, fPath)
	if err != nil {
		if m.AlwaysServe(fPath, m.Always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}

	if m.AlwaysForbid(fPath, m.Forbiden) {
		w.SetResponseCode(404)
		w.AppendBody([]byte("This page does not exists"), "")
		w.Write()
		return true
	}

	root := http.Dir(m.Directory)
	f, err := root.Open(fPath)
	if err != nil {
		if m.AlwaysServe(fPath, m.Always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		if m.AlwaysServe(fPath, m.Always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}

	if d.IsDir() {
		if m.AlwaysServe(fPath, m.Always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}

	http.ServeContent(w.GetWriter(), r.GetRequest(), d.Name(), d.ModTime(), f)
	return true
}

// AlwaysForbid must return true if file extension is not allowed for the upload
func (m *StaticMiddleware) AlwaysForbid(filename string, s []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	return utils.InSSlice(ext[1:], s)
}

// AlwaysServe must indicate that file is expected to be served by static service
func (m *StaticMiddleware) AlwaysServe(filename string, s []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}

	return utils.InSSlice(ext[1:], s)
}

// NewStaticMiddleware creates new static middleware
func NewStaticMiddleware(cfg *wsfhttp.MiddlewareConfig) (mi wsfhttp.Middleware, err error) {
	c := &StaticMiddleware{
		Forbiden: make([]string, 0),
		Always:   make([]string, 0),
	}
	c.Options = cfg
	return c, nil
}
