package db

import (
	"regexp"
	"strconv"
	"strings"
	"wsf/config"
	"wsf/errors"
	"wsf/utils"
)

// General select constants
const (
	TYPEDefaultSelect = "default"

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
	SQLBetween   = "BETWEEN"
)

var (
	// JoinTypes specify legal join types
	JoinTypes = []string{InnerJoin, LeftJoin, RightJoin, FullJoin, CrossJoin, NaturalJoin}

	// UnionTypes specify legal union types
	UnionTypes = []string{SQLUnion, SQLUnionAll}

	buildSelectHandlers = map[string]func(*SelectConfig) (Select, error){}

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
	RegisterSelect(TYPEDefaultSelect, NewDefaultSelect)
}

// Select is a select interface
type Select interface {
	SetAdapter(adapter Adapter) error
	From(name string, cols interface{}) Select
	FromAs(name string, alias string, cols interface{}) Select
	FromSchema(name string, cols interface{}, schema string) Select
	FromAsSchema(name string, alias string, cols interface{}, schema string) Select
	Where(cond string, value interface{}) Select
	OrWhere(cond string, value interface{}) Select
	Limit(count int, offset int) Select
	Order(order string) Select
	Binds() []interface{}
	Err() error
	Assemble() string
	ToString() string
}

// NewSelect creates a new statement
func NewSelect(selectType string, options config.Config) (Select, error) {
	cfg := &SelectConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildSelectHandlers[selectType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database select type \"%v\"", selectType)
}

// NewSelectFromConfig creates a new sql select from config
func NewSelectFromConfig(cfg *SelectConfig) (Select, error) {
	if f, ok := buildSelectHandlers[cfg.Type]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized database select type \"%v\"", cfg.Type)
}

// RegisterSelect registers a handler for database statement creation
func RegisterSelect(selectType string, handler func(*SelectConfig) (Select, error)) {
	buildSelectHandlers[selectType] = handler
}

// DefaultSelect is a db select class
type DefaultSelect struct {
	Options *SelectConfig
	Adapter Adapter
	Bind    map[string]interface{}
	Parts   *selectParts
	Errors  []error
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
func (s *DefaultSelect) SetAdapter(adapter Adapter) error {
	s.Adapter = adapter
	return nil
}

// Distinct makes the query SELECT DISTINCT
func (s *DefaultSelect) Distinct(flag bool) Select {
	s.Parts.Dinstinct = flag
	return s
}

// From adds a FROM table and optional columns to the query
func (s *DefaultSelect) From(name string, cols interface{}) Select {
	return s.prepareJoin(From, name, "", "", cols, "")
}

// FromAs adds a FROM table and optional columns to the query
func (s *DefaultSelect) FromAs(name string, alias string, cols interface{}) Select {
	return s.prepareJoin(From, name, alias, "", cols, "")
}

// FromSchema adds a FROM table and optional columns to the query with specific schema
func (s *DefaultSelect) FromSchema(name string, cols interface{}, schema string) Select {
	return s.prepareJoin(From, name, "", "", cols, schema)
}

// FromAsSchema adds a FROM table and optional columns to the query with specific schema
func (s *DefaultSelect) FromAsSchema(name string, alias string, cols interface{}, schema string) Select {
	return s.prepareJoin(From, name, alias, "", cols, schema)
}

// Columns specifies the columns used in the FROM clause
func (s *DefaultSelect) Columns(cols interface{}, correlationName string) Select {
	if correlationName == "" && len(s.Parts.From) > 0 {
		for key := range s.Parts.From {
			correlationName = key
			break
		}
	}

	if _, ok := s.Parts.From[correlationName]; !ok {
		s.Errors = append(s.Errors, errors.New("No table has been specified for the FROM clause"))
		return s
	}

	s.tableCols(correlationName, cols, "")
	return s
}

// Union adds a UNION clause to the query
func (s *DefaultSelect) Union(sql interface{}, typ string) Select {
	if !utils.InSSlice(typ, UnionTypes) {
		s.Errors = append(s.Errors, errors.Errorf("Invalid union type '%s'", typ))
		return s
	}

	switch t := sql.(type) {
	case []Select:
		for _, target := range sql.([]Select) {
			s.Parts.Union = append(s.Parts.Union, &selectUnion{Target: target.Assemble(), Type: typ})
		}

	case Select:
		s.Parts.Union = append(s.Parts.Union, &selectUnion{Target: sql.(Select).Assemble(), Type: typ})

	case []string:
		for _, target := range sql.([]string) {
			s.Parts.Union = append(s.Parts.Union, &selectUnion{Target: target, Type: typ})
		}

	case string:
		s.Parts.Union = append(s.Parts.Union, &selectUnion{Target: sql.(string), Type: typ})

	default:
		s.Errors = append(s.Errors, errors.Errorf("Unsupported sql type '%t'", t))
		return s
	}

	return s
}

// Where adds a WHERE condition to the query by AND
func (s *DefaultSelect) Where(cond string, value interface{}) Select {
	s.Parts.Where = append(s.Parts.Where, s.where(cond, value, "", true))
	return s
}

// OrWhere adds a WHERE condition to the query by OR
func (s *DefaultSelect) OrWhere(cond string, value interface{}) Select {
	s.Parts.Where = append(s.Parts.Where, s.where(cond, value, "", false))
	return s
}

// Limit sets a limit count and offset to the query
func (s *DefaultSelect) Limit(count int, offset int) Select {
	s.Parts.LimitCount = count
	s.Parts.LimitOffset = offset
	return s
}

// Order sets order to the query
func (s *DefaultSelect) Order(order string) Select {
	return s
}

// Binds returns binds
func (s *DefaultSelect) Binds() []interface{} {
	binds := make([]interface{}, len(s.Bind))
	i := 0
	for _, value := range s.Bind {
		binds[i] = value
		i++
	}

	return binds
}

// Err pops last acuired error or nil if no errors
func (s *DefaultSelect) Err() error {
	if len(s.Errors) > 0 {
		err := s.Errors[0]
		s.Errors = append([]error{}, s.Errors[1:]...)
		return err
	}

	return nil
}

// ToString converts select struct to string
func (s *DefaultSelect) ToString() string {
	return s.Assemble()
}

// Assemble converts this object to an SQL SELECT string
func (s *DefaultSelect) Assemble() string {
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

func (s *DefaultSelect) prepareJoin(typ string, name interface{}, alias string, cond string, cols interface{}, schema string) Select {
	if !utils.InSSlice(typ, JoinTypes) && typ != From {
		s.Errors = append(s.Errors, errors.Errorf("Invalid join type '%s'", typ))
		return s
	}

	if len(s.Parts.Union) > 0 {
		s.Errors = append(s.Errors, errors.Errorf("Invalid use of table with %s", Union))
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

	case *SQLExpr:
		tableName = name.(*SQLExpr).Assemble()
		if alias == "" {
			correlationName = s.uniqueCorrelation("t")
		}

	case Select:
		tableName = name.(Select).Assemble()
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
		s.Errors = append(s.Errors, errors.Errorf("Unsupported join type '%t'", t))
		return s
	}

	// Schema from table name overrides schema argument
	if strings.Index(tableName, ".") > 0 {
		parts := strings.Split(tableName, ".")
		schema, tableName = parts[0], parts[1]
	}

	lastFromCorrelationName := ""
	if correlationName != "" {
		if _, ok := s.Parts.From[correlationName]; ok {
			s.Errors = append(s.Errors, errors.Errorf("You cannot define a correlation name '%s' more than once", correlationName))
			return s
		}

		currentCorrelationName := ""
		tmpFromParts := make(map[string]*selectFrom)
		if typ == From {
			// append this from after the last from joinType
			tmpFromParts := s.Parts.From
			s.Parts.From = make(map[string]*selectFrom)
			// move all the froms onto the stack
			for key, part := range tmpFromParts {
				currentCorrelationName = key
				if part.JoinType != From {
					break
				}

				lastFromCorrelationName = currentCorrelationName
				s.Parts.From[currentCorrelationName] = part
				delete(tmpFromParts, currentCorrelationName)
			}
		}

		s.Parts.From[correlationName] = &selectFrom{
			JoinType:      typ,
			Schema:        schema,
			TableName:     tableName,
			JoinCondition: cond,
		}

		for key, part := range tmpFromParts {
			currentCorrelationName = key
			s.Parts.From[currentCorrelationName] = part
		}
	}

	// add to the columns from this joined table
	return s.tableCols(correlationName, cols, lastFromCorrelationName)
}

func (s *DefaultSelect) where(condition string, value interface{}, typ string, b bool) string {
	if len(s.Parts.Union) > 0 {
		s.Errors = append(s.Errors, errors.Errorf("Invalid use of where clause with %s", Union))
		return ""
	}

	condition = s.Adapter.QuoteIdentifier(condition, true)
	if value != nil {
		condition = s.Adapter.QuoteInto(condition, value, -1)
	}

	cond := ""
	if len(s.Parts.Where) > 0 {
		if b {
			cond = SQLAnd + " "
		} else {
			cond = SQLOr + " "
		}
	}

	return cond + "(" + condition + ")"
}

// Adds to the internal table-to-column mapping array
func (s *DefaultSelect) tableCols(correlationName string, cols interface{}, afterCorrelationName string) Select {
	columns := []interface{}{}
	switch cols.(type) {
	case string:
		columns = append(columns, cols.(string))

	case []*SQLExpr:
		s := make([]interface{}, len(cols.([]*SQLExpr)))
		i := 0
		for _, col := range cols.([]*SQLExpr) {
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
		s.Errors = append(s.Errors, errors.New("Invalid column type"))
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
			tmpColumns = s.Parts.Columns
			s.Parts.Columns = []*selectColumn{}

			for key, currentColumn := range tmpColumns {
				if currentColumn.Alias == afterCorrelationName {
					break
				} else {
					s.Parts.Columns = append(s.Parts.Columns, currentColumn)
					index = key
				}
			}
		}

		// apply current values to current stack
		for _, columnValue := range columnValues {
			s.Parts.Columns = append(s.Parts.Columns, columnValue)
		}

		// finish ensuring that all previous values are applied (if they exist)
		for i := index; i < len(tmpColumns); i++ {
			s.Parts.Columns = append(s.Parts.Columns, tmpColumns[i])
		}
	}

	return s
}

// Generate a unique correlation name
func (s *DefaultSelect) uniqueCorrelation(name string) string {
	// Extract just the last name of a qualified table name
	dot := strings.LastIndex(name, ".")
	c := name
	if dot > -1 {
		c = name[dot+1 : len(name)]
	}

	i := 2
	for {
		if _, ok := s.Parts.From[c]; !ok {
			break
		}

		c = name + "_" + strconv.Itoa(i)
		i++
	}

	return c
}

// Render DISTINCT clause
func (s *DefaultSelect) renderDistinct(sql string) string {
	if s.Parts.Dinstinct {
		sql = sql + " " + SQLDistinct
	}

	return sql
}

// Render Columns
func (s *DefaultSelect) renderColumns(sql string) string {
	if len(s.Parts.Columns) == 0 {
		return sql
	}

	columns := make([]string, 0)
	for _, columnEntry := range s.Parts.Columns {
		correlationName := columnEntry.Table
		column := columnEntry.Column
		alias := columnEntry.Alias

		switch column.(type) {
		case *SQLExpr, Select:
			columns = append(columns, s.Adapter.QuoteColumnAs(column, alias, true))

		case string:
			if column == SQLWildcard {
				column = NewExpr(SQLWildcard).ToString()
				alias = ""
			}

			if correlationName == "" {
				columns = append(columns, s.Adapter.QuoteColumnAs(column.(string), alias, true))
			} else {
				columns = append(columns, s.Adapter.QuoteColumnAs(correlationName+"."+column.(string), alias, true))
			}
		}
	}

	return sql + " " + strings.Join(columns, ", ")
}

// Render FROM clause
func (s *DefaultSelect) renderFrom(sql string) string {
	// If no table specified, use RDBMS-dependent solution
	// for table-less query.  e.g. DUAL in Oracle.
	if len(s.Parts.From) == 0 {
		s.Parts.From = s.dummyTable()
	}

	from := make([]string, 0)
	for correlationName, table := range s.Parts.From {
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
func (s *DefaultSelect) renderWhere(sql string) string {
	if len(s.Parts.From) > 0 && len(s.Parts.Where) > 0 {
		sql = sql + " " + SQLWhere + " " + strings.Join(s.Parts.Where, " ")
	}

	return sql
}

// Render LIMIT clause
func (s *DefaultSelect) renderLimit(sql string) string {
	count := 0
	offset := 0

	if s.Parts.LimitOffset != nil && s.Parts.LimitOffset.(int) > 0 {
		offset = s.Parts.LimitOffset.(int)
		count = int(^uint(0) >> 1)
	}

	if s.Parts.LimitCount != nil && s.Parts.LimitCount.(int) > 0 {
		count = s.Parts.LimitCount.(int)
	}

	if count > 0 {
		sql = strings.TrimSpace(s.Adapter.Limit(sql, count, offset))
	}

	return sql
}

// Render FOR UPDATE clause
func (s *DefaultSelect) renderForupdate(sql string) string {
	if s.Parts.ForUpdate {
		sql = sql + " " + SQLForUpdate
	}

	return sql
}

func (s *DefaultSelect) dummyTable() map[string]*selectFrom {
	return make(map[string]*selectFrom)
}

// Return a quoted schema name
func (s *DefaultSelect) quotedSchema(schema string) string {
	if schema == "" {
		return ""
	}
	return s.Adapter.QuoteIdentifier(schema, true) + "."
}

// Return a quoted table name
func (s *DefaultSelect) quotedTable(tableName string, correlationName string) string {
	return s.Adapter.QuoteTableAs(tableName, correlationName, true)
}

// NewDefaultSelect creates a new select object
func NewDefaultSelect(options *SelectConfig) (Select, error) {
	return &DefaultSelect{
		Options: options,
		Bind:    make(map[string]interface{}),
		Parts: &selectParts{
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
		Errors: make([]error, 0),
	}, nil
}

// NewSelectEmpty creates a new default select object
func NewSelectEmpty() (Select, error) {
	options := &SelectConfig{}
	options.Defaults()

	return &DefaultSelect{
		Options: options,
		Bind:    make(map[string]interface{}),
		Parts: &selectParts{
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
		Errors: make([]error, 0),
	}, nil
}
