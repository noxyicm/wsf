package db

import (
	"database/sql"
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

// Adapter returns a database adapter
func (d *Db) Adapter() Adapter {
	return d.adapter
}

// Select returns a new select object
func (d *Db) Select() Select {
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
func (d *Db) Query(ctx context.Context, sql Select) ([]map[string]interface{}, error) { //(Rowset, error) {
	return d.adapter.Query(ctx, sql)
}

// QueryInto as
func (d *Db) QueryInto(ctx context.Context, dbs Select, o interface{}) ([]interface{}, error) {
	return d.adapter.QueryInto(ctx, dbs, o)
}

// QueryRow runs a query
func (d *Db) QueryRow(ctx context.Context, sql Select) (map[string]interface{}, error) { //(Row, error) {
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
func CreateSelect() Select {
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
func Query(ctx context.Context, sql Select) ([]map[string]interface{}, error) { //(Rowset, error) {
	return ins.Query(ctx, sql)
}

// QueryInto as
func QueryInto(ctx context.Context, dbs Select, o interface{}) ([]interface{}, error) {
	return ins.QueryInto(ctx, dbs, o)
}

// QueryRow runs a query
func QueryRow(ctx context.Context, sql Select) (map[string]interface{}, error) { //(Row, error) {
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

// ColumnData represents a raw byte data from database
type ColumnData struct {
	Column *sql.ColumnType
	Value  []byte
}

// Scan implementation of Scanner interface
func (c *ColumnData) Scan(src interface{}) error {
	if n, ok := c.Column.Nullable(); ok && n {
		if src == nil {
			c.Value = nil
		} else {
			c.Value = src.([]byte)
		}
	} else {
		c.Value = make([]byte, 0)
	}

	return nil
}

// RowData represents a database row
type RowData struct {
	data []*ColumnData
}

// CreateColumns fills data slice with columns
func (rd *RowData) CreateColumns(columns []*sql.ColumnType) {
	rd.data = make([]*ColumnData, len(columns))
	for i := range columns {
		rd.data[i] = &ColumnData{
			Column: columns[i],
		}
	}
}

// Columns returns all row columns
func (rd *RowData) Columns() []*ColumnData {
	return rd.data
}

// ColumnsScan returns all row columns
func (rd *RowData) ColumnsScan() []interface{} {
	srcs := make([]interface{}, len(rd.data))
	for i := range rd.data {
		srcs[i] = rd.data[i]
	}
	return srcs
}

// ToMap returns a row data as mapstruct
func (rd *RowData) ToMap() map[string][]byte {
	d := make(map[string][]byte)
	for i := range rd.data {
		d[rd.data[i].Column.Name()] = rd.data[i].Value
	}

	return d
}

// Unmarshal unmarshals data into struct
func (rd *RowData) Unmarshal(adp Adapter, output interface{}) error {
	//row, err := adp.PrepareRow(rd.Columns())
	//if err != nil {
	//	return errors.Wrap(err, "RowData unmarshal error")
	//}

	//if err := mapstructure.Decode(row, output); err != nil {
	//	return errors.Wrap(err, "RowData unmarshal error")
	//}

	return nil
}
