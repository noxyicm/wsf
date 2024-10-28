package service

import "github.com/noxyicm/wsf/context"

// Interface is a service interface
type Interface interface {
	Priority() int
	AddListener(func(event int, ctx Event))
	Serve(ctx context.Context) error
	Stop()
}
