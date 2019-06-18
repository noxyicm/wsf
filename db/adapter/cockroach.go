package adapter

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"database/sql"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"wsf/db/dbselect"
	"wsf/db/statement"
	"wsf/errors"

	_ "github.com/lib/pq"
)

const (
	// TYPECockroach represents cockroach db adapter
	TYPECockroach = "CockroachDB"
)

func init() {
	Register(TYPECockroach, NewCockroachAdapter)
}

// Cockroach adapter for CockroachDB databeses
type Cockroach struct {
	adapter
	driverConfig *cockroachConfig
}

// Init a connection to database
func (a *Cockroach) Init(ctx context.Context) (err error) {
	db, err := sql.Open("postgres", a.driverConfig.FormatDSN())
	if err != nil {
		return errors.Wrap(err, "CockroachDB Error")
	}

	db.SetConnMaxLifetime(time.Duration(a.options.ConnectionMaxLifeTime) * time.Second)
	db.SetMaxIdleConns(a.options.MaxIdleConnections)
	db.SetMaxOpenConns(a.options.MaxOpenConnections)

	if a.pingTimeout > 0 {
		tctx, cancel := context.WithTimeout(ctx, a.pingTimeout*time.Second)
		defer cancel()

		if err = db.PingContext(tctx); err != nil {
			return errors.Wrap(err, "CockroachDB Error")
		}
	} else {
		if err = db.PingContext(ctx); err != nil {
			return errors.Wrap(err, "CockroachDB Error")
		}
	}

	a.db = db
	a.ctx = ctx
	return nil
}

// SetOptions sets new options for CockroachDB adapter
func (a *Cockroach) SetOptions(options *Config) Interface {
	a.options = options
	return a
}

// Options returns CockroachDB adapter options
func (a *Cockroach) Options() *Config {
	return a.options
}

// Insert inserts new row into table
func (a *Cockroach) Insert(table string, data map[string]interface{}) (int, error) {
	cols := []string{}
	vals := []string{}
	binds := []interface{}{}
	i := 1
	for col, val := range data {
		cols = append(cols, a.QuoteIdentifier(col, true))

		switch val.(type) {
		case *dbselect.Expr:
			vals = append(vals, val.(*dbselect.Expr).ToString())

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

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.pingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.db.PrepareContext(ctx, sql)
	if err != nil {
		return 0, errors.Wrap(err, "CockroachDB insert Error")
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
	defer cancel()

	err = stmt.QueryRowContext(ctx, binds...).Scan(&a.lastInsertID)
	if err != nil {
		return 0, errors.Wrap(err, "CockroachDB insert Error")
	}

	return a.lastInsertID, nil
}

// Update updates rows into table be condition
func (a *Cockroach) Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
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

		set = append(set, a.QuoteIdentifier(col, true)+" = "+value)
	}

	where := a.whereExpr(cond)

	sql := "UPDATE " + a.QuoteIdentifier(table, true) + " SET (" + strings.Join(set, ", ") + ") RETURNING \"id\""
	if where != "" {
		sql = sql + " WHERE " + where
	}

	ctx, cancel := context.WithTimeout(a.ctx, time.Duration(a.pingTimeout)*time.Second)
	defer cancel()

	stmt, err := a.db.PrepareContext(ctx, sql)
	if err != nil {
		return false, errors.Wrap(err, "CockroachDB update Error")
	}
	defer stmt.Close()

	ctx, cancel = context.WithTimeout(a.ctx, time.Duration(a.queryTimeout)*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx, binds...)
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

// NewCockroachAdapter creates a new CockroachDB adapter
func NewCockroachAdapter(options *Config) (ai Interface, err error) {
	adp := &Cockroach{}
	adp.identifierSymbol = `"`
	adp.autoQuoteIdentifiers = true
	adp.defaultStatementType = statement.TYPEDefault
	adp.pingTimeout = time.Duration(options.PingTimeout) * time.Second
	adp.queryTimeout = time.Duration(options.QueryTimeout) * time.Second

	//sql.Register(name string, driver driver.Driver)
	adp.driverConfig = &cockroachConfig{AllowNativePasswords: true}
	adp.driverConfig.User = options.Username
	adp.driverConfig.Passwd = options.Password
	adp.driverConfig.Net = options.Protocol
	adp.driverConfig.Addr = options.Host
	if options.Port > 0 {
		adp.driverConfig.Addr = adp.driverConfig.Addr + ":" + strconv.Itoa(options.Port)
	}
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
		"::",
		"SOME",
		"ANY",
		"ALL",
		"SIMILAR",
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
		" ",
	}

	adp.UnquoteableFunctions = []string{
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
		"IF",
		"IFNULL",
		"NULLIF",
	}

	adp.options = options
	adp.params = map[string]interface{}{
		"positional": true,
		"named":      false,
	}
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
