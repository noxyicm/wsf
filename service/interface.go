package service

// Interface is a service interface
type Interface interface {
	Priority() int
	AddListener(func(event int, ctx interface{}))
	Serve() error
	Stop()
}
