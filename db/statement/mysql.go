package statement

import "wsf/db/dbselect"

const (
	// TYPEMySQL is a type id of statement class
	TYPEMySQL = "MySQL"
)

func init() {
	Register(TYPEMySQL, NewMySQLStatement)
}

// MySQL statement class
type MySQL struct {
	Statement
	keys   map[string]string
	values map[string]string
	meta   map[string]string
}

// Prepare prepares the statement
func (s *MySQL) Prepare(sql dbselect.Interface) error {
	return nil
}

// NewMySQLStatement creates a new mysql adapter statement
func NewMySQLStatement() (Interface, error) {
	stmt := &Statement{
		fetchMode:  0,
		attribute:  make(map[string]interface{}),
		bindColumn: make(map[string]interface{}),
		sqlSplit:   make([]string, 0),
		sqlParam:   make([]string, 0),
	}

	return stmt, nil
}
