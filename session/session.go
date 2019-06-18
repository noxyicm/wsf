package session

import (
	"wsf/config"
	"wsf/errors"
)

const (
	// TYPESessionDefault is a type of controller
	TYPESessionDefault = "default"
)

var (
	buildSessionHandlers = map[string]func(*Config) (Interface, error){}
)

func init() {
	RegisterSession(TYPESessionDefault, NewDefaultSession)
}

// Interface is a session manager interface
type Interface interface {
	IsSecure() bool
	IsDestroyed() bool
	IsStarted() bool
	IsWritable() bool
	IsReadable() bool
}

// NewSession creates a new session from given type and options
func NewSession(sessionType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildSessionHandlers[sessionType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized session type \"%v\"", sessionType)
}

// NewSessionFromConfig creates a new session from given Config
func NewSessionFromConfig(options *Config) (Interface, error) {
	if f, ok := buildSessionHandlers[options.Type]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized session type \"%v\"", options.Type)
}

// RegisterSession registers a handler for session manager creation
func RegisterSession(sessionType string, handler func(*Config) (Interface, error)) {
	if _, ok := buildSessionHandlers[sessionType]; ok {
		panic("[Session] Session of type '" + sessionType + "' is already registered")
	}

	buildSessionHandlers[sessionType] = handler
}

// Session is a default session struct
type Session struct {
	Options           *Config
	Started           bool
	Writable          bool
	Readable          bool
	WriteClosed       bool
	Destroyed         bool
	Strict            bool
	Secure            bool
	RememberMeSeconds int
	Data              map[string]interface{}
}

// IsSecure returns whether session is secure
func (s *Session) IsSecure() bool {
	return s.Secure
}

// IsDestroyed returns whether session is destroyed
func (s *Session) IsDestroyed() bool {
	return s.Destroyed
}

// IsStarted returns whether session is started
func (s *Session) IsStarted() bool {
	return s.Started
}

// IsWritable returns whether session is writable
func (s *Session) IsWritable() bool {
	return s.Writable
}

// IsReadable returns whether session is readable
func (s *Session) IsReadable() bool {
	return s.Readable
}

// NewDefaultSession creates a new default session handler
func NewDefaultSession(options *Config) (Interface, error) {
	return &Session{
		Options: options,
		Data:    make(map[string]interface{}),
	}, nil
}
