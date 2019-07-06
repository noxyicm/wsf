package dbselect

import (
	"regexp"
	"strconv"
	"strings"
	"wsf/errors"
	"wsf/utils"
)

// General select constants
const (
	TYPEDefault = "default"

	Dinstinct   = "distinct"
	Columns     = "columns"
	From        = "from"
	Union       = "union"
	Where       = "where"
	Group       = "group"
	Having      = "having"
	Order       = "order"
	LimitCount  = "limitcount"
	LimitOffset = "limitoffset"
	ForUpdate   = "forupdate"

	InnerJoin   = "inner join"
	LeftJoin    = "inleftner join"
	RightJoin   = "right join"
	FullJoin    = "full join"
	CrossJoin   = "cross join"
	NaturalJoin = "natural join"

	SQLWildcard  = "*"
	SQLSelect    = "SELECT"
	SQLUnion     = "UNION"
	SQLUnionAll  = "UNION ALL"
	SQLFrom      = "FROM"
	SQLWhere     = "WHERE"
	SQLDistinct  = "DISTINCT"
	SQLGroupBy   = "GROUP BY"
	SQLOrderBy   = "ORDER BY"
	SQLHaving    = "HAVING"
	SQLForUpdate = "FOR UPDATE"
	SQLAnd       = "AND"
	SQLAs        = "AS"
	SQLOr        = "OR"
	SQLOn        = "ON"
	SQLAsc       = "ASC"
	SQLDesc      = "DESC"
)

var (
	// JoinTypes specify legal join types
	JoinTypes = []string{InnerJoin, LeftJoin, RightJoin, FullJoin, CrossJoin, NaturalJoin}

	// UnionTypes specify legal union types
	UnionTypes = []string{SQLUnion, SQLUnionAll}

	buildHandlers = map[string]func(*Config) (Interface, error){}

	// RegexColumnAs column AS alias spliter
	RegexColumnAs = regexp.MustCompile(`/^(.+)\s+` + SQLAs + `\s+(.+)$/i`)
	// RegexColumnDot s
	RegexColumnDot = regexp.MustCompile(`/(.+)\.(.+)/`)
	// RegexColumnExpr s
	RegexColumnExpr = regexp.MustCompile(`/^([\w]*\s*\(([^\(\)])*\))$/`)
	// RegexColumnExprOrder s
	RegexColumnExprOrder = regexp.MustCompile(`/^([\w]+\s*\(([^\(\)])*\))$/`)
	// RegexColumnExprGroup s
	RegexColumnExprGroup = regexp.MustCompile(`/^([\w]+\s*\(([^\(\)])*\))$/`)
	// RegexSQLComments s
	/*RegexSQLComments = regexp.MustCompile(`@
	    (([\'"]).*?[^\\\]) # $1 : Skip single & double quoted expressions
	    |(                   # $3 : Match comments
	        (?:\#|\-\-).*?$    # - Single line comments
	        |                # - Multi line (nested) comments
	         /\*             #   . comment open marker
	            (?: [^/*]    #   . non comment-marker characters
	                |/(?!\*) #   . ! not a comment open
	                |\*(?!/) #   . ! not a comment close
	            )*           #   . repeat eventually
	        \*\/             #   . comment close marker
	    )\s*                 # Trim after comments
	    |(?<=;)\s+           # Trim after semi-colon
		@msx`)*/
	// RegexSQLComments s
	RegexSQLComments = regexp.MustCompile(`--(.*?)\r?\n|--(.*?)$|('(('')|[^'])*')|\[((\]\])|[^\]])* \]|(\""((\""\"")|[^""])*\"")`)
)

func init() {
	Register(TYPEDefault, NewDefaultSelect)
}

// Interface is a statement interface
type Interface interface {
	SetAdapter(adapter AdapterInterface) error
	From(name string, cols interface{}) *Select
	FromAs(name string, alias string, cols interface{}) *Select
	FromSchema(name string, cols interface{}, schema string) *Select
	FromAsSchema(name string, alias string, cols interface{}, schema string) *Select
	Where(cond string, value interface{}) *Select
	OrWhere(cond string, value interface{}) *Select
	Binds() []interface{}
	Err() error
	Assemble() string
	ToString() string
}

// NewSelect creates a new statement
func NewSelect(selectType string, cfg *Config) (Interface, error) {
	if f, ok := buildHandlers[selectType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database select type \"%v\"", selectType)
}

// Register registers a handler for database statement creation
func Register(selectType string, handler func(*Config) (Interface, error)) {
	buildHandlers[selectType] = handler
}

// AdapterInterface represents usable adapter interface
type AdapterInterface interface {
	Quote(interface{}) string
	QuoteIdentifier(interface{}, bool) string
	QuoteIdentifierAs(ident interface{}, alias string, auto bool) string
	QuoteIdentifierSymbol() string
	QuoteInto(text string, value interface{}, count int) string
	QuoteColumnAs(ident interface{}, alias string, auto bool) string
	QuoteTableAs(ident interface{}, alias string, auto bool) string
	Limit(sql string, count int, offset int) string
	SupportsParameters(param string) bool
}

// Select is a db select class
type Select struct {
	options *Config
	adapter AdapterInterface
	bind    map[string]interface{}
	parts   *selectParts
	errors  []error
}

// SelectParts is a select object parts holder
type selectParts struct {
	Dinstinct   bool
	Columns     []*selectColumn
	Union       []*selectUnion
	From        map[string]*selectFrom
	Where       []string
	Group       []interface{}
	Having      []interface{}
	Order       []interface{}
	LimitCount  interface{}
	LimitOffset interface{}
	ForUpdate   bool
}

// SelectColumn is a select object column representation
type selectColumn struct {
	Table  string
	Column interface{}
	Alias  string
}

// SelectFrom is a select object from representation
type selectFrom struct {
	JoinType      string
	Schema        string
	TableName     string
	JoinCondition string
}

// SelectUnion is a select object union representation
type selectUnion struct {
	Target string
	Type   string
}

// SelectWhere is a select object where representation
type selectWhere struct {
	Target string
	Type   string
}

// SetAdapter sets the adapter interface to select object
func (s *Select) SetAdapter(adapter AdapterInterface) error {
	s.adapter = adapter
	return nil
}

// Distinct makes the query SELECT DISTINCT
func (s *Select) Distinct(flag bool) *Select {
	s.parts.Dinstinct = flag
	return s
}

// From adds a FROM table and optional columns to the query
func (s *Select) From(name string, cols interface{}) *Select {
	return s.prepareJoin(From, name, "", "", cols, "")
}

// FromAs adds a FROM table and optional columns to the query
func (s *Select) FromAs(name string, alias string, cols interface{}) *Select {
	return s.prepareJoin(From, name, alias, "", cols, "")
}

// FromSchema adds a FROM table and optional columns to the query with specific schema
func (s *Select) FromSchema(name string, cols interface{}, schema string) *Select {
	return s.prepareJoin(From, name, "", "", cols, schema)
}

// FromAsSchema adds a FROM table and optional columns to the query with specific schema
func (s *Select) FromAsSchema(name string, alias string, cols interface{}, schema string) *Select {
	return s.prepareJoin(From, name, alias, "", cols, schema)
}

// Columns specifies the columns used in the FROM clause
func (s *Select) Columns(cols interface{}, correlationName string) *Select {
	if correlationName == "" && len(s.parts.From) > 0 {
		for key := range s.parts.From {
			correlationName = key
			break
		}
	}

	if _, ok := s.parts.From[correlationName]; !ok {
		s.errors = append(s.errors, errors.New("No table has been specified for the FROM clause"))
		return s
	}

	s.tableCols(correlationName, cols, "")
	return s
}

// Union adds a UNION clause to the query
func (s *Select) Union(sql interface{}, typ string) *Select {
	if !utils.InSSlice(typ, UnionTypes) {
		s.errors = append(s.errors, errors.Errorf("Invalid union type '%s'", typ))
		return s
	}

	switch t := sql.(type) {
	case []Interface:
		for _, target := range sql.([]Interface) {
			s.parts.Union = append(s.parts.Union, &selectUnion{Target: target.Assemble(), Type: typ})
		}

	case Interface:
		s.parts.Union = append(s.parts.Union, &selectUnion{Target: sql.(Interface).Assemble(), Type: typ})

	case []string:
		for _, target := range sql.([]string) {
			s.parts.Union = append(s.parts.Union, &selectUnion{Target: target, Type: typ})
		}

	case string:
		s.parts.Union = append(s.parts.Union, &selectUnion{Target: sql.(string), Type: typ})

	default:
		s.errors = append(s.errors, errors.Errorf("Unsupported sql type '%t'", t))
		return s
	}

	return s
}

// Where adds a WHERE condition to the query by AND
func (s *Select) Where(cond string, value interface{}) *Select {
	s.parts.Where = append(s.parts.Where, s.where(cond, value, "", true))
	return s
}

// OrWhere adds a WHERE condition to the query by OR
func (s *Select) OrWhere(cond string, value interface{}) *Select {
	s.parts.Where = append(s.parts.Where, s.where(cond, value, "", false))
	return s
}

// Limit sets a limit count and offset to the query
func (s *Select) Limit(count int, offset int) *Select {
	s.parts.LimitCount = count
	s.parts.LimitOffset = offset
	return s
}

// Binds returns binds
func (s *Select) Binds() []interface{} {
	binds := make([]interface{}, len(s.bind))
	i := 0
	for _, value := range s.bind {
		binds[i] = value
		i++
	}

	return binds
}

// Err pops last acuired error or nil if no errors
func (s *Select) Err() error {
	if len(s.errors) > 0 {
		err := s.errors[0]
		s.errors = append([]error{}, s.errors[1:]...)
		return err
	}

	return nil
}

// ToString converts select struct to string
func (s *Select) ToString() string {
	return s.Assemble()
}

// Assemble converts this object to an SQL SELECT string
func (s *Select) Assemble() string {
	sql := SQLSelect
	sql = s.renderDistinct(sql)
	sql = s.renderColumns(sql)
	//sql = s.renderUnion(sql)
	sql = s.renderFrom(sql)
	sql = s.renderWhere(sql)
	//sql = s.renderGroup(sql)
	//sql = s.renderHaving(sql)
	//sql = s.renderOrder(sql)
	sql = s.renderLimit(sql)
	//sql = s.renderLimitOffset(sql)
	//sql = s.renderForupdate(sql)

	return sql
}

func (s *Select) prepareJoin(typ string, name interface{}, alias string, cond string, cols interface{}, schema string) *Select {
	if !utils.InSSlice(typ, JoinTypes) && typ != From {
		s.errors = append(s.errors, errors.Errorf("Invalid join type '%s'", typ))
		return s
	}

	if len(s.parts.Union) > 0 {
		s.errors = append(s.errors, errors.Errorf("Invalid use of table with %s", Union))
		return s
	}

	correlationName := alias
	tableName := ""
	switch t := name.(type) {
	case map[string]string:
		for tmpCorrelationName, tmpTableName := range name.(map[string]string) {
			tableName = tmpTableName
			correlationName = tmpCorrelationName
			break
		}

	case []string:
		for _, tmpTableName := range name.([]string) {
			tableName = tmpTableName
			correlationName = s.uniqueCorrelation(tableName)
			break
		}

	case *Expr:
		tableName = name.(*Expr).Assemble()
		if alias == "" {
			correlationName = s.uniqueCorrelation("t")
		}

	case Interface:
		tableName = name.(Interface).Assemble()
		if alias == "" {
			correlationName = s.uniqueCorrelation("t")
		}

	case string:
		if m := RegexColumnAs.FindAllString(name.(string), -1); len(m) > 0 {
			tableName = m[1]
			correlationName = m[2]
		} else {
			tableName = name.(string)
			if alias == "" {
				correlationName = s.uniqueCorrelation(tableName)
			}
		}

	default:
		s.errors = append(s.errors, errors.Errorf("Unsupported join type '%t'", t))
		return s
	}

	// Schema from table name overrides schema argument
	if strings.Index(tableName, ".") > 0 {
		parts := strings.Split(tableName, ".")
		schema, tableName = parts[0], parts[1]
	}

	lastFromCorrelationName := ""
	if correlationName != "" {
		if _, ok := s.parts.From[correlationName]; ok {
			s.errors = append(s.errors, errors.Errorf("You cannot define a correlation name '%s' more than once", correlationName))
			return s
		}

		currentCorrelationName := ""
		tmpFromParts := make(map[string]*selectFrom)
		if typ == From {
			// append this from after the last from joinType
			tmpFromParts := s.parts.From
			s.parts.From = make(map[string]*selectFrom)
			// move all the froms onto the stack
			for key, part := range tmpFromParts {
				currentCorrelationName = key
				if part.JoinType != From {
					break
				}

				lastFromCorrelationName = currentCorrelationName
				s.parts.From[currentCorrelationName] = part
				delete(tmpFromParts, currentCorrelationName)
			}
		}

		s.parts.From[correlationName] = &selectFrom{
			JoinType:      typ,
			Schema:        schema,
			TableName:     tableName,
			JoinCondition: cond,
		}

		for key, part := range tmpFromParts {
			currentCorrelationName = key
			s.parts.From[currentCorrelationName] = part
		}
	}

	// add to the columns from this joined table
	return s.tableCols(correlationName, cols, lastFromCorrelationName)
}

func (s *Select) where(condition string, value interface{}, typ string, b bool) string {
	if len(s.parts.Union) > 0 {
		s.errors = append(s.errors, errors.Errorf("Invalid use of where clause with %s", Union))
		return ""
	}

	condition = s.adapter.QuoteIdentifier(condition, true)
	if value != nil {
		condition = s.adapter.QuoteInto(condition, value, -1)
	}

	cond := ""
	if len(s.parts.Where) > 0 {
		if b {
			cond = SQLAnd + " "
		} else {
			cond = SQLOr + " "
		}
	}

	return cond + "(" + condition + ")"
}

// Adds to the internal table-to-column mapping array
func (s *Select) tableCols(correlationName string, cols interface{}, afterCorrelationName string) *Select {
	columns := []interface{}{}
	switch cols.(type) {
	case string:
		columns = append(columns, cols.(string))

	case []*Expr:
		s := make([]interface{}, len(cols.([]*Expr)))
		i := 0
		for _, col := range cols.([]*Expr) {
			s[i] = col
			i++
		}
		columns = append(columns, s...)

	case []string:
		s := make([]interface{}, len(cols.([]string)))
		i := 0
		for _, col := range cols.([]string) {
			s[i] = col
			i++
		}
		columns = append(columns, s...)

	default:
		s.errors = append(s.errors, errors.New("Invalid column type"))
		return s
	}

	columnValues := []*selectColumn{}
	for _, col := range columns {
		if col == nil {
			continue
		}

		currentCorrelationName := correlationName
		column := ""
		alias := ""
		switch col.(type) {
		case string:
			// Check for a column matching "<column> AS <alias>" and extract the alias name
			column = strings.Trim(strings.ReplaceAll(col.(string), "\n", " "), "")
			if m := RegexColumnAs.FindAllString(column, -1); len(m) > 0 {
				column = m[1]
				alias = m[2]
			}

			// Check for columns that look like functions and convert to dbselect.Expr
			if RegexColumnExpr.MatchString(column) {
				column = NewExpr(column).ToString()
			} else if m := RegexColumnDot.FindAllString(column, -1); len(m) > 0 {
				currentCorrelationName = m[1]
				column = m[2]
			}
		}

		columnValues = append(columnValues, &selectColumn{Table: currentCorrelationName, Column: column, Alias: alias})
	}

	if len(columnValues) > 0 {
		tmpColumns := []*selectColumn{}
		index := 0
		// should we attempt to prepend or insert these values?
		if afterCorrelationName != "" {
			tmpColumns = s.parts.Columns
			s.parts.Columns = []*selectColumn{}

			for key, currentColumn := range tmpColumns {
				if currentColumn.Alias == afterCorrelationName {
					break
				} else {
					s.parts.Columns = append(s.parts.Columns, currentColumn)
					index = key
				}
			}
		}

		// apply current values to current stack
		for _, columnValue := range columnValues {
			s.parts.Columns = append(s.parts.Columns, columnValue)
		}

		// finish ensuring that all previous values are applied (if they exist)
		for i := index; i < len(tmpColumns); i++ {
			s.parts.Columns = append(s.parts.Columns, tmpColumns[i])
		}
	}

	return s
}

// Generate a unique correlation name
func (s *Select) uniqueCorrelation(name string) string {
	// Extract just the last name of a qualified table name
	dot := strings.LastIndex(name, ".")
	c := name
	if dot > -1 {
		c = name[dot+1 : len(name)]
	}

	i := 2
	for {
		if _, ok := s.parts.From[c]; !ok {
			break
		}

		c = name + "_" + strconv.Itoa(i)
		i++
	}

	return c
}

// Render DISTINCT clause
func (s *Select) renderDistinct(sql string) string {
	if s.parts.Dinstinct {
		sql = sql + " " + SQLDistinct
	}

	return sql
}

// Render Columns
func (s *Select) renderColumns(sql string) string {
	if len(s.parts.Columns) == 0 {
		return sql
	}

	columns := make([]string, 0)
	for _, columnEntry := range s.parts.Columns {
		correlationName := columnEntry.Table
		column := columnEntry.Column
		alias := columnEntry.Alias

		switch column.(type) {
		case Expr, Interface:
			columns = append(columns, s.adapter.QuoteColumnAs(column, alias, true))

		case string:
			if column == SQLWildcard {
				column = NewExpr(SQLWildcard).ToString()
				alias = ""
			}

			if correlationName == "" {
				columns = append(columns, s.adapter.QuoteColumnAs(column.(string), alias, true))
			} else {
				columns = append(columns, s.adapter.QuoteColumnAs(correlationName+"."+column.(string), alias, true))
			}
		}
	}

	return sql + " " + strings.Join(columns, ", ")
}

// Render FROM clause
func (s *Select) renderFrom(sql string) string {
	// If no table specified, use RDBMS-dependent solution
	// for table-less query.  e.g. DUAL in Oracle.
	if len(s.parts.From) == 0 {
		s.parts.From = s.dummyTable()
	}

	from := make([]string, 0)
	for correlationName, table := range s.parts.From {
		tmp := ""
		joinType := table.JoinType
		if table.JoinType == From {
			joinType = InnerJoin
		}

		// Add join clause (if applicable)
		if len(from) > 0 {
			tmp = tmp + " " + strings.ToUpper(joinType) + " "
		}

		tmp = tmp + s.quotedSchema(table.Schema)
		if table.TableName == correlationName {
			tmp = tmp + s.quotedTable(table.TableName, "")
		} else {
			tmp = tmp + s.quotedTable(table.TableName, correlationName)
		}

		// Add join conditions (if applicable)
		if len(from) > 0 && table.JoinCondition != "" {
			tmp = tmp + " " + SQLOn + " " + table.JoinCondition
		}

		// Add the table name and condition add to the list
		from = append(from, tmp)
	}

	// Add the list of all joins
	if len(from) > 0 {
		sql = sql + " " + SQLFrom + " " + strings.Join(from, "\n")
	}

	return sql
}

// Render WHERE clause
func (s *Select) renderWhere(sql string) string {
	if len(s.parts.From) > 0 && len(s.parts.Where) > 0 {
		sql = sql + " " + SQLWhere + " " + strings.Join(s.parts.Where, " ")
	}

	return sql
}

// Render LIMIT clause
func (s *Select) renderLimit(sql string) string {
	count := 0
	offset := 0

	if s.parts.LimitOffset != nil && s.parts.LimitOffset.(int) > 0 {
		offset = s.parts.LimitOffset.(int)
		count = 100000000000000
	}

	if s.parts.LimitCount != nil && s.parts.LimitCount.(int) > 0 {
		count = s.parts.LimitCount.(int)
	}

	if count > 0 {
		sql = strings.TrimSpace(s.adapter.Limit(sql, count, offset))
	}

	return sql
}

// Render FOR UPDATE clause
func (s *Select) renderForupdate(sql string) string {
	if s.parts.ForUpdate {
		sql = sql + " " + SQLForUpdate
	}

	return sql
}

func (s *Select) dummyTable() map[string]*selectFrom {
	return make(map[string]*selectFrom)
}

// Return a quoted schema name
func (s *Select) quotedSchema(schema string) string {
	if schema == "" {
		return ""
	}
	return s.adapter.QuoteIdentifier(schema, true) + "."
}

// Return a quoted table name
func (s *Select) quotedTable(tableName string, correlationName string) string {
	return s.adapter.QuoteTableAs(tableName, correlationName, true)
}

// NewDefaultSelect creates a new select object
func NewDefaultSelect(options *Config) (Interface, error) {
	return &Select{
		options: options,
		bind:    make(map[string]interface{}),
		parts: &selectParts{
			Dinstinct:   false,
			Columns:     []*selectColumn{},
			Union:       []*selectUnion{},
			From:        map[string]*selectFrom{},
			Where:       []string{},
			Group:       []interface{}{},
			Having:      []interface{}{},
			Order:       []interface{}{},
			LimitCount:  nil,
			LimitOffset: nil,
			ForUpdate:   false,
		},
		errors: make([]error, 0),
	}, nil
}

// New creates a new default select object
func New() (Interface, error) {
	options := &Config{}
	options.Defaults()

	return &Select{
		options: options,
		bind:    make(map[string]interface{}),
		parts: &selectParts{
			Dinstinct:   false,
			Columns:     []*selectColumn{},
			Union:       []*selectUnion{},
			From:        map[string]*selectFrom{},
			Where:       []string{},
			Group:       []interface{}{},
			Having:      []interface{}{},
			Order:       []interface{}{},
			LimitCount:  nil,
			LimitOffset: nil,
			ForUpdate:   false,
		},
		errors: make([]error, 0),
	}, nil
}
