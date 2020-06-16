package acl

import "wsf/errors"

const (
	// TYPERoleDefault resource name
	TYPERoleDefault = "default"
)

var (
	buildRoleHandlers = map[string]func(string) (Role, error){}
)

func init() {
	RegisterRole(TYPERoleDefault, NewRoleDefault)
}

// Role defines an acl role
type Role interface {
	ID() string
}

// DefaultRole is a default acl role
type DefaultRole struct {
	id string
}

// ID returns role identifier
func (r *DefaultRole) ID() string {
	return r.id
}

// NewRole creates a new role of type typ
func NewRole(typ string, name string) (Role, error) {
	if f, ok := buildRoleHandlers[typ]; ok {
		return f(name)
	}

	return nil, errors.Errorf("Unrecognized acl role type '%s'", typ)
}

// NewRoleDefault creates a new role of type default
func NewRoleDefault(rolename string) (Role, error) {
	return &DefaultRole{
		id: rolename,
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
	rp.Children[role.Instance.ID()] = role
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
		rp.Parents[parent.ID()] = parent
	}

	return rp
}

// RegisterRole registers a handler for acl role creation
func RegisterRole(typ string, handler func(string) (Role, error)) {
	buildRoleHandlers[typ] = handler
}
