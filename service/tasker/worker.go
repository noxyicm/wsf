package tasker

import (
	"sync"
	"time"
	"wsf/config"
	"wsf/context"
	"wsf/db"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/service"
	"wsf/utils"
)

const (
	// ID of service
	ID = "tasker"

	// TaskStatusReady represents task that is ready do work
	TaskStatusReady = 1
	// TaskStatusInProgress represents task that is in work
	TaskStatusInProgress = 100
	// TaskStatusOver represents task that is over
	TaskStatusOver = 200
	// TaskStatusFail represents task that is failed
	TaskStatusFail = 500
)

var handlers = map[string]Handler{}

// Service is Worker service
type Service struct {
	options               *Config
	Db                    *db.Db
	Logger                *log.Log
	ctx                   context.Context
	StopChan              chan bool
	ExitChan              chan bool
	IntervalChan          chan bool
	NotyChan              chan int
	OutChan               chan TaskState
	NotifyedTasks         []int
	Tasks                 map[int]time.Time
	InExitSequence        bool
	MaxConsequetiveErrors int
	lsns                  []func(event int, ctx interface{})
	mu                    sync.Mutex
	mux                   sync.Mutex
	serving               bool
	priority              int
}

// Init Worker service
func (s *Service) Init(options *Config) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.options = options
	dbResource := registry.GetResource("db")
	if dbResource == nil {
		return false, errors.New("DB resource is not configured")
	}

	s.Db = dbResource.(*db.Db)

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("Log resource is not configured")
	}

	s.Logger = logResource.(*log.Log)
	return true, nil
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// AddListener attaches server event watcher
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

// Serve serves the service
func (s *Service) Serve(ctx context.Context) (err error) {
	if s.Db == nil {
		return errors.New("Db resource is not configured")
	}

	s.mu.Lock()
	s.serving = true
	s.ctx = ctx
	s.mu.Unlock()

	s.Logger.Info("[Tasker] Started", nil)

	tasksInRoutines := make([]int, 0)
	tasksConsequetiveErrors := make(map[int]int)
	go func() {
		for obj := range s.OutChan {
			s.mu.Lock()

			if !obj.State {
				if v, ok := tasksConsequetiveErrors[obj.ID]; ok {
					if !utils.InISlice(obj.ID, s.NotifyedTasks) && v > s.MaxConsequetiveErrors {
						s.NotyChan <- obj.ID
					}

					tasksConsequetiveErrors[obj.ID] = v + 1
				} else {
					tasksConsequetiveErrors[obj.ID] = 1
				}

				_, _ = s.Db.Update(s.Db.Context(), "tasks", map[string]interface{}{"lastrun": nil}, map[string]interface{}{"id = ?": obj.ID})
			} else {
				delete(tasksConsequetiveErrors, obj.ID)
				key, hasKey := utils.IKey(obj.ID, s.NotifyedTasks)
				if hasKey {
					s.NotifyedTasks = append(s.NotifyedTasks[:key], s.NotifyedTasks[key+1:]...)
				}
			}

			if obj.Error != nil {
				s.Logger.Error(obj.Error.Error(), nil)
			}

			key, hasKey := utils.IKey(obj.ID, tasksInRoutines)
			if hasKey {
				tasksInRoutines = append(tasksInRoutines[:key], tasksInRoutines[key+1:]...)
			}

			if s.InExitSequence && len(tasksInRoutines) == 0 {
				close(s.ExitChan)
			}

			s.mu.Unlock()
		}

		return
	}()

	go func() {
		for aid := range s.NotyChan {
			s.SendAlert(aid)
		}

		return
	}()

Mainloop:
	for {
		select {
		case <-s.StopChan:
			break Mainloop
		default:
			s.mu.Lock()
			now := time.Now()

			sel, err := s.Db.Select()
			if err != nil {
				s.Logger.Error(err.Error(), nil)
			}

			sel.From("tasks", "*").Where("state = ?", TaskStatusReady)
			preresult, err := s.Db.Query(s.ctx, sel)
			if err != nil {
				s.Logger.Error(err.Error(), nil)
			} else {
				s.Logger.Debugf("[Tasker] Found %d tasks", nil, preresult.Count())

				for preresult.Next() {
					row := preresult.Get()

					if utils.InISlice(row.GetInt("id"), tasksInRoutines) {
						s.Logger.Debugf("[Tasker] Task #%d already processing. Ignoring", nil, row.GetInt("id"))
						continue
					}

					var task *Task
					if err := row.Unmarshal(&task); err != nil {
						s.Logger.Error(err.Error(), nil)
					}

					if _, ok := handlers[task.Handler]; !ok {
						s.Logger.Errorf("[Tasker] Handler by name '%s' is not registered", nil, task.Handler)
						continue
					}

					if task.Interval {
						tasksInRoutines = append(tasksInRoutines, task.ID)
						go func(t *Task) {
							itrvl := time.Duration(t.Second + (60 * t.Minute) + (3600 * t.Hour) + (86400 * t.Day))
							select {
							case <-time.After(itrvl * time.Second):
								go handlers[task.Handler].StartRoutine(s.ctx, task, s.OutChan)
							}
						}(task)
					} else {
						if task.Date.IsZero() {
							year := now.Year()
							month := now.Month()
							day := now.Day()

							if task.Year != 0 {
								year = task.Year
							}

							switch task.Month {
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

							if task.Day != 0 {
								day = task.Day
							}

							exectime := time.Date(year, month, day, task.Hour, task.Minute, task.Second, 0, time.Local)
							if exectime.Equal(now.Round(time.Second)) {
								go handlers[task.Handler].StartRoutine(s.ctx, task, s.OutChan)
								tasksInRoutines = append(tasksInRoutines, task.ID)
							} else {
								s.Logger.Debugf("[Tasker] Task #%d scheduled for %s. Ignoring", nil, task.ID, exectime.Format("15:04:05"))
							}
							//} else if task.Date.Equal(now.Round(time.Second)) {
						} else if now.After(task.Date) {
							go handlers[task.Handler].StartRoutine(s.ctx, task, s.OutChan)
							tasksInRoutines = append(tasksInRoutines, task.ID)
						} else {
							s.Logger.Debugf("[Tasker] Task #%d scheduled for %s. Ignoring", nil, task.ID, task.Date.Format(time.RFC3339))
						}
					}

					//_, _ = s.Db.Update("tasks", map[string]interface{}{"lastrun": time.Now()}, map[string]interface{}{"id = ?": task.ID})
				}
			}

			s.mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}

	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.mu.Lock()
	s.InExitSequence = true
	s.ctx.Cancel()

	if s.serving {
		close(s.StopChan)
		close(s.IntervalChan)
		s.OutChan <- TaskState{ID: 0, State: false, Error: nil}
		close(s.NotyChan)
		s.Logger.Info("[Tasker] Waiting for all routines to exit...", nil)
	}
	s.mu.Unlock()

	<-s.ExitChan
	close(s.OutChan)
	s.Logger.Info("[Tasker] Stoped", nil)
}

// SendAlert sends message to administrator about bad tasks
func (s *Service) SendAlert(aid int) {
	s.NotifyedTasks = append(s.NotifyedTasks, aid)
}

// NewService creates a new service of type Tasker
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{
		StopChan:              make(chan bool, 1),
		ExitChan:              make(chan bool),
		IntervalChan:          make(chan bool, 1),
		NotyChan:              make(chan int),
		OutChan:               make(chan TaskState, 1),
		NotifyedTasks:         make([]int, 0),
		Tasks:                 make(map[int]time.Time),
		InExitSequence:        false,
		MaxConsequetiveErrors: 10,
		serving:               false,
		priority:              5,
	}, nil
}

// RegisterHandler registers a new handler for worker tasks
func RegisterHandler(handlerName string, hndl Handler) error {
	if _, ok := handlers[handlerName]; ok {
		return errors.Errorf("[Tasker] Handler with name '%s' already registered", handlerName)
	}

	handlers[handlerName] = hndl
	return nil
}
