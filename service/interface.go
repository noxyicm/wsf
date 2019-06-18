package service

// Interface is a service interface
type Interface interface {
	Priority() int
	Serve() error
	Stop()
}
