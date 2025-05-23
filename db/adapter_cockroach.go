package db

import (
	"bytes"
	goctx "context"
	"crypto/rsa"
	"crypto/tls"
	"database/sql"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"

	// CockroachDB uses postgres package for tcp connections
	//_ "github.com/lib/pq"
	//_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/v4"
	// pgx needs a connections pool
	//_ "github.com/jackc/pgx/pgxpool"
	// CockroachDB uses pgx package for tcp connections
	_ "github.com/jackc/pgx/stdlib"
)

const (
	// TYPEAdapterCockroach represents cockroach db adapter
	TYPEAdapterCockroach = "CockroachDB"
)

func init() {
	RegisterAdapter(TYPEAdapterCockroach, NewCockroachAdapter)
}

// Cockroach adapter for CockroachDB databeses
type Cockroach struct {
	DefaultAdapter
	driverConfig *cockroachConfig
}

// Setup the adapter
func (a *Cockroach) Setup() {
	a.IdentifierSymbol = `"`
	a.AutoQuoteIdentifiers = true
	a.PingTimeout = time.Duration(a.Options.PingTimeout) * time.Second
	a.QueryTimeout = time.Duration(a.Options.QueryTimeout) * time.Second

	//sql.Register(name string, driver driver.Driver)
	a.driverConfig = &cockroachConfig{AllowNativePasswords: true}
	a.driverConfig.User = a.Options.Username
	a.driverConfig.Passwd = a.Options.Password
	a.driverConfig.Net = a.Options.Protocol
	a.driverConfig.Addr = a.Options.Host
	if a.Options.Port > 0 {
		a.driverConfig.Addr = a.driverConfig.Addr + ":" + strconv.Itoa(a.Options.Port)
	}
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
		"::",
		"SOME",
		"ANY",
		"ALL",
		"SIMILAR",
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
		" ",
	}

	a.UnquoteableFunctions = []string{
		"concat",
		"concat_ws",
		"lower",
		"upper",
		"md5",
		"btrim",
		"max",
		"min",
		"avg",
		"sum",
		"abs",
		"round",
		"ceil",
		"floor",
		"div",
		"count",
		"random",
		"current_timestamp",
		"greatest",
		"least",
		"if",
		"ifnull",
		"nullif",
	}

	a.Params = map[string]interface{}{
		"positional": true,
		"named":      false,
	}
}

// Init a connection to database
func (a *Cockroach) Init() (err error) {
	db, err := sql.Open("pgx", a.driverConfig.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "CockroachDB Error")
	}

	db.SetConnMaxLifetime(time.Duration(a.Options.ConnectionMaxLifeTime) * time.Second)
	db.SetMaxIdleConns(a.Options.MaxIdleConnections)
	db.SetMaxOpenConns(a.Options.MaxOpenConnections)

	if a.PingTimeout > 0 {
		tctx, cancel := goctx.WithTimeout(goctx.Background(), a.PingTimeout*time.Second)
		defer cancel()

		if err = db.PingContext(tctx); err != nil {
			return errors.Wrap(err, "CockroachDB Error")
		}
	} else {
		if err = db.PingContext(goctx.Background()); err != nil {
			return errors.Wrap(err, "CockroachDB Error")
		}
	}

	a.Db = db
	return nil
}

// SetOptions sets new options for CockroachDB adapter
func (a *Cockroach) SetOptions(options *AdapterConfig) error {
	a.Options = options
	return nil
}

// GetOptions returns CockroachDB adapter options
func (a *Cockroach) GetOptions() *AdapterConfig {
	return a.Options
}

// Select creates a new adapter specific select object
func (a *Cockroach) Select() Select {
	sel := NewSelectFromConfig(Options().Select)
	sel.SetAdapter(a)
	return sel
}

// Insert inserts new row into table
func (a *Cockroach) Insert(ctx context.Context, table string, data map[string]interface{}) (int, error) {
	cols := []string{}
	vals := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		cols = append(cols, a.QuoteIdentifier(col, true))

		switch v := val.(type) {
		case *SQLExpr:
			vals = append(vals, v.ToString())

		default:
			if a.SupportsParameters("positional") {
				vals = append(vals, "$"+strconv.Itoa(i))
				binds = append(binds, val)
				i++
			} else if a.SupportsParameters("named") {
				vals = append(vals, ":col"+strconv.Itoa(i))
				binds = append(binds, sql.Named("col"+strconv.Itoa(i), val))
				i++
			} else {
				return 0, errors.New("Adapter doesn't support positional or named binding")
			}
		}
	}

	sql := "INSERT INTO " + a.QuoteIdentifier(table, true) + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(vals, ", ") + ") RETURNING \"id\""

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return 0, errors.Wrap(err, "CockroachDB insert Error")
	}
	defer stmt.Close()

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	err = stmt.QueryRowContext(qctx, binds...).Scan(&a.LastInsertID)
	if err != nil {
		return 0, errors.Wrap(err, "CockroachDB insert Error")
	}

	return a.LastInsertID, nil
}

// Update updates rows into table be condition
func (a *Cockroach) Update(ctx context.Context, table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	set := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		var value string

		switch v := val.(type) {
		case *SQLExpr:
			value = v.ToString()

		default:
			if a.SupportsParameters("positional") {
				value = "$" + strconv.Itoa(i)
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

		set = append(set, a.QuoteIdentifier(col, true)+"="+value)
	}

	where := a.whereExpr(cond)

	sql := "UPDATE " + a.QuoteIdentifier(table, true) + " SET " + strings.Join(set, ", ") + ""
	if where != "" {
		sql = sql + " WHERE " + where
	}
	sql = sql + " RETURNING \"id\""

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return false, errors.Wrap(err, "CockroachDB update Error")
	}
	defer stmt.Close()

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx, binds...)
	if err != nil {
		return false, errors.Wrap(err, "CockroachDB update Error")
	}
	defer rows.Close()

	for rows.Next() {
		var updatedID int
		if err := rows.Scan(&updatedID); err != nil {
			return true, errors.Wrap(err, "CockroachDB update Error")
		}
	}

	return true, nil
}

// Delete removes rows from table
func (a *Cockroach) Delete(ctx context.Context, table string, cond map[string]interface{}) (bool, error) {
	where := a.whereExpr(cond)

	sql := "DELETE FROM " + a.QuoteIdentifier(table, true)
	if where != "" {
		sql = sql + " WHERE " + where
	}
	sql = sql + " RETURNING \"id\""

	pctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(pctx, sql)
	if err != nil {
		return false, errors.Wrap(err, "CockroachDB Error")
	}
	defer stmt.Close()

	qctx, cancel := goctx.WithTimeout(ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(qctx)
	if err != nil {
		return false, errors.Wrap(err, "CockroachDB Error")
	}
	defer rows.Close()

	deletedIDs := make([]int, 0)
	for rows.Next() {
		var deletedID int
		if err := rows.Scan(&deletedID); err != nil {
			return true, errors.Wrap(err, "CockroachDB delete Error")
		}

		deletedIDs = append(deletedIDs, deletedID)
	}

	if len(deletedIDs) > 0 {
		return true, nil
	}

	return false, nil
}

// DescribeTable returns information about columns in table
func (a *Cockroach) DescribeTable(table string, schema string) (map[string]*TableColumn, error) {
	var sqlstr string
	if schema != "" {
		sqlstr = "SHOW COLUMNS FROM " + a.QuoteIdentifier(schema, true) + "." + a.QuoteIdentifier(table, true)
	} else {
		sqlstr = "SHOW COLUMNS FROM " + a.QuoteIdentifier(table, true)
	}

	ctx, cancel := goctx.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.Db.PrepareContext(ctx, sqlstr)
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}
	defer stmt.Close()

	ctx, cancel = goctx.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	scanArgs := make([]interface{}, len(columns))
	for i := range columns {
		if _, ok := columns[i].Nullable(); !ok {
			scanArgs[i] = a.ReferenceNulls(columns[i].ScanType())
		} else {
			scanArgs[i] = a.Reference(columns[i].ScanType())
		}
	}

	desc := make(map[string]*TableColumn)
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, errors.Wrap(err, "CockroachDB Error")
		}

		d := make(map[string]interface{})
		for i := range columns {
			d[columns[i].Name()] = a.Dereference(scanArgs[i])
		}

		row := &TableColumn{
			TableSchema:  schema,
			TableName:    table,
			Name:         d["column_name"].(string),
			Default:      d["column_default"],
			Position:     0,
			DataType:     d["data_type"].(string),
			Length:       0,
			Precision:    0,
			Scale:        0,
			CharacterSet: "",
			Collation:    "",
			//ColumnType:   values["COLUMN_TYPE"].(string),
			//ColumnKey:    values["COLUMN_KEY"].(string),
			//Extra:        values["EXTRA"].(string),
		}

		if d["is_nullable"].(bool) {
			row.IsNullable = true
		}

		desc[row.Name] = row
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	if schema != "" {
		sqlstr = "SELECT constraint_name, table_schema, table_name, column_name FROM information_schema.key_column_usage WHERE " + a.QuoteInto("table_name = ?", table, -1) + " AND " + a.QuoteInto("table_schema = ?", schema, -1)
	} else {
		sqlstr = "SELECT constraint_name, table_schema, table_name, column_name FROM information_schema.key_column_usage WHERE " + a.QuoteInto("table_name = ?", table, -1)
	}

	ctx, cancel = goctx.WithTimeout(a.Ctx, time.Duration(a.PingTimeout)*time.Second)
	defer cancel()

	stmt2, err := a.Db.PrepareContext(ctx, sqlstr)
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}
	defer stmt2.Close()

	ctx, cancel = goctx.WithTimeout(a.Ctx, time.Duration(a.QueryTimeout)*time.Second)
	defer cancel()

	rows2, err := stmt2.QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}
	defer rows2.Close()

	columns, err = rows2.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	scanArgs = make([]interface{}, len(columns))
	for i := range columns {
		if _, ok := columns[i].Nullable(); !ok {
			scanArgs[i] = a.ReferenceNulls(columns[i].ScanType())
		} else {
			scanArgs[i] = a.Reference(columns[i].ScanType())
		}
	}

	var i int64
	for rows2.Next() {
		if err := rows2.Scan(scanArgs...); err != nil {
			return nil, errors.Wrap(err, "CockroachDB Error")
		}

		d := make(map[string]interface{})
		for i := range columns {
			d[columns[i].Name()] = a.Dereference(scanArgs[i])
		}

		if d["constraint_name"].(string) == "primary" {
			desc[d["column_name"].(string)].Primary = true
			desc[d["column_name"].(string)].PrimaryPosition = i
			i++
		}
	}

	if err := rows2.Err(); err != nil {
		return nil, errors.Wrap(err, "CockroachDB Error")
	}

	return desc, nil
}

// Limit ands a limit clause to statement
func (a *Cockroach) Limit(sql string, count int, offset int) string {
	if count > 0 {
		sql = sql + " LIMIT " + strconv.Itoa(count)

		if offset > 0 {
			sql = sql + " OFFSET " + strconv.Itoa(offset)
		}
	}

	return sql
}

// NextSequenceID returns nex value from sequence
func (a *Cockroach) NextSequenceID(sequence string) int {
	return 0
}

// BeginTransaction creates a new database transaction
func (a *Cockroach) BeginTransaction(ctx context.Context) (Transaction, error) {
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

	trns.SetAdapter(a)
	trns.SetContext(ctx)
	return trns, err
}

// FormatDSN returns a formated dsn string
func (a *Cockroach) FormatDSN() string {
	return a.driverConfig.FormatDSN()
}

// NewCockroachAdapter creates a new CockroachDB adapter
func NewCockroachAdapter(options *AdapterConfig) (ai Adapter, err error) {
	adp := &Cockroach{}
	adp.Options = options
	adp.Setup()

	return adp, nil
}

type cockroachConfig struct {
	User             string            // Username
	Passwd           string            // Password (requires User)
	Net              string            // Network type
	Addr             string            // Network address (requires Net)
	DBName           string            // Database name
	Params           map[string]string // Connection parameters
	Collation        string            // Connection collation
	Loc              *time.Location    // Location for time.Time values
	MaxAllowedPacket int               // Max packet size allowed
	ServerPubKey     string            // Server public key name
	pubKey           *rsa.PublicKey    // Server public key
	TLSConfig        string            // TLS configuration name
	tls              *tls.Config       // TLS configuration
	Timeout          time.Duration     // Dial timeout
	ReadTimeout      time.Duration     // I/O read timeout
	WriteTimeout     time.Duration     // I/O write timeout

	AllowAllFiles           bool // Allow all files to be used with LOAD DATA LOCAL INFILE
	AllowCleartextPasswords bool // Allows the cleartext client side plugin
	AllowNativePasswords    bool // Allows the native password authentication method
	AllowOldPasswords       bool // Allows the old insecure password method
	ClientFoundRows         bool // Return number of matching rows instead of rows changed
	ColumnsWithAlias        bool // Prepend table alias to column names
	InterpolateParams       bool // Interpolate placeholders into query string
	MultiStatements         bool // Allow multiple statements in one query
	ParseTime               bool // Parse time values to time.Time
	RejectReadOnly          bool // Reject read-only connections
}

// FormatDSN formats the given Config into a DSN string which can be passed to the driver
func (cfg *cockroachConfig) FormatDSN() string {
	var buf bytes.Buffer

	buf.WriteString("postgresql")
	buf.WriteByte(':')
	buf.WriteByte('/')
	buf.WriteByte('/')
	// [username[:password]@]
	if len(cfg.User) > 0 {
		buf.WriteString(cfg.User)
		if len(cfg.Passwd) > 0 {
			buf.WriteByte(':')
			buf.WriteString(cfg.Passwd)
		}
		buf.WriteByte('@')
	}

	// [protocol[(address)]]
	if len(cfg.Net) > 0 {
		//buf.WriteString(cfg.Net)
		if len(cfg.Addr) > 0 {
			//buf.WriteByte('(')
			buf.WriteString(cfg.Addr)
			//buf.WriteByte(')')
		}
	}

	// /dbname
	buf.WriteByte('/')
	buf.WriteString(cfg.DBName)

	// [?param1=value1&...&paramN=valueN]
	hasParam := false

	if cfg.AllowAllFiles {
		hasParam = true
		buf.WriteString("?allowAllFiles=true")
	}

	if cfg.AllowCleartextPasswords {
		if hasParam {
			buf.WriteString("&allowCleartextPasswords=true")
		} else {
			hasParam = true
			buf.WriteString("?allowCleartextPasswords=true")
		}
	}

	if !cfg.AllowNativePasswords {
		if hasParam {
			buf.WriteString("&allowNativePasswords=false")
		} else {
			hasParam = true
			buf.WriteString("?allowNativePasswords=false")
		}
	}

	if cfg.AllowOldPasswords {
		if hasParam {
			buf.WriteString("&allowOldPasswords=true")
		} else {
			hasParam = true
			buf.WriteString("?allowOldPasswords=true")
		}
	}

	if cfg.ClientFoundRows {
		if hasParam {
			buf.WriteString("&clientFoundRows=true")
		} else {
			hasParam = true
			buf.WriteString("?clientFoundRows=true")
		}
	}

	if col := cfg.Collation; len(col) > 0 {
		if hasParam {
			buf.WriteString("&collation=")
		} else {
			hasParam = true
			buf.WriteString("?collation=")
		}
		buf.WriteString(col)
	}

	if cfg.ColumnsWithAlias {
		if hasParam {
			buf.WriteString("&columnsWithAlias=true")
		} else {
			hasParam = true
			buf.WriteString("?columnsWithAlias=true")
		}
	}

	if cfg.InterpolateParams {
		if hasParam {
			buf.WriteString("&interpolateParams=true")
		} else {
			hasParam = true
			buf.WriteString("?interpolateParams=true")
		}
	}

	if cfg.Loc != time.UTC && cfg.Loc != nil {
		if hasParam {
			buf.WriteString("&loc=")
		} else {
			hasParam = true
			buf.WriteString("?loc=")
		}
		buf.WriteString(url.QueryEscape(cfg.Loc.String()))
	}

	if cfg.MultiStatements {
		if hasParam {
			buf.WriteString("&multiStatements=true")
		} else {
			hasParam = true
			buf.WriteString("?multiStatements=true")
		}
	}

	if cfg.ParseTime {
		if hasParam {
			buf.WriteString("&parseTime=true")
		} else {
			hasParam = true
			buf.WriteString("?parseTime=true")
		}
	}

	if cfg.ReadTimeout > 0 {
		if hasParam {
			buf.WriteString("&readTimeout=")
		} else {
			hasParam = true
			buf.WriteString("?readTimeout=")
		}
		buf.WriteString(cfg.ReadTimeout.String())
	}

	if cfg.RejectReadOnly {
		if hasParam {
			buf.WriteString("&rejectReadOnly=true")
		} else {
			hasParam = true
			buf.WriteString("?rejectReadOnly=true")
		}
	}

	if len(cfg.ServerPubKey) > 0 {
		if hasParam {
			buf.WriteString("&serverPubKey=")
		} else {
			hasParam = true
			buf.WriteString("?serverPubKey=")
		}
		buf.WriteString(url.QueryEscape(cfg.ServerPubKey))
	}

	if cfg.Timeout > 0 {
		if hasParam {
			buf.WriteString("&timeout=")
		} else {
			hasParam = true
			buf.WriteString("?timeout=")
		}
		buf.WriteString(cfg.Timeout.String())
	}

	if len(cfg.TLSConfig) > 0 {
		if hasParam {
			buf.WriteString("&sslmode=")
		} else {
			hasParam = true
			buf.WriteString("?sslmode=")
		}
		buf.WriteString(url.QueryEscape(cfg.TLSConfig))
	} else {
		if hasParam {
			buf.WriteString("&sslmode=disable")
		} else {
			hasParam = true
			buf.WriteString("?sslmode=disable")
		}
	}

	if cfg.WriteTimeout > 0 {
		if hasParam {
			buf.WriteString("&writeTimeout=")
		} else {
			hasParam = true
			buf.WriteString("?writeTimeout=")
		}
		buf.WriteString(cfg.WriteTimeout.String())
	}

	// other params
	if cfg.Params != nil {
		var params []string
		for param := range cfg.Params {
			params = append(params, param)
		}
		sort.Strings(params)
		for _, param := range params {
			if hasParam {
				buf.WriteByte('&')
			} else {
				hasParam = true
				buf.WriteByte('?')
			}

			buf.WriteString(param)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(cfg.Params[param]))
		}
	}

	return buf.String()
}
