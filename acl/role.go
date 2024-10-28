package acl

import "github.com/noxyicm/wsf/errors"

const (
	// TYPERoleDefault resource name
	TYPERoleDefault = "default"
)

var (
	buildRoleHandlers = map[string]func(int, string) (Role, error){}
)

func init() {
	RegisterRole(TYPERoleDefault, NewRoleDefault)
}

// Role defines an acl role
type Role interface {
	ID() int
	Alias() string
	Name() string
	SetName(name string)
	Description() string
	SetDescription(desc string)
}

// DefaultRole is a default acl role
type DefaultRole struct {
	id          int
	alias       string
	name        string
	description string
}

// ID returns role identifier
func (r *DefaultRole) ID() int {
	return r.id
}

// Alias returns role string identifier
func (r *DefaultRole) Alias() string {
	return r.alias
}

// Name returns role name
func (r *DefaultRole) Name() string {
	return r.name
}

// SetName sets role name
func (r *DefaultRole) SetName(name string) {
	r.name = name
}

// Description returns role description
func (r *DefaultRole) Description() string {
	return r.description
}

// SetDescription sets role description
func (r *DefaultRole) SetDescription(desc string) {
	r.description = desc
}

// NewRole creates a new role of type typ
func NewRole(typ string, id int, alias string) (Role, error) {
	if f, ok := buildRoleHandlers[typ]; ok {
		return f(id, alias)
	}

	return nil, errors.Errorf("Unrecognized acl role type '%s'", typ)
}

// NewRoleDefault creates a new role of type default
func NewRoleDefault(id int, rolename string) (Role, error) {
	return &DefaultRole{
		id:    id,
		alias: rolename,
	}, nil
}

// RolePack holds an information for acl role
type RolePack struct {
	Instance Role
	Parents  map[string]Role
	Children map[string]*RolePack
}

// AddChild adds a new child to role
func (rp *RolePack) AddChild(role *RolePack) {
	rp.Children[role.Instance.Alias()] = role
}

// UnsetChild removes a role from children list
func (rp *RolePack) UnsetChild(roleID string) {
	if _, ok := rp.Children[roleID]; ok {
		delete(rp.Children, roleID)
	}
}

// UnsetParent removes a role from parents list
func (rp *RolePack) UnsetParent(roleID string) {
	if _, ok := rp.Parents[roleID]; ok {
		delete(rp.Parents, roleID)
	}
}

// NewRolePack creates and initializes a new role pack
func NewRolePack(inst Role, parents []Role) *RolePack {
	rp := &RolePack{
		Instance: inst,
		Parents:  make(map[string]Role),
		Children: make(map[string]*RolePack),
	}

	for _, parent := range parents {
		rp.Parents[parent.Alias()] = parent
	}

	return rp
}

// RegisterRole registers a handler for acl role creation
func RegisterRole(typ string, handler func(int, string) (Role, error)) {
	buildRoleHandlers[typ] = handler
}
