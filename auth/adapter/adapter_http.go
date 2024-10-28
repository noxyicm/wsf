package adapter

import (
	"bufio"
	"os"
	"strconv"
	"wsf/auth"
	"wsf/context"
	"wsf/errors"
)

// Public constants
const (
	TYPEAuthAdapterHTTP = "http"
)

func init() {
	auth.RegisterAdapter(TYPEAuthAdapterHTTP, NewAuthAdapterHTTP)
}

// AdapterHTTP is a http based auth adapter
type AdapterHTTP struct {
	Options  *auth.AdapterConfig
	Filename string
}

// Setup the object
func (a *AdapterHTTP) Setup() error {
	a.Filename = a.Options.Source
	_, err := os.Stat(a.Filename)
	if err != nil {
		return errors.Wrap(err, "Unable to setup auth AdapterHTTP adapter")
	}

	return nil
}

// Authenticate performs an authentication attempt
func (a *AdapterHTTP) Authenticate(ctx context.Context) auth.Result {
	res, err := auth.NewResult(auth.TYPEAuthResultDefault, auth.ResultFailure, nil, make([]error, 0))
	if err != nil {
		res = auth.NewResultDefault(auth.ResultFailure, nil, make([]error, 0))
	}

	iden := ctx.ParamString(context.IdentityKey)
	if iden == "" {
		res.AddError(errors.New("Authentication falied: Identity is not set"))
		return res
	}

	cred := ctx.ParamString(context.CredentialKey)
	if iden == "" {
		res.AddError(errors.New("Authentication falied: Credentials is not set"))
		return res
	}

	fd, err := os.Open(a.Filename)
	if err != nil {
		res.AddError(errors.Wrap(err, "Authentication falied"))
		return res
	}

	found := false
	line := iden + ":" + cred
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if line == scanner.Text() {
			found = true
			break
		}
	}
	fd.Close()

	if err := scanner.Err(); err != nil {
		res.AddError(errors.Wrap(err, "Authentication falied"))
		return res
	}

	if !found {
		res.SetCode(auth.ResultFailureIdentityNotFound)
		res.AddError(errors.New("Authentication falied: Identity not found"))
		return res
	}

	idenID, _ := strconv.Atoi(iden)
	resultIdentity, err := auth.NewIdentityFromConfig(auth.Options().Identity, map[string]interface{}{
		"id":         idenID,
		"role":       auth.ROLEUser,
		"roleID":     1,
		"instanceID": 0,
		"name":       "APIUser",
	})
	if err != nil {
		res.AddError(errors.Wrap(err, "Authentication falied"))
		return res
	}

	res.SetCode(auth.ResultSuccess)
	res.SetIdentity(resultIdentity)
	return res
}

// NewAuthAdapterHTTP creates a new http auth adapter
func NewAuthAdapterHTTP(options *auth.AdapterConfig) (auth.Adapter, error) {
	a := &AdapterHTTP{}
	a.Options = options
	a.Setup()

	return a, nil
}
