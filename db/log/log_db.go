package log

import (
	"wsf/auth"
	"wsf/config"
	"wsf/context"
	"wsf/db"
	"wsf/utils"
)

// Log is a db log table object
type DBLog struct {
	*db.DefaultTable

	enabled  bool
	advanced bool
	excludes []string
}

// Enable the log
func (t *DBLog) Enable() db.Log {
	t.enabled = true
	return t
}

// Disable the log
func (t *DBLog) Disable() db.Log {
	t.enabled = false
	return t
}

// IsEnabled returns true if log is enabled
func (t *DBLog) IsEnabled() bool {
	return t.enabled
}

// SetAdvanced sets wethere the log is advanced or not
func (t *DBLog) SetAdvanced(value bool) db.Log {
	t.advanced = value
	return t
}

// IsAdvanced returns true if log is advanced
func (t *DBLog) IsAdvanced() bool {
	return t.advanced
}

// SetExcludes sets excluded from log tables
func (t *DBLog) SetExcludes(tables []string) db.Log {
	t.excludes = tables
	return t
}

// AddExclude add a table to excludes
func (t *DBLog) AddExclude(table string) db.Log {
	if table != "" {
		t.excludes = append(t.excludes, table)
	}

	return t
}

// Excludes returns slice of excluded tables
func (t *DBLog) Excludes() []string {
	return t.excludes
}

// IsExcluded returns true if table should be excluded
func (t *DBLog) IsExcluded(table string) bool {
	return utils.InSSlice(table, t.excludes)
}

// IsWritable returns true if table is valid for log
func (t *DBLog) IsWritable(table string) bool {
	if table == t.DefaultTable.Name {
		return false
	}

	if !t.IsEnabled() {
		return false
	}

	if t.IsExcluded(table) {
		return false
	}

	return true
}

// Write log data
func (t *DBLog) Write(ctx context.Context, operation string, table string, id int, data map[string]interface{}) bool {
	if !t.IsWritable(table) {
		return false
	}

	var instanceID int
	var userID int
	//idnt := ctx.Value("session").(string)
	//if authResource := registry.Get("auth"); authResource != nil {
	//	if idnt, err := authResource.(auth.Interface).Identity(idnt); err == nil {
	//		instanceID = idnt.GetInt("instanceID")
	//		userID = idnt.GetInt("id")
	//	}
	//}
	if idnt, ok := auth.IdentityFromContext(ctx); ok {
		instanceID = idnt.GetInt("instanceID")
		userID = idnt.GetInt("id")
	}

	insertData := map[string]interface{}{
		"instanceId": instanceID,
		"object":     table,
		"itemId":     id,
		"userId":     userID,
		"operation":  operation,
	}

	if t.IsAdvanced() {
		insertData["advanced"] = data
	}

	if _, err := t.DefaultTable.Insert(ctx, insertData); err == nil {
		return true
	}

	return false
}

// NewDbLog creates a db log object
func NewDbLog(options config.Config) (db.Log, error) {
	cfg := &db.TableConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	t := &DBLog{
		DefaultTable: db.NewEmptyDefaultTable(cfg),
	}

	t.DefaultTable.Name = "logs"
	t.enabled = options.GetBoolDefault("enabled", false)
	t.advanced = options.GetBoolDefault("advanced", false)
	t.excludes = options.GetStringSlice("excludes")
	t.DefaultTable.Setup()
	t.DefaultTable.Init()
	return t, nil
}
