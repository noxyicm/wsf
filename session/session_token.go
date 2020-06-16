package session

const (
	// TYPESessionToken is a type of controller
	TYPESessionToken = "token"

	// HTTPHeaderBearer is
	HTTPHeaderBearer = "Token"
)

func init() {
	RegisterSession(TYPESessionToken, NewSessionToken)
}

// Token is a session handler
type Token struct {
	Session
}

// NewSessionToken creates a new token session handler
func NewSessionToken(options *Config) (Interface, error) {
	s := &Token{}
	s.Options = options
	s.Data = make(map[string]interface{})

	return s, nil
}
