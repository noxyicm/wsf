package db

import (
	goctx "context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wsf/config"
	"wsf/context"
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

// Adapter represents database adapter interface
type Adapter interface {
	Setup()
	Init() error
	Context() context.Context
	SetContext(ctx context.Context) error
	Connection(ctx context.Context) (Connection, error)
	Select() Select
	//Query(ctx context.Context, sql Select) (Rowset, error)
	Query(ctx context.Context, sql Select) ([]map[string]interface{}, error)
	QueryInto(ctx context.Context, dbs Select, o interface{}) ([]interface{}, error)
	//QueryRow(ctx context.Context, sql Select) (Row, error)
	QueryRow(ctx context.Context, sql Select) (map[string]interface{}, error)
	//PrepareRowset(rows *sql.Rows) ([]map[string]interface{}, error)
	PrepareRowset(rows *sql.Rows) ([]map[string]interface{}, error)
	//PrepareRow(row []sql.RawBytes, columns []*sql.ColumnType) (data map[string]interface{}, err error)
	PrepareRow(rows *sql.Rows) (map[string]interface{}, error)
	//PrepareRow(row *sql.Row) (*RowData, error)
	//PrepareRow(row []*ColumnData) (map[string]interface{}, error)
	//Profiler()
	Insert(ctx context.Context, table string, data map[string]interface{}) (int, error)
	Update(ctx context.Context, table string, data map[string]interface{}, cond map[string]interface{}) (bool, error)
	Delete(ctx context.Context, table string, cond map[string]interface{}) (bool, error)
	//Select
	//FetchAll
	//FetchAssoc
	//FetchCol
	//FetchPairs
	//FetchOne
	NextSequenceID(sequence string) int
	BeginTransaction(ctx context.Context) (Transaction, error)
	DescribeTable(table string, schema string) (map[string]*TableColumn, error)
	Quote(interface{}) string
	QuoteIdentifier(interface{}, bool) string
	QuoteIdentifierAs(ident interface{}, alias string, auto bool) string
	QuoteIdentifierSymbol() string
	QuoteInto(text string, value interface{}, count int) string
	QuoteColumnAs(ident interface{}, alias string, auto bool) string
	QuoteTableAs(ident interface{}, alias string, auto bool) string
	SupportsParameters(param string) bool
	SetOptions(options *AdapterConfig) error
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
func (a *DefaultAdapter) Connection(ctx context.Context) (conn Connection, err error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	conn, err = NewConnectionFromConfig(a.Options.Connection.Type, a.Options.Connection)
	if err != nil {
		return nil, err
	}

	conn.SetContext(ctx)
	return conn, nil
}

// Query runs a query
func (a *DefaultAdapter) Query(ctx context.Context, dbs Select) ([]map[string]interface{}, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := a.Db.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	return a.PrepareRowset(rows)
}

// QueryInto runs a query and returns sql.Rows
func (a *DefaultAdapter) QueryInto(ctx context.Context, dbs Select, o interface{}) ([]interface{}, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := a.Db.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	t := reflect.TypeOf(o)
	var v reflect.Value
	if t.Kind() == reflect.Ptr {
		v = reflect.ValueOf(t.Elem()).Elem()
	} else if t.Kind() == reflect.Struct {
		v = reflect.New(t)
	} else {
		return nil, errors.New("Nope")
	}

	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
		return nil, err
	}

	rt := make([]interface{}, 0)
	for rows.Next() {
		values, err := a.resolveValues(columns, v.Elem())
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
			return nil, err
		}

		if err := rows.Scan(values...); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		ptr := v.Interface()
		rt = append(rt, ptr)
	}
	return rt, nil
}

// resolveValues returns slice of call arguments for service Init method
func (a *DefaultAdapter) resolveValues(columns []string, o reflect.Value) (values []interface{}, err error) {
	values = make([]interface{}, len(columns))
	var valueField reflect.Value
	for i := range columns {
		if columns[i] == "id" {
			valueField = o.FieldByName("ID")
		} else {
			valueField = o.FieldByName(strings.ToTitle(columns[i][:1]) + columns[i][1:])
		}
		//valueField := o.FieldByName(columns[i])
		values[i] = valueField.Addr().Interface()
	}
	return
}

// QueryRow runs a query
func (a *DefaultAdapter) QueryRow(ctx context.Context, dbs Select) (map[string]interface{}, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := a.Db.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "Database query Error")
	}
	defer rows.Close()

	return a.PrepareRow(rows)
}

// Insert inserts new row into table
func (a *DefaultAdapter) Insert(ctx context.Context, table string, data map[string]interface{}) (int, error) {
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

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	result, err := stmt.ExecContext(qctx, binds...)
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
func (a *DefaultAdapter) Update(ctx context.Context, table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
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

	sql := "UPDATE " + a.QuoteIdentifier(table, true) + " SET " + strings.Join(set, ", ") + ""
	if where != "" {
		sql = sql + " WHERE " + where
	}

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx, binds...)
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
func (a *DefaultAdapter) Delete(ctx context.Context, table string, cond map[string]interface{}) (bool, error) {
	where := a.whereExpr(cond)

	sql := "DELETE FROM " + a.QuoteIdentifier(table, true)
	if where != "" {
		sql = sql + " WHERE " + where
	}

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return true, nil
}

// BeginTransaction creates a new database transaction
func (a *DefaultAdapter) BeginTransaction(ctx context.Context) (Transaction, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	tx, err := a.Db.BeginTx(ctx, &sql.TxOptions{Isolation: a.Options.Transaction.IsolationLevel, ReadOnly: a.Options.Transaction.ReadOnly})
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

	case []int:
		v := value.([]int)
		sl := make([]string, len(v))
		i := 0
		for _, val := range v {
			sl[i] = a.Quote(val)
			i++
		}

		return strings.Join(sl, ", ")

	case []int64:
		v := value.([]int64)
		sl := make([]string, len(v))
		i := 0
		for _, val := range v {
			sl[i] = a.Quote(val)
			i++
		}

		return strings.Join(sl, ", ")

	case []int32:
		v := value.([]int32)
		sl := make([]string, len(v))
		i := 0
		for _, val := range v {
			sl[i] = a.Quote(val)
			i++
		}

		return strings.Join(sl, ", ")

	case []float32:
		v := value.([]float32)
		sl := make([]string, len(v))
		i := 0
		for _, val := range v {
			sl[i] = a.Quote(val)
			i++
		}

		return strings.Join(sl, ", ")

	case []float64:
		v := value.([]float64)
		sl := make([]string, len(v))
		i := 0
		for _, val := range v {
			sl[i] = a.Quote(val)
			i++
		}

		return strings.Join(sl, ", ")

	case int:
		return strconv.Itoa(value.(int))
	case int8:
		return strconv.Itoa(int(value.(int8)))
	case int16:
		return strconv.Itoa(int(value.(int16)))
	case int32:
		return strconv.Itoa(int(value.(int32)))
	case int64:
		return strconv.Itoa(int(value.(int64)))
	case uint:
		return strconv.Itoa(int(value.(uint)))
	case uint8:
		return strconv.Itoa(int(value.(uint8)))
	case uint16:
		return strconv.Itoa(int(value.(uint16)))
	case uint32:
		return strconv.Itoa(int(value.(uint32)))
	case uint64:
		return strconv.Itoa(int(value.(uint64)))

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
			reg, err := regexp.Compile(find)
			if err == nil {
				v = reg.ReplaceAllString(v, `{`+strconv.Itoa((len(literals)-1))+`}`)
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
			quoted = strings.ReplaceAll(quoted, `{`+strconv.Itoa(key)+`}`, literal)
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

// PrepareRowset parses sql.Rows into mapstructure slice
func (a *DefaultAdapter) PrepareRowset(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "Database prepare result Error")
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range columns {
		scanArgs[i] = a.reference(columns[i].ScanType())
	}

	data := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err = rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		rowdata := make(map[string]interface{})
		for i := range columns {
			rowdata[columns[i].Name()] = a.dereference(scanArgs[i])
		}
		data = append(data, rowdata)
	}

	if err := rows.Err(); err != nil {
		if err == sql.ErrNoRows {
			return data, nil
		}

		return nil, errors.Wrap(err, "Database prepare result Error")
	}

	return data, nil
}

// PrepareRow parses a RawBytes into map structure
func (a *DefaultAdapter) PrepareRow(rows *sql.Rows) (map[string]interface{}, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "Database prepare result error")
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range columns {
		scanArgs[i] = a.reference(columns[i].ScanType())
	}

	data := make(map[string]interface{})
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, errors.Wrap(err, "Database prepare result error")
		}

		return nil, nil
	}

	if err = rows.Scan(scanArgs...); err != nil {
		return nil, errors.Wrap(err, "Database prepare result error")
	}

	for i := range columns {
		data[columns[i].Name()] = a.dereference(scanArgs[i])
	}

	return data, nil
}

// PrepareRow parses a RawBytes into map structure
/*func (a *DefaultAdapter) PrepareRow(row []sql.RawBytes, columns []*sql.ColumnType) (data map[string]interface{}, err error) {
	err = nil
	data = make(map[string]interface{})
	for i, col := range row {
		if columns[i].ScanType() == reflect.TypeOf(sql.NullBool{}) {
			v := sql.NullBool{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Bool
				} else {
					data[columns[i].Name()] = false
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullInt64{}) {
			v := sql.NullInt64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Int64
				} else {
					data[columns[i].Name()] = 0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullFloat64{}) {
			v := sql.NullFloat64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Float64
				} else {
					data[columns[i].Name()] = 0.0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullString{}) {
			v := sql.NullString{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.String
				} else {
					data[columns[i].Name()] = ""
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullTime{}) {
			v := sql.NullTime{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					t := v.Time
					data[columns[i].Name()] = t.Local()
				} else {
					data[columns[i].Name()] = time.Time{}
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.RawBytes{}) {
			data[columns[i].Name()] = string(col)
		} else if columns[i].ScanType() == reflect.TypeOf((int)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = 0
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 0)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int8)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int8(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 8)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int16)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int16(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 16)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int32)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int32(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int64)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int64(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float32)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = float32(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseFloat(string(col), 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float64)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = float64(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseFloat(string(col), 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf(time.Time{}) {
			if string(col) == "" {
				data[columns[i].Name()] = time.Time{}
			} else {
				var t time.Time
				t, err = time.Parse("2006-01-02T15:04:05Z", string(col))
				data[columns[i].Name()] = t.Local()
			}
		} else if columns[i].ScanType() == reflect.TypeOf(true) {
			data[columns[i].Name()], err = strconv.ParseBool(string(col))
		} else {
			data[columns[i].Name()] = string(col)
		}

		if err != nil {
			return data, err
		}
	}

	return data, err
}*/

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

// returns a value from pointer
func (a *DefaultAdapter) dereference(v interface{}) interface{} {
	switch t := v.(type) {
	case *bool:
		return *t
	case *sql.NullBool:
		return t.Bool
	case *[]byte:
		return string(*t)
	case *string:
		return *t
	case *sql.NullString:
		return t.String
	case *int:
		return *t
	case *int8:
		return *t
	case *int16:
		return *t
	case *int32:
		return *t
	case *int64:
		return *t
	case *sql.NullInt64:
		return t.Int64
	case *float32:
		return *t
	case *float64:
		return *t
	case *sql.NullFloat64:
		return t.Float64
	case *time.Time:
		return *t
	default:
		return nil
	}
}

// creates a pointer to value
func (a *DefaultAdapter) reference(tp reflect.Type) interface{} {
	if tp == reflect.TypeOf(sql.NullBool{}) {
		var v sql.NullBool
		return &v
	} else if tp == reflect.TypeOf(sql.NullInt64{}) {
		var v sql.NullInt64
		return &v
	} else if tp == reflect.TypeOf(sql.NullFloat64{}) {
		var v sql.NullFloat64
		return &v
	} else if tp == reflect.TypeOf(sql.NullString{}) {
		var v sql.NullString
		return &v
	} else if tp == reflect.TypeOf(sql.NullTime{}) {
		var v time.Time
		return &v
	} else if tp == reflect.TypeOf(sql.RawBytes{}) {
		var v []byte
		return &v
	} else if tp == reflect.TypeOf((int)(0)) {
		var v int
		return &v
	} else if tp == reflect.TypeOf((int8)(0)) {
		var v int8
		return &v
	} else if tp == reflect.TypeOf((int16)(0)) {
		var v int16
		return &v
	} else if tp == reflect.TypeOf((int32)(0)) {
		var v int32
		return &v
	} else if tp == reflect.TypeOf((int64)(0)) {
		var v int64
		return &v
	} else if tp == reflect.TypeOf((float32)(0)) {
		var v float32
		return &v
	} else if tp == reflect.TypeOf((float64)(0)) {
		var v float64
		return &v
	} else if tp == reflect.TypeOf(time.Time{}) {
		var v time.Time
		return &v
	} else if tp == reflect.TypeOf(true) {
		var v bool
		return &v
	} else {
		var v string
		return &v
	}
}
