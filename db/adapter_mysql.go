package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
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
func (a *MySQL) Init(ctx context.Context) (err error) {
	db, err := sql.Open("mysql", a.driverConfig.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "MySQL Error")
	}

	db.SetConnMaxLifetime(time.Duration(a.Options.ConnectionMaxLifeTime) * time.Second)
	db.SetMaxIdleConns(a.Options.MaxIdleConnections)
	db.SetMaxOpenConns(a.Options.MaxOpenConnections)

	if a.PingTimeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, a.PingTimeout*time.Second)
		defer cancel()

		if err = db.PingContext(tctx); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	} else {
		if err = db.PingContext(ctx); err != nil {
			return errors.Wrap(err, "MySQL Error")
		}
	}

	a.Db = db
	a.Ctx = ctx
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
func (a *MySQL) Select() (Select, error) {
	sel, err := NewSelectFromConfig(Options().Select)
	if err != nil {
		return nil, err
	}

	sel.SetAdapter(a)
	return sel, nil
}

// DescribeTable returns information about columns in table
func (a *MySQL) DescribeTable(table string, schema string) (map[string]*TableColumn, error) {
	var sqlstr string
	if schema != "" {
		sqlstr = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE " + a.QuoteInto("table_name = ?", table, -1) + " AND " + a.QuoteInto("table_schema = ?", schema, -1)
	} else {
		sqlstr = "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE " + a.QuoteInto("table_name = ?", table, -1)
	}

	ctx, cancel := context.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
		fmt.Println(columns[i].Name())
	}

	desc := make(map[string]*TableColumn)
	var i int64
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, errors.Wrap(err, "CockroachDB Error")
		}

		d, err := PrepareRow(values, columns)
		if err != nil {
			return nil, errors.Wrap(err, "CockroachDB Error")
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
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	return desc, nil
}

// Limit is
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

// NewMySQLAdapter creates a new MySQL adapter
func NewMySQLAdapter(options *AdapterConfig) (ai Adapter, err error) {
	adp := &MySQL{}
	adp.Options = options
	adp.Setup()

	return adp, nil
}
