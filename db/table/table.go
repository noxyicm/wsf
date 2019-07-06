package table

import (
	"wsf/db/dbselect"
	"wsf/registry"
	"wsf/db/adapter"
	"wsf/utils"
	"wsf/cache/backend"
	"wsf/db"
)

// Public constants
const (
	Adapter          = "db"
    Definition        = "definition"
    DefinitionName = "definitionName"
    Schema           = "schema"
    Name             = "name"
	Primary          = "primary"
	Cols             = "cols"
	MetaData         = "metadata"
    MetaDataCache   = "metadataCache"
    RowType        = "rowType"
    RowsetType     = "rowsetType"
    ReferenceMap    = "referenceMap"
    DependentTables = "dependentTables"
    Sequence         = "sequence"

    Columns          = "columns"
    RefTable  = "refTable"
    RefColumns      = "refColumns"
	OnDelete        = "onDelete"
	OnUpdate        = "onUpdate"

    Cascade          = "cascade"
    CascadeRecurse  = "cascadeRecurse"
    Restrict         = "restrict"
    SetNull         = "setNull"

    DefaultNone     = "defaultNone"
    DefaultType    = "defaultType"
    DefaultDb       = "defaultDb"

    SelectWithFormPart    = true
    SelectWithoutFormPart = false
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	DefaultDb adapter.Interface
	DefaultMetadataCache backend.Interface
)

// Interface is an interface for SQL table
type Interface interface {
	SetOptions(options *Config) Interface
	SetRowType(tp string) Interface
	GetRowClass() string
	SetRowsetType(tp string) Interface
	GetRowsetClass() string
}

// NewAdapter creates a new table from given type and options
func NewTable(tableType string, options config.Config) (ti Interface, err error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[tableType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database table type \"%v\"", tableType)
}

// Register registers a handler for database table creation
func Register(tableType string, handler func(*Config) (Interface, error)) {
	buildHandlers[tableType] = handler
}

// SetDefaultAdapter sets the default db.adapter.Interface for all db.table.Interface objects
func SetDefaultAdapter(db interface{}) {
	DefaultDb = SetupAdapter(db)
}

// DefaultAdapter gets the default db.adapter.Interface for all db.table.Interface objects
func DefaultAdapter() adapter.Interface {
	return DefaultDb
}

// SetupAdapter checks if db is a valid database adapter
func SetupAdapter(db interface{}) (adapter.Interface, error) {
	if db == nil {
		return nil, errors.New("Argument must be of type db.adapter.Interface, or a Registry key where a db.adapter.Interface object is stored")
	}

	switch db.(type) {
	case string:
		if dba := registry.Get(db.(string)); dba != nil {
			return dba.(adapter.Interface), nil
		}

	case adapter.Interface:
		return db.(adapter.Interface), nil
	}

	return nil, errors.New("Argument must be of type db.adapter.Interface, or a Registry key where a db.adapter.Interface object is stored")
}

// Sets the default metadata cache for information returned by adapter.DescribeTable()
func SetDefaultMetadataCache(metadataCache backend.Interface) {
	DefaultMetadataCache = SetupMetadataCache(metadataCache)
}

// DefaultMetadataCache gets the default metadata cache for information returned by adapter.DescribeTable()
func DefaultMetadataCache() backend.Interface {
	return DefaultMetadataCache
}

// SetupMetadataCache setups the metadata cache
func SetupMetadataCache(metadataCache interface{}) (backend.Interface, error) {
	if metadataCache == nil {
		return nil, errors.New("Argument must be of type cache.Interface, or a Registry key where a cache.Interface object is stored")
	}

	switch metadataCache.(type) {
	case string:
		if mdc := registry.Get(metadataCache.(string)); mdc != nil {
			return mdc.(backend.Interface), nil
		}

	case backend.Interface:
		return metadataCache.(backend.Interface), nil
	}

	return nil, errors.New("Argument must be of type cache.Interface, or a Registry key where a cache.Interface object is stored")
}

// Table is a default SQL Table
type Table struct {
	Options *Config
    Db adapter.Interface
    Schema string
    Name string
    Cols []string
    Primary         []string
    Identity int64
	ReferenceMap map[string]TableReference
	Sequence string
	Metadata map[string]*adapter.ColumnMetadata
	MetadataCache backend.Interface
	RowType string
	RowsetType string
}

// SetOptions sets object options
func (t *Table) SetOptions(options *Config) Interface {
	t.SetAdapter(options.Adapter)
	t.Options = options
	return t
}

// SetRowType sets a table's representation row type
func (t *Table) SetRowType(tp string) Interface {
	t.RowType = tp
	return t
}

// GetRowClass returns representaion row type
func (t *Table) GetRowClass() string {
	return t.RowType
}

// SetRowsetType sets a table's representation rowset type
func (t *Table) SetRowsetType(tp string) Interface {
	t.RowType = tp
	return t
}

// GetRowsetClass returns representaion rowset type
func (t *Table) GetRowsetClass() string {
	return t.RowType
}

// AddReference adds a reference to the reference map
func (t *Table) AddReference(ruleKey string, columns Columns, refTable string, refColumns Columns, onDelete string, onUpdate string) Interface {
	reference := TableReference{
		Columns: columns,
		Table: refTable,
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
func (t *Table) SetReferences(referenceMap map[string]TableReference) Interface {
	t.ReferenceMap = referenceMap
	return t
}

// Reference returns a table reference
func (t *Table) Reference(table string, ruleKey string) (TableReference, error) {
	refMap := t.getReferenceMapNormalized()
	if ruleKey != "" {
		if _, ok := refMap[ruleKey]; !ok {
			return nil, errors.Errorf("No reference rule '%s' from table '%s' to table '%s'", ruleKey, t.Options.DefinitionName, table)
		}

		if refMap[ruleKey].Table != table {
			return nil, errors.Errorf("Reference rule '%s' does not reference table '%s'", ruleKey, table)
		}

		return refMap[ruleKey]
	}

	for _, reference := range refMap {
		if reference.Table == table {
			return reference
		}
	}

	return nil, errors.Errorf("No reference from table '%s' to table '%s'", t.Options.DefinitionName, table);
}

// SetDependentTables sets a dependant tables map
func (t *Table) SetDependentTables(dependentTables map[string]interface{}) Interface {
	t.DependentTables = dependentTables
	return t
}

// GetDependentTables returns a map of dependant tables
func (t *Table) GetDependentTables() map[string]interface{} {
	return t.DependentTables
}

// SetDefaultValues set the default values for the table object
func (t *Table) SetDefaultValues(defaultValues map[string]interface{}) Interface {
	for defaultName, defaultValue := range defaultValues {
		if _, ok := t.Metadata[defaultName]; ok {
			t.DefaultValues[defaultName] = defaultValue
		}
	}

	return t
}

// GetDefaultValues returns a map of default values
func (t *Table) GetDefaultValues() map[string]interface{} {
	return t.DefaultValues
}

// SetAdapter sets a database adapter for table
func (t *Table) SetAdapter(db interface{}) Interface {
	t.Db = SetupAdapter(db)
	return t
}

// Adapter returns a database adapter for table
func (t *Table) Adapter() adapter.Interface {
	return t.Db
}

// Sets the metadata cache for information returned by table.Describe()
func (t *Table) SetMetadataCache(metadataCache interface{}) Interface {
	t.MetaDataCache = SetupMetadataCache(metadataCache)
	return t
}

// GetMetadataCache gets the metadata cache for information returned by table.Describe()
func (t *Table) GetMetadataCache() backend.Interface {
	return t.MetadataCache
}

// SetMetadataCacheInStruct indicate whether metadata should be cached in the struct for the duration of the instance
func (t *Table) SetMetadataCacheInStruct(flag bool) Interface {
	t.MetadataCacheInStruct = flag
	return t
}

// IsMetadataCacheInClass retrieve flag indicating if metadata should be cached for duration of instance
func (t *Table) IsMetadataCacheInClass() bool {
	return t.MetadataCacheInStruct
}

// Setup is turnkey for initialization of a table object
// Calls other protected methods for individual tasks, to make it easier
// for a subclass to override part of the setup logic
func (t *Table) Setup() error {
	if err := t.SetupDatabaseAdapter(); err != nil {
		return err
	}

	t.SetupTableName()
}

// SetupDatabaseAdapter initializes database adapter
func (t *Table) SetupDatabaseAdapter() error {
	if t.Db == nil {
		t.Db = DefaultAdapter()
		if t.Db != nil {
			return errors.Errorf("No adapter found for '%t'", reflect.TypeOf(t))
		}
	}

	return nil
}

// SetupTableName initialize table and schema names
// If the table name is not set in the class definition, use the class name itself as the table name
//
// A schema name provided with the table name (e.g., "schema.table") overrides any existing value for t.Schema
func (t *Table) SetupTableName() {
	if t.Name == nil {
		r := reflect.TypeOf(t)
		parts := strings.Split(string(r), ".")
		t.Name = parts[len(parts) - 1]
	} else if strings.Contains(t.Name, ".") {
		parts := strings.Split(string(t.Name), ".")
		t.Schema = parts[0]
		t.Name = parts[1]
	}
}

// SetupMetadata initializes metadata
// If metadata cannot be loaded from cache, adapter's DescribeTable() method is called to discover metadata
// information. Returns true if and only if the metadata are loaded from cache
func (t *Table) SetupMetadata() error {
	if t.IsMetadataCacheInStruct() && len(t.MetaData) > 0 {
		return true
	}

	isMetadataFromCache := true
	if t.MetadataCache == nil && DefaultMetadataCache != nil {
		t.SetMetadataCache(DefaultMetadataCache)
	}

	var metadata map[string]*adapter.ColumnMetadata
	if t.MetadataCache != nil {
		hashed := md5.New()
		hashed.Write([]byte(t.Db.FormatDSN() +":"+ t.Schema +"."+ t.Name))
		cacheId := hex.EncodeToString(hashed.Sum(nil))

		metadata, err := t.MetadataCache.Load(cacheId)
		if err != nil {
			isMetadataFromCache = false
		}
	} else {
		isMetadataFromCache = false
		metadata, err := t.Db.DescribeTable(t.Name, t.Schema)
		if err != nil {
			return err
		}

		if err := t.MetadataCache.Save(metadata, cacheId, []string{cacheId}, 0); err != nil {
			return errors.New("Failed saving metadata to metadataCache")
		}
	}

	t.Metadata = metadata
	return isMetadataFromCache
}

// GetCols retrieves table columns
func (t *Table) GetCols() []string {
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
func (t *Table) SetupPrimaryKey() error {
    if t.Primary == nil || len(t.Primary) == 0 {
        if err := t.SetupMetadata(); err != nil {
            return err
        }

        s.Primary = []string{}
        for _, col := range t.Metadata {
            if col.Primary {
                t.Primary = append(t.Primary, col.Name)
                if col.Identity {
                    t.Identity = col.PrimaryPosition
                }
            }
        }

        if len(s.Primary) == 0 {
            return errors.Errorf("A table must have a primary key, but none was found for table '%s'", t.Name);
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
func (t *Table) Init() error {
    return nil
}

// Info returns table information
func (t *Table) Info() TableInfo {
    t.SetupPrimaryKey()

    return TableInfo{
        Schema: t.Schema,
        Name: t.Name,
        Cols: t.GetCols(),
        Primary: t.Primary,
        RowType: t.GetRowType(),
        RowsetType: t.GetRowsetType(),
        ReferenceMap: t.ReferenceMap,
        Sequence: t.Sequence,
    }
}

// Select returns an instance of a dbselect.Interface object
func (t *Table) Select(withFromPart bool) (dbtable.Interface, error) {
    slct, err := dbselect.New()
    if err != nil {
        return nil, err
    }

    if withFromPart == SelectWithFormPart {
        slct.From(t.Name, dbselect.SQLWildcard, t.Schema)
    }

    return slct, nil
}

// Insert inserts a new row
func (t *Table) Insert(data map[string]interface{}) (int, error) {
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
		data[pkIdentity] = t.Adapter.NextSequenceId(t.Sequence)
	}

	tableSpec := t.Name
	if t.Schema != "" {
		tableSpec = t.Schema+"."+tableSpec
	}

	insertedId, err := t.Adapter.Insert(tableSpec, data)
	if err != nil {
		return 0, err
	}

	if _, ok := data[pkIdentity]; !ok && t.Sequence != "" {
		data[pkIdentity] = t.Adapter.LastInsertId()
	}

	return data[pkIdentity].(int), nil
}

// IsIdentity check if the provided column is an identity of the table
func (t *Table) IsIdentity(column string) bool {
	if err := t.SetupPrimaryKey(); err != nil {
		return false
	}

	if _, ok := t.Metadata[column]; !ok {
		errors.Errorf("Column '%s' not found in table", column)
	}

	return t.Metadata[column].Identity
}

    /**
     * Updates existing rows.
     *
     * @param  array        $data  Column-value pairs.
     * @param  array|string $where An SQL WHERE clause, or an array of SQL WHERE clauses.
     * @return int          The number of rows updated.
     */
    public function update(array $data, $where)
    {
        $tableSpec = ($this->_schema ? $this->_schema . '.' : '') . $this->_name;
        return $this->_db->update($tableSpec, $data, $where);
    }

    /**
     * Called by a row object for the parent table's class during save() method.
     *
     * @param  string $parentTableClassname
     * @param  array  $oldPrimaryKey
     * @param  array  $newPrimaryKey
     * @return int
     */
    public function _cascadeUpdate($parentTableClassname, array $oldPrimaryKey, array $newPrimaryKey)
    {
        $this->_setupMetadata();
        $rowsAffected = 0;
        foreach ($this->_getReferenceMapNormalized() as $map) {
            if ($map[self::REF_TABLE_CLASS] == $parentTableClassname && isset($map[self::ON_UPDATE])) {
                switch ($map[self::ON_UPDATE]) {
                    case self::CASCADE:
                        $newRefs = array();
                        $where = array();
                        for ($i = 0; $i < count($map[self::COLUMNS]); ++$i) {
                            $col = $this->_db->foldCase($map[self::COLUMNS][$i]);
                            $refCol = $this->_db->foldCase($map[self::REF_COLUMNS][$i]);
                            if (array_key_exists($refCol, $newPrimaryKey)) {
                                $newRefs[$col] = $newPrimaryKey[$refCol];
                            }
                            $type = $this->_metadata[$col]['DATA_TYPE'];
                            $where[] = $this->_db->quoteInto(
                                $this->_db->quoteIdentifier($col, true) . ' = ?',
                                $oldPrimaryKey[$refCol], $type);
                        }
                        $rowsAffected += $this->update($newRefs, $where);
                        break;
                    default:
                        // no action
                        break;
                }
            }
        }
        return $rowsAffected;
    }

    /**
     * Deletes existing rows.
     *
     * @param  array|string $where SQL WHERE clause(s).
     * @return int          The number of rows deleted.
     */
    public function delete($where)
    {
        $depTables = $this->getDependentTables();
        if (!empty($depTables)) {
            $resultSet = $this->fetchAll($where);
            if (count($resultSet) > 0 ) {
                foreach ($resultSet as $row) {
                    /**
                     * Execute cascading deletes against dependent tables
                     */
                    foreach ($depTables as $tableClass) {
                        $t = self::getTableFromString($tableClass, $this);
                        $t->_cascadeDelete(
                            get_class($this), $row->getPrimaryKey()
                        );
                    }
                }
            }
        }

        $tableSpec = ($this->_schema ? $this->_schema . '.' : '') . $this->_name;
        return $this->_db->delete($tableSpec, $where);
    }

    /**
     * Called by parent table's class during delete() method.
     *
     * @param  string $parentTableClassname
     * @param  array  $primaryKey
     * @return int    Number of affected rows
     */
    public function _cascadeDelete($parentTableClassname, array $primaryKey)
    {
        // setup metadata
        $this->_setupMetadata();

        // get this class name
        $thisClass = get_class($this);
        if ($thisClass === 'Zend_Db_Table') {
            $thisClass = $this->_definitionConfigName;
        }

        $rowsAffected = 0;

        foreach ($this->_getReferenceMapNormalized() as $map) {
            if ($map[self::REF_TABLE_CLASS] == $parentTableClassname && isset($map[self::ON_DELETE])) {

                $where = array();

                // CASCADE or CASCADE_RECURSE
                if (in_array($map[self::ON_DELETE], array(self::CASCADE, self::CASCADE_RECURSE))) {
                    for ($i = 0; $i < count($map[self::COLUMNS]); ++$i) {
                        $col = $this->_db->foldCase($map[self::COLUMNS][$i]);
                        $refCol = $this->_db->foldCase($map[self::REF_COLUMNS][$i]);
                        $type = $this->_metadata[$col]['DATA_TYPE'];
                        $where[] = $this->_db->quoteInto(
                            $this->_db->quoteIdentifier($col, true) . ' = ?',
                            $primaryKey[$refCol], $type);
                    }
                }

                // CASCADE_RECURSE
                if ($map[self::ON_DELETE] == self::CASCADE_RECURSE) {

                    /**
                     * Execute cascading deletes against dependent tables
                     */
                    $depTables = $this->getDependentTables();
                    if (!empty($depTables)) {
                        foreach ($depTables as $tableClass) {
                            $t = self::getTableFromString($tableClass, $this);
                            foreach ($this->fetchAll($where) as $depRow) {
                                $rowsAffected += $t->_cascadeDelete($thisClass, $depRow->getPrimaryKey());
                            }
                        }
                    }
                }

                // CASCADE or CASCADE_RECURSE
                if (in_array($map[self::ON_DELETE], array(self::CASCADE, self::CASCADE_RECURSE))) {
                    $rowsAffected += $this->delete($where);
                }

            }
        }
        return $rowsAffected;
    }

    /**
     * Fetches rows by primary key.  The argument specifies one or more primary
     * key value(s).  To find multiple rows by primary key, the argument must
     * be an array.
     *
     * This method accepts a variable number of arguments.  If the table has a
     * multi-column primary key, the number of arguments must be the same as
     * the number of columns in the primary key.  To find multiple rows in a
     * table with a multi-column primary key, each argument must be an array
     * with the same number of elements.
     *
     * The find() method always returns a Rowset object, even if only one row
     * was found.
     *
     * @param  mixed $key The value(s) of the primary keys.
     * @return Zend_Db_Table_Rowset_Abstract Row(s) matching the criteria.
     * @throws Zend_Db_Table_Exception
     */
    public function find()
    {
        $this->_setupPrimaryKey();
        $args = func_get_args();
        $keyNames = array_values((array) $this->_primary);

        if (count($args) < count($keyNames)) {
            // require_once 'Zend/Db/Table/Exception.php';
            throw new Zend_Db_Table_Exception("Too few columns for the primary key");
        }

        if (count($args) > count($keyNames)) {
            // require_once 'Zend/Db/Table/Exception.php';
            throw new Zend_Db_Table_Exception("Too many columns for the primary key");
        }

        $whereList = array();
        $numberTerms = 0;
        foreach ($args as $keyPosition => $keyValues) {
            $keyValuesCount = count($keyValues);
            // Coerce the values to an array.
            // Don't simply typecast to array, because the values
            // might be Zend_Db_Expr objects.
            if (!is_array($keyValues)) {
                $keyValues = array($keyValues);
            }
            if ($numberTerms == 0) {
                $numberTerms = $keyValuesCount;
            } else if ($keyValuesCount != $numberTerms) {
                // require_once 'Zend/Db/Table/Exception.php';
                throw new Zend_Db_Table_Exception("Missing value(s) for the primary key");
            }
            $keyValues = array_values($keyValues);
            for ($i = 0; $i < $keyValuesCount; ++$i) {
                if (!isset($whereList[$i])) {
                    $whereList[$i] = array();
                }
                $whereList[$i][$keyPosition] = $keyValues[$i];
            }
        }

        $whereClause = null;
        if (count($whereList)) {
            $whereOrTerms = array();
            $tableName = $this->_db->quoteTableAs($this->_name, null, true);
            foreach ($whereList as $keyValueSets) {
                $whereAndTerms = array();
                foreach ($keyValueSets as $keyPosition => $keyValue) {
                    $type = $this->_metadata[$keyNames[$keyPosition]]['DATA_TYPE'];
                    $columnName = $this->_db->quoteIdentifier($keyNames[$keyPosition], true);
                    $whereAndTerms[] = $this->_db->quoteInto(
                        $tableName . '.' . $columnName . ' = ?',
                        $keyValue, $type);
                }
                $whereOrTerms[] = '(' . implode(' AND ', $whereAndTerms) . ')';
            }
            $whereClause = '(' . implode(' OR ', $whereOrTerms) . ')';
        }

        // issue ZF-5775 (empty where clause should return empty rowset)
        if ($whereClause == null) {
            $rowsetClass = $this->getRowsetClass();
            if (!class_exists($rowsetClass)) {
                // require_once 'Zend/Loader.php';
                Zend_Loader::loadClass($rowsetClass);
            }
            return new $rowsetClass(array('table' => $this, 'rowClass' => $this->getRowClass(), 'stored' => true));
        }

        return $this->fetchAll($whereClause);
    }

    /**
     * Fetches all rows.
     *
     * Honors the Zend_Db_Adapter fetch mode.
     *
     * @param string|array|Zend_Db_Table_Select $where  OPTIONAL An SQL WHERE clause or Zend_Db_Table_Select object.
     * @param string|array                      $order  OPTIONAL An SQL ORDER clause.
     * @param int                               $count  OPTIONAL An SQL LIMIT count.
     * @param int                               $offset OPTIONAL An SQL LIMIT offset.
     * @return Zend_Db_Table_Rowset_Abstract The row results per the Zend_Db_Adapter fetch mode.
     */
    public function fetchAll($where = null, $order = null, $count = null, $offset = null)
    {
        if (!($where instanceof Zend_Db_Table_Select)) {
            $select = $this->select();

            if ($where !== null) {
                $this->_where($select, $where);
            }

            if ($order !== null) {
                $this->_order($select, $order);
            }

            if ($count !== null || $offset !== null) {
                $select->limit($count, $offset);
            }

        } else {
            $select = $where;
        }

        $rows = $this->_fetch($select);

        $data  = array(
            'table'    => $this,
            'data'     => $rows,
            'readOnly' => $select->isReadOnly(),
            'rowClass' => $this->getRowClass(),
            'stored'   => true
        );

        $rowsetClass = $this->getRowsetClass();
        if (!class_exists($rowsetClass)) {
            // require_once 'Zend/Loader.php';
            Zend_Loader::loadClass($rowsetClass);
        }
        return new $rowsetClass($data);
    }

    /**
     * Fetches one row in an object of type Zend_Db_Table_Row_Abstract,
     * or returns null if no row matches the specified criteria.
     *
     * @param string|array|Zend_Db_Table_Select $where  OPTIONAL An SQL WHERE clause or Zend_Db_Table_Select object.
     * @param string|array                      $order  OPTIONAL An SQL ORDER clause.
     * @param int                               $offset OPTIONAL An SQL OFFSET value.
     * @return Zend_Db_Table_Row_Abstract|null The row results per the
     *     Zend_Db_Adapter fetch mode, or null if no row found.
     */
    public function fetchRow($where = null, $order = null, $offset = null)
    {
        if (!($where instanceof Zend_Db_Table_Select)) {
            $select = $this->select();

            if ($where !== null) {
                $this->_where($select, $where);
            }

            if ($order !== null) {
                $this->_order($select, $order);
            }

            $select->limit(1, ((is_numeric($offset)) ? (int) $offset : null));

        } else {
            $select = $where->limit(1, $where->getPart(Zend_Db_Select::LIMIT_OFFSET));
        }

        $rows = $this->_fetch($select);

        if (count($rows) == 0) {
            return null;
        }

        $data = array(
            'table'   => $this,
            'data'     => $rows[0],
            'readOnly' => $select->isReadOnly(),
            'stored'  => true
        );

        $rowClass = $this->getRowClass();
        if (!class_exists($rowClass)) {
            // require_once 'Zend/Loader.php';
            Zend_Loader::loadClass($rowClass);
        }
        return new $rowClass($data);
    }

    /**
     * Fetches a new blank row (not from the database).
     *
     * @return Zend_Db_Table_Row_Abstract
     * @deprecated since 0.9.3 - use createRow() instead.
     */
    public function fetchNew()
    {
        return $this->createRow();
    }

    /**
     * Fetches a new blank row (not from the database).
     *
     * @param  array $data OPTIONAL data to populate in the new row.
     * @param  string $defaultSource OPTIONAL flag to force default values into new row
     * @return Zend_Db_Table_Row_Abstract
     */
    public function createRow(array $data = array(), $defaultSource = null)
    {
        $cols     = $this->_getCols();
        $defaults = array_combine($cols, array_fill(0, count($cols), null));

        // nothing provided at call-time, take the class value
        if ($defaultSource == null) {
            $defaultSource = $this->_defaultSource;
        }

        if (!in_array($defaultSource, array(self::DEFAULT_CLASS, self::DEFAULT_DB, self::DEFAULT_NONE))) {
            $defaultSource = self::DEFAULT_NONE;
        }

        if ($defaultSource == self::DEFAULT_DB) {
            foreach ($this->_metadata as $metadataName => $metadata) {
                if (($metadata['DEFAULT'] != null) &&
                    ($metadata['NULLABLE'] !== true || ($metadata['NULLABLE'] === true && isset($this->_defaultValues[$metadataName]) && $this->_defaultValues[$metadataName] === true)) &&
                    (!(isset($this->_defaultValues[$metadataName]) && $this->_defaultValues[$metadataName] === false))) {
                    $defaults[$metadataName] = $metadata['DEFAULT'];
                }
            }
        } elseif ($defaultSource == self::DEFAULT_CLASS && $this->_defaultValues) {
            foreach ($this->_defaultValues as $defaultName => $defaultValue) {
                if (array_key_exists($defaultName, $defaults)) {
                    $defaults[$defaultName] = $defaultValue;
                }
            }
        }

        $config = array(
            'table'    => $this,
            'data'     => $defaults,
            'readOnly' => false,
            'stored'   => false
        );

        $rowClass = $this->getRowClass();
        if (!class_exists($rowClass)) {
            // require_once 'Zend/Loader.php';
            Zend_Loader::loadClass($rowClass);
        }
        $row = new $rowClass($config);
        $row->setFromArray($data);
        return $row;
    }

    /**
     * Generate WHERE clause from user-supplied string or array
     *
     * @param  string|array $where  OPTIONAL An SQL WHERE clause.
     * @return Zend_Db_Table_Select
     */
    protected function _where(Zend_Db_Table_Select $select, $where)
    {
        $where = (array) $where;

        foreach ($where as $key => $val) {
            // is $key an int?
            if (is_int($key)) {
                // $val is the full condition
                $select->where($val);
            } else {
                // $key is the condition with placeholder,
                // and $val is quoted into the condition
                $select->where($key, $val);
            }
        }

        return $select;
    }

    /**
     * Generate ORDER clause from user-supplied string or array
     *
     * @param  string|array $order  OPTIONAL An SQL ORDER clause.
     * @return Zend_Db_Table_Select
     */
    protected function _order(Zend_Db_Table_Select $select, $order)
    {
        if (!is_array($order)) {
            $order = array($order);
        }

        foreach ($order as $val) {
            $select->order($val);
        }

        return $select;
    }

    /**
     * Support method for fetching rows.
     *
     * @param  Zend_Db_Table_Select $select  query options.
     * @return array An array containing the row results in FETCH_ASSOC mode.
     */
    protected function _fetch(Zend_Db_Table_Select $select)
    {
        $stmt = $this->_db->query($select);
        $data = $stmt->fetchAll(Zend_Db::FETCH_ASSOC);
        return $data;
    }

    /**
     * Get table gateway object from string
     *
     * @param  string                 $tableName
     * @param  Zend_Db_Table_Abstract $referenceTable
     * @throws Zend_Db_Table_Row_Exception
     * @return Zend_Db_Table_Abstract
     */
    public static function getTableFromString($tableName, Zend_Db_Table_Abstract $referenceTable = null)
    {
        if ($referenceTable instanceof Zend_Db_Table_Abstract) {
            $tableDefinition = $referenceTable->getDefinition();

            if ($tableDefinition !== null && $tableDefinition->hasTableConfig($tableName)) {
                return new Zend_Db_Table($tableName, $tableDefinition);
            }
        }

        // assume the tableName is the class name
        if (!class_exists($tableName)) {
            try {
                // require_once 'Zend/Loader.php';
                Zend_Loader::loadClass($tableName);
            } catch (Zend_Exception $e) {
                // require_once 'Zend/Db/Table/Row/Exception.php';
                throw new Zend_Db_Table_Row_Exception($e->getMessage(), $e->getCode(), $e);
            }
        }

        $options = array();

        if ($referenceTable instanceof Zend_Db_Table_Abstract) {
            $options['db'] = $referenceTable->getAdapter();
        }

        if (isset($tableDefinition) && $tableDefinition !== null) {
            $options[Zend_Db_Table_Abstract::DEFINITION] = $tableDefinition;
        }

        return new $tableName($options);
    }

}

// NewDefaultTable creates a default table object
// Supported params for $config are:
// - db              = user-supplied instance of database connector,
//                     or key name of registry instance.
// - name            = table name.
// - primary         = string or array of primary key(s).
// - rowClass        = row class name.
// - rowsetClass     = rowset class name.
// - referenceMap    = array structure to declare relationship
//                     to parent tables.
// - dependentTables = array of child tables.
// - metadataCache   = cache for information from adapter describeTable()
func NewDefaultTable(options *Config) (Interface, error) {
	t := &Table{}
	if options != nil {
		t.SetOptions(options)
	}

	t.Setup()
	t.Init()
}

// TableReference holds table references 
type TableReference struct {
	Columns Columns
	Table string
	RefColumns Columns
	OnDelete string
	OnUpdate string
}

// TableInfo holds table info 
type TableInfo struct {
	Schema string
    Name string
    Cols []string
    Primary []string
    RowType string
    RowsetType string
    ReferenceMap map[string]TableReference
    DependantTables map[string]TableReference
    Sequence string
}