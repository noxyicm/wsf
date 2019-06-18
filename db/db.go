package db

import (
	"context"
	"wsf/config"
	"wsf/db/adapter"
	"wsf/db/connection"
	"wsf/db/dbselect"
	"wsf/db/table/rowset"
	"wsf/db/transaction"
)

var ins *Db

// Db type resource
type Db struct {
	options        *Config
	adapter        adapter.Interface
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
func (d *Db) Connection() (connection.Interface, error) {
	return d.adapter.Connection()
}

// Select returns a new select object
func (d *Db) Select() (dbselect.Interface, error) {
	return d.adapter.Select()
}

// Insert inserts new row into table
func (d *Db) Insert(table string, data map[string]interface{}) (int, error) {
	return d.adapter.Insert(table, data)
}

// Update inserts new row into table
func (d *Db) Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	return d.adapter.Update(table, data, cond)
}

// Query runs a query
func (d *Db) Query(sql dbselect.Interface) (rowset.Interface, error) {
	return d.adapter.Query(sql)
}

// BeginTransaction creates a new database transaction
func (d *Db) BeginTransaction() (transaction.Interface, error) {
	return d.adapter.BeginTransaction()
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

// NewDB creates new Db handler
func NewDB(options config.Config) (db *Db, err error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	var a adapter.Interface
	acfg := options.Get("adapter")
	if acfg != nil {
		adapterType := acfg.GetString("type")
		a, err = adapter.NewAdapter(adapterType, acfg)
		if err != nil {
			return nil, err
		}
	}

	err = a.Init(context.Background())
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

// Select returns a new select object
func Select() (dbselect.Interface, error) {
	return ins.Select()
}

// Insert inserts new row into table
func Insert(table string, data map[string]interface{}) (int, error) {
	return ins.Insert(table, data)
}

// Update inserts new row into table
func Update(table string, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	return ins.Update(table, data, cond)
}

// Query runs a query
func Query(sql dbselect.Interface) (rowset.Interface, error) {
	return ins.Query(sql)
}
