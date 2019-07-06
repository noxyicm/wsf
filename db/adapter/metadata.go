package adapter

// ColumnMetadata represents information schema column
type ColumnMetadata struct {
	TableSchema     string
	TableName       string
	Name            string
	Default         interface{}
	Position        int64
	Type            string
	Length          int64
	Precision       int64
	Scale           int64
	CharacterSet    string
	Collation       string
	ColumnType      string
	ColumnKey       string
	Extra           string
	Unsigned        bool
	IsNullable      bool
	Primary         bool
	PrimaryPosition int64
	Identity        bool
}
