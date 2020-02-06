package tasker

import "time"

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
