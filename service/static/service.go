package static

import (
	"net/http"
	"path"
	"path/filepath"
	"wsf/config"
	"wsf/context"
	"wsf/service"
	wsfhttp "wsf/service/http"
)

// ID of service
const ID = "static"

// Service serves static files
type Service struct {
	options  *Config
	root     http.Dir
	priority int
}

// Init Static service
func (s *Service) Init(options *Config, h *wsfhttp.Service) (bool, error) {
	if h == nil {
		return false, nil
	}

	s.options = options
	s.root = http.Dir(s.options.Dir)
	h.AddMiddleware(s.middleware)
	return true, nil
}

// AddListener attaches server event watcher
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	return
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// Serve the service
func (s *Service) Serve(ctx context.Context) (err error) {
	return nil
}

// Stop the service
func (s *Service) Stop() {
	return
}

// middleware must return true if request/response pair is handled within the middleware
func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.handleStatic(w, r) {
			f(w, r)
		}
	}
}

func (s *Service) handleStatic(w http.ResponseWriter, r *http.Request) bool {
	fPath := path.Clean(r.URL.Path)

	_, err := filepath.Rel(s.options.Dir, fPath)
	if err != nil {
		return false
	}

	if s.options.AlwaysForbid(fPath) {
		return false
	}

	f, err := s.root.Open(fPath)
	if err != nil {
		if s.options.AlwaysServe(fPath) {
			w.WriteHeader(404)
			return true
		}

		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return false
	}

	if d.IsDir() {
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}

// NewService creates a new service of type Static
func NewService(options config.Config) (service.Interface, error) {
	return &Service{
		priority: 2,
	}, nil
}
