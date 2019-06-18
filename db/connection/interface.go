package connection

import (
	"context"
	"database/sql"
	"time"
	"wsf/config"
	"wsf/db/transaction"
	"wsf/errors"
)

const (
	// TYPEDefault is a type id of connection class
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultConnection)
}

// Interface represents connection interface
type Interface interface {
	Context() context.Context
	SetContext(ctx context.Context) error
	Query(sql string, bind ...interface{}) (map[int]map[string]interface{}, error)
	BeginTransaction() (transaction.Interface, error)
	Close()
}

// NewConnection creates a new connection
func NewConnection(connectionType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[connectionType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database connection type \"%v\"", connectionType)
}

// NewConnectionFromConfig creates a new connection from connection.Config
func NewConnectionFromConfig(connectionType string, options *Config) (Interface, error) {
	if f, ok := buildHandlers[connectionType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database connection type \"%v\"", connectionType)
}

// Register registers a handler for database rowset creation
func Register(connectionType string, handler func(*Config) (Interface, error)) {
	buildHandlers[connectionType] = handler
}

// Connection represents database connection
type Connection struct {
	options      *Config
	conn         *sql.Conn
	ctx          context.Context
	pingTimeout  time.Duration
	queryTimeout time.Duration
}

// Context returns connection specific context
func (c *Connection) Context() context.Context {
	return c.ctx
}

// SetContext sets connection specific context
func (c *Connection) SetContext(ctx context.Context) error {
	c.ctx = ctx
	return nil
}

// Query runs the query
func (c *Connection) Query(sql string, bind ...interface{}) (map[int]map[string]interface{}, error) {
	stmt, err := c.conn.PrepareContext(c.ctx, sql)
	if err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}
	defer stmt.Close()

	ctx, cancel := context.WithTimeout(c.ctx, c.queryTimeout*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx, bind)
	if err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}
	defer rows.Close()

	result, err := c.processRows(rows)
	if err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}
	return result, nil
}

// BeginTransaction creates a new transaction
func (c *Connection) BeginTransaction() (transaction.Interface, error) {
	if c.conn == nil {
		return nil, errors.New("Database connection is not initialized")
	}

	tx, err := c.conn.BeginTx(c.ctx, &sql.TxOptions{Isolation: c.options.Transaction.IsolationLevel, ReadOnly: c.options.Transaction.ReadOnly})
	if err != nil {
		return nil, err
	}

	trns, err := transaction.NewTransaction(c.options.Transaction.Type, tx)
	if err != nil {
		return nil, err
	}

	trns.SetContext(c.ctx)
	return trns, nil
}

// Close closes the connection
func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) processRows(rows *sql.Rows) (map[int]map[string]interface{}, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	result := make(map[int]map[string]interface{})
	j := 0
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		value := make(map[string]interface{})
		for i, col := range values {
			if nullable, ok := columns[i].Nullable(); ok && nullable {
				value[columns[i].Name()] = nil
			} else {
				value[columns[i].Name()] = string(col)
			}
		}

		result[j] = value
		j++
	}

	return result, nil
}

// NewDefaultConnection creates default connection
func NewDefaultConnection(options *Config) (Interface, error) {
	return &Connection{
		options:      options,
		ctx:          context.Background(),
		pingTimeout:  time.Duration(options.PingTimeout) * time.Second,
		queryTimeout: time.Duration(options.QueryTimeout) * time.Second,
	}, nil
}
