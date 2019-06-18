package loader

// ResourceLoader Structure
type ResourceLoader struct {
}

// NewResourceLoader ResourceLoader Constructor
func NewResourceLoader() (r *ResourceLoader, err error) {
	r = new(ResourceLoader)
	return r, nil
}
