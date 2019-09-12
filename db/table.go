package db

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"wsf/cache"
	"wsf/config"
	"wsf/errors"
	"wsf/registry"
	"wsf/utils"
)

// Public constants
const (
	DataAdapter     = "db"
	Definition      = "definition"
	DefinitionName  = "definitionName"
	Schema          = "schema"
	Name            = "name"
	Primary         = "primary"
	Cols            = "cols"
	MetaData        = "metadata"
	MetaDataCache   = "metadataCache"
	RowType         = "rowType"
	RowsetType      = "rowsetType"
	ReferenceMap    = "referenceMap"
	DependentTables = "dependentTables"
	Sequence        = "sequence"

	RefTable   = "refTable"
	RefColumns = "refColumns"
	OnDelete   = "onDelete"
	OnUpdate   = "onUpdate"

	Cascade        = "cascade"
	CascadeRecurse = "cascadeRecurse"
	Restrict       = "restrict"
	SetNull        = "setNull"

	DefaultNone   = "defaultNone"
	DefaultObject = "defaultObject"
	DefaultDb     = "defaultDb"

	SelectWithFormPart    = true
	SelectWithoutFormPart = false
)

var (
	buildTableHandlers = map[string]func(*TableConfig) (Table, error){}

	defaultMetadataCache cache.Interface
)

// Table is an interface for db table
type Table interface {
	SetOptions(options *TableConfig) Table
	SetRowType(typ string) Table
	GetRowType() string
	SetRowsetType(typ string) Table
	GetRowsetType() string
	SetAdapter(db interface{}) Table
	GetAdapter() Adapter
	SetMetadataCache(metadataCache interface{}) Table
	GetMetadataCache() cache.Interface
	SetMetadataCacheInStruct(flag bool) Table
	IsMetadataCacheInClass() bool
	Setup() error
	SetupAdapter() error
	SetupTableName()
	SetupMetadata() (bool, error)
	SetupPrimaryKey() error
	GetCols() []string
	Init() error
	Info() TableInfo
	Select(withFromPart bool) (Select, error)
	Insert(ctx context.Context, data map[string]interface{}) (int, error)
	IsIdentity(column string) bool
	Update(ctx context.Context, data map[string]interface{}, cond map[string]interface{}) (bool, error)
	//CascadeUpdate(parentTable string, oldPrimaryKey map[string]string, newPrimaryKey map[string]string) (int, error)
	Delete(ctx context.Context, cond map[string]interface{}) (bool, error)
	//CascadeDelete(parentTable string, primaryKey map[string]interface{}) (int, error)
	Find(ctx context.Context, args ...interface{}) (Rowset, error)
}

// NewTable creates a new table from given type and options
func NewTable(tableType string, options config.Config) (ti Table, err error) {
	cfg := &TableConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildTableHandlers[tableType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database table type \"%v\"", tableType)
}

// RegisterTable registers a handler for database table creation
func RegisterTable(tableType string, handler func(*TableConfig) (Table, error)) {
	buildTableHandlers[tableType] = handler
}

// IsRegisteredTable returns true if handler registered for provided type
func IsRegisteredTable(tableType string) bool {
	for k := range buildTableHandlers {
		if k == tableType {
			return true
		}
	}

	return false
}

// SetDefaultMetadataCache sets the default metadata cache for information returned by adapter.DescribeTable()
func SetDefaultMetadataCache(metadataCache cache.Interface) {
	defaultMetadataCache, _ = SetupMetadataCache(metadataCache)
}

// DefaultMetadataCache gets the default metadata cache for information returned by adapter.DescribeTable()
func DefaultMetadataCache() cache.Interface {
	return defaultMetadataCache
}

// SetupMetadataCache setups the metadata cache
func SetupMetadataCache(metadataCache interface{}) (cache.Interface, error) {
	if metadataCache == nil {
		return nil, errors.New("Argument must be of type cache.Interface, or a Registry key where a cache.Interface object is stored")
	}

	switch metadataCache.(type) {
	case string:
		if mdc := registry.Get(metadataCache.(string)); mdc != nil {
			return mdc.(cache.Interface), nil
		}

	case cache.Interface:
		return metadataCache.(cache.Interface), nil
	}

	return nil, errors.New("Argument must be of type cache.Interface, or a Registry key where a cache.Interface object is stored")
}

// DefaultTable is a default SQL Table
type DefaultTable struct {
	Options               *TableConfig
	Adapter               Adapter
	Schema                string
	Name                  string
	Cols                  []string
	Primary               []string
	Identity              int64
	ReferenceMap          map[string]TableReference
	DependentTables       map[string]TableReference
	DefaultSource         string
	DefaultValues         map[string]interface{}
	Sequence              string
	Metadata              map[string]*TableColumn
	MetadataCache         cache.Interface
	MetadataCacheInStruct bool
	RowType               string
	RowsetType            string
}

// SetOptions sets object options
func (t *DefaultTable) SetOptions(options *TableConfig) Table {
	t.SetAdapter(options.Adapter)
	t.Options = options
	return t
}

// SetRowType sets a table's representation row type
func (t *DefaultTable) SetRowType(tp string) Table {
	t.RowType = tp
	return t
}

// GetRowType returns representaion row type
func (t *DefaultTable) GetRowType() string {
	return t.RowType
}

// SetRowsetType sets a table's representation rowset type
func (t *DefaultTable) SetRowsetType(tp string) Table {
	t.RowType = tp
	return t
}

// GetRowsetType returns representaion rowset type
func (t *DefaultTable) GetRowsetType() string {
	return t.RowType
}

// AddReference adds a reference to the reference map
func (t *DefaultTable) AddReference(ruleKey string, columns []string, refTable string, refColumns []string, onDelete string, onUpdate string) Table {
	reference := TableReference{
		Columns:    columns,
		Table:      refTable,
		RefColumns: refColumns,
	}

	if onDelete != "" {
		reference.OnDelete = onDelete
	}

	if onUpdate != "" {
		reference.OnUpdate = onUpdate
	}

	t.ReferenceMap[ruleKey] = reference

	return t
}

// SetReferences sets reference map
func (t *DefaultTable) SetReferences(referenceMap map[string]TableReference) Table {
	t.ReferenceMap = referenceMap
	return t
}

// Reference returns a table reference
func (t *DefaultTable) Reference(table string, ruleKey string) (TableReference, error) {
	refMap := t.getReferenceMapNormalized()
	if ruleKey != "" {
		if _, ok := refMap[ruleKey]; !ok {
			return TableReference{}, errors.Errorf("No reference rule '%s' from table '%s' to table '%s'", ruleKey, t.Options.DefinitionName, table)
		}

		if refMap[ruleKey].Table != table {
			return TableReference{}, errors.Errorf("Reference rule '%s' does not reference table '%s'", ruleKey, table)
		}

		return refMap[ruleKey], nil
	}

	for _, reference := range refMap {
		if reference.Table == table {
			return reference, nil
		}
	}

	return TableReference{}, errors.Errorf("No reference from table '%s' to table '%s'", t.Options.DefinitionName, table)
}

// SetDependentTables sets a dependant tables map
func (t *DefaultTable) SetDependentTables(dependentTables map[string]TableReference) Table {
	t.DependentTables = dependentTables
	return t
}

// GetDependentTables returns a map of dependant tables
func (t *DefaultTable) GetDependentTables() map[string]TableReference {
	return t.DependentTables
}

// SetDefaultValues set the default values for the table object
func (t *DefaultTable) SetDefaultValues(defaultValues map[string]interface{}) Table {
	for defaultName, defaultValue := range defaultValues {
		if _, ok := t.Metadata[defaultName]; ok {
			t.DefaultValues[defaultName] = defaultValue
		}
	}

	return t
}

// GetDefaultValues returns a map of default values
func (t *DefaultTable) GetDefaultValues() map[string]interface{} {
	return t.DefaultValues
}

// SetAdapter sets a database adapter for table
func (t *DefaultTable) SetAdapter(db interface{}) Table {
	t.Adapter, _ = SetupAdapter(db)
	return t
}

// GetAdapter returns a database adapter for table
func (t *DefaultTable) GetAdapter() Adapter {
	return t.Adapter
}

// SetMetadataCache sets the metadata cache for information returned by table.Describe()
func (t *DefaultTable) SetMetadataCache(metadataCache interface{}) Table {
	t.MetadataCache, _ = SetupMetadataCache(metadataCache)
	return t
}

// GetMetadataCache gets the metadata cache for information returned by table.Describe()
func (t *DefaultTable) GetMetadataCache() cache.Interface {
	return t.MetadataCache
}

// SetMetadataCacheInStruct indicate whether metadata should be cached in the struct for the duration of the instance
func (t *DefaultTable) SetMetadataCacheInStruct(flag bool) Table {
	t.MetadataCacheInStruct = flag
	return t
}

// IsMetadataCacheInClass retrieve flag indicating if metadata should be cached for duration of instance
func (t *DefaultTable) IsMetadataCacheInClass() bool {
	return t.MetadataCacheInStruct
}

// Setup is turnkey for initialization of a table object
// Calls other protected methods for individual tasks, to make it easier
// for a subclass to override part of the setup logic
func (t *DefaultTable) Setup() error {
	if err := t.SetupAdapter(); err != nil {
		return err
	}

	t.SetupTableName()
	return nil
}

// SetupAdapter initializes database adapter
func (t *DefaultTable) SetupAdapter() error {
	if t.Adapter == nil {
		t.Adapter = GetDefaultAdapter()
		if t.Adapter != nil {
			return errors.Errorf("No adapter found for '%t'", reflect.TypeOf(t))
		}
	}

	return nil
}

// SetupTableName initialize table and schema names
// If the table name is not set in the class definition, use the class name itself as the table name
//
// A schema name provided with the table name (e.g., "schema.table") overrides any existing value for t.Schema
func (t *DefaultTable) SetupTableName() {
	if t.Name == "" {
		r := reflect.TypeOf(t)
		parts := strings.Split(r.String(), ".")
		t.Name = parts[len(parts)-1]
	} else if strings.Contains(t.Name, ".") {
		parts := strings.Split(string(t.Name), ".")
		t.Schema = parts[0]
		t.Name = parts[1]
	}
}

// SetupMetadata initializes metadata
// If metadata cannot be loaded from cache, adapter's DescribeTable() method is called to discover metadata
// information. Returns true if and only if the metadata are loaded from cache
func (t *DefaultTable) SetupMetadata() (fromCache bool, err error) {
	if t.MetadataCacheInStruct && len(t.Metadata) > 0 {
		return true, nil
	}

	if t.MetadataCache == nil && defaultMetadataCache != nil {
		t.SetMetadataCache(defaultMetadataCache)
	}

	hashed := md5.New()
	hashed.Write([]byte(t.Adapter.FormatDSN() + ":" + t.Schema + "." + t.Name))
	cacheID := hex.EncodeToString(hashed.Sum(nil))

	if t.MetadataCache != nil {
		if metadata, ok := t.MetadataCache.Load(cacheID, false); ok {
			json.Unmarshal(metadata, &t.Metadata)
			return ok, nil
		}
	}

	t.Metadata, err = t.Adapter.DescribeTable(t.Name, t.Schema)
	if err != nil {
		return false, err
	}

	if t.MetadataCache != nil {
		serialized, _ := json.Marshal(t.Metadata)
		if ok := t.MetadataCache.Save(serialized, cacheID, []string{cacheID}, 0); !ok {
			return false, errors.New("Failed saving metadata to metadataCache")
		}
	}

	return false, nil
}

// GetCols retrieves table columns
func (t *DefaultTable) GetCols() []string {
	if t.Cols == nil {
		t.SetupMetadata()
		keys := make([]string, len(t.Metadata))
		i := 0
		for k := range t.Metadata {
			keys[i] = k
			i++
		}

		t.Cols = keys
	}

	return t.Cols
}

// SetupPrimaryKey initializes primary key from metadata
// If t.Primary is not defined, discover primary keys from the information returned by adapter.DescribeTable()
func (t *DefaultTable) SetupPrimaryKey() error {
	if t.Primary == nil || len(t.Primary) == 0 {
		if _, err := t.SetupMetadata(); err != nil {
			return err
		}

		t.Primary = []string{}
		for _, col := range t.Metadata {
			if col.Primary {
				t.Primary = append(t.Primary, col.Name)
				if col.Identity {
					t.Identity = col.PrimaryPosition
				}
			}
		}

		if len(t.Primary) == 0 {
			return errors.Errorf("A table must have a primary key, but none was found for table '%s'", t.Name)
		}
	}

	cols := t.GetCols()
	intersect, ok := utils.IntersectSSlice(t.Primary, cols)
	if !ok || !utils.EqualSSlice(t.Primary, intersect) {
		return errors.Errorf("Primary key column(s) ( %s ) are not columns in this table ( %s )", strings.Join(t.Primary, ", "), strings.Join(cols, ", "))
	}

	return nil
}

// Init initializes object
func (t *DefaultTable) Init() error {
	return nil
}

// Info returns table information
func (t *DefaultTable) Info() TableInfo {
	t.SetupPrimaryKey()

	return TableInfo{
		Schema:       t.Schema,
		Name:         t.Name,
		Cols:         t.GetCols(),
		Primary:      t.Primary,
		RowType:      t.GetRowType(),
		RowsetType:   t.GetRowsetType(),
		ReferenceMap: t.ReferenceMap,
		Sequence:     t.Sequence,
	}
}

// Select returns an instance of a dbselect.Interface object
func (t *DefaultTable) Select(withFromPart bool) (Select, error) {
	slct, err := NewSelectFromConfig(Options().Select)
	if err != nil {
		return nil, err
	}

	if withFromPart == SelectWithFormPart {
		tableSpec := t.Name
		if t.Schema != "" {
			tableSpec = t.Schema + "." + tableSpec
		}
		slct.From(tableSpec, SQLWildcard)
	}

	return slct, nil
}

// Insert inserts a new row
func (t *DefaultTable) Insert(ctx context.Context, data map[string]interface{}) (int, error) {
	if err := t.SetupPrimaryKey(); err != nil {
		return 0, err
	}

	primary := t.Primary
	pkIdentity := primary[t.Identity]

	if _, ok := data[pkIdentity]; ok {
		if data[pkIdentity] == nil || data[pkIdentity] == `` {
			delete(data, pkIdentity)
		}
	}

	if _, ok := data[pkIdentity]; !ok && t.Sequence != "" {
		data[pkIdentity] = t.Adapter.NextSequenceID(t.Sequence)
	}

	tableSpec := t.Name
	if t.Schema != "" {
		tableSpec = t.Schema + "." + tableSpec
	}

	_, err := t.Adapter.Insert(ctx, tableSpec, data)
	if err != nil {
		return 0, err
	}

	//if _, ok := data[pkIdentity]; !ok && t.Sequence != "" {
	//	data[pkIdentity] = t.Adapter.LastInsertID()
	//}

	return data[pkIdentity].(int), nil
}

// IsIdentity check if the provided column is an identity of the table
func (t *DefaultTable) IsIdentity(column string) bool {
	if err := t.SetupPrimaryKey(); err != nil {
		return false
	}

	if _, ok := t.Metadata[column]; !ok {
		errors.Errorf("Column '%s' not found in table", column)
	}

	return t.Metadata[column].Identity
}

// Update updates existing rows
func (t *DefaultTable) Update(ctx context.Context, data map[string]interface{}, cond map[string]interface{}) (bool, error) {
	tableSpec := t.Name
	if t.Schema != "" {
		tableSpec = t.Schema + "." + tableSpec
	}

	return t.Adapter.Update(ctx, tableSpec, data, cond)
}

// CascadeUpdate called by a row object for the parent table's class during save() method
func (t *DefaultTable) CascadeUpdate(ctx context.Context, parentTable string, oldPrimaryKey map[string]string, newPrimaryKey map[string]string) (int, error) {
	t.SetupMetadata()
	rowsAffected := 0
	for _, ref := range t.getReferenceMapNormalized() {
		if ref.Table == parentTable && ref.OnUpdate != "" {
			switch ref.OnUpdate {
			case Cascade:
				newRefs := map[string]interface{}{}
				where := map[string]interface{}{}
				for i := 0; i < len(ref.Columns); i++ {
					col := t.Adapter.FoldCase(ref.Columns[i])
					refCol := t.Adapter.FoldCase(ref.RefColumns[i])
					if v, ok := newPrimaryKey[refCol]; ok {
						newRefs[col] = v
					}

					where[t.Adapter.QuoteIdentifier(col, true)+" = ?"] = oldPrimaryKey[refCol]
				}

				ok, err := t.Adapter.Update(ctx, ref.Table, newRefs, where)
				if err == nil {
					return rowsAffected, err
				}

				if ok {
					rowsAffected++
				}
			}
		}
	}

	return rowsAffected, nil
}

// Delete deletes existing rows
func (t *DefaultTable) Delete(ctx context.Context, cond map[string]interface{}) (bool, error) {
	depTables := t.GetDependentTables()
	if len(depTables) > 0 {
		resultSet, err := t.FetchAll(ctx, cond)
		if err != nil {
			return false, err
		}

		if !resultSet.IsEmpty() {
			for resultSet.Next() {
				/*row := resultSet.Get()
				// Execute cascading deletes against dependent tables
				for _, tbl := range depTables {
					tbl.CascadeDelete(t.Name, row.PrimaryKey())
				}*/
			}
		}
	}

	tableSpec := t.Name
	if t.Schema != "" {
		tableSpec = t.Schema + "." + tableSpec
	}

	return t.Adapter.Delete(ctx, tableSpec, cond)
}

// CascadeDelete called by parent table's object during delete() method
func (t *DefaultTable) CascadeDelete(ctx context.Context, parentTable string, primaryKey map[string]interface{}) (int, error) {
	t.SetupMetadata()

	rowsAffected := 0
	for _, ref := range t.getReferenceMapNormalized() {
		if ref.Table == parentTable && ref.OnDelete != "" {
			cond := make(map[string]interface{})

			if ref.OnDelete == Cascade || ref.OnDelete == CascadeRecurse {
				for i := 0; i < len(ref.Columns); i++ {
					col := t.Adapter.FoldCase(ref.Columns[i])
					refCol := t.Adapter.FoldCase(ref.RefColumns[i])
					cond[t.Adapter.QuoteIdentifier(col, true)+" = ?"] = primaryKey[refCol]
				}
			}

			if ref.OnDelete == CascadeRecurse {
				// Execute cascading deletes against dependent tables
				depTables := t.GetDependentTables()
				if len(depTables) > 0 {
					/*for _, tbl := range depTables {
						resultSet, err := t.FetchAll(cond)
						for resultSet.Next() {
							row := resultSet.Get()

							ok, err := d.Adapter.CascadeDelete(t.Name, row.PrimaryKey())
							if err != nil {
								return rowsAffected, err
							}

							if ok {
								rowsAffected++
							}
						}
					}*/
				}
			}

			if ref.OnDelete == Cascade || ref.OnDelete == CascadeRecurse {
				ok, err := t.Adapter.Delete(ctx, t.Name, cond)
				if err != nil {
					return rowsAffected, err
				}

				if ok {
					rowsAffected++
				}
			}
		}
	}

	return rowsAffected, nil
}

// Find fetches rows by primary key.  The argument specifies one or more primary
// key value(s).  To find multiple rows by primary key, the argument must
// be an array.
//
// This method accepts a variable number of arguments.  If the table has a
// multi-column primary key, the number of arguments must be the same as
// the number of columns in the primary key. To find multiple rows in a
// table with a multi-column primary key, each argument must be an array
// with the same number of elements.
//
// The Find() method always returns a rowset.Interface object, even if only one row
// was found.
func (t *DefaultTable) Find(ctx context.Context, args ...interface{}) (Rowset, error) {
	t.SetupPrimaryKey()

	if len(args) < len(t.Primary) {
		return NewEmptyRowset(t.GetRowsetType()), errors.New("Too few columns for the primary key")
	}

	if len(args) > len(t.Primary) {
		return NewEmptyRowset(t.GetRowsetType()), errors.New("Too many columns for the primary key")
	}

	var whereList [][]interface{}
	numberTerms := 0
	for keyPosition, keyValues := range args {
		var v []interface{}
		keyValuesCount := 0
		switch keyValues.(type) {
		case []interface{}:
			v = keyValues.([]interface{})
			keyValuesCount = len(v)

		default:
			v = []interface{}{keyValues}
			keyValuesCount = 1
		}

		if numberTerms == 0 {
			numberTerms = keyValuesCount
			whereList = make([][]interface{}, keyValuesCount, len(args))
		} else if keyValuesCount != numberTerms {
			return nil, errors.New("Missing value(s) for the primary key")
		}

		for i := 0; i < keyValuesCount; i++ {
			whereList[i][keyPosition] = v[i]
		}
	}

	whereClause := ""
	if len(whereList) > 0 {
		whereOrTerms := make([]string, 0)
		tableName := t.Adapter.QuoteTableAs(t.Name, "", true)
		for _, keyValueSets := range whereList {
			whereAndTerms := make([]string, 0)
			for keyPosition, keyValue := range keyValueSets {
				//typ := t.Metadata[t.Primary[keyPosition]].DataType
				columnName := t.Adapter.QuoteIdentifier(t.Primary[keyPosition], true)
				whereAndTerms = append(whereAndTerms, t.Adapter.QuoteInto(tableName+"."+columnName+" = ?", keyValue, 1))
			}

			whereOrTerms = append(whereOrTerms, "("+strings.Join(whereAndTerms, " AND ")+")")
		}

		whereClause = "(" + strings.Join(whereOrTerms, " OR ") + ")"
	}

	if whereClause == "" {
		return NewEmptyRowset(t.GetRowsetType()), nil
	}

	return t.FetchAll(ctx, whereClause)
}

// FetchAll fetches all rows
func (t *DefaultTable) FetchAll(ctx context.Context, cond interface{}) (Rowset, error) {
	return t.fetchAll(ctx, cond, nil, 0, 0)
}

// FetchAllOrder fetches all rows with specified order
func (t *DefaultTable) FetchAllOrder(ctx context.Context, cond interface{}, order interface{}) (Rowset, error) {
	return t.fetchAll(ctx, cond, order, 0, 0)
}

// FetchAllLimit fetches all rows with limit and offset
func (t *DefaultTable) FetchAllLimit(ctx context.Context, cond interface{}, count int, offset int) (Rowset, error) {
	return t.fetchAll(ctx, cond, nil, count, offset)
}

// FetchAllOrderLimit fetches all rows with with specified order, limit and offset
func (t *DefaultTable) FetchAllOrderLimit(ctx context.Context, cond interface{}, order interface{}, count int, offset int) (Rowset, error) {
	return t.fetchAll(ctx, cond, order, count, offset)
}

func (t *DefaultTable) fetchAll(ctx context.Context, cond interface{}, order interface{}, count int, offset int) (Rowset, error) {
	var slct Select
	var err error
	switch cond.(type) {
	case Select:
		slct = cond.(Select)

	case string:
		slct, err = t.Select(true)
		if err != nil {
			return NewEmptyRowset(t.GetRowsetType()), err
		}

		t.where(slct, cond)
	}

	if order != nil {
		t.order(slct, order)
	}

	if count > 0 || offset > 0 {
		slct.Limit(count, offset)
	}

	rctx := RowsetConfigToContext(ctx, &RowsetConfig{
		Type: t.GetRowsetType(),
		Tbl:  t.Name,
		Row: &RowConfig{
			Type:  t.GetRowType(),
			Table: t.Name,
		},
	})
	rows, err := t.Adapter.Query(rctx, slct)
	if err != nil {
		return NewEmptyRowset(t.GetRowsetType()), err
	}

	rows.SetTable(t)
	return rows, nil
}

// FetchRow fetches one row in an object of type db.Row,
// or returns nil if no row matches the specified criteria
func (t *DefaultTable) FetchRow(ctx context.Context, cond interface{}) (Row, error) {
	return t.fetchRow(ctx, cond, nil, 0)
}

// FetchRowOrder fetches one row in an object of type db.Row,
// or returns nil if no row matches the specified criteria
func (t *DefaultTable) FetchRowOrder(ctx context.Context, cond interface{}, order interface{}) (Row, error) {
	return t.fetchRow(ctx, cond, order, 0)
}

// FetchRowOffset fetches one row in an object of type db.Row,
// or returns nil if no row matches the specified criteria
func (t *DefaultTable) FetchRowOffset(ctx context.Context, cond interface{}, offset int) (Row, error) {
	return t.fetchRow(ctx, cond, nil, offset)
}

// FetchRowOrderOffset fetches one row in an object of type db.Row,
// or returns nil if no row matches the specified criteria
func (t *DefaultTable) FetchRowOrderOffset(ctx context.Context, cond interface{}, order interface{}, offset int) (Row, error) {
	return t.fetchRow(ctx, cond, order, offset)
}

func (t *DefaultTable) fetchRow(ctx context.Context, cond interface{}, order interface{}, offset int) (Row, error) {
	var slct Select
	var err error
	switch cond.(type) {
	case Select:
		slct = cond.(Select)

	case string:
		slct, err = t.Select(true)
		if err != nil {
			return t.CreateRow(nil, DefaultNone), err
		}

		t.where(slct, cond)
	}

	if order != nil {
		t.order(slct, order)
	}

	if offset > 0 {
		slct.Limit(1, offset)
	} else {
		slct.Limit(1, 0)
	}
	fmt.Println(t.GetRowType())
	os.Exit(2)
	rctx := RowConfigToContext(ctx, &RowConfig{
		Type:  t.GetRowType(),
		Table: t.Name,
	})
	row, err := t.Adapter.QueryRow(rctx, slct)
	if err != nil {
		return t.CreateRow(nil, DefaultNone), err
	}

	return row, nil
}

// CreateRow fetches a new blank row (not from the database)
func (t *DefaultTable) CreateRow(data map[string]interface{}, defaultSource string) Row {
	cols := t.GetCols()
	defaults := make(map[string]interface{})
	for _, col := range cols {
		defaults[col] = nil
	}

	if defaultSource == "" {
		defaultSource = t.DefaultSource
	}

	if !utils.InSSlice(defaultSource, []string{DefaultNone, DefaultObject, DefaultDb}) {
		defaultSource = DefaultNone
	}

	if defaultSource == DefaultDb {
		for columnName, metadata := range t.Metadata {
			v, ok := t.DefaultValues[columnName]
			if metadata.Default != nil && (!metadata.IsNullable || metadata.IsNullable && ok && v.(bool)) && (!ok && !v.(bool)) {
				defaults[columnName] = metadata.Default
			}
		}
	} else if defaultSource == DefaultObject && len(t.DefaultValues) > 0 {
		for defaultName, defaultValue := range t.DefaultValues {
			if _, ok := defaults[defaultName]; ok {
				defaults[defaultName] = defaultValue
			}
		}
	}

	config := &RowConfig{
		Type:     t.GetRowType(),
		Table:    t.Name,
		Data:     defaults,
		ReadOnly: false,
		Stored:   false,
	}

	row, err := NewRow(t.GetRowType(), config)
	if err != nil {
		return NewEmptyRow(t.GetRowType())
	}

	row.SetTable(t)
	row.Populate(data)
	return row
}

// CreateRowConfig creates a row config with default values from table
func (t *DefaultTable) CreateRowConfig(defaultSource string) *RowConfig {
	cols := t.GetCols()
	defaults := make(map[string]interface{})
	for _, col := range cols {
		defaults[col] = nil
	}

	if defaultSource == "" {
		defaultSource = t.DefaultSource
	}

	if !utils.InSSlice(defaultSource, []string{DefaultNone, DefaultObject, DefaultDb}) {
		defaultSource = DefaultNone
	}

	if defaultSource == DefaultDb {
		for columnName, metadata := range t.Metadata {
			v, ok := t.DefaultValues[columnName]
			if metadata.Default != nil && (!metadata.IsNullable || metadata.IsNullable && ok && v.(bool)) && (!ok && !v.(bool)) {
				defaults[columnName] = metadata.Default
			}
		}
	} else if defaultSource == DefaultObject && len(t.DefaultValues) > 0 {
		for defaultName, defaultValue := range t.DefaultValues {
			if _, ok := defaults[defaultName]; ok {
				defaults[defaultName] = defaultValue
			}
		}
	}

	return &RowConfig{
		Type:     t.GetRowType(),
		Table:    t.Name,
		Data:     defaults,
		ReadOnly: false,
		Stored:   false,
	}
}

// Generate WHERE clause from user-supplied string or array
func (t *DefaultTable) where(slct Select, cond interface{}) Select {
	switch cond.(type) {
	case string:
		slct.Where(cond.(string), nil)

	case map[string]interface{}:
		for c, b := range cond.(map[string]interface{}) {
			slct.Where(c, b)
		}
	}

	return slct
}

// Generate ORDER clause from user-supplied string or array
func (t *DefaultTable) order(slct Select, order interface{}) Select {
	orderSlice := make([]string, 0)
	switch order.(type) {
	case []string:
		orderSlice = order.([]string)

	case string:
		orderSlice = []string{order.(string)}
	}

	for _, val := range orderSlice {
		slct.Order(val)
	}

	return slct
}

func (t *DefaultTable) getReferenceMapNormalized() map[string]TableReference {
	return t.ReferenceMap
}

// NewDefaultTable creates a default table object
func NewDefaultTable(options *TableConfig) (Table, error) {
	t := &DefaultTable{}
	if options != nil {
		t.SetOptions(options)
	}

	t.Setup()
	t.Init()
	return t, nil
}

// TableReference holds table references
type TableReference struct {
	Columns    []string
	Table      string
	RefColumns []string
	OnDelete   string
	OnUpdate   string
}

// TableInfo holds table info
type TableInfo struct {
	Schema          string
	Name            string
	Cols            []string
	Primary         []string
	RowType         string
	RowsetType      string
	ReferenceMap    map[string]TableReference
	DependantTables map[string]TableReference
	Sequence        string
}

// TableColumn represents information schema column
type TableColumn struct {
	TableSchema     string
	TableName       string
	Name            string
	Default         interface{}
	Position        int64
	DataType        string
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
