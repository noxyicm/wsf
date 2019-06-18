package router

const (
	// TYPEDefault represents default router
	TYPEDefault = "default"
)

func init() {
	Register(TYPEDefault, NewDefaultRouter)
}

// Default is a default router
type Default struct {
	router
}

// NewDefaultRouter creates new default router
func NewDefaultRouter(cfg *Config) (ri Interface, err error) {
	r := &Default{}
	r.routes = make(map[string]RouteInterface)
	r.globalParams = make(map[string]interface{})
	return r, nil
}
