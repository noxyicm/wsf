package service

import "wsf/context"

// Interface is a service interface
type Interface interface {
	Priority() int
	AddListener(func(event int, ctx interface{}))
	Serve(ctx context.Context) error
	Stop()
}
