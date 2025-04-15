package controller

import (
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/view"
)

const (
	// TYPEHelperURL represents URL action helper
	TYPEHelperURL = "url"
)

func init() {
	RegisterHelper(TYPEHelperURL, NewURLHelper)
}

// URL is a action helper that handles urls
type URL struct {
	name            string
	View            view.Interface
	throwExceptions bool
}

// Name returns helper name
func (h *URL) Name() string {
	return h.name
}

// Init the helper
func (h *URL) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *URL) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *URL) PostDispatch(ctx context.Context) error {
	return nil
}

// Assemble returns an url
func (h *URL) Assemble(ctx context.Context, params map[string]interface{}, name string, reset bool, encode bool) string {
	url, err := Router().Assemble(ctx, params, name, reset, encode)
	if err != nil && h.throwExceptions {
		panic(err)
	}

	return url
}

// NewURLHelper creates new URL action helper
func NewURLHelper(name string) (HelperInterface, error) {
	return &URL{
		name:            name,
		throwExceptions: false,
	}, nil
}
