package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/noxyicm/wsf/config"
	wsfctx "github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/service"
	wsfhttp "github.com/noxyicm/wsf/service/http"
	"github.com/noxyicm/wsf/session"
)

const (
	// ID of service
	ID = "auth"

	// TYPENone type represents no authorization
	TYPENone = "none"

	// TYPEBasic type represents basic authorization
	TYPEBasic = "basic"

	// TYPEAccessToken type represents authorization by access token
	TYPEAccessToken = "accessToken"

	// TYPEBearerToken type represents authorization by bearer token
	TYPEBearerToken = "bearerToken"
)

// Auth provides server specofoc authorization
type Auth struct {
	options  *Config
	priority int
}

// Init Auth service
func (s *Auth) Init(options *Config, h *wsfhttp.Service) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	if h == nil {
		return false, nil
	}

	s.options = options
	return true, nil
}

// AddListener attaches server event watcher
func (s *Auth) AddListener(l func(event int, ctx service.Event)) {
	return
}

// Priority returns predefined service priority
func (s *Auth) Priority() int {
	return s.priority
}

// Serve the service
func (s *Auth) Serve(ctx wsfctx.Context) (err error) {
	return nil
}

// Stop the service
func (s *Auth) Stop() {
	return
}

// middleware must return true if request/response pair is handled within the middleware
func (s *Auth) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, rqs *http.Request) {
		f(w, s.handleAuth(w, rqs))
	}
}

func (s *Auth) handleAuth(w http.ResponseWriter, rqs *http.Request) *http.Request {
	switch s.options.Type {
	case TYPENone:
		return rqs

	case TYPEBasic:
		return rqs

	case TYPEAccessToken:
		if tk := rqs.FormValue("access_token"); tk != "" {
			decoded, err := base64.URLEncoding.DecodeString(strings.TrimSpace(tk))
			if err != nil {
				ctx := context.WithValue(rqs.Context(), session.AutostartKey, false)
				ctx = context.WithValue(ctx, session.SetCookieKey, false)
				return rqs.Clone(ctx)
			}
			sid := string(decoded)
			ctx := context.WithValue(rqs.Context(), session.Instance().Options().SessionName, sid)
			ctx = context.WithValue(ctx, session.AutostartKey, false)
			ctx = context.WithValue(ctx, session.SetCookieKey, false)
			return rqs.Clone(ctx)
		}

		return rqs

	case TYPEBearerToken:
		a := rqs.Header.Get(s.options.Header)
		if a == "" {
			return rqs
		}

		if strings.Contains(a, "Bearer") {
			encoded := strings.TrimSpace(strings.Replace(a, "Bearer", "", 1))
			decoded, err := base64.URLEncoding.DecodeString(encoded)
			if err != nil {
				ctx := context.WithValue(rqs.Context(), session.AutostartKey, false)
				ctx = context.WithValue(ctx, session.SetCookieKey, false)
				return rqs.Clone(ctx)
			}

			sid := string(decoded)
			ctx := context.WithValue(rqs.Context(), session.Instance().Options().SessionName, sid)
			ctx = context.WithValue(ctx, session.AutostartKey, false)
			ctx = context.WithValue(ctx, session.SetCookieKey, false)
			return rqs.Clone(ctx)
		}

		return rqs
	}

	return rqs
}

// NewService creates a new service of type Auth
func NewService(options config.Config) (service.Interface, error) {
	return &Auth{
		priority: 3,
	}, nil
}
