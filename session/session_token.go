package session

const (
	// TYPESessionToken is a type of controller
	TYPESessionToken = "token"

	// HTTPHeaderBearer is
	HTTPHeaderBearer = "Token"
)

func init() {
	RegisterSession(TYPESessionToken, NewTokenSession)
}

// Token is a session handler
type Token struct {
	Session
}

// NewTokenSession creates a new token session handler
func NewTokenSession(options *Config) (Interface, error) {
	s := &Token{}
	s.Options = options
	s.Data = make(map[string]interface{})

	return s, nil
}
