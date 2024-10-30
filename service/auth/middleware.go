package auth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	wsfhttp "github.com/noxyicm/wsf/service/http"
	"github.com/noxyicm/wsf/session"
)

const (
	// TYPEAuthMiddleware is a name of this middleware
	TYPEAuthMiddleware = "auth"
)

func init() {
	wsfhttp.RegisterMiddleware(TYPEAuthMiddleware, NewAuthMiddleware)
}

type AuthMiddleware struct {
	Options *wsfhttp.MiddlewareConfig
}

// Handle middleware
func (m *AuthMiddleware) Handle(s *wsfhttp.Service, r request.Interface, w response.Interface) bool {
	var tp string
	if utp, ok := m.Options.Params["type"]; ok {
		if tp, ok = utp.(string); !ok {
			return false
		}
	}

	var hdr string
	if uhdr, ok := m.Options.Params["header"]; ok {
		if hdr, ok = uhdr.(string); !ok {
			return false
		}
	}

	switch tp {
	case TYPENone:
		return false

	case TYPEBasic:
		return false

	case TYPEAccessToken:
		if tk := r.ParamString("access_token"); tk != "" {
			decoded, err := base64.URLEncoding.DecodeString(strings.TrimSpace(tk))
			if err != nil {
				ctx := context.WithValue(r.Context(), session.AutostartKey, false)
				ctx = context.WithValue(ctx, session.SetCookieKey, false)
				r.SetContext(ctx)
				return false
			}
			sid := string(decoded)
			ctx := context.WithValue(r.Context(), session.Instance().Options().SessionName, sid)
			ctx = context.WithValue(ctx, session.AutostartKey, false)
			ctx = context.WithValue(ctx, session.SetCookieKey, false)
			r.SetContext(ctx)
			return false
		}

		return false

	case TYPEBearerToken:
		a := r.Header(hdr)
		if a == "" {
			return false
		}

		if strings.Contains(a, "Bearer") {
			encoded := strings.TrimSpace(strings.Replace(a, "Bearer", "", 1))
			decoded, err := base64.URLEncoding.DecodeString(encoded)
			if err != nil {
				ctx := context.WithValue(r.Context(), session.AutostartKey, false)
				ctx = context.WithValue(ctx, session.SetCookieKey, false)
				r.SetContext(ctx)
				return false
			}

			sid := string(decoded)
			ctx := context.WithValue(r.Context(), session.Instance().Options().SessionName, sid)
			ctx = context.WithValue(ctx, session.AutostartKey, false)
			ctx = context.WithValue(ctx, session.SetCookieKey, false)
			r.SetContext(ctx)
			return false
		}

		return false
	}

	return false
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(cfg *wsfhttp.MiddlewareConfig) (mi wsfhttp.Middleware, err error) {
	c := &AuthMiddleware{}
	c.Options = cfg
	return c, nil
}
