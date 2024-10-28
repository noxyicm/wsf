package resource

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/session"
)

// TYPESession id of resource
const TYPESession = "session"

func init() {
	Register(TYPESession, NewSessionResource)
}

// NewSessionResource creates a new resource of type Session
func NewSessionResource(cfg config.Config) (Interface, error) {
	handlerType := cfg.GetString("type")
	ses, err := session.NewSessionManager(handlerType, cfg)
	if err != nil {
		return nil, err
	}

	session.SetInstance(ses)
	return ses, nil
}
