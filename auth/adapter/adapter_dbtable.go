package adapter

import (
	"wsf/auth"
	"wsf/context"
	"wsf/db"
	"wsf/errors"

	"github.com/jamesruan/sodium"
)

// Public constants
const (
	TYPEAdapterDbTable = "DbTable"
)

func init() {
	auth.RegisterAdapter(TYPEAdapterDbTable, NewAdapterDbTable)
}

// DbTableAdapter is a database based auth adapter
type DbTableAdapter struct {
	Options             *auth.AdapterConfig
	Db                  db.Adapter
	TableName           string
	IdentityColumn      string
	CredentialColumn    string
	CredentialTreatment string
	AmbiguityIdentity   bool
	prefix              string
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
func (a *DbTableAdapter) Authenticate(ctx context.Context) auth.Result {
	var identity string
	var credential string
	res, err := auth.NewResult(auth.TYPEAuthResultDefault, auth.ResultFailure, nil, make([]error, 0))
	if err != nil {
		res = auth.NewResultDefault(auth.ResultFailure, nil, make([]error, 0))
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

	//credentialTreatment := a.CredentialTreatment
	//if a.CredentialTreatment == "" || !strings.Contains(a.CredentialTreatment, "?") {
	//	credentialTreatment = "?"
	//}

	//credentialExpression := db.NewExpr("(CASE WHEN " + a.Db.QuoteInto(a.Db.QuoteIdentifier(a.CredentialColumn, true)+" = "+credentialTreatment, credential, 1) + " THEN 1 ELSE 0 END) AS " + a.Db.QuoteIdentifier(a.Db.FoldCase("wsf_auth_credential_match"), true))
	dbSelect := a.Db.Select()
	dbSelect.From(a.TableName, []interface{}{db.SQLWildcard})
	dbSelect.JoinLeft("roles", "roles.id = users.roleId", map[string]string{
		"role": "alias",
	})
	dbSelect.Where(a.Db.QuoteIdentifier(a.IdentityColumn, true)+" = ?", identity)

	resultIdentities, err := a.Db.Query(ctx, dbSelect)
	if err != nil {
		res.AddError(errors.Wrap(err, "The supplied parameters to wsf.auth.DbTable failed to produce a valid sql statement, please check table and column names for validity"))
		return res
	}

	if len(resultIdentities) < 1 {
		res.SetCode(auth.ResultFailureIdentityNotFound)
		res.AddError(errors.New("A record with the supplied identity could not be found"))
		return res
	} else if len(resultIdentities) > 1 && !a.AmbiguityIdentity {
		res.SetCode(auth.ResultFailureIdentityAmbiguous)
		res.AddError(errors.New("More than one record matches the supplied identity"))
		return res
	}

	resultIdentityRow := resultIdentities[0]
	invalid := false
	if p, ok := resultIdentityRow[a.CredentialColumn]; ok {
		if pw, ok := p.(string); ok {
			pwb := make([]byte, 128, 128)
			copy(pwb, []byte(a.prefix+pw))
			pwd := sodium.LoadPWHashStr(pwb)
			if err = pwd.PWHashVerify(credential); err != nil {
				res.SetCode(auth.ResultFailureCredentialInvalid)
				res.AddError(errors.New("Supplied credential is invalid"))
				return res
			}
		} else {
			invalid = true
		}
	} else {
		invalid = true
	}

	if invalid {
		res.SetCode(auth.ResultFailureIdentityNotFound)
		res.AddError(errors.New("Supplied credential is invalid"))
		return res
	}

	authCredentialStatusColumn := a.Db.FoldCase("state")
	if st, ok := resultIdentityRow[authCredentialStatusColumn].(int); ok && st == 9 {
		res.SetCode(auth.ResultFailureIdentityBlocked)
		res.AddError(errors.New("Supplied credential is blocked"))
		return res
	} else if st, ok := resultIdentityRow[authCredentialStatusColumn].(int); ok && st == -1 {
		res.SetCode(auth.ResultFailureIdentityInactive)
		res.AddError(errors.New("Supplied credential is not confirmed"))
		return res
	} else if st, ok := resultIdentityRow[authCredentialStatusColumn].(int); ok && st == 0 {
		res.SetCode(auth.ResultFailureIdentityBlocked)
		res.AddError(errors.New("Supplied credential is blocked"))
		return res
	}

	delete(resultIdentityRow, a.CredentialColumn)

	resultIdentity, err := auth.NewIdentityFromConfig(auth.Options().Identity, resultIdentityRow)
	if err != nil {
		res.AddError(errors.Wrap(err, "Authentication falied"))
		return res
	}

	res.SetCode(auth.ResultSuccess)
	res.SetIdentity(resultIdentity)
	return res
}

// NewAdapterDbTable creates a new dbtable adapter
func NewAdapterDbTable(options *auth.AdapterConfig) (auth.Adapter, error) {
	a := &DbTableAdapter{
		prefix: "$argon2id$v=19$m=65536,t=2,p=1$",
	}
	a.Options = options
	a.Setup()

	return a, nil
}
