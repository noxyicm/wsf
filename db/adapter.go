package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wsf/config"
	"wsf/errors"
	"wsf/utils"
)

// Public constants
const (
	LOGInsert     = "insert"
	LOGPreUpdate  = "preupdate"
	LOGPostUpdate = "preupdate"
	LOGPreDelete  = "predelete"
	LOGPostDelete = "predelete"
)

var (
	buildAdapterHandlers = map[string]func(*AdapterConfig) (Adapter, error){}

	// RegexpSingleQuote matches single quote in column name
	RegexpSingleQuote = regexp.MustCompile(`('.+?')`)
)

// Adapter database adapter
type Adapter interface {
	Init(ctx context.Context) error
	Context() context.Context
	SetContext(ctx context.Context) error
	Connection() (Connection, error)
	Select() (Select, error)
	Query(sql Select) (Rowset, error)
	QueryRow(sql Select) (Row, error)
	//Profiler()
	Insert(table string, data map[string]interface{}) (int, error)
	Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error)
	Delete(table string, cond map[string]interface{}) (bool, error)
	//Select
	//FetchAll
	//FetchAssoc
	//FetchCol
	//FetchPairs
	//FetchOne
	LastInsertID() int
	NextSequenceID(sequence string) int
	BeginTransaction() (Transaction, error)
	DescribeTable(table string, schema string) (map[string]*TableColumn, error)
	Quote(interface{}) string
	QuoteIdentifier(interface{}, bool) string
	QuoteIdentifierAs(ident interface{}, alias string, auto bool) string
	QuoteIdentifierSymbol() string
	QuoteInto(text string, value interface{}, count int) string
	QuoteColumnAs(ident interface{}, alias string, auto bool) string
	QuoteTableAs(ident interface{}, alias string, auto bool) string
	SupportsParameters(param string) bool
	SetOptions(options *AdapterConfig) Adapter
	GetOptions() *AdapterConfig
	FormatDSN() string
	Limit(sql string, count int, offset int) string
	FoldCase(s string) string
}

// NewAdapter creates a new adapter from given type and options
func NewAdapter(adapterType string, options config.Config) (ai Adapter, err error) {
	cfg := &AdapterConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildAdapterHandlers[adapterType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database adapter type \"%v\"", adapterType)
}

// RegisterAdapter registers a handler for database adapter creation
func RegisterAdapter(adapterType string, handler func(*AdapterConfig) (Adapter, error)) {
	buildAdapterHandlers[adapterType] = handler
}

// DefaultAdapter is a base object for adapters
type DefaultAdapter struct {
	Options                    *AdapterConfig
	Db                         *sql.DB
	Ctx                        context.Context
	Params                     map[string]interface{}
	DefaultStatementType       string
	PingTimeout                time.Duration
	QueryTimeout               time.Duration
	ConnectionMaxLifeTime      int
	MaxIdleConnections         int
	MaxOpenConnections         int
	inTransaction              bool
	identifierSymbol           string
	AutoQuoteIdentifiers       bool
	AllowSerialization         bool
	AutoReconnectOnUnserialize bool
	lastInsertID               int
	lastInsertUUID             string

	Unquoteable          []string
	Spliters             []string
	UnquoteableFunctions []string
}

// Context returns adapter specific context
func (a *DefaultAdapter) Context() context.Context {
	return a.Ctx
}

// SetContext sets adapter specific context
func (a *DefaultAdapter) SetContext(ctx context.Context) error {
	a.Ctx = ctx
	return nil
}

// Connection returns a connection to database
func (a *DefaultAdapter) Connection() (conn Connection, err error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	conn, err = NewConnectionFromConfig(a.Options.Connection.Type, a.Options.Connection)
	if err != nil {
		return nil, err
	}

	conn.SetContext(a.Ctx)
	return conn, nil
}

// Query runs a query
func (a *DefaultAdapter) Query(sql Select) (Rowset, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	if err := sql.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	rows, err := a.Db.QueryContext(ctx, sql.Assemble(), sql.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	rst, err := NewRowset(Options().Rowset.Type, Options().Rowset)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	if err := rst.Prepare(rows); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	return rst, nil
}

// QueryRow runs a query
func (a *DefaultAdapter) QueryRow(dbs Select) (Row, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	rows, err := a.Db.QueryContext(ctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		rw, err := NewRow(Options().Row.Type, Options().Row)
		if err != nil {
			return nil, errors.Wrap(err, "Database query Error")
		}

		if err := rw.Prepare(values, columns); err != nil {
			return nil, err
		}

		return rw, nil
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	return nil, nil
}

// LastInsertID returns the last insert query ID
func (a *DefaultAdapter) LastInsertID() int {
	return a.lastInsertID
}

// LastInsertUUID returns the last insert query UUID
func (a *DefaultAdapter) LastInsertUUID() string {
	return a.lastInsertUUID
}

// Insert inserts new row into table
func (a *DefaultAdapter) Insert(table string, data map[string]interface{}) (int, error) {
	cols := []string{}
	vals := []string{}
	binds := []interface{}{}
	i := 0
	for col, val := range data {
		cols = append(cols, a.QuoteIdentifier(col, true))

		switch val.(type) {
		case *SQLExpr:
			vals = append(vals, val.(*SQLExpr).ToString())

		default:
			if a.SupportsParameters("positional") {
				vals = append(vals, "?")
				binds = append(binds, val)
			} else if a.SupportsParameters("named") {
				vals = append(vals, ":col"+strconv.Itoa(i))
				binds = append(binds, sql.Named("col"+strconv.Itoa(i), val))
				i++
			} else {
				return 0, errors.New("Adapter doesn't support positional or named binding")
			}
		}
	}

	sql := "INSERT INTO " + a.QuoteIdentifier(table, true) + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(vals, ", ") + ")"

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sql)
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	ctx, cancel = context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	result, err := stmt.ExecContext(ctx, binds...)
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	lastInsertID, err := result.LastInsertId()
	if err == nil {
		a.lastInsertID = int(lastInsertID)
	}

	return a.lastInsertID, nil
}

// Update updates rows into table be condition
func (a *DefaultAdapter) Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	set := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		var value string

		switch val.(type) {
		case *SQLExpr:
			value = val.(*SQLExpr).ToString()

		default:
			if a.SupportsParameters("positional") {
				value = "?"
				binds = append(binds, val)
				i++
			} else if a.SupportsParameters("named") {
				value = ":col" + strconv.Itoa(i)
				binds = append(binds, sql.Named("col"+strconv.Itoa(i), val))
				i++
			} else {
				return false, errors.New("Adapter doesn't support positional or named binding")
			}
		}

		set = append(set, a.QuoteIdentifier(col, true)+" = "+value)
	}

	where := a.whereExpr(cond)

	sql := "UPDATE " + a.QuoteIdentifier(table, true) + " SET (" + strings.Join(set, ", ") + ")"
	if where != "" {
		sql = sql + " WHERE " + where
	}

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx, binds...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var updatedID int
		if err := rows.Scan(&updatedID); err != nil {
			return true, err
		}
	}

	return true, nil
}

// Delete removes rows from table
func (a *DefaultAdapter) Delete(table string, cond map[string]interface{}) (bool, error) {
	where := a.whereExpr(cond)

	sql := "DELETE FROM " + a.QuoteIdentifier(table, true)
	if where != "" {
		sql = sql + " WHERE " + where
	}

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return true, nil
}

// BeginTransaction creates a new database transaction
func (a *DefaultAdapter) BeginTransaction() (Transaction, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	tx, err := a.Db.BeginTx(a.Ctx, &sql.TxOptions{Isolation: a.Options.Transaction.IsolationLevel, ReadOnly: a.Options.Transaction.ReadOnly})
	if err != nil {
		return nil, err
	}

	trns, err := NewTransaction(a.Options.Transaction.Type, tx)
	if err != nil {
		return nil, err
	}

	trns.SetContext(a.Ctx)
	return trns, err
}

// Quote a string
func (a *DefaultAdapter) Quote(value interface{}) string {
	switch value.(type) {
	case Select:
		return "(" + value.(Select).Assemble() + ")"

	case *SQLExpr:
		return value.(*SQLExpr).ToString()

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
func (a *DefaultAdapter) QuoteInto(text string, value interface{}, count int) string {
	return strings.Replace(text, "?", a.Quote(value), count)
}

// QuoteIdentifier re
func (a *DefaultAdapter) QuoteIdentifier(ident interface{}, auto bool) string {
	return a.QuoteIdentifierAs(ident, "", auto)
}

// QuoteIdentifierAs a
func (a *DefaultAdapter) QuoteIdentifierAs(ident interface{}, alias string, auto bool) string {
	as := " " + strings.ToUpper(SQLAs) + " "
	quoted := ""
	literals := make([]string, 0)
	idents := make([]interface{}, 0)

	switch ident.(type) {
	case Select:
		quoted = "(" + ident.(Select).Assemble() + ")"

	case *SQLExpr:
		quoted = ident.(*SQLExpr).ToString()

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
		reg, err := regexp.Compile(`(?i)(` + strings.ToLower(strings.Join(a.UnquoteableFunctions, "|")) + `)[\s]*\((.+?)\)`)
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
				case Select:
					segments = append(segments, "("+segment.(Select).Assemble()+")")

				case *SQLExpr:
					segments = append(segments, segment.(*SQLExpr).ToString())

				case string:
					split := []string{}
					spliters := []string{}
					subsplit := [][]string{}
					for _, spliter := range a.Spliters {
						spliters = append(spliters, regexp.QuoteMeta(spliter))
					}

					reg, err := regexp.Compile(`(?i)([^` + strings.Join(spliters, "") + `]*)([` + strings.Join(spliters, "]|[") + `]+)([^` + strings.Join(spliters, "") + `]*)`)
					if err == nil {
						subsplit = reg.FindAllStringSubmatch(segment.(string), -1)
						for _, subsplitmatch := range subsplit {
							for subkey, match := range subsplitmatch {
								if subkey == 0 {
									continue
								}

								split = append(split, match)
							}
							//break
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
func (a *DefaultAdapter) QuoteIdentifierSymbol() string {
	return a.identifierSymbol
}

// QuoteColumnAs quote a column identifier and alias
func (a *DefaultAdapter) QuoteColumnAs(ident interface{}, alias string, auto bool) string {
	return a.QuoteIdentifierAs(ident, alias, auto)
}

// QuoteTableAs quotes a table identifier and alias
func (a *DefaultAdapter) QuoteTableAs(ident interface{}, alias string, auto bool) string {
	return a.QuoteIdentifierAs(ident, alias, auto)
}

// SupportsParameters returns true if adapter supports
func (a *DefaultAdapter) SupportsParameters(param string) bool {
	if v, ok := a.Params[param]; ok {
		return v.(bool)
	}

	return false
}

// FoldCase folds a case
func (a *DefaultAdapter) FoldCase(s string) string {
	return s
}

// quotes identifier
func (a *DefaultAdapter) quoteIdentifier(ident string, auto bool) string {
	if !auto || a.AutoQuoteIdentifiers {
		q := a.QuoteIdentifierSymbol()
		return q + strings.ReplaceAll(ident, q, q+q) + q
	}

	return ident
}

// quotes specifics
func (a *DefaultAdapter) quoteIdentifierSpec(idents []string, auto bool) string {
	if !auto || a.AutoQuoteIdentifiers {
		for key, segment := range idents {
			if !utils.InSSlice(strings.TrimSpace(segment), a.Unquoteable) && segment != "" && segment != " " {
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
func (a *DefaultAdapter) quoteString(value string) string {
	return `'` + utils.Addslashes(value) + `'`
}

// Convert an array, string, or Expr object into a string to put in a WHERE clause
func (a *DefaultAdapter) whereExpr(cond interface{}) string {
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
			case *SQLExpr:
				where = append(where, "( "+term.(*SQLExpr).ToString()+" )")

			case Select:
				where = append(where, "( "+term.(Select).Assemble()+" )")

			case string:
				where = append(where, "( "+term.(string)+" )")
			}
		}
	}

	return strings.Join(where, " AND ")
}
