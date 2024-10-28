package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"sync"
	"time"
	"wsf/cache"
	"wsf/config"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/session/validator"
)

// Public contants
const (
	// TYPESessionManagerDefault is a type of session manager
	TYPESessionManagerDefault = "default"

	IDKey        = "sessionID"
	Key          = "session"
	AutostartKey = "autostart"
	SetCookieKey = "setcookie"
)

var (
	buildManagerHandlers = map[string]func(*ManagerConfig) (ManagerInterface, error){}

	ses ManagerInterface

	// ErrorUndefinedSessionID is a error
	ErrorUndefinedSessionID = errors.New("Undefined session id")

	// ErrorNoSessionIDInRequest is a error
	ErrorNoSessionIDInRequest = errors.New("No session ID in request")
)

func init() {
	Register(TYPESessionManagerDefault, NewDefaultSessionManager)
}

// ManagerInterface represents session manager interface
type ManagerInterface interface {
	Priority() int
	Init(options *ManagerConfig) (bool, error)
	Options() *ManagerConfig
	IsStarted() bool
	GetSID(req request.Interface) (string, error)
	NewSID() (string, error)
	SessionStart(req request.Interface, rsp response.Interface) (Interface, string, error)
	SessionClose(sid string) error
	SessionGet(sid string) (Interface, bool)
	SessionDestroy(rqs request.Interface, rsp response.Interface)
	SessionExist(sid string) bool
	SessionSave(sid string) error
	SessionLoad(sid string, s Interface) error
	SessionAll() int
	RegisterValidator(v validator.Interface) error
}

// NewSessionManager creates a new session manager from given type and options
func NewSessionManager(managerType string, options config.Config) (sm ManagerInterface, err error) {
	cfg := &ManagerConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildManagerHandlers[managerType]; ok {
		sm, err = f(cfg)
		if err != nil {
			return nil, err
		}

		ses = sm
		return sm, nil
	}

	return nil, errors.Errorf("Unrecognized session manager type \"%v\"", managerType)
}

// Register registers a handler for session manager creation
func Register(managerType string, handler func(*ManagerConfig) (ManagerInterface, error)) {
	if _, ok := buildManagerHandlers[managerType]; ok {
		panic("[Session] Session manager of type '" + managerType + "' is already registered")
	}

	buildManagerHandlers[managerType] = handler
}

// Manager is a default session manager
type Manager struct {
	Opts       *ManagerConfig
	Started    bool
	Secure     bool
	Strict     bool
	Sessions   sync.Map
	Storage    cache.Interface
	Validators []validator.Interface
	mu         sync.Mutex
}

// Priority returns a priority of resource
func (m *Manager) Priority() int {
	return m.Opts.Priority
}

// Init the session manager
func (m *Manager) Init(options *ManagerConfig) (bool, error) {
	m.Opts = options

	ccfg := &cache.Config{}
	ccfg.Defaults()
	ccfg.Populate(options.Storage)

	cch, err := cache.NewCore(options.Storage.GetString("type"), options.Storage)
	if err != nil {
		return false, errors.Wrap(err, "[DefaultSessionManager] Storage creation error")
	}
	m.Storage = cch

	if ok, err := m.Storage.Init(ccfg); !ok {
		return ok, err
	}

	m.mu.Lock()
	m.Started = true
	m.mu.Unlock()

	return true, nil
}

// Options returns session manager config options
func (m *Manager) Options() *ManagerConfig {
	return m.Opts
}

// IsStarted returns true if manager initialized
func (m *Manager) IsStarted() bool {
	return m.Started
}

// GetSID returns a session id if registered
func (m *Manager) GetSID(rqs request.Interface) (string, error) {
	if sid := rqs.Context().Value(m.Opts.SessionName); sid != nil {
		return sid.(string), nil
	}

	if sid := rqs.Cookie(m.Opts.SessionName); sid != "" {
		return url.QueryUnescape(sid)
	}

	if m.Opts.EnableSidInURLQuery {
		if sid := rqs.Param(m.Opts.SessionName); sid != nil {
			return sid.(string), nil
		}
	}

	if m.Opts.EnableSidInHTTPHeader {
		if sid := rqs.Header(m.Opts.SessionNameInHTTPHeader); sid != "" {
			return sid, nil
		}
	}

	return "", ErrorNoSessionIDInRequest
}

// NewSID returns a new session id
func (m *Manager) NewSID() (string, error) {
	b := make([]byte, m.Opts.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", errors.New("[Session] Could not successfully read from the system CSPRNG")
	}

	return m.Opts.SessionIDPrefix + hex.EncodeToString(b), nil
}

// SessionStart starts a new session
func (m *Manager) SessionStart(rqs request.Interface, rsp response.Interface) (Interface, string, error) {
	if !m.Started {
		return nil, "", errors.New("[Session] Manager is not initialized")
	}

	autostart := m.Opts.SessionAutostart
	if v := rqs.Context().Value(AutostartKey); v != nil {
		autostart = v.(bool)
	}

	setcookie := m.Opts.EnableSetCookie
	if v := rqs.Context().Value(SetCookieKey); v != nil {
		setcookie = v.(bool)
	}

	sid, err := m.GetSID(rqs)
	if err != nil {
		sid, err = m.NewSID()
		if err != nil {
			return nil, "", errors.Wrap(err, "[Session] Unable to start session")
		}
	}

	if s, ok := m.Sessions.Load(sid); ok {
		return s.(Interface), sid, nil
	}

	s, err := NewSession(m.Opts.Session.GetString("type"), m.Opts.Session)
	if err != nil {
		return nil, "", errors.Wrap(err, "[Session] Unable to start session")
	}

	if m.SessionExist(sid) {
		if err = m.SessionLoad(sid, s); err != nil {
			sid, err = m.NewSID()
			if err != nil {
				return nil, "", errors.Wrap(err, "[Session] Unable to start session")
			}
		}
	} else if autostart {
		sid, err = m.NewSID()
		if err != nil {
			return nil, "", errors.Wrap(err, "[Session] Unable to start session")
		}
	} else {
		return nil, "", nil
	}

	if setcookie {
		cookie := &http.Cookie{
			Name:     m.Opts.SessionName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: m.Opts.HTTPOnly,
			Secure:   rqs.IsSecure() && m.Opts.Secure,
			Domain:   config.App.GetString("application.Domain"),
		}

		if m.Opts.SessionLifeTime > 0 {
			cookie.MaxAge = int(m.Opts.SessionLifeTime)
			cookie.Expires = time.Now().Add(time.Duration(m.Opts.SessionLifeTime) * time.Second)
		}

		rqs.AddCookie(cookie)
		rsp.AddCookie(cookie)
	}

	if m.Opts.EnableSidInHTTPHeader {
		rqs.AddHeader(m.Opts.SessionNameInHTTPHeader, sid)
		rsp.AddHeader(m.Opts.SessionNameInHTTPHeader, sid)
	}

	m.Sessions.Store(sid, s)
	return s, sid, nil
}

// SessionClose writes and closes a session
func (m *Manager) SessionClose(sid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.Sessions.Load(sid); ok {
		encoded, err := s.(Interface).Marshal()
		if err != nil {
			return errors.Wrap(err, "Unable to save sassion")
		}

		if !m.Storage.Save(encoded, sid, []string{sid}, m.Opts.SessionLifeTime) {
			return errors.Wrap(m.Storage.Error(), "Unable to save sassion")
		}
	}

	m.Sessions.Delete(sid)
	return nil
}

// SessionSave writes a session
func (m *Manager) SessionSave(sid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.Sessions.Load(sid); ok {
		encoded, err := s.(Interface).Marshal()
		if err != nil {
			return errors.Wrap(err, "Unable to save sassion")
		}

		if !m.Storage.Save(encoded, sid, []string{sid}, m.Opts.SessionLifeTime) {
			return errors.Wrap(m.Storage.Error(), "Unable to save sassion")
		}
	}

	return nil
}

// SessionGet returns a cached session
func (m *Manager) SessionGet(sid string) (Interface, bool) {
	if s, ok := m.Sessions.Load(sid); ok {
		return s.(Interface), true
	}

	return nil, false
}

// SessionDestroy destroys existing session
func (m *Manager) SessionDestroy(rqs request.Interface, rsp response.Interface) {
	sid, err := m.GetSID(rqs)
	if err != nil {
		return
	}

	m.Sessions.Delete(sid)
	m.Storage.Remove(sid)

	if m.Opts.EnableSetCookie {
		cookie := rqs.RawCookie(m.Opts.SessionName)
		if cookie == nil {
			return
		}

		cookie.Value = ""
		cookie.MaxAge = -1
		cookie.Expires = time.Now()
		rqs.AddCookie(cookie)
		rsp.AddCookie(cookie)
	}

	if m.Opts.EnableSidInHTTPHeader {
		rqs.RemoveHeader(m.Opts.SessionNameInHTTPHeader)
		rsp.RemoveHeader(m.Opts.SessionNameInHTTPHeader)
	}

	return
}

// SessionExist returns true if session by id exists
func (m *Manager) SessionExist(sid string) bool {
	return m.Storage.Test(sid)
}

// SessionLoad loads session from storage and populates its data to s
func (m *Manager) SessionLoad(sid string, s Interface) error {
	data, ok := m.Storage.Load(sid, false)
	if !ok {
		return m.Storage.Error()
	}

	if err := s.Unmarshal(data); err != nil {
		return err
	}

	for _, vld := range m.Validators {
		if err := vld.Valid(s.All()); err != nil {
			return errors.Wrap(err, "Unable to load session")
		}
	}

	return nil
}

// SessionAll returns number of registered sessions
func (m *Manager) SessionAll() int {
	return -1
}

// RegisterValidator registers a session validator
func (m *Manager) RegisterValidator(v validator.Interface) error {
	m.Validators = append(m.Validators, v)
	return nil
}

// NewDefaultSessionManager creates a new default session manager
func NewDefaultSessionManager(options *ManagerConfig) (ManagerInterface, error) {
	sm := &Manager{
		Opts: options,
	}

	for _, vcfg := range options.Valds {
		v, err := validator.NewValidatorFromConfig(vcfg)
		if err != nil {
			return nil, err
		}

		v.Setup()
		sm.Validators = append(sm.Validators, v)
	}

	if options.Storage == nil {
		return nil, errors.New("[DefaultSessionManager] Storage is not configured")
	}

	return sm, nil
}

// SetInstance sets session instance
func SetInstance(s ManagerInterface) {
	ses = s
}

// Instance returns a session instance
func Instance() ManagerInterface {
	return ses
}

// Start new session
func Start(rqs request.Interface, rsp response.Interface) (Interface, string, error) {
	return ses.SessionStart(rqs, rsp)
}

// Get returns a session
func Get(sid string) (Interface, bool) {
	return ses.SessionGet(sid)
}

// Close session
func Close(sid string) {
	ses.SessionClose(sid)
}
