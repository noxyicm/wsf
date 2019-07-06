package adapter

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
	"wsf/db/dbselect"
	"wsf/db/statement"
	"wsf/errors"

	"github.com/go-sql-driver/mysql"
)

const (
	// TYPEMySQL represents mysql adapter
	TYPEMySQL = "MySQL"
)

func init() {
	Register(TYPEMySQL, NewMySQLAdapter)
}

// MySQL adapter for MySQL databeses
type MySQL struct {
	adapter
	driverConfig *mysql.Config
}

// Init a connection to database
func (a *MySQL) Init(ctx context.Context) (err error) {
	db, err := sql.Open("mysql", a.driverConfig.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "MySQL Error")
	}

	db.SetConnMaxLifetime(time.Duration(a.options.ConnectionMaxLifeTime) * time.Second)
	db.SetMaxIdleConns(a.options.MaxIdleConnections)
	db.SetMaxOpenConns(a.options.MaxOpenConnections)

	if a.pingTimeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, a.pingTimeout*time.Second)
		defer cancel()

		if err = db.PingContext(tctx); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	} else {
		if err = db.PingContext(ctx); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	}

	a.db = db
	a.ctx = ctx
	return nil
}

// SetOptions sets new options for MySQL adapter
func (a *MySQL) SetOptions(options *Config) Interface {
	a.options = options
	return a
}

// Options returns MySQL adapter options
func (a *MySQL) Options() *Config {
	return a.options
}

// Select creates a new adapter specific select object
func (a *MySQL) Select() (dbselect.Interface, error) {
	sel, err := dbselect.NewSelect(a.options.Select.Type, a.options.Select)
	if err != nil {
		return nil, err
	}

	sel.SetAdapter(a)
	return sel, nil
}

// Insert inserts new row into table
func (a *MySQL) Insert(table string, data map[string]interface{}) (int, error) {
	cols := []string{}
	vals := []string{}
	binds := []interface{}{}
	i := 0
	for col, val := range data {
		cols = append(cols, a.QuoteIdentifier(col, true))

		switch val.(type) {
		case *dbselect.Expr:
			vals = append(vals, val.(*dbselect.Expr).ToString())

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

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.pingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.db.PrepareContext(ctx, sql)
	if err != nil {
		return 0, errors.Wrap(err, "Database insert Error")
	}

	ctx, cancel = context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
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
func (a *MySQL) Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	set := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		var value string

		switch val.(type) {
		case *dbselect.Expr:
			value = val.(*dbselect.Expr).ToString()

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

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.pingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.db.PrepareContext(ctx, sql)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
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

// DescribeTable returns information about columns in table
func (a *MySQL) DescribeTable(table string, schema string) (map[string]*ColumnMetadata, error) {
	var sql string
	if schema != "" {
		sql = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = " + a.QuoteIdentifier(table, true) + " AND table_schema = " + a.QuoteIdentifier(schema, true)
	} else {
		sql = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = " + a.QuoteIdentifier(table, true)
	}

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.pingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	desc := make(map[string]*ColumnMetadata)
	var i int64
	for rows.Next() {
		def := map[string]interface{}{
			"TABLE_CATALOG":            "",
			"TABLE_SCHEMA":             "",
			"TABLE_NAME":               "",
			"COLUMN_NAME":              "",
			"ORDINAL_POSITION":         0,
			"COLUMN_DEFAULT":           nil,
			"IS_NULLABLE":              "NO",
			"DATA_TYPE":                "",
			"CHARACTER_MAXIMUM_LENGTH": 0,
			"CHARACTER_OCTET_LENGTH":   0,
			"NUMERIC_PRECISION":        0,
			"NUMERIC_SCALE":            0,
			"DATETIME_PRECISION":       0,
			"CHARACTER_SET_NAME":       "",
			"COLLATION_NAME":           "",
			"COLUMN_TYPE":              "",
			"COLUMN_KEY":               "",
			"EXTRA":                    "",
		}

		if err := rows.Scan(&def); err != nil {
			return nil, err
		}

		row := &ColumnMetadata{
			TableSchema:  def["TABLE_CATALOG"].(string),
			TableName:    def["TABLE_NAME"].(string),
			Name:         def["COLUMN_NAME"].(string),
			Default:      def["COLUMN_DEFAULT"],
			Position:     def["ORDINAL_POSITION"].(int64),
			Type:         def["DATA_TYPE"].(string),
			Length:       def["CHARACTER_MAXIMUM_LENGTH"].(int64),
			Precision:    def["NUMERIC_PRECISION"].(int64),
			Scale:        def["NUMERIC_SCALE"].(int64),
			CharacterSet: def["CHARACTER_SET_NAME"].(string),
			Collation:    def["COLLATION_NAME"].(string),
			ColumnType:   def["COLUMN_TYPE"].(string),
			ColumnKey:    def["COLUMN_KEY"].(string),
			Extra:        def["EXTRA"].(string),
		}

		if def["IS_NULLABLE"].(string) == "YES" {
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

		desc[def["COLUMN_NAME"].(string)] = row
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return desc, nil
}

// Limit is
func (a *MySQL) Limit(sql string, count int, offset int) string {
	if count <= 0 {
		return sql
	}

	if offset < 0 {
		return sql
	}

	sql = sql + " LIMIT " + strconv.Itoa(count)
	if offset > 0 {
		sql = sql + " OFFSET " + strconv.Itoa(offset)
	}

	return sql
}

// FormatDSN returns a formated dsn string
func (a *MySQL) FormatDSN() string {
	return a.driverConfig.FormatDSN()
}

// NewMySQLAdapter creates a new MySQL adapter
func NewMySQLAdapter(options *Config) (ai Interface, err error) {
	adp := &MySQL{}
	adp.identifierSymbol = "`"
	adp.autoQuoteIdentifiers = true
	adp.defaultStatementType = statement.TYPEDefault
	adp.pingTimeout = time.Duration(options.PingTimeout) * time.Second
	adp.queryTimeout = time.Duration(options.QueryTimeout) * time.Second

	//sql.Register(name string, driver driver.Driver)
	adp.driverConfig = mysql.NewConfig()
	adp.driverConfig.User = options.Username
	adp.driverConfig.Passwd = options.Password
	adp.driverConfig.Net = options.Protocol
	adp.driverConfig.Addr = options.Host
	adp.driverConfig.DBName = options.DBname
	adp.driverConfig.Loc = options.TimeFormat
	adp.driverConfig.Collation = options.Charset
	//TLSConfig

	adp.Unquoteable = []string{
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

	adp.Spliters = []string{
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

	adp.UnquoteableFunctions = []string{
		"CONCAT",
		"LOWER",
		"UPPER",
		"DATE",
		"UNIX_TIMESTAMP",
		"AVG",
		"SUM",
		"COUNT",
	}

	adp.options = options
	adp.params = map[string]interface{}{
		"positional": true,
		"named":      false,
	}
	return adp, nil
}
