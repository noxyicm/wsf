package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wsf/config"
	"wsf/db/connection"
	"wsf/db/dbselect"
	"wsf/db/table/rowset"
	"wsf/db/transaction"
	"wsf/errors"
	"wsf/utils"
)

// Public constants
const (
	SQLAs      = "as"
	SQLAnd     = "and"
	SQLOr      = "or"
	SQLBetween = "between"

	LOGInsert     = "insert"
	LOGPreUpdate  = "preupdate"
	LOGPostUpdate = "preupdate"
	LOGPreDelete  = "predelete"
	LOGPostDelete = "predelete"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	// RegexpSingleQuote matches single quote in column name
	RegexpSingleQuote = regexp.MustCompile(`('.+?')`)
)

// Interface database adapter
type Interface interface {
	Init(ctx context.Context) error
	Context() context.Context
	SetContext(ctx context.Context) error
	Connection() (connection.Interface, error)
	Select() (dbselect.Interface, error)
	Query(sql dbselect.Interface) (rowset.Interface, error)
	//Profiler()
	Insert(table string, data map[string]interface{}) (int, error)
	Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error)
	//Delete
	//Select
	//FetchAll
	//FetchRow
	//FetchAssoc
	//FetchCol
	//FetchPairs
	//FetchOne
	LastInsertID() int
	BeginTransaction() (transaction.Interface, error)
	Quote(interface{}) string
	QuoteIdentifier(interface{}, bool) string
	QuoteIdentifierAs(ident interface{}, alias string, auto bool) string
	QuoteIdentifierSymbol() string
	QuoteInto(text string, value interface{}, count int) string
	QuoteColumnAs(ident interface{}, alias string, auto bool) string
	QuoteTableAs(ident interface{}, alias string, auto bool) string
	SupportsParameters(param string) bool
	Options() *Config
	SetOptions(options *Config) Interface
}

type adapter struct {
	options                    *Config
	db                         *sql.DB
	ctx                        context.Context
	params                     map[string]interface{}
	defaultStatementType       string
	pingTimeout                time.Duration
	queryTimeout               time.Duration
	connectionMaxLifeTime      int
	maxIdleConnections         int
	maxOpenConnections         int
	inTransaction              bool
	identifierSymbol           string
	autoQuoteIdentifiers       bool
	allowSerialization         bool
	autoReconnectOnUnserialize bool
	lastInsertID               int
	lastInsertUUID             string

	Unquoteable          []string
	Spliters             []string
	UnquoteableFunctions []string
}

// Context returns adapter specific context
func (a *adapter) Context() context.Context {
	return a.ctx
}

// SetContext sets adapter specific context
func (a *adapter) SetContext(ctx context.Context) error {
	a.ctx = ctx
	return nil
}

// Connection returns a connection to database
func (a *adapter) Connection() (conn connection.Interface, err error) {
	if a.db == nil {
		return nil, errors.New("Database is not initialized")
	}

	conn, err = connection.NewConnectionFromConfig(a.options.Connection.Type, a.options.Connection)
	if err != nil {
		return nil, err
	}

	conn.SetContext(a.ctx)
	return conn, nil
}

// Select creates a new adapter specific select object
func (a *adapter) Select() (dbselect.Interface, error) {
	sel, err := dbselect.NewSelect(a.options.Select.Type, a.options.Select)
	if err != nil {
		return nil, err
	}

	sel.SetAdapter(a)
	return sel, nil
}

// Query runs a query
func (a *adapter) Query(sql dbselect.Interface) (rowset.Interface, error) {
	if a.db == nil {
		return nil, errors.New("Database is not initialized")
	}

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
	defer cancel()

	if err := sql.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	rows, err := a.db.QueryContext(ctx, sql.Assemble(), sql.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	rst, err := rowset.NewRowset(a.options.Rowset.Type, a.options.Rowset)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	if err := rst.Prepare(rows); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	return rst, nil
}

// LastInsertID returns the last insert query ID
func (a *adapter) LastInsertID() int {
	return a.lastInsertID
}

// LastInsertUUID returns the last insert query UUID
func (a *adapter) LastInsertUUID() string {
	return a.lastInsertUUID
}

// BeginTransaction creates a new database transaction
func (a *adapter) BeginTransaction() (transaction.Interface, error) {
	if a.db == nil {
		return nil, errors.New("Database is not initialized")
	}

	tx, err := a.db.BeginTx(a.ctx, &sql.TxOptions{Isolation: a.options.Transaction.IsolationLevel, ReadOnly: a.options.Transaction.ReadOnly})
	if err != nil {
		return nil, err
	}

	trns, err := transaction.NewTransaction(a.options.Transaction.Type, tx)
	if err != nil {
		return nil, err
	}

	trns.SetContext(a.ctx)
	return trns, err
}

// Quote a string
func (a *adapter) Quote(value interface{}) string {
	switch value.(type) {
	case dbselect.Interface:
		return "(" + value.(dbselect.Interface).Assemble() + ")"

	case *dbselect.Expr:
		return value.(*dbselect.Expr).ToString()

	case map[string]interface{}:
		sl := make([]string, 0)
		for _, val := range value.(map[string]interface{}) {
			sl = append(sl, a.Quote(val))
		}

		return strings.Join(sl, ", ")

	case []string:
		sl := make([]string, 0)
		for _, val := range value.([]string) {
			sl = append(sl, a.Quote(val))
		}

		return strings.Join(sl, ", ")

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return strconv.Itoa(value.(int))

	case float32:
		return fmt.Sprintf("%f", value.(float32))

	case float64:
		return fmt.Sprintf("%f", value.(float64))

	case string:
		return a.quoteString(value.(string))
	}

	return ""
}

// QuoteInto quotes a value and places into a piece of text at a placeholder
func (a *adapter) QuoteInto(text string, value interface{}, count int) string {
	return strings.Replace(text, "?", a.Quote(value), count)
}

// QuoteIdentifier re
func (a *adapter) QuoteIdentifier(ident interface{}, auto bool) string {
	return a.QuoteIdentifierAs(ident, "", auto)
}

// QuoteIdentifierAs a
func (a *adapter) QuoteIdentifierAs(ident interface{}, alias string, auto bool) string {
	as := " " + strings.ToUpper(SQLAs) + " "
	quoted := ""
	literals := make([]string, 0)
	idents := make([]interface{}, 0)

	switch ident.(type) {
	case dbselect.Interface:
		quoted = "(" + ident.(dbselect.Interface).Assemble() + ")"

	case *dbselect.Expr:
		quoted = ident.(*dbselect.Expr).ToString()

	case string:
		functions := make([]string, 0)
		declarations := make([]string, 0)
		initialIdent := ident.(string)
		v := ident.(string)
		if find := RegexpSingleQuote.FindString(v); find != "" {
			literals = append(literals, find)
			reg, err := regexp.Compile(`\{\$` + strconv.Itoa((len(literals) - 1)) + `\}`)
			if err == nil {
				quoted = reg.ReplaceAllString(v, find)
			}
		}

		matches := []string{}
		reg, err := regexp.Compile(`(` + strings.ToLower(strings.Join(a.UnquoteableFunctions, "|")) + `)\((.+?)\)`)
		if err == nil {
			matches = reg.FindStringSubmatch(v)
		}

		if len(matches) > 0 {
			functions = append(functions, matches[0])
			declarations = append(declarations, matches[1])
			idents = append(idents, strings.TrimSpace(matches[2]))
		} else {
			idents = append(idents, v)
		}

		if len(idents) > 0 {
			segments := make([]string, 0)
			for _, segment := range idents {
				switch segment.(type) {
				case dbselect.Interface:
					segments = append(segments, "("+segment.(dbselect.Interface).Assemble()+")")

				case *dbselect.Expr:
					segments = append(segments, segment.(*dbselect.Expr).ToString())

				case string:
					split := []string{}
					spliters := []string{}
					subsplit := [][]string{}
					for _, spliter := range a.Spliters {
						spliters = append(spliters, regexp.QuoteMeta(spliter))
					}

					reg, err := regexp.Compile(`(.+?)([` + strings.Join(spliters, "|") + `]+)(.+)`)
					if err == nil {
						subsplit = reg.FindAllStringSubmatch(segment.(string), -1)
						for _, subsplitmatch := range subsplit {
							for subkey, match := range subsplitmatch {
								if subkey == 0 {
									continue
								}

								split = append(split, match)
							}
							break
						}
					}

					if len(split) > 0 {
						segments = append(segments, a.quoteIdentifierSpec(split, auto))
					} else {
						cleanSegment := strings.ReplaceAll(segment.(string), a.QuoteIdentifierSymbol(), "")
						segments = append(segments, a.quoteIdentifier(cleanSegment, auto))
					}
				}
			}

			if alias != "" && segments[len(segments)-1] == alias {
				alias = ""
			}

			if len(functions) > 0 {
				quoted = initialIdent
				for key, segment := range segments {
					quoted = strings.ReplaceAll(quoted, functions[key], declarations[key]+"("+segment+")")
				}
			} else {
				quoted = strings.Join(segments, ".")
			}
		} else {
			quoted = a.quoteIdentifier(ident.(string), auto)
		}

	case map[string]interface{}:
		sl := make([]string, 0)
		for _, val := range ident.(map[string]interface{}) {
			sl = append(sl, a.Quote(val))
		}

		return strings.Join(sl, ", ")

	case []interface{}:
		return ""
	}

	if alias != "" {
		quoted = quoted + as + a.quoteIdentifier(alias, auto)
	}

	if len(literals) > 0 {
		for key, literal := range literals {
			quoted = strings.ReplaceAll(quoted, `{$`+strconv.Itoa(key)+`}`, literal)
		}
	}

	return quoted
}

// QuoteIdentifierSymbol returns symbol of identifier quote
func (a *adapter) QuoteIdentifierSymbol() string {
	return a.identifierSymbol
}

// QuoteColumnAs quote a column identifier and alias
func (a *adapter) QuoteColumnAs(ident interface{}, alias string, auto bool) string {
	return a.QuoteIdentifierAs(ident, alias, auto)
}

// Quote a table identifier and alias
func (a *adapter) QuoteTableAs(ident interface{}, alias string, auto bool) string {
	return a.QuoteIdentifierAs(ident, alias, auto)
}

// SupportsParameters returns true if adapter supports
func (a *adapter) SupportsParameters(param string) bool {
	if v, ok := a.params[param]; ok {
		return v.(bool)
	}

	return false
}

// quotes identifier
func (a *adapter) quoteIdentifier(ident string, auto bool) string {
	if !auto || a.autoQuoteIdentifiers {
		q := a.QuoteIdentifierSymbol()
		return q + strings.ReplaceAll(ident, q, q+q) + q
	}

	return ident
}

// quotes specifics
func (a *adapter) quoteIdentifierSpec(idents []string, auto bool) string {
	if !auto || a.autoQuoteIdentifiers {
		for key, segment := range idents {
			if !utils.InSSlice(strings.Trim(segment, " "), a.Unquoteable) && segment != "" && segment != " " {
				segment = strings.ReplaceAll(segment, a.QuoteIdentifierSymbol(), "")
				if segment[0:1] == `:` || segment[0:1] == `'` || segment[0:1] == `\\` || segment[0:1] == `{` || utils.InSSlice(segment[0:1], []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}) {
					idents[key] = segment
				} else {
					idents[key] = a.quoteIdentifier(segment, auto)
				}
			} else {
				idents[key] = segment
			}
		}
	}

	return strings.Join(idents, "")
}

// quoteString qoutes string value
func (a *adapter) quoteString(value string) string {
	return `'` + utils.Addslashes(value) + `'`
}

// Convert an array, string, or Expr object into a string to put in a WHERE clause
func (a *adapter) whereExpr(cond interface{}) string {
	if cond == nil {
		return ""
	}

	where := make([]string, 0)
	switch cond.(type) {
	case string:
		where = append(where, "( "+cond.(string)+" )")

	case []string:
		for _, term := range cond.([]string) {
			where = append(where, "( "+term+" )")
		}

	case map[string]interface{}:
		for cnd, term := range cond.(map[string]interface{}) {
			where = append(where, "( "+a.QuoteInto(cnd, term, -1)+" )")
		}

	case []interface{}:
		for _, term := range cond.([]interface{}) {
			switch term.(type) {
			case *dbselect.Expr:
				where = append(where, "( "+term.(*dbselect.Expr).ToString()+" )")

			case dbselect.Interface:
				where = append(where, "( "+term.(dbselect.Interface).Assemble()+" )")

			case string:
				where = append(where, "( "+term.(string)+" )")
			}
		}
	}

	return strings.Join(where, " AND ")
}

// NewAdapter creates a new adapter from given type and options
func NewAdapter(adapterType string, options config.Config) (ai Interface, err error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[adapterType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database adapter type \"%v\"", adapterType)
}

// Register registers a handler for database adapter creation
func Register(adapterType string, handler func(*Config) (Interface, error)) {
	buildHandlers[adapterType] = handler
}
