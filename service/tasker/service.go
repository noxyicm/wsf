package tasker

import (
	goctx "context"
	"sync"
	"time"
	"wsf/config"
	"wsf/context"
	"wsf/db"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/service"
)

const (
	// ID of service
	ID = "tasker"
)

var (
	handlers            = map[string]func(*Task) (Handler, error){}
	buildWorkerHandlers = map[string]func(config.Config) (Worker, error){}
)

// Service is Worker service
type Service struct {
	Options               *Config
	Db                    *db.Db
	Logger                *log.Log
	ctx                   context.Context
	cancel                goctx.CancelFunc
	workers               map[string][]Worker
	StopChan              chan bool
	ExitChan              chan bool
	IntervalChan          chan bool
	NotyChan              chan int
	OutChan               chan<- Waiter
	watchChan             chan<- context.Context
	doneChan              chan<- bool
	watching              bool
	runingWorkers         int
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

	s.Options = options
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
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.beginWatch()
	if err := s.watchContext(s.ctx); err != nil {
		return errors.Wrap(err, "[Tasker] Unable to serve service")
	}

	potentialWorkers := 0
	for wrkType, wrkCfg := range s.Options.Workers {
		for i := 0; i < wrkCfg.GetIntDefault("instances", 0); i++ {
			wrk, err := NewWorker(wrkType, wrkCfg)
			if err != nil {
				s.Logger.Warning(errors.Wrapf(err, "[Tasker] Unable to create worker of type '%s'", wrkType).Error(), nil)
				continue
			}

			wrk.SetLogger(s.Logger)
			s.workers[wrkType] = append(s.workers[wrkType], wrk)
			potentialWorkers++
		}
	}

	if potentialWorkers == 0 {
		return errors.New("No workers defined")
	}

	s.Logger.Info("[Tasker] Started", nil)

	for typ := range s.workers {
		for i := range s.workers[typ] {
			go func(wrk Worker) {
				s.mu.Lock()
				s.OutChan <- wrk
				s.runingWorkers++
				s.mu.Unlock()

				if err := wrk.Start(s.ctx); err != nil {
					s.Logger.Warning(err.Error(), nil)
					s.mu.Lock()
					s.runingWorkers--
					s.mu.Unlock()
				}

				return
			}(s.workers[typ][i])
		}
	}

	<-s.ExitChan

	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	s.Logger.Info("[Tasker] Stoped", nil)
	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.mu.Lock()
	s.InExitSequence = true

	if s.serving {
		s.Logger.Info("[Tasker] Waiting for all routines to exit...", nil)
		s.ctx.Cancel()
	}
	s.mu.Unlock()

	<-s.ExitChan
}

// SendAlert sends message to administrator about bad tasks
func (s *Service) SendAlert(aid int) {
	s.NotifyedTasks = append(s.NotifyedTasks, aid)
}

func (s *Service) beginWatch() {
	s.mu.Lock()
	watcher := make(chan context.Context, 1)
	s.watchChan = watcher
	out := make(chan Waiter)
	s.OutChan = out
	s.mu.Unlock()

	go func() {
		for {
			var ctx context.Context
			select {
			case ctx = <-watcher:
			case <-s.StopChan:
				return
			}

			select {
			case <-ctx.Done():
				close(s.StopChan)
			case <-s.StopChan:
				return
			}
		}
	}()

	go func() {
		for wt := range out {
			go func(wt Waiter) {
				for msg := range wt.Wait() {
					if msg.Error == nil {
						s.Logger.Info(msg.Text, nil)
					} else {
						s.Logger.Log(msg.Error, msg.Priority, nil)
					}
				}

				s.mu.Lock()
				s.runingWorkers--
				if s.runingWorkers < 1 && s.InExitSequence {
					close(out)
				}
				s.mu.Unlock()
				return
			}(wt)
		}

		s.mu.Lock()
		close(s.ExitChan)
		s.mu.Unlock()
		return
	}()
}

func (s *Service) watchContext(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watching {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if ctx.Done() == nil {
		return nil
	}

	if s.watchChan == nil {
		return nil
	}

	s.watching = true
	s.watchChan <- ctx
	return nil
}

// NewService creates a new service of type Tasker
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{
		workers:        make(map[string][]Worker),
		StopChan:       make(chan bool, 1),
		ExitChan:       make(chan bool),
		IntervalChan:   make(chan bool, 1),
		NotyChan:       make(chan int),
		NotifyedTasks:  make([]int, 0),
		Tasks:          make(map[int]time.Time),
		InExitSequence: false,
		runingWorkers:  0,
		serving:        false,
		priority:       5,
	}, nil
}

// RegisterHandler registers a handler for worker tasks
func RegisterHandler(handlerName string, hndl func(*Task) (Handler, error)) error {
	if _, ok := handlers[handlerName]; ok {
		return errors.Errorf("[Tasker] Handler with name '%s' already registered", handlerName)
	}

	handlers[handlerName] = hndl
	return nil
}

// HasHandler return true if handler by name handlerName is in handlers map
func HasHandler(handlerName string) bool {
	if _, ok := handlers[handlerName]; ok {
		return true
	}

	return false
}

// NewHandler return new instance of handler by name handlerName
func NewHandler(handlerName string, task *Task) (Handler, error) {
	if f, ok := handlers[handlerName]; ok {
		return f(task)
	}

	return nil, errors.Errorf("Unrecognized handler \"%v\"", handlerName)
}

// RegisterWorker registers a worker
func RegisterWorker(workerName string, wrkr func(config.Config) (Worker, error)) error {
	if _, ok := buildWorkerHandlers[workerName]; ok {
		return errors.Errorf("[Tasker] Worker with name '%s' already registered", workerName)
	}

	buildWorkerHandlers[workerName] = wrkr
	return nil
}

// NewWorker creates a new worker
func NewWorker(workerType string, options config.Config) (Worker, error) {
	if f, ok := buildWorkerHandlers[workerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized worker type \"%v\"", workerType)
}
