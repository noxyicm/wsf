package row

import (
	"database/sql"
	"reflect"
	"strconv"
	"time"
	"wsf/errors"

	"github.com/mitchellh/mapstructure"
)

const (
	// TYPEDefault is a type id of rowset class
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultRow)
}

// Interface represents row interface
type Interface interface {
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetFloat(key string) float64
	GetBool(key string) bool
	GetTime(key string) time.Time
	Unmarshal(output interface{}) error
	Prepare(row []sql.RawBytes, columns []*sql.ColumnType) error
	SetTable(table string) error
	Table() string
}

// Row holds and operates over row
type Row struct {
	Options   *Config
	Data      map[string]interface{}
	Tbl       string
	Connected bool
	Stored    bool
	ReadOnly  bool
}

// Get returns a value by its key
func (r *Row) Get(key string) interface{} {
	if v, ok := r.Data[key]; ok {
		return v
	}

	return nil
}

// GetString returns a value by its key as string
func (r *Row) GetString(key string) string {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// GetInt returns a value by its key as int
func (r *Row) GetInt(key string) int {
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
func (r *Row) GetFloat(key string) float64 {
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
func (r *Row) GetBool(key string) bool {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(bool); ok {
			return v
		}
	}

	return false
}

// GetTime returns a value by its key as time
func (r *Row) GetTime(key string) time.Time {
	if v, ok := r.Data[key]; ok {
		if v, ok := v.(time.Time); ok {
			return v
		}
	}

	return time.Time{}
}

// Unmarshal unmarshals data into struct
func (r *Row) Unmarshal(output interface{}) error {
	if err := mapstructure.Decode(r.Data, output); err != nil {
		return err
	}

	return nil
}

// Prepare initializes row
func (r *Row) Prepare(row []sql.RawBytes, columns []*sql.ColumnType) (err error) {
	err = nil
	for i, col := range row {
		if columns[i].ScanType() == reflect.TypeOf(sql.NullBool{}) {
			v := sql.NullBool{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					r.Data[columns[i].Name()] = v.Bool
				} else {
					r.Data[columns[i].Name()] = false
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullInt64{}) {
			v := sql.NullInt64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					r.Data[columns[i].Name()] = v.Int64
				} else {
					r.Data[columns[i].Name()] = 0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullFloat64{}) {
			v := sql.NullFloat64{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					r.Data[columns[i].Name()] = v.Float64
				} else {
					r.Data[columns[i].Name()] = 0.0
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.NullString{}) {
			v := sql.NullString{}
			v.Scan(string(col))
			if v.Valid {
				if v.Valid {
					r.Data[columns[i].Name()] = v.String
				} else {
					r.Data[columns[i].Name()] = ""
				}
			}
		} else if columns[i].ScanType() == reflect.TypeOf(sql.RawBytes{}) {
			r.Data[columns[i].Name()] = string(col)
		} else if columns[i].ScanType() == reflect.TypeOf((int)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 0)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int8)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 8)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int16)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 16)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int32)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((int64)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseInt(string(col), 10, 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float32)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseFloat(string(col), 32)
			}
		} else if columns[i].ScanType() == reflect.TypeOf((float64)(0)) {
			if len(col) == 0 {
				r.Data[columns[i].Name()] = nil
			} else {
				r.Data[columns[i].Name()], err = strconv.ParseFloat(string(col), 64)
			}
		} else if columns[i].ScanType() == reflect.TypeOf(time.Time{}) {
			if string(col) == "" {
				r.Data[columns[i].Name()] = time.Time{}
			} else {
				var t time.Time
				t, err = time.Parse("2006-01-02T15:04:05Z", string(col))
				r.Data[columns[i].Name()] = t.Local()
			}
		} else if columns[i].ScanType() == reflect.TypeOf(true) {
			r.Data[columns[i].Name()], err = strconv.ParseBool(string(col))
		} else {
			r.Data[columns[i].Name()] = string(col)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// SetTable sets the table object
func (r *Row) SetTable(table string) error {
	r.Tbl = table
	return nil
}

// Table return table
func (r *Row) Table() string {
	return r.Tbl
}

// NewDefaultRow creates default row
func NewDefaultRow(options *Config) (Interface, error) {
	return &Row{
		Options:   options,
		Data:      make(map[string]interface{}),
		Connected: false,
	}, nil
}

// NewRow creates a new row
func NewRow(rowType string, options *Config) (Interface, error) {
	if f, ok := buildHandlers[rowType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized database row type \"%v\"", rowType)
}

// Register registers a handler for database row creation
func Register(rowType string, handler func(*Config) (Interface, error)) {
	buildHandlers[rowType] = handler
}
