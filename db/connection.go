package db

import (
	"context"
	"database/sql"
	"time"
	"wsf/config"
	"wsf/errors"
)

const (
	// TYPEDefaultConnection is a type id of connection class
	TYPEDefaultConnection = "default"
)

var (
	buildConnectionHandlers = map[string]func(*ConnectionConfig) (Connection, error){}
)

func init() {
	RegisterConnection(TYPEDefaultConnection, NewDefaultConnection)
}

// Connection represents connection interface
type Connection interface {
	Context() context.Context
	SetContext(ctx context.Context) error
	Query(sql string, bind ...interface{}) (map[int]map[string]interface{}, error)
	BeginTransaction() (Transaction, error)
	Close()
}

// NewConnection creates a new connection
func NewConnection(connectionType string, options config.Config) (Connection, error) {
	cfg := &ConnectionConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildConnectionHandlers[connectionType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database connection type \"%v\"", connectionType)
}

// NewConnectionFromConfig creates a new connection from connection.Config
func NewConnectionFromConfig(connectionType string, options *ConnectionConfig) (Connection, error) {
	if f, ok := buildConnectionHandlers[connectionType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database connection type \"%v\"", connectionType)
}

// RegisterConnection registers a handler for database rowset creation
func RegisterConnection(connectionType string, handler func(*ConnectionConfig) (Connection, error)) {
	buildConnectionHandlers[connectionType] = handler
}

// DefaultConnection represents database connection
type DefaultConnection struct {
	Options      *ConnectionConfig
	Conn         *sql.Conn
	Ctx          context.Context
	PingTimeout  time.Duration
	QueryTimeout time.Duration
}

// Context returns connection specific context
func (c *DefaultConnection) Context() context.Context {
	return c.Ctx
}

// SetContext sets connection specific context
func (c *DefaultConnection) SetContext(ctx context.Context) error {
	c.Ctx = ctx
	return nil
}

// Query runs the query
func (c *DefaultConnection) Query(sql string, bind ...interface{}) (map[int]map[string]interface{}, error) {
	stmt, err := c.Conn.PrepareContext(c.Ctx, sql)
	if err != nil {
		return nil, errors.Wrap(err, "Database connection Error")
	}
	defer stmt.Close()

	ctx, cancel := context.WithTimeout(c.Ctx, c.QueryTimeout*time.Second)
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
func (c *DefaultConnection) BeginTransaction() (Transaction, error) {
	if c.Conn == nil {
		return nil, errors.New("Database connection is not initialized")
	}

	tx, err := c.Conn.BeginTx(c.Ctx, &sql.TxOptions{Isolation: c.Options.Transaction.IsolationLevel, ReadOnly: c.Options.Transaction.ReadOnly})
	if err != nil {
		return nil, err
	}

	trns, err := NewTransaction(c.Options.Transaction.Type, tx)
	if err != nil {
		return nil, err
	}

	trns.SetContext(c.Ctx)
	return trns, nil
}

// Close closes the connection
func (c *DefaultConnection) Close() {
	c.Conn.Close()
}

func (c *DefaultConnection) processRows(rows *sql.Rows) (map[int]map[string]interface{}, error) {
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
func NewDefaultConnection(options *ConnectionConfig) (Connection, error) {
	return &DefaultConnection{
		Options:      options,
		Ctx:          context.Background(),
		PingTimeout:  time.Duration(options.PingTimeout) * time.Second,
		QueryTimeout: time.Duration(options.QueryTimeout) * time.Second,
	}, nil
}
