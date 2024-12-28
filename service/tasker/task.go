package tasker

import (
	"time"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

// Task represents task
type Task struct {
	ID           int64
	State        int
	Name         string
	Date         time.Time
	Created      time.Time
	Year         int
	Month        int
	Day          int
	Hour         int
	Minute       int
	Second       int
	Intervaled   int
	Uniq         int
	Data         string
	ParsedData   map[string]interface{}
	Extras       string
	ParsedExtras map[string]interface{}
	Handler      string
	Worker       string
	Dataset      string
	LastID       int
	LastUpdate   time.Time
	Datahash     string
}

// Interval returns a task interval
func (t *Task) Interval() int {
	return t.Second + (60 * t.Minute) + (3600 * t.Hour) + (86400 * t.Day)
}

// ExecutionTime return a task planed execution time
func (t *Task) ExecutionTime() time.Time {
	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()

	if t.Year != 0 {
		year = t.Year
	}

	switch t.Month {
	case 1:
		month = time.January

	case 2:
		month = time.February

	case 3:
		month = time.March

	case 4:
		month = time.April

	case 5:
		month = time.May

	case 6:
		month = time.June

	case 7:
		month = time.July

	case 8:
		month = time.August

	case 9:
		month = time.September

	case 10:
		month = time.October

	case 11:
		month = time.November

	case 12:
		month = time.December
	}

	if t.Day != 0 {
		day = t.Day
	}

	return time.Date(year, month, day, t.Hour, t.Minute, t.Second, 0, time.Local)
}

// Clone the task
func (t *Task) Clone() Task {
	return Task{
		ID:           t.ID,
		State:        t.State,
		Name:         t.Name,
		Date:         t.Date,
		Created:      time.Now(),
		Year:         t.Year,
		Month:        t.Month,
		Day:          t.Day,
		Hour:         t.Hour,
		Minute:       t.Minute,
		Second:       t.Second,
		Intervaled:   t.Intervaled,
		Uniq:         t.Uniq,
		Data:         t.Data,
		ParsedData:   utils.MapSCopy(t.ParsedData),
		Extras:       t.Extras,
		ParsedExtras: utils.MapSCopy(t.ParsedExtras),
		Handler:      t.Handler,
		Worker:       t.Worker,
		LastUpdate:   time.Time{},
	}
}

// NewTask create a new task
func NewTask(id int64, name, handler, worker string, data map[string]interface{}) *Task {
	return &Task{
		ID:           id,
		State:        1,
		Name:         name,
		Date:         time.Time{},
		Created:      time.Now(),
		ParsedData:   data,
		ParsedExtras: make(map[string]interface{}),
		Handler:      handler,
		Worker:       worker,
		LastUpdate:   time.Time{},
	}
}

// NewTaskFromConfig create a new task from config
func NewTaskFromConfig(cfg config.Config) (*Task, error) {
	hndlr := cfg.GetString("handler")
	wrk := cfg.GetString("worker")
	if hndlr == "" {
		return nil, errors.New("Unable to create task. Handler is not defined")
	} else if wrk == "" {
		return nil, errors.New("Unable to create task. Worker is not defined")
	}

	return &Task{
		ID:           cfg.GetInt64Default("id", time.Now().UnixNano()),
		State:        1,
		Name:         cfg.GetStringDefault("name", "unnamed task"),
		Date:         cfg.GetTimeDefault("date", time.Time{}),
		Created:      time.Now(),
		Year:         cfg.GetIntDefault("year", 0),
		Month:        cfg.GetIntDefault("month", 0),
		Day:          cfg.GetIntDefault("day", 0),
		Hour:         cfg.GetIntDefault("hour", 0),
		Minute:       cfg.GetIntDefault("minute", 0),
		Second:       cfg.GetIntDefault("second", 0),
		Intervaled:   cfg.GetIntDefault("intervaled", 0),
		Uniq:         cfg.GetIntDefault("uniq", 0),
		ParsedData:   cfg.GetStringMap("data"),
		ParsedExtras: make(map[string]interface{}),
		Handler:      hndlr,
		Worker:       wrk,
		LastUpdate:   time.Time{},
	}, nil
}
