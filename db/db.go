package db

import (
	"wsf/config"
	"wsf/context"
	"wsf/errors"
	"wsf/registry"
)

type contextKey int

var (
	ins *Db

	defaultAdapter Adapter
)

// Db type resource
type Db struct {
	options        *Config
	adapter        Adapter
	db             string
	defaultAdapter string
}

// SetOptions sets new options for resource and reinits it
func (d *Db) SetOptions(options *Config) error {
	d.options = options
	return nil
}

// Options returns resource options
func (d *Db) Options() *Config {
	return d.options
}

// Priority returns resource initialization priority
func (d *Db) Priority() int {
	return d.options.Priority
}

// Context returns a database adapter specific context
func (d *Db) Context() context.Context {
	return d.adapter.Context()
}

// Connection returns a connection to database
func (d *Db) Connection(ctx context.Context) (Connection, error) {
	return d.adapter.Connection(ctx)
}

// Select returns a new select object
func (d *Db) Select() (Select, error) {
	return d.adapter.Select()
}

// Insert inserts new row into table
func (d *Db) Insert(ctx context.Context, table string, data map[string]interface{}) (int, error) {
	return d.adapter.Insert(ctx, table, data)
}

// Update inserts new row into table
func (d *Db) Update(ctx context.Context, table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	return d.adapter.Update(ctx, table, data, cond)
}

// Delete removes row from table
func (d *Db) Delete(ctx context.Context, table string, cond map[string]interface{}) (bool, error) {
	return d.adapter.Delete(ctx, table, cond)
}

// Query runs a query
func (d *Db) Query(ctx context.Context, sql Select) (Rowset, error) {
	return d.adapter.Query(ctx, sql)
}

// QueryRow runs a query
func (d *Db) QueryRow(ctx context.Context, sql Select) (Row, error) {
	return d.adapter.QueryRow(ctx, sql)
}

// BeginTransaction creates a new database transaction
func (d *Db) BeginTransaction(ctx context.Context) (Transaction, error) {
	return d.adapter.BeginTransaction(ctx)
}

// Quote a string
func (d *Db) Quote(value interface{}) string {
	return d.adapter.Quote(value)
}

// QuoteIdentifier quotes a specific identifier
func (d *Db) QuoteIdentifier(ident interface{}, auto bool) string {
	return d.adapter.QuoteIdentifier(ident, auto)
}

// QuoteIdentifierAs a
func (d *Db) QuoteIdentifierAs(ident interface{}, alias string, auto bool) string {
	return d.adapter.QuoteIdentifierAs(ident, alias, auto)
}

// QuoteIdentifierSymbol returns symbol of identifier quote
func (d *Db) QuoteIdentifierSymbol() string {
	return d.adapter.QuoteIdentifierSymbol()
}

// SupportsParameters returns true if adapter supports
func (d *Db) SupportsParameters(param string) bool {
	return d.adapter.SupportsParameters(param)
}

// GetOrCreateTable gets from registry or creates a new table object
func (d *Db) GetOrCreateTable(table string) Table {
	return nil
}

// NewDB creates new Db handler
func NewDB(options config.Config) (db *Db, err error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	var a Adapter
	acfg := options.Get("adapter")
	if acfg != nil {
		adapterType := acfg.GetString("type")
		a, err = NewAdapter(adapterType, acfg)
		if err != nil {
			return nil, err
		}
	}

	err = a.Init()
	if err != nil {
		return nil, err
	}

	db = &Db{
		options: cfg,
		adapter: a,
	}

	return db, nil
}

// SetInstance sets a main db instance
func SetInstance(d *Db) {
	ins = d
}

// Instance returns a db controller instance
func Instance() *Db {
	return ins
}

// CreateSelect returns a select configured by db instance
func CreateSelect() (Select, error) {
	return ins.Select()
}

// Insert inserts new row into table
func Insert(ctx context.Context, table string, data map[string]interface{}) (int, error) {
	return ins.Insert(ctx, table, data)
}

// Update inserts new row into table
func Update(ctx context.Context, table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	return ins.Update(ctx, table, data, cond)
}

// Delete removes a row from table
func Delete(ctx context.Context, table string, cond map[string]interface{}) (bool, error) {
	return ins.Delete(ctx, table, cond)
}

// Query runs a query
func Query(ctx context.Context, sql Select) (Rowset, error) {
	return ins.Query(ctx, sql)
}

// QueryRow runs a query
func QueryRow(ctx context.Context, sql Select) (Row, error) {
	return ins.QueryRow(ctx, sql)
}

// SetDefaultAdapter sets the default db.Adapter
func SetDefaultAdapter(db interface{}) {
	defaultAdapter, _ = SetupAdapter(db)
}

// GetDefaultAdapter gets the default db.Adapter
func GetDefaultAdapter() Adapter {
	return defaultAdapter
}

// SetupAdapter checks if db is a valid database adapter
func SetupAdapter(db interface{}) (Adapter, error) {
	if db == nil {
		return nil, errors.New("Argument must be of type db.adapter.Interface, or a Registry key where a db.adapter.Interface object is stored")
	}

	switch db.(type) {
	case string:
		if dba := registry.Get(db.(string)); dba != nil {
			return dba.(Adapter), nil
		}

	case Adapter:
		return db.(Adapter), nil
	}

	return nil, errors.New("Argument must be of type db.adapter.Interface, or a Registry key where a db.adapter.Interface object is stored")
}

// Options return db options
func Options() *Config {
	return ins.Options()
}
