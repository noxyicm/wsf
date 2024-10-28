package db

import (
	"database/sql"
	"sync"
	"wsf/errors"
)

const (
	// TYPEDefaultRowset is a type id of rowset class
	TYPEDefaultRowset = "default"
)

var (
	buildRowsetHandlers = map[string]func(*RowsetConfig) (Rowset, error){}
)

func init() {
	RegisterRowset(TYPEDefaultRowset, NewDefaultRowset)
}

// Rowset represents rows interface
type Rowset interface {
	Setup() error
	Push(row Row)
	Get() Row
	GetOffset(key int) Row
	OffsetExists(key int) bool
	Next() bool
	Populate(data []Row)
	PopulateMap(data []map[string]interface{}) error
	Prepare(rows []map[string]interface{}) error
	PrepareRaw(rows *sql.Rows) error
	SetRowType(typ string) error
	SetTable(table Table) error
	Table() Table
	Count() int
	IsEmpty() bool
}

// DefaultRowset holds and operates over rows
type DefaultRowset struct {
	Options           *RowsetConfig
	Data              []Row
	Tbl               Table
	Connected         bool
	Pointer           uint32
	Cnt               uint32
	CurrentKeyUnseted bool
	Pointing          bool
	Stored            bool
	ReadOnly          bool
	mu                sync.Mutex
}

// Setup the object
func (r *DefaultRowset) Setup() error {
	return nil
}

// Push row to rowset
func (r *DefaultRowset) Push(row Row) {
	r.Data = append(r.Data, row)
}

// Get returns row
func (r *DefaultRowset) Get() Row {
	return r.Data[r.Pointer]
}

// GetOffset returns row in a specific offset
func (r *DefaultRowset) GetOffset(key int) Row {
	if key < 0 && key >= int(r.Cnt) {
		return nil
	}

	return r.Data[key]
}

// OffsetExists returns true if key is in rows data
func (r *DefaultRowset) OffsetExists(key int) bool {
	return key >= 0 && key < len(r.Data)
}

// Next moves pointer further
func (r *DefaultRowset) Next() bool {
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

// Populate the object with provided data
func (r *DefaultRowset) Populate(data []Row) {
	r.Data = append(r.Data, data...)
	r.Cnt = uint32(len(r.Data))
}

// PopulateMap populates the object with provided mapstruct
func (r *DefaultRowset) PopulateMap(data []map[string]interface{}) error {
	for _, rowdata := range data {
		row, err := NewRow(r.Options.Row.Type, r.Options.Row)
		if err != nil {
			return errors.Wrap(err, "Database rowset populate Error")
		}

		row.Populate(rowdata)
		r.Data = append(r.Data, row)
	}

	r.Cnt = uint32(len(r.Data))
	return nil
}

// Prepare initializes rowset
func (r *DefaultRowset) Prepare(rows []map[string]interface{}) error {
	if r.Tbl == nil {
		return errors.New("Database rowset result Error: Reference table must be set")
	}

	for i := range rows {
		row, err := NewRow(r.Options.Row.Type, r.Options.Row)
		if err != nil {
			return errors.Wrap(err, "Database rowset result Error")
		}

		//if err := row.Prepare(values, columns); err != nil {
		//rowdata, err := r.Tbl.GetAdapter().PrepareRow(rows[i].Columns())
		//if err != nil {
		//	return err
		//}
		//row.Populate(rowdata)
		row.Populate(rows[i])

		r.Data = append(r.Data, row)
	}

	r.Cnt = uint32(len(r.Data))
	return nil
}

// PrepareRaw initializes rowset
func (r *DefaultRowset) PrepareRaw(rows *sql.Rows) error {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return errors.Wrap(err, "Database rowset result Error")
	}

	if r.Tbl == nil {
		return errors.New("Database rowset result Error: Reference table must be set")
	}

	for rows.Next() {
		rd := &RowData{}
		rd.CreateColumns(columns)

		if err = rows.Scan(rd.ColumnsScan()...); err != nil {
			return err
		}

		row, err := NewRow(r.Options.Row.Type, r.Options.Row)
		if err != nil {
			return errors.Wrap(err, "Database rowset result Error")
		}

		//if err := row.Prepare(values, columns); err != nil {
		//rowdata, err := r.Tbl.GetAdapter().PrepareRow(rd.Columns())
		//if err != nil {
		//	return err
		//}
		//row.Populate(rowdata)

		r.Data = append(r.Data, row)
	}

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "Database rowset result Error")
	}

	r.Cnt = uint32(len(r.Data))
	return nil
}

// SetRowType sets the this rowset row
func (r *DefaultRowset) SetRowType(typ string) error {
	r.Options.Row.Type = typ
	return nil
}

// SetTable sets the table object
func (r *DefaultRowset) SetTable(table Table) error {
	r.Tbl = table
	return nil
}

// Table returns table
func (r *DefaultRowset) Table() Table {
	return r.Tbl
}

// Count returns count of rows
func (r *DefaultRowset) Count() int {
	return int(r.Cnt)
}

// IsEmpty returns true if rowset is empty
func (r *DefaultRowset) IsEmpty() bool {
	return r.Cnt == 0
}

// NewDefaultRowset creates default rowset
func NewDefaultRowset(options *RowsetConfig) (Rowset, error) {
	return &DefaultRowset{
		Options:   options,
		Data:      make([]Row, 0),
		Connected: false,
		Pointer:   0,
		Cnt:       0,
		Pointing:  false,
	}, nil
}

// EmptyDefaultRowset creates default rowset
func EmptyDefaultRowset(options *RowsetConfig) *DefaultRowset {
	return &DefaultRowset{
		Options:   options,
		Data:      make([]Row, 0),
		Connected: false,
		Pointer:   0,
		Cnt:       0,
		Pointing:  false,
	}
}

// NewRowset creates a new rowset
func NewRowset(rowsetType string, options *RowsetConfig) (Rowset, error) {
	if f, ok := buildRowsetHandlers[rowsetType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database rowset type \"%v\"", rowsetType)
}

// NewEmptyRowset creates a new empty rowset
func NewEmptyRowset(rowsetType string) Rowset {
	options := &RowsetConfig{}
	options.Defaults()

	if f, ok := buildRowsetHandlers[rowsetType]; ok {
		if rowset, err := f(options); err == nil {
			return rowset
		}
	}

	rowset, _ := NewDefaultRowset(options)
	return rowset
}

// RegisterRowset registers a handler for database rowset creation
func RegisterRowset(rowsetType string, handler func(*RowsetConfig) (Rowset, error)) {
	buildRowsetHandlers[rowsetType] = handler
}
