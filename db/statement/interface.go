package statement

import (
	"database/sql"
	"regexp"
	"strconv"
	"wsf/db/dbselect"
	"wsf/errors"
	"wsf/utils"
	"wsf/utils/stack"
)

const (
	// TYPEDefault is a type id of statement class
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func() (Interface, error){}
)

func init() {
	Register(TYPEDefault, NewDefaultStatement)
}

// Interface is a statement interface
type Interface interface {
	SetAdapter(AdapterInterface) error
	//Prepare(sql dbselect.Interface) error
	BindColumn(column string, param interface{}, typ int) error
	BindParam(parameter string, variable interface{}, typ int, length int, options map[string]interface{}) error
	BindValue(parameter string, value interface{}, typ int) error
	//CloseCursor() error
	//ColumnCount() int
	//ErrorCode() int
	//ErrorInfo() string
	//Execute(params []interface{}) error
	//Fetch(style int, cursor Interface, offset int)
	//FetchAll(style int, col int)

	//Returns a single column from the next row of a result set
	//FetchColumn(col int) interface{}

	//Set a statement attribute
	//SetAttribute(key string, val interface{}) error

	//Retrieve a statement attribute
	//Attribute(key string) error

	//Retrieves the next rowset (result set) for a SQL statement that has
	//multiple result sets.  An example is a stored procedure that returns
	//the results of multiple queries
	//NextRowset()

	//Returns the number of rows affected by the execution of the
	//last INSERT, DELETE, or UPDATE statement executed by this
	//statement object.
	//RowCount() int

	//Set the default fetch mode for this statement
	//SetFetchMode(mode int) error
}

// NewStatement creates a new statement
func NewStatement(statementType string) (Interface, error) {
	if f, ok := buildHandlers[statementType]; ok {
		return f()
	}

	return nil, errors.Errorf("Unrecognized database statement type \"%v\"", statementType)
}

// Register registers a handler for database statement creation
func Register(statementType string, handler func() (Interface, error)) {
	buildHandlers[statementType] = handler
}

// Statement is a default statement class
type Statement struct {
	options    *Config
	stmt       *sql.Stmt
	adapter    AdapterInterface
	fetchMode  int
	attribute  map[string]interface{}
	bindColumn map[string]interface{}
	bindParams *stack.Referenced
	sqlSplit   []string
	sqlParam   []string
	queryID    int
}

// SetAdapter sets statement adapter reference
func (s *Statement) SetAdapter(adp AdapterInterface) error {
	s.adapter = adp
	return nil
}

// BindColumn binds a column of the statement result set to a variable
func (s *Statement) BindColumn(column string, param interface{}, typ int) error {
	s.bindColumn[column] = param
	return nil
}

// BindParam binds a parameter to the specified variable name
func (s *Statement) BindParam(parameter string, variable interface{}, typ int, length int, options map[string]interface{}) error {
	binded := false
	intval, err := strconv.Atoi(parameter)
	if s.adapter.SupportsParameters("positional") && err == nil && intval > 0 {
		if intval >= 1 || intval <= len(s.sqlParam) {
			s.bindParams.Set(parameter, variable)
			binded = true
		}
	} else if s.adapter.SupportsParameters("named") {
		if parameter[0:1] != `:` {
			parameter = `:` + parameter
		}

		if utils.InSSlice(parameter, s.sqlParam) {
			s.bindParams.Set(parameter, variable)
			binded = true
		}
	}

	if !binded {
		return errors.Errorf("Invalid bind-variable position '%v'", parameter)
	}

	//return s.bindParam($position, $variable, $type, $length, $options);*/
	return nil
}

// BindValue binds a value to a parameter
func (s *Statement) BindValue(parameter string, value interface{}, typ int) error {
	return s.BindParam(parameter, value, typ, 0, map[string]interface{}{})
}

// Executes a prepared statement
/*func Execute(array $params = null)
{
	// Simple case - no query profiler to manage
	if ($this->_queryId === null) {
		return $this->_execute($params);
	}

	// Do the same thing, but with query profiler management before and after the execute
	$prof = $this->_adapter->getProfiler();
	$qp = $prof->getQueryProfile($this->_queryId);
	if ($qp->hasEnded()) {
		$this->_queryId = $prof->queryClone($qp);
		$qp = $prof->getQueryProfile($this->_queryId);
	}
	if ($params !== null) {
		$qp->bindParams($params);
	} else {
		$qp->bindParams($this->_bindParam);
	}
	$qp->start($this->_queryId);

	$retval = $this->_execute($params);

	$prof->queryEnd($this->_queryId);

	return $retval;
}*/

// Returns an array containing all of the result set rows
/*func FetchAll($style = null, $col = null)
{
	$data = array();
	if ($style === Zend_Db::FETCH_COLUMN && $col === null) {
		$col = 0;
	}
	if ($col === null) {
		while ($row = $this->fetch($style)) {
			$data[] = $row;
		}
	} else {
		while (false !== ($val = $this->fetchColumn($col))) {
			$data[] = $val;
		}
	}
	return $data;
}*/

// Returns a single column from the next row of a result set
/*func FetchColumn($col = 0)
{
	$data = array();
	$col = (int) $col;
	$row = $this->fetch(Zend_Db::FETCH_NUM);
	if (!is_array($row)) {
		return false;
	}
	return $row[$col];
}*/

// Fetches the next row and returns it as an object
/*func FetchObject($class = 'stdClass', array $config = array())
{
	$obj = new $class($config);
	$row = $this->fetch(Zend_Db::FETCH_ASSOC);
	if (!is_array($row)) {
		return false;
	}
	foreach ($row as $key => $val) {
		$obj->$key = $val;
	}
	return $obj;
}*/

// Retrieve a statement attribute
/*func getAttribute($key)
{
	if (array_key_exists($key, $this->_attribute)) {
		return $this->_attribute[$key];
	}
}*/

// Set a statement attribute
/*func SetAttribute($key, $val)
{
	$this->_attribute[$key] = $val;
}*/

// Set the default fetch mode for this statement
/*func SetFetchMode($mode)
{
	switch ($mode) {
		case Zend_Db::FETCH_NUM:
		case Zend_Db::FETCH_ASSOC:
		case Zend_Db::FETCH_BOTH:
		case Zend_Db::FETCH_OBJ:
			$this->_fetchMode = $mode;
			break;
		case Zend_Db::FETCH_BOUND:
		default:
			$this->closeCursor();
			// require_once 'Zend/Db/Statement/Exception.php';
			throw new Zend_Db_Statement_Exception('invalid fetch mode');
			break;
	}
}*/

// Helper function to map retrieved row to bound column variables
/*func fetchBound($row)
{
	foreach ($row as $key => $value) {
		// bindColumn() takes 1-based integer positions
		// but fetch() returns 0-based integer indexes
		if (is_int($key)) {
			$key++;
		}
		// set results only to variables that were bound previously
		if (isset($this->_bindColumn[$key])) {
			$this->_bindColumn[$key] = $value;
		}
	}
	return true;
}*/

// Gets the Zend_Db_Adapter_Abstract for this particular Zend_Db_Statement object
/*func Adapter()
{
	return $this->_adapter;
}*/

// Gets the resource or object setup by the parse method
/*func DriverStatement()
{
	return $this->_stmt;
}*/

func (s *Statement) parseParameters(sql dbselect.Interface) error {
	quotedsql, err := s.stripQuoted(sql.Assemble())
	if err != nil {
		return err
	}

	// split into text and params
	reg := regexp.MustCompile(`(\?|\:[a-zA-Z0-9_]+)`)
	s.sqlSplit = reg.Split(quotedsql, -1)

	// map params
	s.sqlParam = make([]string, 0)
	for _, val := range s.sqlSplit {
		if val == "?" && !s.options.SupportsParameters("positional") {
			return errors.Errorf("Invalid bind-variable position '%s'", val)
		} else if val[0:1] == ":" && !s.options.SupportsParameters("named") {
			return errors.Errorf("Invalid bind-variable name '%s'", val)
		}

		s.sqlParam = append(s.sqlParam, val)
	}

	// set up for binding
	s.bindParams = stack.NewReferenced()
	return nil
}

func (s *Statement) stripQuoted(sql string) (string, error) {
	// get the character for value quoting. this should be '
	q := s.adapter.Quote("a")
	q = q[0:1]

	// get the value used as an escaped quote, e.g. \' or ''
	qe := s.adapter.Quote(q)
	qe = qe[1:2] //substr($qe, 1, 2);
	qe = regexp.QuoteMeta(qe)
	escapeChar := qe[0:1] //substr($qe,0,1);
	// remove 'foo\'bar'
	if q != "" {
		escapeChar = regexp.QuoteMeta(escapeChar)
		reg := regexp.MustCompile(q + `([^` + q + `{` + escapeChar + `}]*|(` + qe + `)*)*` + q)
		sql = reg.ReplaceAllString(sql, "")
	}

	// get a version of the SQL statement with all quoted
	// values and delimited identifiers stripped out
	// remove "foo\"bar"
	reg := regexp.MustCompile(`\"(\\\\\"|[^\"])*\"`)
	sql = reg.ReplaceAllString(sql, "")

	// get the character for delimited id quotes, this is usually " but in MySQL is `
	d := s.adapter.QuoteIdentifier("a")
	d = d[0:1]

	// get the value used as an escaped delimited id quote, e.g. \" or "" or \`
	de := s.adapter.QuoteIdentifier(d)
	de = de[1:2] //substr($de, 1, 2);
	de = regexp.QuoteMeta(de)

	// Note: de and d where never used..., now they are:
	reg = regexp.MustCompile(d + `(` + de + `|\\\\{2}|[^` + d + `])*` + d)
	sql = reg.ReplaceAllString(sql, "")
	return sql, nil
}

// NewDefaultStatement Creates new default statement object
func NewDefaultStatement() (Interface, error) {
	stmt := &Statement{
		fetchMode:  0,
		attribute:  make(map[string]interface{}),
		bindColumn: make(map[string]interface{}),
		sqlSplit:   make([]string, 0),
		sqlParam:   make([]string, 0),
	}

	return stmt, nil
}

// AdapterInterface represents usable adapter interface
type AdapterInterface interface {
	Quote(string) string
	QuoteIdentifier(string) string
	SupportsParameters(param string) bool
}
