package session

import (
	"encoding/json"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
)

// Public constants
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
	Has(key string) bool
	Get(key string) interface{}
	Set(key string, data interface{}) error
	Unset(key string) bool
	All() map[string]interface{}
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
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
func RegisterSession(sessionType string, handler func(*Config) (Interface, error)) error {
	if _, ok := buildSessionHandlers[sessionType]; ok {
		return errors.Errorf("Session of type '%s' is already registered", sessionType)
	}

	buildSessionHandlers[sessionType] = handler
	return nil
}

// Session is a default session
type Session struct {
	Options     *Config
	Started     bool
	Writable    bool
	Readable    bool
	WriteClosed bool
	Destroyed   bool
	Strict      bool
	Secure      bool
	Data        map[string]interface{}
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

// Has returns true if session contains provided key
func (s *Session) Has(key string) bool {
	if _, ok := s.Data[key]; ok {
		return true
	}

	return false
}

// Get returns a session value by its key
func (s *Session) Get(key string) interface{} {
	if !s.Has(key) {
		return nil
	}

	return s.Data[key]
}

// Set session value
func (s *Session) Set(key string, data interface{}) error {
	s.Data[key] = data
	return nil
}

// Unset the value by its key
func (s *Session) Unset(key string) bool {
	if !s.Has(key) {
		return false
	}

	delete(s.Data, key)
	return true
}

// All returns all session params
func (s *Session) All() map[string]interface{} {
	m := s.Data
	return m
}

// Marshal session into json
func (s *Session) Marshal() ([]byte, error) {
	return json.Marshal(s.Data)
}

// Unmarshal session from json
func (s *Session) Unmarshal(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	s.Data = m
	return nil
}

// NewDefaultSession creates a new default session handler
func NewDefaultSession(options *Config) (Interface, error) {
	return &Session{
		Options: options,
		Data:    make(map[string]interface{}),
	}, nil
}
