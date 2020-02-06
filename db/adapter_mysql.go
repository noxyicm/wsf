package db

import (
	goctx "context"
	"database/sql"
	"reflect"
	"strconv"
	"strings"
	"time"
	"wsf/context"
	"wsf/errors"

	"github.com/go-sql-driver/mysql"
)

const (
	// TYPEAdapterMySQL represents mysql adapter
	TYPEAdapterMySQL = "MySQL"
)

func init() {
	RegisterAdapter(TYPEAdapterMySQL, NewMySQLAdapter)
}

// MySQL adapter for MySQL databeses
type MySQL struct {
	DefaultAdapter
	driverConfig *mysql.Config
}

// Setup the adapter
func (a *MySQL) Setup() {
	a.identifierSymbol = "`"
	a.AutoQuoteIdentifiers = true
	a.PingTimeout = time.Duration(a.Options.PingTimeout) * time.Second
	a.QueryTimeout = time.Duration(a.Options.QueryTimeout) * time.Second

	//sql.Register(name string, driver driver.Driver)
	a.driverConfig = mysql.NewConfig()
	a.driverConfig.User = a.Options.Username
	a.driverConfig.Passwd = a.Options.Password
	a.driverConfig.Net = a.Options.Protocol
	a.driverConfig.Addr = a.Options.Host
	a.driverConfig.DBName = a.Options.DBname
	a.driverConfig.Loc = a.Options.TimeFormat
	a.driverConfig.Collation = a.Options.Charset
	//TLSConfig

	a.Unquoteable = []string{
		"BETWEEN",
		"LIKE",
		"AND",
		"OR",
		"=",
		"!=",
		">",
		">=",
		"<",
		"<=",
		"<>",
		"/",
		"+",
		"-",
		"?",
		"*",
		"(",
		")",
		"IS",
		"NOT",
		"NULL",
		"IN",
		"IN(",
		" ",
		".",
	}

	a.Spliters = []string{
		"=",
		"!=",
		">",
		">=",
		"<",
		"<=",
		"<>",
		"/",
		"+",
		"-",
		".",
		` `,
		"(",
		")",
	}

	a.UnquoteableFunctions = []string{
		"CONCAT",
		"LOWER",
		"UPPER",
		"DATE",
		"UNIX_TIMESTAMP",
		"AVG",
		"SUM",
		"COUNT",
	}

	a.Params = map[string]interface{}{
		"positional": true,
		"named":      false,
	}
}

// Init a connection to database
func (a *MySQL) Init() (err error) {
	db, err := sql.Open("mysql", a.driverConfig.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "MySQL Error")
	}

	db.SetConnMaxLifetime(time.Duration(a.Options.ConnectionMaxLifeTime) * time.Second)
	db.SetMaxIdleConns(a.Options.MaxIdleConnections)
	db.SetMaxOpenConns(a.Options.MaxOpenConnections)

	if a.PingTimeout > 0 {
		tctx, cancel := goctx.WithTimeout(goctx.Background(), a.PingTimeout*time.Second)
		defer cancel()

		if err = db.PingContext(tctx); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	} else {
		if err = db.PingContext(goctx.Background()); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	}

	a.Db = db
	return nil
}

// SetOptions sets new options for MySQL adapter
func (a *MySQL) SetOptions(options *AdapterConfig) error {
	a.Options = options
	return nil
}

// GetOptions returns MySQL adapter options
func (a *MySQL) GetOptions() *AdapterConfig {
	return a.Options
}

// Select creates a new adapter specific select object
func (a *MySQL) Select() Select {
	sel := NewSelectFromConfig(Options().Select)
	sel.SetAdapter(a)
	return sel
}

// Query runs a query
func (a *MySQL) Query(ctx context.Context, dbs Select) ([]map[string]interface{}, error) {
	if a.Db == nil {
		return nil, errors.New("Database is not initialized")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "MySQL query error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := a.Db.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "MySQL query error")
	}
	defer rows.Close()

	return a.PrepareRowset(rows)
}

// QueryRow runs a query
func (a *MySQL) QueryRow(ctx context.Context, dbs Select) (map[string]interface{}, error) {
	if a.Db == nil {
		return nil, errors.New("MySQL is not initialized")
	}

	if err := dbs.Err(); err != nil {
		return nil, errors.Wrap(err, "MySQL query Error")
	}

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := a.Db.QueryContext(qctx, dbs.Assemble(), dbs.Binds()...)
	if err != nil {
		return nil, errors.Wrap(err, "MySQL query Error")
	}
	defer rows.Close()

	return a.PrepareRow(rows)
}

// PrepareRowset parses sql.Rows into mapstructure slice
func (a *MySQL) PrepareRowset(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "MySQL prepare result Error")
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

		return nil, errors.Wrap(err, "MySQL prepare result Error")
	}

	return data, nil
}

// PrepareRow parses a RawBytes into map structure
func (a *MySQL) PrepareRow(rows *sql.Rows) (map[string]interface{}, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "MySQL prepare result error")
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range columns {
		scanArgs[i] = a.reference(columns[i].ScanType())
	}

	data := make(map[string]interface{})
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, errors.Wrap(err, "MySQL prepare result error")
		}

		return nil, nil
	}

	if err = rows.Scan(scanArgs...); err != nil {
		return nil, errors.Wrap(err, "MySQL prepare result error")
	}

	for i := range columns {
		data[columns[i].Name()] = a.dereference(scanArgs[i])
	}

	return data, nil
}

// DescribeTable returns information about columns in table
func (a *MySQL) DescribeTable(table string, schema string) (map[string]*TableColumn, error) {
	var sqlstr string
	if schema != "" {
		sqlstr = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE " + a.QuoteInto("table_name = ?", table, -1) + " AND " + a.QuoteInto("table_schema = ?", schema, -1)
	} else {
		sqlstr = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE " + a.QuoteInto("table_name = ?", table, -1)
	}

	ctx, cancel := goctx.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ctx, cancel = goctx.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "MySQL Error")
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range columns {
		scanArgs[i] = a.reference(columns[i].ScanType())
	}

	desc := make(map[string]*TableColumn)
	var i int64
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, errors.Wrap(err, "MySQL Error")
		}

		d := make(map[string]interface{})
		for i := range columns {
			d[columns[i].Name()] = a.dereference(scanArgs[i])
		}

		row := &TableColumn{
			TableSchema:  d["TABLE_SCHEMA"].(string),
			TableName:    d["TABLE_NAME"].(string),
			Name:         d["COLUMN_NAME"].(string),
			Default:      d["COLUMN_DEFAULT"],
			DataType:     d["DATA_TYPE"].(string),
			Length:       d["CHARACTER_MAXIMUM_LENGTH"].(int64),
			Precision:    d["NUMERIC_PRECISION"].(int64),
			Scale:        d["NUMERIC_SCALE"].(int64),
			CharacterSet: d["CHARACTER_SET_NAME"].(string),
			Collation:    d["COLLATION_NAME"].(string),
			ColumnType:   d["COLUMN_TYPE"].(string),
			ColumnKey:    d["COLUMN_KEY"].(string),
			Extra:        d["EXTRA"].(string),
		}

		row.Position, _ = strconv.ParseInt(d["ORDINAL_POSITION"].(string), 10, 64)
		if d["IS_NULLABLE"] == "YES" {
			row.IsNullable = true
		}

		if strings.ToUpper(row.ColumnKey) == "PRI" {
			row.Primary = true
			row.PrimaryPosition = i
			if row.Extra == "auto_increment" {
				row.Identity = true
			}
			i++
		}

		desc[row.Name] = row
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "MySQL Error")
	}

	return desc, nil
}

// Limit adds a limit clause to query
func (a *MySQL) Limit(sql string, count int, offset int) string {
	if count > 0 {
		sql = sql + " LIMIT " + strconv.Itoa(count)

		if offset > 0 {
			sql = sql + " OFFSET " + strconv.Itoa(offset)
		}
	}

	return sql
}

// NextSequenceID returns nex value from sequence
func (a *MySQL) NextSequenceID(sequence string) int {
	return 0
}

// FormatDSN returns a formated dsn string
func (a *MySQL) FormatDSN() string {
	return a.driverConfig.FormatDSN()
}

// returns a value from pointer
func (a *MySQL) dereference(v interface{}) interface{} {
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
	case *mysql.NullTime:
		return t.Time
	default:
		return nil
	}
}

// creates a pointer to value
func (a *MySQL) reference(tp reflect.Type) interface{} {
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
	} else if tp == reflect.TypeOf(mysql.NullTime{}) {
		var v mysql.NullTime
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
		var v *time.Time
		return &v
	} else if tp == reflect.TypeOf(true) {
		var v bool
		return &v
	} else {
		var v string
		return &v
	}
}

// NewMySQLAdapter creates a new MySQL adapter
func NewMySQLAdapter(options *AdapterConfig) (ai Adapter, err error) {
	adp := &MySQL{}
	adp.Options = options
	adp.Setup()

	return adp, nil
}
