package rowset

import (
	"database/sql"
	"sync"
	"wsf/db/table/row"
	"wsf/errors"
)

const (
	// TYPEDefault is a type id of rowset class
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultRowset)
}

// Interface represents rows interface
type Interface interface {
	Get() row.Interface
	GetOffset(key int) row.Interface
	Next() bool
	Prepare(rows *sql.Rows) error
	SetRowType(typ string) error
	SetTable(table string) error
	Table() string
	Count() int
	IsEmpty() bool
}

// Rowset holds and operates over rows
type Rowset struct {
	Options   *Config
	Data      []row.Interface
	Tbl       string
	Connected bool
	Pointer   uint32
	Cnt       uint32
	Pointing  bool
	Stored    bool
	ReadOnly  bool
	mu        sync.Mutex
}

// Get returns row
func (r *Rowset) Get() row.Interface {
	return r.Data[r.Pointer]
}

// GetOffset returns row in a specific offset
func (r *Rowset) GetOffset(key int) row.Interface {
	if key < 0 && key >= int(r.Cnt) {
		return nil
	}

	return r.Data[key]
}

// Next moves pointer further
func (r *Rowset) Next() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Cnt == 0 {
		return false
	}

	if r.Pointing {
		r.Pointer++
	} else {
		r.Pointing = true
	}

	if r.Pointer >= r.Cnt {
		if r.Pointer-1 < 0 {
			r.Pointer = 0
		} else {
			r.Pointer--
		}

		return false
	}

	return true
}

// Prepare initializes rowset
func (r *Rowset) Prepare(rows *sql.Rows) error {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return errors.Wrap(err, "Database rowset result Error")
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return err
		}

		row, err := row.NewRow(r.Options.Row.Type, r.Options.Row)
		if err != nil {
			return errors.Wrap(err, "Database rowset result Error")
		}

		if err := row.Prepare(values, columns); err != nil {
			return err
		}

		r.Data = append(r.Data, row)
	}

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "Database rowset result Error")
	}

	r.Cnt = uint32(len(r.Data))
	return nil
}

// SetRowType sets the this rowset row
func (r *Rowset) SetRowType(typ string) error {
	r.Options.Row.Type = typ
	return nil
}

// SetTable sets the table object
func (r *Rowset) SetTable(table string) error {
	r.Tbl = table
	return nil
}

// Table returns table
func (r *Rowset) Table() string {
	return r.Tbl
}

// Count returns count of rows
func (r *Rowset) Count() int {
	return int(r.Cnt)
}

// IsEmpty returns true if rowset is empty
func (r *Rowset) IsEmpty() bool {
	return r.Cnt == 0
}

// NewDefaultRowset creates default rowset
func NewDefaultRowset(options *Config) (Interface, error) {
	return &Rowset{
		Options:   options,
		Data:      make([]row.Interface, 0),
		Connected: false,
		Pointer:   0,
		Cnt:       0,
		Pointing:  false,
	}, nil
}

// NewRowset creates a new rowset
func NewRowset(rowsetType string, options *Config) (Interface, error) {
	if f, ok := buildHandlers[rowsetType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database rowset type \"%v\"", rowsetType)
}

// Register registers a handler for database rowset creation
func Register(rowsetType string, handler func(*Config) (Interface, error)) {
	buildHandlers[rowsetType] = handler
}
