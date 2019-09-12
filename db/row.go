package db

import (
	"context"
	"database/sql"
	"reflect"
	"strconv"
	"time"
	"wsf/errors"
	"wsf/registry"

	"github.com/mitchellh/mapstructure"
)

const (
	// TYPEDefaultRow is a type id of rowset class
	TYPEDefaultRow = "default"
)

var (
	buildRowHandlers = map[string]func(*RowConfig) (Row, error){}

	rowCfgContextKey contextKey
)

func init() {
	RegisterRow(TYPEDefaultRow, NewDefaultRow)
}

// Row represents row interface
type Row interface {
	Setup() error
	Set(key string, value interface{})
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetFloat(key string) float64
	GetBool(key string) bool
	GetTime(key string) time.Time
	Unmarshal(output interface{}) error
	Populate(data map[string]interface{})
	Prepare(row []sql.RawBytes, columns []*sql.ColumnType) error
	SetTable(table Table) error
	Table() Table
	IsEmpty() bool
}

// DefaultRow holds data and operates over row
type DefaultRow struct {
	Row
	Options   *RowConfig
	Data      map[string]interface{}
	Tbl       Table
	Connected bool
	Stored    bool
	ReadOnly  bool
}

// Setup the object
func (r *DefaultRow) Setup() error {
	return nil
}

// Set value v to row data with key k
func (r *DefaultRow) Set(k string, v interface{}) {
	r.Data[k] = v
}

// Get returns a value by its key
func (r *DefaultRow) Get(key string) interface{} {
	if v, ok := r.Data[key]; ok {
		return v
	}

	return nil
}

// GetString returns a value by its key as string
func (r *DefaultRow) GetString(key string) string {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// GetInt returns a value by its key as int
func (r *DefaultRow) GetInt(key string) int {
	if v, ok := r.Data[key]; ok {
		switch v.(type) {
		case int:
			return v.(int)

		case int8:
			return int(v.(int8))

		case int16:
			return int(v.(int16))

		case int32:
			return int(v.(int32))

		case int64:
			return int(v.(int64))
		}
	}

	return 0
}

// GetFloat returns a value by its key as float64
func (r *DefaultRow) GetFloat(key string) float64 {
	if v, ok := r.Data[key]; ok {
		switch v.(type) {
		case float64:
			return v.(float64)

		case float32:
			return float64(v.(float32))
		}
	}

	return 0.0
}

// GetBool returns a value by its key as bool
func (r *DefaultRow) GetBool(key string) bool {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(bool); ok {
			return v
		}
	}

	return false
}

// GetTime returns a value by its key as time
func (r *DefaultRow) GetTime(key string) time.Time {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(time.Time); ok {
			return v
		}
	}

	return time.Time{}
}

// GetAll returns all row columns as map
func (r *DefaultRow) GetAll() map[string]interface{} {
	return r.Data
}

// Unmarshal unmarshals data into struct
func (r *DefaultRow) Unmarshal(output interface{}) error {
	if err := mapstructure.Decode(r.Data, output); err != nil {
		return err
	}

	return nil
}

// Populate the row object with provided data
func (r *DefaultRow) Populate(data map[string]interface{}) {
	for key, value := range data {
		r.Data[key] = value
	}
}

// Prepare initializes row
func (r *DefaultRow) Prepare(row []sql.RawBytes, columns []*sql.ColumnType) (err error) {
	if r.Data, err = PrepareRow(row, columns); err != nil {
		return err
	}

	return nil
}

// SetTable sets the table object
func (r *DefaultRow) SetTable(table Table) error {
	r.Tbl = table
	return nil
}

// Table return table
func (r *DefaultRow) Table() Table {
	return r.Tbl
}

// IsEmpty returns true if object has no data
func (r *DefaultRow) IsEmpty() bool {
	return len(r.Data) == 0
}

// NewDefaultRow creates default row
func NewDefaultRow(options *RowConfig) (Row, error) {
	r := &DefaultRow{
		Options:   options,
		Data:      make(map[string]interface{}),
		Connected: false,
	}

	if options.Table != "" {
		dbResource := registry.Get("db")
		if dbResource != nil {
			db := dbResource.(*Db)
			tbl := db.GetOrCreateTable(options.Table)
			if tbl != nil {
				r.Tbl = tbl
			}
		}
	}

	r.Setup()
	return r, nil
}

// NewRow creates a new row
func NewRow(rowType string, options *RowConfig) (Row, error) {
	if f, ok := buildRowHandlers[rowType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database row type \"%v\"", rowType)
}

// NewEmptyRow creates a new empty row
func NewEmptyRow(rowType string) Row {
	options := &RowConfig{}
	options.Defaults()

	if f, ok := buildRowHandlers[rowType]; ok {
		if row, err := f(options); err == nil {
			return row
		}
	}

	row, _ := NewDefaultRow(options)
	return row
}

// RegisterRow registers a handler for database row creation
func RegisterRow(rowType string, handler func(*RowConfig) (Row, error)) {
	buildRowHandlers[rowType] = handler
}

// RowTypeRegistered returns true if type rowType registered
func RowTypeRegistered(rowType string) bool {
	if _, ok := buildRowHandlers[rowType]; ok {
		return true
	}

	return false
}

// PrepareRow parses a RawBytes into map structure
func PrepareRow(row []sql.RawBytes, columns []*sql.ColumnType) (data map[string]interface{}, err error) {
	err = nil
	data = make(map[string]interface{})
	for i, col := range row {
		if columns[i].ScanType() == reflect.TypeOf(sql.NullBool{}) {
			v := sql.NullBool{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Bool
				} else {
					data[columns[i].Name()] = false
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullInt64{}) {
			v := sql.NullInt64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Int64
				} else {
					data[columns[i].Name()] = 0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullFloat64{}) {
			v := sql.NullFloat64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.Float64
				} else {
					data[columns[i].Name()] = 0.0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullString{}) {
			v := sql.NullString{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					data[columns[i].Name()] = v.String
				} else {
					data[columns[i].Name()] = ""
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.RawBytes{}) {
			data[columns[i].Name()] = string(col)
		} else if columns[i].ScanType() == reflect.TypeOf((int)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = 0
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 0)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int8)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int8(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 8)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int16)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int16(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 16)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int32)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int32(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int64)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = int64(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float32)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = float32(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseFloat(string(col), 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float64)(0)) {
			if len(col) == 0 {
				data[columns[i].Name()] = float64(0)
			} else {
				data[columns[i].Name()], err = strconv.ParseFloat(string(col), 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf(time.Time{}) {
			if string(col) == "" {
				data[columns[i].Name()] = time.Time{}
			} else {
				var t time.Time
				t, err = time.Parse("2006-01-02T15:04:05Z", string(col))
				data[columns[i].Name()] = t.Local()
			}
		} else if columns[i].ScanType() == reflect.TypeOf(true) {
			data[columns[i].Name()], err = strconv.ParseBool(string(col))
		} else {
			data[columns[i].Name()] = string(col)
		}

		if err != nil {
			return data, err
		}
	}

	return data, err
}

// RowConfigToContext returns a new context with stored row config
func RowConfigToContext(ctx context.Context, cfg *RowConfig) context.Context {
	return context.WithValue(ctx, rowCfgContextKey, cfg)
}

// RowConfigFromContext returns a row config stored in context
func RowConfigFromContext(ctx context.Context) (*RowConfig, bool) {
	v, ok := ctx.Value(rowCfgContextKey).(*RowConfig)
	return v, ok
}
