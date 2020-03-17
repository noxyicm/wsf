package tasker

import (
	goctx "context"
	"sync"
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
	Options             *Config
	Logger              *log.Log
	ctx                 context.Context
	cancel              goctx.CancelFunc
	workers             map[string][]Worker
	stopChan            chan bool
	exitChan            chan bool
	doneChan            chan bool
	returnChan          chan bool
	outChan             chan<- Waiter
	watchChan           chan<- context.Context
	inChan              chan *Message
	watching            bool
	watchingCtx         bool
	runingWorkers       int
	autostartingWorkers int
	autostartedWorkers  int
	inExitSequence      bool
	lsns                []func(event int, ctx interface{})
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
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.beginWatch()
	if err := s.watchContext(s.ctx); err != nil {
		s.stopWatch()
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
			s.mu.Lock()
			s.workers[wrkType] = append(s.workers[wrkType], wrk)
			s.mu.Unlock()
			potentialWorkers++
		}
	}

	if potentialWorkers == 0 {
		s.stopWatch()
		return errors.New("[Tasker] No workers defined")
	}

	s.mu.Lock()
	workers := s.workers
	s.mu.Unlock()

	for typ := range workers {
		for i := range workers[typ] {
			s.mu.Lock()
			s.outChan <- workers[typ][i]
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

	s.Logger.Info("[Tasker] Started", nil)

	if s.runingWorkers < 1 && !s.Options.Persistent {
		s.stopWatch()
		s.Logger.Info("[Tasker] Stoped. No persistent tasks", nil)
		return nil
	}

	s.mu.Lock()
	s.serving = true
	s.inChan = make(chan *Message, 1)
	s.mu.Unlock()

MainLoop:
	for {
		select {
		case <-s.stopChan:
			s.done()
			break MainLoop

		case msg, ok := <-s.inChan:
			if !ok {
				break MainLoop
			}

			switch msg.Type {
			case MessageModifyTask:
				wrk, err := s.Worker(msg.Task.Worker, 0)
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "[Tasker] Unable to handle task modification"), nil)
					continue
				}

				wrkInChan, err := wrk.InChannel()
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "[Tasker] Unable to handle task modification"), nil)
					continue
				}

				wrkInChan <- msg

			case MessageAddTask:
				wrk, err := s.lazyWorker(msg.Task.Worker)
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "[Tasker] Unable to handle new task"), nil)
					continue
				}

				wrkInChan, err := wrk.InChannel()
				if err != nil {
					s.Logger.Error(errors.Wrap(err, "[Tasker] Unable to handle new task"), nil)
					continue
				}

				wrkInChan <- msg

			case MessageStopTask:
				s.Logger.Error(errors.New("[Tasker] Not implemented"), nil)
				continue
			}
		}
	}

	<-s.exitChan
	s.mu.Lock()
	s.serving = false
	close(s.inChan)
	s.inChan = nil
	s.mu.Unlock()

	s.Logger.Info("[Tasker] Stoped", nil)
	close(s.returnChan)
	s.returnChan = nil
	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.mu.Lock()
	s.inExitSequence = true
	serving := s.serving
	s.mu.Unlock()

	if serving {
		s.Logger.Info("[Tasker] Waiting for all routines to exit...", nil)
		s.ctx.Cancel()
	}

	<-s.returnChan
}

// InChannel returns channel for receiving new tasks
func (s *Service) InChannel() (chan<- *Message, error) {
	if !s.serving {
		return nil, errors.New("[Tasker] Service is not serving. Incoming channel is not avaliable at this time")
	}

	return s.inChan, nil
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

	return nil, errors.Errorf("[Tasker] Worker by name '%s' is not registered", workerName)
}

func (s *Service) postStartPhase() error {
	if len(s.Options.Tasks) > 0 {
		for _, tcfg := range s.Options.Tasks {
			task, err := NewTaskFromConfig(tcfg)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[Tasker] Unable to handle root task '%d'", tcfg.GetInt64("id")), nil)
				continue
			}

			msg := &Message{Type: MessageAddTask, Task: *task}
			wrk, err := s.lazyWorker(msg.Task.Worker)
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[Tasker] Unable to handle root task '%d'", tcfg.GetInt64("id")), nil)
				continue
			}

			wrkInChan, err := wrk.InChannel()
			if err != nil {
				s.Logger.Error(errors.Wrapf(err, "[Tasker] Unable to handle root task '%d'", tcfg.GetInt64("id")), nil)
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
		return nil, errors.New("Service is stoping, can't add new tasks")
	}

	var ok bool
	var wrkCfg config.Config
	if wrkCfg, ok = s.Options.Workers[workerName]; !ok {
		return nil, errors.Errorf("Unrecognized worker type '%s'", workerName)
	}

	if !WorkerSupport(workerName, "canReceiveTasks") {
		return nil, errors.New("Worker specified by task does not support receiving tasks over channels")
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
		return nil, errors.Errorf("Unrecognized worker type '%s'", workerName)
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

	return nil, errors.New("No avaliable workers")
}

func (s *Service) createWorker(workerName string, wrkCfg config.Config) (Worker, error) {
	wrk, err := NewWorker(workerName, wrkCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create worker of type '%s'", workerName)
	}

	wrk.SetLogger(s.Logger)
	s.mu.Lock()
	s.workers[workerName] = append(s.workers[workerName], wrk)
	s.outChan <- wrk
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
		s.Logger.Error(errors.Wrap(err, "[Tasker] Unable to handle new task"), nil)
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
	s.outChan = out
	s.watching = true
	s.mu.Unlock()

	go func() {
		for {
			var ctx context.Context
			select {
			case ctx = <-watcher:
			case <-s.stopChan:
				return
			}

			select {
			case <-ctx.Done():
				close(s.stopChan)
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
								s.Logger.Error(errors.Wrapf(err, "[Tasker] Post start phase failed"), nil)
							}
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

		close(s.exitChan)
		s.mur.Lock()
		s.exitChan = nil
		s.mur.Unlock()
		return
	}()

	go func() {
		for {
			select {
			case <-s.doneChan:
				s.stopWatch()
				return
			}
		}
	}()
}

func (s *Service) stopWatch() {
	s.mur.Lock()
	watching := s.watching
	s.mur.Unlock()

	if !watching {
		return
	}

	close(s.outChan)

	s.mu.Lock()
	s.watchChan = nil
	s.outChan = nil
	s.stopChan = nil
	s.watching = false
	s.mu.Unlock()
}

func (s *Service) done() {
	s.mur.Lock()
	runningWorkers := s.runingWorkers
	exiting := s.inExitSequence
	s.mur.Unlock()

	if runningWorkers < 1 && (exiting || !s.Options.Persistent) {
		close(s.doneChan)
		s.mur.Lock()
		s.doneChan = nil
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

	return nil, errors.Errorf("[Tasker] Unrecognized handler \"%v\"", handlerName)
}

// RegisterWorker registers a worker
func RegisterWorker(workerName string, wrkr func(config.Config) (Worker, error), workerSupports map[string]bool) error {
	if _, ok := buildWorkerHandlers[workerName]; ok {
		return errors.Errorf("[Tasker] Worker with name '%s' already registered", workerName)
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
