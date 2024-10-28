package acl

import "github.com/noxyicm/wsf/errors"

const (
	// TYPEResourceDefault resource name
	TYPEResourceDefault = "default"
)

var (
	buildResourceHandlers = map[string]func(string) (Resource, error){}
)

func init() {
	RegisterResource(TYPEResourceDefault, NewResourceDefault)
}

// Resource defines an acl resource
type Resource interface {
	ID() string
}

// ResourceDefault is a default acl resource
type ResourceDefault struct {
	id string
}

// ID returns resource id
func (r *ResourceDefault) ID() string {
	return r.id
}

// NewResource creates a new acl resource of type typ
func NewResource(typ string, name string) (Resource, error) {
	if f, ok := buildResourceHandlers[typ]; ok {
		return f(name)
	}

	return nil, errors.Errorf("Unrecognized acl resource type '%s'", typ)
}

// NewResourceDefault creates a new role of type default
func NewResourceDefault(name string) (Resource, error) {
	return &ResourceDefault{
		id: name,
	}, nil
}

// RegisterResource registers a handler for acl resource creation
func RegisterResource(typ string, handler func(string) (Resource, error)) {
	buildResourceHandlers[typ] = handler
}

// ResourcePack holds an information for acl resource
type ResourcePack struct {
	Instance Resource
	Parent   Resource
	Children map[string]*ResourcePack
}

// AddChild adds a new child to role
func (rp *ResourcePack) AddChild(res *ResourcePack) {
	rp.Children[res.Instance.ID()] = res
}

// UnsetChild removes a role from children list
func (rp *ResourcePack) UnsetChild(resourceID string) {
	if _, ok := rp.Children[resourceID]; ok {
		delete(rp.Children, resourceID)
	}
}

// NewResourcePack creates and initializes a new resource pack
func NewResourcePack(inst Resource, parent Resource) *ResourcePack {
	return &ResourcePack{
		Instance: inst,
		Parent:   parent,
		Children: make(map[string]*ResourcePack),
	}
}
