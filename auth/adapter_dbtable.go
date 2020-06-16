package auth

import (
	"strings"
	"wsf/context"
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

	a.TableName = a.Options.Source
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
	res, err := NewResult(TYPEAuthResultDefault, ResultFailure, nil, make([]error, 0))
	if err != nil {
		res = NewResultDefault(ResultFailure, nil, make([]error, 0))
	}

	if a.TableName == "" {
		res.AddError(errors.New("A table must be supplied for the wsf.auth.DbTable authentication adapter"))
	} else if a.IdentityColumn == "" {
		res.AddError(errors.New("An identity column must be supplied for the wsf.auth.DbTable authentication adapter"))
	} else if a.CredentialColumn == "" {
		res.AddError(errors.New("A credential column must be supplied for the wsf.auth.DbTable authentication adapter"))
	}

	idnt := ctx.Param("auth.identity")
	if v, ok := idnt.(string); ok {
		identity = v
	} else {
		res.AddError(errors.New("A value for the identity was not provided prior to authentication with wsf.auth.DbTable"))
	}

	crdntl := ctx.Param("auth.credential")
	if v, ok := crdntl.(string); ok {
		credential = v
	} else {
		res.AddError(errors.New("A credential value was not provided prior to authentication with wsf.auth.DbTable"))
	}

	if len(res.GetErrors()) > 0 {
		return res
	}

	credentialTreatment := a.CredentialTreatment
	if a.CredentialTreatment == "" || !strings.Contains(a.CredentialTreatment, "?") {
		credentialTreatment = "?"
	}

	credentialExpression := db.NewExpr("(CASE WHEN " + a.Db.QuoteInto(a.Db.QuoteIdentifier(a.CredentialColumn, true)+" = "+credentialTreatment, credential, 1) + " THEN 1 ELSE 0 END) AS " + a.Db.QuoteIdentifier(a.Db.FoldCase("wsf_auth_credential_match"), true))
	dbSelect := a.Db.Select()
	dbSelect.From(a.TableName, []interface{}{db.SQLWildcard, credentialExpression})
	dbSelect.Where(a.Db.QuoteIdentifier(a.IdentityColumn, true)+" = ?", identity)

	resultIdentities, err := a.Db.Query(ctx, dbSelect)
	if err != nil {
		res.AddError(errors.New("The supplied parameters to wsf.auth.DbTable failed to produce a valid sql statement, please check table and column names for validity"))
		return res
	}

	if len(resultIdentities) < 1 {
		res.SetCode(ResultFailureIdentityNotFound)
		res.AddError(errors.New("A record with the supplied identity could not be found"))
		return res
	} else if len(resultIdentities) > 1 && !a.AmbiguityIdentity {
		res.SetCode(ResultFailureIdentityAmbiguous)
		res.AddError(errors.New("More than one record matches the supplied identity"))
		return res
	}

	var resultIdentityRow map[string]interface{}
	authCredentialMatchColumn := a.Db.FoldCase("wsf_auth_credential_match")
	if a.AmbiguityIdentity {
		validIdentities := make([]map[string]interface{}, 0)
		for _, idnt := range resultIdentities {
			if v, ok := idnt[authCredentialMatchColumn]; ok && v == 1 {
				validIdentities = append(validIdentities, idnt)
			}
		}

		resultIdentityRow = validIdentities[0]
	} else {
		resultIdentityRow = resultIdentities[0]
	}

	if v, ok := resultIdentityRow[authCredentialMatchColumn]; ok && v != 1 {
		res.SetCode(ResultFailureCredentialInvalid)
		res.AddError(errors.New("Supplied credential is invalid"))
		return res
	}

	delete(resultIdentityRow, authCredentialMatchColumn)

	resultIdentity, err := NewIdentityFromConfig(Options().Identity, resultIdentityRow)
	if err != nil {
		res.AddError(errors.Wrap(err, "Authentication falied"))
		return res
	}

	res.SetCode(ResultSuccess)
	res.SetIdentity(resultIdentity)
	return res
}

// NewAdapterDbTable creates a new dbtable adapter
func NewAdapterDbTable(options *AdapterConfig) (Adapter, error) {
	a := &DbTableAdapter{}
	a.Options = options
	a.Setup()

	return a, nil
}
