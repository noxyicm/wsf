package static

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
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
	Options *wsfhttp.MiddlewareConfig
}

// Handle middleware
func (m *StaticMiddleware) Handle(s *wsfhttp.Service, r request.Interface, w response.Interface) bool {
	var dr string
	if udr, ok := m.Options.Params["dir"]; ok {
		if dr, ok = udr.(string); !ok {
			return false
		}
	}

	var forbid []string
	if uforbid, ok := m.Options.Params["forbid"]; ok {
		if forbid, ok = uforbid.([]string); !ok {
			return false
		}
	}

	var always []string
	if ualways, ok := m.Options.Params["always"]; ok {
		if always, ok = ualways.([]string); !ok {
			return false
		}
	}

	fPath := path.Clean(r.PathInfo())
	_, err := filepath.Rel(dr, fPath)
	if err != nil {
		if m.AlwaysServe(fPath, always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}

	if m.AlwaysForbid(fPath, forbid) {
		w.SetResponseCode(404)
		w.AppendBody([]byte("This page does not exists"), "")
		w.Write()
		return true
	}

	root := http.Dir(dr)
	f, err := root.Open(fPath)
	if err != nil {
		if m.AlwaysServe(fPath, always) {
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
		if m.AlwaysServe(fPath, always) {
			w.SetResponseCode(404)
			w.AppendBody([]byte("This page does not exists"), "")
			w.Write()
			return true
		}

		return false
	}

	if d.IsDir() {
		if m.AlwaysServe(fPath, always) {
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
	c := &StaticMiddleware{}
	c.Options = cfg
	return c, nil
}
