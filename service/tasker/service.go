package tasker

import (
	goctx "context"
	"fmt"
	"sync"
	"time"
	"wsf/config"
	"wsf/context"
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
	workersSupport      = map[string]map[string]bool{}
)

// Service is Worker service
type Service struct {
	id                  int64
	Options             *Config
	Logger              *log.Log
	name                string
	ctx                 context.Context
	cancel              goctx.CancelFunc
	workers             map[string][]Worker
	stopChan            chan bool
	exitChan            chan bool
	doneChan            chan bool
	returnChan          chan bool
	waiterChan          chan<- Waiter
	watchChan           chan<- context.Context
	inChan              chan *Message
	outChan             chan *Message
	watching            bool
	watchingCtx         bool
	runingWorkers       int
	autostartingWorkers int
	autostartedWorkers  int
	inExitSequence      bool
	lsns                []func(event int, ctx service.Event)
	mu                  sync.Mutex
	mur                 sync.RWMutex
	mux                 sync.Mutex
	serving             bool
	priority            int
}

// Init Worker service
func (s *Service) Init(options *Config) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.Options = options
	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("[" + s.name + "] Log resource is not configured")
	}

	s.Logger = logResource.(*log.Log)
	return true, nil
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// AddListener attaches server event watcher
func (s *Service) AddListener(l func(event int, ctx service.Event)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx service.Event) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

// Serve serves the service
func (s *Service) Serve(ctx context.Context) (err error) {
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.inChan = make(chan *Message)
	s.serving = true
	s.mu.Unlock()

	s.beginWatch()
	if err := s.watchContext(s.ctx); err != nil {
		s.stop()
		return errors.Wrap(err, "["+s.name+"] Unable to serve service")
	}

	potentialWorkers := 0
	for wrkType, wrkCfg := range s.Options.Workers {
		for i := 0; i < wrkCfg.GetIntDefault("instances", 0); i++ {
			wrk, err := NewWorker(wrkType, wrkCfg)
			if err != nil {
				s.Logger.Warning(errors.Wrapf(err, "[%s] Unable to create worker of type '%s'", s.name, wrkType).Error(), nil)
				continue
			}

			wrk.SetLogger(s.Logger)
			s.mu.Lock()
			s.workers[wrkType] = append(s.workers[wrkType], wrk)
			s.mu.Unlock()
			potentialWorkers++
		}
	}

	if potentialWorkers == 0 {
		s.stop()
		return errors.New("[" + s.name + "] Stopped. No workers defined")
	}

	s.mu.Lock()
	workers := s.workers
	s.mu.Unlock()

	for typ := range workers {
		for i := range workers[typ] {
			s.mu.Lock()
			s.waiterChan <- workers[typ][i]
			s.mu.Unlock()

			if workers[typ][i].IsAutoStart() {
				s.mu.Lock()
				s.runingWorkers++
				s.autostartingWorkers++
				s.mu.Unlock()

				if err := workers[typ][i].Start(s.ctx); err != nil {
					s.mu.Lock()
					s.runingWorkers--
					s.autostartingWorkers--
					s.mu.Unlock()
					s.Logger.Warning(err.Error(), nil)
				}
			}
		}
	}

	s.Logger.Info("["+s.name+"] Started", nil)

	if s.runingWorkers < 1 && !s.Options.Persistent {
		s.stop()
		s.Logger.Info("["+s.name+"] All tasks done. Stoped", nil)
		return nil
	}

MainLoop:
	for {
		select {
		case <-s.stopChan:
			break MainLoop

		case msg, ok := <-s.inChan:
			if !ok {
				break MainLoop
			}

			switch msg.Type {
			case MessageModifyTask:
				wrk, err := s.Worker(msg.Task.Worker, 0)
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "["+s.name+"] Unable to handle task modification"), nil)
					continue
				}

				wrkInChan, err := wrk.InChannel()
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "["+s.name+"] Unable to handle task modification"), nil)
					continue
				}

				wrkInChan <- msg

			case MessageAddTask:
				fallthrough
			case MessageStartTask:
				wrk, err := s.lazyWorker(msg.Task.Worker)
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "["+s.name+"] Unable to handle new task"), nil)
					continue
				}

				wrkInChan, err := wrk.InChannel()
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "["+s.name+"] Unable to handle new task"), nil)
					continue
				}

				wrkInChan <- msg

			case MessageStopTask:
				s.Logger.Error(errors.New("["+s.name+"] Not implemented"), nil)

			default:
				s.Logger.Error(errors.Errorf("[%s] Unrecognized message type %d", s.name, msg.Type), nil)
			}
		}
	}

	s.stop()

	s.Logger.Info("["+s.name+"] Stoped", nil)
	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.mu.Lock()
	s.inExitSequence = true
	serving := s.serving
	s.mu.Unlock()

	if serving {
		s.Logger.Info("["+s.name+"] Waiting for all routines to exit...", nil)
		s.ctx.Cancel()
		<-s.exitChan
	}

	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()
}

// ID implements waiter interface
func (s *Service) ID() int64 {
	return s.id
}

// InChannel returns channel for receiving messages
func (s *Service) InChannel() (chan<- *Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.serving {
		return nil, errors.New("[" + s.name + "] Service is not serving. Income channel is not avaliable at this time")
	}

	if s.inChan == nil {
		s.inChan = make(chan *Message)
	}

	return s.inChan, nil
}

// SetOutChannel sets channel for sending messages
func (s *Service) SetOutChannel(out chan *Message) error {
	s.outChan = out
	return nil
}

// Wait returns channel for sending messages
func (s *Service) Wait() <-chan *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.serving || s.inExitSequence {
		return nil
	}

	if s.outChan == nil {
		s.outChan = make(chan *Message, 2)
	}

	c := s.outChan
	return c
}

// Worker returns a registered worker, error otherwise
func (s *Service) Worker(workerName string, indx int) (Worker, error) {
	s.mu.Lock()
	workers := s.workers
	s.mu.Unlock()

	if wrks, ok := workers[workerName]; ok {
		if len(wrks) >= indx {
			return wrks[indx], nil
		}
	}

	return nil, errors.Errorf("[%s] Worker by name '%s' is not registered", s.name, workerName)
}

func (s *Service) stop() {
	s.mu.Lock()
	s.inExitSequence = true
	s.mu.Unlock()

	if s.inChan != nil {
		close(s.inChan)
	}

	s.stopWatch()

	<-s.exitChan

	if s.outChan != nil {
		close(s.outChan)
	}
}

func (s *Service) postStartPhase() error {
	if len(s.Options.Tasks) > 0 {
		for _, tcfg := range s.Options.Tasks {
			task, err := NewTaskFromConfig(tcfg)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
				continue
			}

			msg := &Message{Type: MessageStartTask, Task: *task}
			wrk, err := s.lazyWorker(msg.Task.Worker)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
				continue
			}

			wrkInChan, err := wrk.InChannel()
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
				continue
			}

			wrkInChan <- msg
		}
	}

	return nil
}

func (s *Service) lazyWorker(workerName string) (wrk Worker, err error) {
	s.mur.Lock()
	exiting := s.inExitSequence
	s.mur.Unlock()

	if exiting {
		return nil, errors.New("[" + s.name + "] Service is stoping, can't add new tasks")
	}

	var ok bool
	var wrkCfg config.Config
	if wrkCfg, ok = s.Options.Workers[workerName]; !ok {
		return nil, errors.Errorf("[%s] Unrecognized worker type '%s'", s.name, workerName)
	}

	if !WorkerSupport(workerName, "canReceiveTasks") {
		return nil, errors.New("[" + s.name + "] Worker specified by task does not support receiving tasks over channels")
	}

	var wrks []Worker
	if wrks, ok = s.workers[workerName]; !ok && s.Options.TryStartNewWorkers {
		wrk, err := s.createWorker(workerName, wrkCfg)
		if err != nil {
			return nil, err
		}

		go s.startWorker(wrk)
		return wrk, nil
	} else if !ok && !s.Options.TryStartNewWorkers {
		return nil, errors.Errorf("[%s] Unrecognized worker type '%s'", s.name, workerName)
	}

	for i := range wrks {
		if wrks[i].IsWorking() && wrks[i].CanHandleMore() {
			return wrks[i], nil
		} else if wrks[i].CanHandleMore() {
			go s.startWorker(wrks[i])
			return wrks[i], nil
		}
	}

	maxWorkers := wrkCfg.GetIntDefault("instances", -1)
	if s.Options.TryStartNewWorkers && (maxWorkers == 0 || maxWorkers > len(wrks)) {
		wrk, err := s.createWorker(workerName, wrkCfg)
		if err != nil {
			return nil, err
		}

		go s.startWorker(wrk)
		return wrk, nil
	}

	return nil, errors.New("[" + s.name + "] No avaliable workers")
}

func (s *Service) createWorker(workerName string, wrkCfg config.Config) (Worker, error) {
	wrk, err := NewWorker(workerName, wrkCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "[%s] Unable to create worker of type '%s'", s.name, workerName)
	}

	wrk.SetLogger(s.Logger)
	s.mu.Lock()
	s.workers[workerName] = append(s.workers[workerName], wrk)
	s.waiterChan <- wrk
	s.mu.Unlock()

	return wrk, nil
}

func (s *Service) startWorker(wrk Worker) {
	s.mu.Lock()
	s.runingWorkers++
	s.mu.Unlock()

	if err := wrk.Start(s.ctx); err != nil {
		s.mu.Lock()
		s.runingWorkers--
		s.mu.Unlock()
		s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle new task", s.name), nil)
		return
	}
}

func (s *Service) beginWatch() {
	s.mu.Lock()
	s.stopChan = make(chan bool)
	s.exitChan = make(chan bool)
	s.doneChan = make(chan bool)
	watcher := make(chan context.Context, 1)
	s.watchChan = watcher
	out := make(chan Waiter)
	s.waiterChan = out
	s.watching = true
	s.mu.Unlock()

	go func() {
		for {
			var ctx context.Context
			var ok bool
			select {
			case ctx, ok = <-watcher:
				if !ok {
					return
				}
			case <-s.stopChan:
				return
			}

			select {
			case <-ctx.Done():
				fmt.Println("Service.beginWatch() <-ctx.Done()")
				s.stopWatch()
			case <-s.stopChan:
				return
			}
		}
	}()

	go func() {
		for wt := range out {
			go func(wt Waiter) {
				for msg := range wt.Wait() {
					switch msg.Type {
					case MessageWorkerStart:
						s.mu.Lock()
						s.autostartedWorkers++
						starting := s.autostartingWorkers
						started := s.autostartedWorkers
						s.mu.Unlock()

						if starting == started {
							if err := s.postStartPhase(); err != nil {
								s.Logger.Error(errors.Wrapf(err, "[%s] Post start phase failed", s.name), nil)
							}
						}

					case MessageTaskAdded:
						fallthrough
					case MessageTaskModified:
						fallthrough
					case MessageTaskStoped:
						fallthrough
					case MessageTaskRemoved:
						fallthrough
					case MessageTaskDone:
						if s.outChan != nil {
							s.outChan <- msg
						}
					}

					if msg.Error == nil {
						s.Logger.Info(msg.Text, nil)
					} else {
						s.Logger.Log(msg.Error, msg.Priority, nil)
					}
				}

				s.mu.Lock()
				s.runingWorkers--
				s.mu.Unlock()

				s.done()
				return
			}(wt)
		}

		s.done()
		return
	}()
}

func (s *Service) stopWatch() {
	s.mur.Lock()
	watching := s.watching
	s.mur.Unlock()

	if !watching {
		return
	}

	s.mu.Lock()
	close(s.waiterChan)
	s.waiterChan = nil
	close(s.watchChan)
	s.watchChan = nil
	close(s.stopChan)
	s.stopChan = nil
	s.watching = false
	s.mu.Unlock()
}

func (s *Service) done() {
	s.mur.Lock()
	runningWorkers := s.runingWorkers
	exiting := s.inExitSequence
	serving := s.serving
	s.mur.Unlock()

	if !serving || s.exitChan == nil {
		return
	}

	if runningWorkers < 1 && (exiting || s.waiterChan == nil || !s.Options.Persistent) {
		s.mur.Lock()
		close(s.exitChan)
		s.exitChan = nil
		s.mur.Unlock()
	}
}

func (s *Service) watchContext(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watchingCtx {
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

	s.watchingCtx = true
	s.watchChan <- ctx
	return nil
}

// NewService creates a new service of type Tasker
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{
		id:             time.Now().UnixNano(),
		name:           "Tasker",
		workers:        make(map[string][]Worker),
		inExitSequence: false,
		returnChan:     make(chan bool),
		runingWorkers:  0,
		serving:        false,
		priority:       5,
	}, nil
}

// RegisterHandler registers a handler for worker tasks
func RegisterHandler(handlerName string, hndl func(*Task) (Handler, error)) error {
	if _, ok := handlers[handlerName]; ok {
		return errors.Errorf("Handler with name '%s' already registered", handlerName)
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
func RegisterWorker(workerName string, wrkr func(config.Config) (Worker, error), workerSupports map[string]bool) error {
	if _, ok := buildWorkerHandlers[workerName]; ok {
		return errors.Errorf("Worker with name '%s' already registered", workerName)
	}

	buildWorkerHandlers[workerName] = wrkr
	workersSupport[workerName] = workerSupports
	return nil
}

// NewWorker creates a new worker
func NewWorker(workerType string, options config.Config) (Worker, error) {
	if f, ok := buildWorkerHandlers[workerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized worker type \"%v\"", workerType)
}

// WorkerSupport return true if worker support specified functionality
func WorkerSupport(workerName string, param string) bool {
	if w, ok := workersSupport[workerName]; ok {
		if b, ok := w[param]; ok {
			return b
		}
	}

	return false
}
