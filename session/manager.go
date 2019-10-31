package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	// TYPEDefaultSessionManager is a type of session manager
	TYPEDefaultSessionManager = "default"

	IDKey = "sessionID"
	Key   = "session"
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
	Register(TYPEDefaultSessionManager, NewDefaultSessionManager)
}

// ManagerInterface represents session manager interface
type ManagerInterface interface {
	Priority() int
	Init(options *ManagerConfig) (bool, error)
	IsStarted() bool
	GetSID(req request.Interface) (string, error)
	NewSID() (string, error)
	SessionStart(req request.Interface, rsp response.Interface) (Interface, string, error)
	SessionClose(sid string)
	SessionGet(sid string) (Interface, bool)
	SessionDestroy(rqs request.Interface, rsp response.Interface)
	SessionExist(sid string) bool
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
	Options    *ManagerConfig
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
	return m.Options.Priority
}

// Init the session manager
func (m *Manager) Init(options *ManagerConfig) (bool, error) {
	m.mu.Lock()
	m.Started = true
	m.mu.Unlock()

	ccfg := &cache.Config{}
	ccfg.Defaults()
	ccfg.Populate(options.Storage)
	if ok, err := m.Storage.Init(ccfg); !ok {
		m.mu.Lock()
		m.Started = false
		m.mu.Unlock()

		return ok, err
	}

	return true, nil
}

// IsStarted returns true if manager initialized
func (m *Manager) IsStarted() bool {
	return m.Started
}

// GetSID returns a session id if registered
func (m *Manager) GetSID(rqs request.Interface) (string, error) {
	if sid := rqs.Cookie(m.Options.SessionName); sid != "" {
		return url.QueryUnescape(sid)
	}

	if m.Options.EnableSidInURLQuery {
		if sid := rqs.Param(m.Options.SessionName); sid != nil {
			return sid.(string), nil
		}
	}

	if m.Options.EnableSidInHTTPHeader {
		if sid := rqs.Header(m.Options.SessionNameInHTTPHeader); sid != "" {
			return sid, nil
		}
	}

	return "", ErrorNoSessionIDInRequest
}

// NewSID returns a new session id
func (m *Manager) NewSID() (string, error) {
	b := make([]byte, m.Options.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", errors.New("[Session] Could not successfully read from the system CSPRNG")
	}

	return m.Options.SessionIDPrefix + hex.EncodeToString(b), nil
}

// SessionStart starts a new session
func (m *Manager) SessionStart(rqs request.Interface, rsp response.Interface) (Interface, string, error) {
	if !m.Started {
		return nil, "", errors.New("[Session] Manager is not initialized")
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

	s, err := NewSession(m.Options.Session.GetString("type"), m.Options.Session)
	if err != nil {
		return nil, "", errors.Wrap(err, "[Session] Unable to start session")
	}

	if m.SessionExist(sid) {
		if err = m.SessionLoad(sid, s); err != nil {
			return nil, "", errors.Wrap(err, "[Session] Unable to load session")
		}
	}

	if m.Options.EnableSetCookie {
		cookie := &http.Cookie{
			Name:     m.Options.SessionName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: m.Options.HTTPOnly,
			Secure:   rqs.IsSecure() && m.Options.Secure,
			Domain:   config.App.GetString("application.Domain"),
		}

		if m.Options.SessionLifeTime > 0 {
			cookie.MaxAge = int(m.Options.SessionLifeTime)
			cookie.Expires = time.Now().Add(time.Duration(m.Options.SessionLifeTime) * time.Second)
		}

		rqs.AddCookie(cookie)
		rsp.AddCookie(cookie)
	}

	if m.Options.EnableSidInHTTPHeader {
		rqs.AddHeader(m.Options.SessionNameInHTTPHeader, sid)
		rsp.AddHeader(m.Options.SessionNameInHTTPHeader, sid)
	}

	m.Sessions.Store(sid, s)
	return s, sid, nil
}

// SessionClose writes and cleses a session
func (m *Manager) SessionClose(sid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.Sessions.Load(sid); ok {
		encoded, _ := json.Marshal(s.(Interface))
		m.Storage.Save(encoded, sid, []string{sid}, m.Options.SessionLifeTime)
	}

	m.Sessions.Delete(sid)
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

	if m.Options.EnableSetCookie {
		cookie := rqs.RawCookie(m.Options.SessionName)
		if cookie == nil {
			return
		}

		cookie.Value = ""
		cookie.MaxAge = -1
		cookie.Expires = time.Now()
		rqs.AddCookie(cookie)
		rsp.AddCookie(cookie)
	}

	if m.Options.EnableSidInHTTPHeader {
		rqs.RemoveHeader(m.Options.SessionNameInHTTPHeader)
		rsp.RemoveHeader(m.Options.SessionNameInHTTPHeader)
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

	if err := json.Unmarshal(data, s); err != nil {
		return err
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
		Options: options,
	}

	for _, vcfg := range options.Valds {
		v, err := validator.NewValidatorFromConfig(vcfg)
		if err != nil {
			return nil, err
		}

		sm.Validators = append(sm.Validators, v)
	}

	if options.Storage == nil {
		return nil, errors.New("[DefaultSessionManager] Storage is not configured")
	}

	str, err := cache.NewCore(options.Storage.GetString("type"), options.Storage)
	if err != nil {
		return nil, errors.Wrap(err, "[DefaultSessionManager] Storage creation error")
	}

	sm.Storage = str
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
