package tasker

import (
	goctx "context"
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
	ctxCancel           goctx.CancelFunc
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
	serving             bool
	priority            int
	lsns                []func(event int, ctx service.Event)
	mu                  sync.Mutex
	mur                 sync.RWMutex
	mux                 sync.Mutex
	wgp                 sync.WaitGroup
	wrkwgp              sync.WaitGroup
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
	defer s.recover()

	s.mur.Lock()
	s.serving = true
	s.ctx, s.ctxCancel = context.WithCancel(ctx)
	s.mur.Unlock()

	s.beginWatch()
	if err := s.watchContext(s.ctx); err != nil {
		s.stopWatch()

		s.wrkwgp.Wait()
		s.wgp.Wait()
		return errors.Wrapf(err, "[%s] Unable to serve service", s.name)
	}

	s.wgp.Add(1)
	go s.watchInput()

	s.Logger.Infof("[%s] Started", nil, s.name)

	go s.serve()
	return nil
}

func (s *Service) recover() {
	if r := recover(); r != nil {
		s.mur.RLock()
		serving := s.serving
		s.mur.RUnlock()

		if serving {
			switch err := r.(type) {
			case error:
				s.Logger.Error(errors.Wrapf(err, "[%s] Service failed", s.name), nil)

			default:
				s.Logger.Error(errors.Errorf("[%s] Service failed: %v", s.name, err), nil)
			}
		}
	}

	s.reset()
}

func (s *Service) serve() {
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
		s.stopWatchContext()
		s.stopWatch()

		s.wrkwgp.Wait()
		s.wgp.Wait()
		s.Logger.Errorf(errors.New("[%s] No workers defined. Stopped"), nil, s.name)
		return
	}

	s.mu.Lock()
	workers := s.workers
	s.mu.Unlock()

	for typ := range workers {
		for i := range workers[typ] {
			s.waiterChan <- workers[typ][i]

			if workers[typ][i].IsAutoStart() {
				s.mur.Lock()
				s.runingWorkers++
				s.autostartingWorkers++
				s.mur.Unlock()

				if err := workers[typ][i].Start(s.ctx); err != nil {
					s.mur.Lock()
					s.runingWorkers--
					s.autostartingWorkers--
					s.mur.Unlock()
					s.Logger.Warning(err.Error(), nil)
				}
			}
		}
	}

	s.mur.RLock()
	runingWorkers := s.runingWorkers
	s.mur.RUnlock()

	s.Logger.Infof("[%s] Started %d workers", nil, s.name, runingWorkers)

	if runingWorkers < 1 {
		if err := s.postStartPhase(); err != nil {
			s.Logger.Errorf("[%s] Unable to initiate configurated tasks: %s", nil, s.name, err)
		}
	}

	s.mur.RLock()
	runingWorkers = s.runingWorkers
	s.mur.RUnlock()

	if runingWorkers < 1 && !s.Options.Persistent {
		s.stopWatchContext()
		s.stopWatch()

		s.wrkwgp.Wait()
		s.wgp.Wait()
		s.Logger.Info("["+s.name+"] No persistent works. Stoped", nil)
		return
	}

	if s.Options.Persistent {
		<-s.exitChan
	}

	s.wrkwgp.Wait()
	s.wgp.Wait()
	s.Logger.Infof("[%s] Stoped. All works done", nil, s.name)

	if s.outChan != nil {
		close(s.outChan)
		s.outChan = nil
	}
}

// Stop stops the service
func (s *Service) Stop() {
	s.mur.RLock()
	serving := s.serving
	watchingCtx := s.watchingCtx
	s.mur.RUnlock()

	if serving {
		s.Logger.Infof("[%s] Waiting for all routines to exit...", nil, s.name)
		if watchingCtx {
			s.ctxCancel()
		} else {
			s.stop()
		}
	}
}

// ID implements waiter interface
func (s *Service) ID() int64 {
	return s.id
}

// InChannel returns channel for receiving messages
func (s *Service) InChannel() (chan<- *Message, error) {
	s.mur.RLock()
	serving := s.serving
	s.mur.RUnlock()

	if !serving {
		return nil, errors.Errorf("[%s] Service is not serving. Income channel is not avaliable at this time", s.name)
	}

	var inChan chan *Message
	s.mu.Lock()
	if s.inChan != nil {
		inChan = s.inChan
	} else {
		inChan = make(chan *Message, 1)
		s.inChan = inChan
	}
	s.mu.Unlock()

	return inChan, nil
}

// SetOutChannel sets channel for sending messages
func (s *Service) SetOutChannel(out chan *Message) error {
	s.mu.Lock()
	if s.outChan != nil {
		close(s.outChan)
	}

	s.outChan = out
	s.mu.Unlock()

	//go func() {
	//	for range out {
	//	}
	//}()
	return nil
}

// UnsetOutChannel unsets channel for sending messages
func (s *Service) UnsetOutChannel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.outChan != nil {
		close(s.outChan)
	}

	s.outChan = nil
}

// Wait returns channel for sending messages
func (s *Service) Wait() <-chan *Message {
	s.mur.RLock()
	serving := s.serving
	exiting := s.inExitSequence
	s.mur.RUnlock()

	if !serving || exiting {
		return nil
	}

	s.mu.Lock()
	outChan := s.outChan
	s.mu.Unlock()

	if outChan == nil {
		outChan = make(chan *Message, 1)
		s.SetOutChannel(outChan)
	}

	return outChan
}

// Worker returns a registered worker, error otherwise
func (s *Service) Worker(workerName string, indx int) (Worker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if wrks, ok := s.workers[workerName]; ok {
		if len(wrks) >= indx {
			return wrks[indx], nil
		}
	}

	return nil, errors.Errorf("[%s] Worker by name '%s' is not registered", s.name, workerName)
}

// Workers returns a registered workers, error otherwise
func (s *Service) Workers(workerName string) ([]Worker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if wrks, ok := s.workers[workerName]; ok {
		return wrks, nil
	}

	return nil, errors.Errorf("[%s] Worker by name '%s' is not registered", s.name, workerName)
}

func (s *Service) stop() {
	s.mur.Lock()
	serving := s.serving
	s.mur.Unlock()

	if !serving {
		return
	}

	s.mu.Lock()
	s.inExitSequence = true
	if s.inChan != nil {
		close(s.inChan)
	}

	if s.exitChan != nil {
		close(s.exitChan)
	}
	s.mu.Unlock()

	s.stopWatchContext()
	s.stopWatch()
}

// resets all service internals
func (s *Service) reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serving {
		return errors.New("Working services can not be reseted")
	}

	s.ctx = nil
	s.ctxCancel = nil
	s.workers = make(map[string][]Worker)
	s.stopChan = nil
	s.exitChan = nil
	s.doneChan = make(chan bool)
	s.returnChan = make(chan bool)
	s.waiterChan = nil
	s.watchChan = nil
	s.inChan = make(chan *Message, 1)
	s.outChan = nil
	s.watching = false
	s.watchingCtx = false
	s.runingWorkers = 0
	s.autostartingWorkers = 0
	s.autostartedWorkers = 0
	s.inExitSequence = false
	s.serving = false

	return nil
}

func (s *Service) postStartPhase() error {
	if len(s.Options.Tasks) > 0 {
		for _, tcfg := range s.Options.Tasks {
			task, err := NewTaskFromConfig(tcfg)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
				continue
			}

			wrk, err := s.lazyWorker(task.Worker)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
				continue
			}

			if wrk.IsWorking() {
				if err := wrk.StartHandler(task); err != nil {
					s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
					continue
				}
			} else {
				if err := wrk.StartTask(s.ctx, task); err != nil {
					s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle root task '%d'", s.name, tcfg.GetInt64("id")), nil)
					continue
				}
			}
		}
	}

	return nil
}

func (s *Service) lazyWorker(workerName string) (wrk Worker, err error) {
	s.mur.RLock()
	exiting := s.inExitSequence
	s.mur.RUnlock()

	if exiting {
		return nil, errors.Errorf("[%s] Service is stoping, can't add new tasks", s.name)
	}

	var ok bool
	var wrkCfg config.Config
	if wrkCfg, ok = s.Options.Workers[workerName]; !ok {
		return nil, errors.Errorf("[%s] Unrecognized worker type '%s'", s.name, workerName)
	}

	if wrkCfg.GetBool("persistent") && !WorkerSupport(workerName, "canReceiveTasks") {
		return nil, errors.Errorf("[%s] Worker specified by task does not support receiving tasks over channels", s.name)
	}

	var wrks []Worker
	s.mu.Lock()
	if wrks, ok = s.workers[workerName]; !ok && s.Options.TryStartNewWorkers {
		s.mu.Unlock()
		wrk, err := s.createWorker(workerName, wrkCfg)
		if err != nil {
			return nil, err
		}

		//s.startWorker(wrk)
		return wrk, nil
	} else if !ok && !s.Options.TryStartNewWorkers {
		s.mu.Unlock()
		return nil, errors.Errorf("[%s] Unrecognized worker type '%s'", s.name, workerName)
	} else {
		s.mu.Unlock()
	}

	for i := range wrks {
		if wrks[i].IsWorking() && wrks[i].CanHandleMore() {
			return wrks[i], nil
		} else if wrks[i].CanHandleMore() {
			//s.startWorker(wrks[i])
			return wrks[i], nil
		}
	}

	maxWorkers := wrkCfg.GetIntDefault("instances", -1)
	if s.Options.TryStartNewWorkers && (maxWorkers == 0 || maxWorkers > len(wrks)) {
		wrk, err := s.createWorker(workerName, wrkCfg)
		if err != nil {
			return nil, err
		}

		//s.startWorker(wrk)
		return wrk, nil
	}

	return nil, errors.Errorf("[%s] No avaliable workers", s.name)
}

func (s *Service) createWorker(workerName string, wrkCfg config.Config) (Worker, error) {
	wrk, err := NewWorker(workerName, wrkCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "[%s] Unable to create worker of type '%s'", s.name, workerName)
	}

	wrk.SetLogger(s.Logger)
	s.mu.Lock()
	s.workers[workerName] = append(s.workers[workerName], wrk)
	s.mu.Unlock()
	s.waiterChan <- wrk

	return wrk, nil
}

func (s *Service) startWorker(wrk Worker) {
	s.mur.Lock()
	s.runingWorkers++
	s.mur.Unlock()

	if err := wrk.Start(s.ctx); err != nil {
		s.mur.Lock()
		s.runingWorkers--
		s.mur.Unlock()
		s.Logger.Error(errors.Wrapf(err, "[%s] Unable to start worker", s.name), nil)
		return
	}
}

func (s *Service) startWorkerWithTask(wrk Worker, tsk *Task) error {
	s.mur.Lock()
	s.runingWorkers++
	s.mur.Unlock()

	if err := wrk.StartTask(s.ctx, tsk); err != nil {
		s.mur.Lock()
		s.runingWorkers--
		s.mur.Unlock()
		s.Logger.Error(errors.Wrapf(err, "[%s] Unable to start worker", s.name), nil)
		return err
	}

	return nil
}

func (s *Service) beginWatch() {
	s.mu.Lock()
	s.stopChan = make(chan bool)
	s.exitChan = make(chan bool)
	watcher := make(chan context.Context, 1)
	s.watchChan = watcher
	out := make(chan Waiter)
	s.waiterChan = out
	s.watching = true
	s.mu.Unlock()

	s.wgp.Add(1)
	go func() {
		defer s.wgp.Done()

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
				s.stop()
				return
			case <-s.stopChan:
				return
			}
		}
	}()

	s.wgp.Add(1)
	go func() {
		defer s.wgp.Done()

		for wt := range out {
			s.wrkwgp.Add(1)
			go func(wt Waiter) {
				defer s.wrkwgp.Done()

				for msg := range wt.Wait() {
					s.mu.Lock()
					outChan := s.outChan
					s.mu.Unlock()

					if outChan != nil && msg.Scope <= ScopeGlobal {
						outChan <- msg
					}

					switch msg.Type {
					case MessageWorkerStarted:
						s.mur.Lock()
						s.autostartedWorkers++
						s.mur.Unlock()

						s.mur.RLock()
						starting := s.autostartingWorkers
						started := s.autostartedWorkers
						s.mur.RUnlock()

						if starting == started {
							if err := s.postStartPhase(); err != nil {
								s.Logger.Error(errors.Wrapf(err, "[%s] Post start phase failed", s.name), nil)
							}
						}
					}

					if msg.Error == nil {
						s.Logger.Info(msg.Text, nil)
					} else {
						s.Logger.Log(msg.Error, msg.Priority, nil)
					}
				}

				s.mur.Lock()
				s.runingWorkers--
				s.mur.Unlock()
				return
			}(wt)
		}

		return
	}()

	time.Sleep(time.Millisecond * 100)
}

func (s *Service) stopWatch() {
	s.mur.Lock()
	watching := s.watching
	s.mur.Unlock()

	if !watching {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.waiterChan)
	close(s.stopChan)
	s.watching = false
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

// Stop watching for context changes
func (s *Service) stopWatchContext() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.watchingCtx {
		return
	}

	s.watchingCtx = false
	if s.watchChan != nil {
		close(s.watchChan)
	}
}

func (s *Service) watchInput() {
	s.Logger.Debugf("[%s] watching input", nil, s.name)
	defer s.restartWatchInput()

	s.mu.Lock()
	inChan := s.inChan
	s.mu.Unlock()

	for msg := range inChan {
		s.Logger.Debugf("[%s] watchInput() : message '%d' of type '%v' received", nil, s.name, msg.ID, msg.Type)
		s.mur.RLock()
		serving := s.serving
		exiting := s.inExitSequence
		s.mur.RUnlock()

		if !serving || exiting {
			s.Logger.Debugf("[%s] watchInput() : not serving or in exit sequence", nil, s.name)
			// Maybe add a message about pushing to closing channel
			break
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
				s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle new task '%d'", s.name, msg.Task.ID), nil)
				continue
			}

			if wrk.IsWorking() {
				wrkInChan, err := wrk.InChannel()
				if err != nil {
					s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle new task '%d'", s.name, msg.Task.ID), nil)
					continue
				}

				wrkInChan <- msg
			} else {
				if err := s.startWorkerWithTask(wrk, &msg.Task); err != nil {
					s.Logger.Error(errors.Wrapf(err, "[%s] Unable to handle new task '%d'", s.name, msg.Task.ID), nil)
					continue
				}
			}

		case MessageStopTask:
			s.Logger.Error(errors.Errorf("[%s] Not implemented", s.name), nil)

		default:
			s.Logger.Error(errors.Errorf("[%s] Unrecognized message type %d", s.name, msg.Type), nil)
		}
	}
}

func (s *Service) restartWatchInput() {
	if r := recover(); r != nil {
		s.Logger.Debugf("[%s] watchInput() recovered from error: %v", nil, s.name, r)
		s.mur.RLock()
		serving := s.serving
		exiting := s.inExitSequence
		s.mur.RUnlock()

		if serving && !exiting {
			s.Logger.Debugf("[%s] watchInput() restarting...", nil, s.name)
			go s.watchInput()
			return
		} else if !serving {
			s.Logger.Debugf("[%s] restartWatchInput() : not serving", nil, s.name)
			return
		}
	}

	s.wgp.Done()
}

// NewService creates a new service of type Tasker
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{
		id:             time.Now().UnixNano(),
		name:           "Tasker",
		workers:        make(map[string][]Worker),
		inExitSequence: false,
		returnChan:     make(chan bool),
		inChan:         make(chan *Message, 1),
		runingWorkers:  0,
		serving:        false,
		priority:       5,
		wgp:            sync.WaitGroup{},
		wrkwgp:         sync.WaitGroup{},
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
