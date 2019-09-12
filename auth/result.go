package auth

import "wsf/errors"

// Public constants
const (
	TYPEResultDefault = "default"

	// General Failure
	ResultFailure = 0

	// Failure due to identity not being found
	ResultFailureIdentityNotFound = -1

	// Failure due to identity being ambiguous
	ResultFailureIdentityAmbiguous = -2

	// Failure due to invalid credential being supplied
	ResultFailureCredentialInvalid = -3

	// Failure due to uncategorized reasons
	ResultFailureUncategorized = -4

	// Authentication success
	ResultSuccess = 1
)

var (
	buildResultHandlers = map[string]func(int, Identity, []error) (Result, error){}
)

func init() {
	RegisterResult(TYPEResultDefault, NewResultDefault)
}

// Result represent auth result interface
type Result interface {
	Setup() error
	Valid() bool
	SetCode(code int)
	GetCode() int
	SetIdentity(idnt Identity)
	GetIdentity() Identity
	GetErrors() []error
	AddError(error)
	SetErrors([]error)
}

// NewResult creates a new auth result from given type and options
func NewResult(resultType string, code int, identity Identity, messages []error) (Result, error) {
	if f, ok := buildResultHandlers[resultType]; ok {
		return f(code, identity, messages)
	}

	return nil, errors.Errorf("Unrecognized auth result type \"%v\"", resultType)
}

// RegisterResult registers a handler for auth result creation
func RegisterResult(resultType string, handler func(int, Identity, []error) (Result, error)) {
	buildResultHandlers[resultType] = handler
}

// DefaultResult is a default auth result object
type DefaultResult struct {
	Code     int
	Identity Identity
	Errors   []error
}

// Setup the object
func (r *DefaultResult) Setup() error {
	return nil
}

// Valid returns true on validation success
func (r *DefaultResult) Valid() bool {
	if r.Code > 0 {
		return true
	}

	return false
}

// SetCode sets a result code
func (r *DefaultResult) SetCode(code int) {
	r.Code = code
}

// GetCode returns the result code for this authentication attempt
func (r *DefaultResult) GetCode() int {
	return r.Code
}

// SetIdentity sets identity
func (r *DefaultResult) SetIdentity(idnt Identity) {
	r.Identity = idnt
}

// GetIdentity returns the identity used in the authentication attempt
func (r *DefaultResult) GetIdentity() Identity {
	return r.Identity
}

// GetErrors returns a slice of string reasons why the authentication attempt was unsuccessful
func (r *DefaultResult) GetErrors() []error {
	return r.Errors
}

// AddError adds an error to the result object
func (r *DefaultResult) AddError(err error) {
	r.Errors = append(r.Errors, err)
}

// SetErrors sets the errors for result object
func (r *DefaultResult) SetErrors(errs []error) {
	r.Errors = errs
}

// NewResultDefault creates a new default auth result object
func NewResultDefault(code int, identity Identity, messages []error) (Result, error) {
	if code < ResultFailureUncategorized {
		code = ResultFailure
	} else if code > ResultSuccess {
		code = ResultSuccess
	}

	return &DefaultResult{
		Code:     code,
		Identity: identity,
		Errors:   messages,
	}, nil
}
