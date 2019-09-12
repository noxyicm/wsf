package auth

import (
	"context"
	"strings"
	"wsf/db"
	"wsf/errors"
)

// Public constants
const (
	TYPEAdapterDbTable = "DbTable"
)

func init() {
	RegisterAdapter(TYPEAdapterDbTable, NewAdapterDbTable)
}

// DbTableAdapter is a database based auth adapter
type DbTableAdapter struct {
	Options             *AdapterConfig
	Db                  db.Adapter
	TableName           string
	IdentityColumn      string
	CredentialColumn    string
	CredentialTreatment string
	AmbiguityIdentity   bool
}

// Setup the object
func (a *DbTableAdapter) Setup() error {
	if a.Db == nil {
		a.Db = db.GetDefaultAdapter()
	}

	a.TableName = a.Options.TableName
	a.IdentityColumn = a.Options.IdentityColumn
	a.CredentialColumn = a.Options.CredentialColumn
	a.CredentialTreatment = a.Options.CredentialTreatment
	a.AmbiguityIdentity = a.Options.AmbiguityIdentity

	return nil
}

// Authenticate performs an authentication attempt
func (a *DbTableAdapter) Authenticate(ctx context.Context) Result {
	var identity string
	var credential string
	r, _ := NewResultDefault(ResultFailure, nil, []error{})

	if a.TableName == "" {
		r.AddError(errors.New("A table must be supplied for the wsf.auth.DbTable authentication adapter"))
	} else if a.IdentityColumn == "" {
		r.AddError(errors.New("An identity column must be supplied for the wsf.auth.DbTable authentication adapter"))
	} else if a.CredentialColumn == "" {
		r.AddError(errors.New("A credential column must be supplied for the wsf.auth.DbTable authentication adapter"))
	}

	idnt := ctx.Value(a.IdentityColumn)
	if v, ok := idnt.(string); ok {
		identity = v
	} else {
		r.AddError(errors.New("A value for the identity was not provided prior to authentication with wsf.auth.DbTable"))
	}

	crdntl := ctx.Value(a.CredentialTreatment)
	if v, ok := crdntl.(string); ok {
		credential = v
	} else {
		r.AddError(errors.New("A credential value was not provided prior to authentication with wsf.auth.DbTable"))
	}

	if len(r.GetErrors()) > 0 {
		return r
	}

	credentialTreatment := a.CredentialTreatment
	if a.CredentialTreatment == "" || !strings.Contains(a.CredentialTreatment, "?") {
		credentialTreatment = "?"
	}

	credentialExpression := db.NewExpr("(CASE WHEN " + a.Db.QuoteInto(a.Db.QuoteIdentifier(a.CredentialColumn, true)+" = "+credentialTreatment, credential, 1) + " THEN 1 ELSE 0 END) AS " + a.Db.QuoteIdentifier(a.Db.FoldCase("wsf_auth_credential_match"), true))
	dbSelect, err := a.Db.Select()
	if err != nil {
		r.AddError(errors.New("Unable to create select object for authentication with wsf.auth.DbTable"))
		return r
	}

	dbSelect.From(a.TableName, []interface{}{db.SQLWildcard, credentialExpression})
	dbSelect.Where(a.Db.QuoteIdentifier(a.IdentityColumn, true)+" = ?", identity)

	resultIdentities, err := a.Db.Query(ctx, dbSelect)
	if err != nil {
		r.AddError(errors.New("The supplied parameters to wsf.auth.DbTable failed to produce a valid sql statement, please check table and column names for validity"))
		return r
	}

	if resultIdentities.Count() < 1 {
		r.SetCode(ResultFailureIdentityNotFound)
		r.AddError(errors.New("A record with the supplied identity could not be found"))
		return r
	} else if resultIdentities.Count() > 1 && !a.AmbiguityIdentity {
		r.SetCode(ResultFailureIdentityAmbiguous)
		r.AddError(errors.New("More than one record matches the supplied identity"))
		return r
	}

	var resultIdentityRow db.Row
	authCredentialMatchColumn := a.Db.FoldCase("wsf_auth_credential_match")
	if a.AmbiguityIdentity {
		validIdentities := make([]db.Row, 0)
		for resultIdentities.Next() {
			idnt := resultIdentities.Get()
			if idnt.GetInt(authCredentialMatchColumn) == 1 {
				validIdentities = append(validIdentities, idnt)
			}
		}

		resultIdentityRow = validIdentities[0]
	} else {
		resultIdentities.Next()
		resultIdentityRow = resultIdentities.Get()
	}

	if resultIdentityRow.GetInt(authCredentialMatchColumn) != 1 {
		r.SetCode(ResultFailureCredentialInvalid)
		r.AddError(errors.New("Supplied credential is invalid"))
		return r
	}

	var m map[string]interface{}
	resultIdentityRow.Unmarshal(&m)

	delete(m, authCredentialMatchColumn)

	resultIdentity, _ := NewIdentityFromConfig(Options().Identity, m)
	r.SetCode(ResultSuccess)
	r.SetIdentity(resultIdentity)
	r.AddError(errors.New("Authentication successful"))

	return r
}

// NewAdapterDbTable creates a new dbtable adapter
func NewAdapterDbTable(options *AdapterConfig) (Adapter, error) {
	a := &DbTableAdapter{}
	a.Options = options
	a.Setup()

	return a, nil
}
