package acl

import (
	"wsf/config"
	"wsf/errors"
)

const (
	// TYPEDefaultRoleRegistry resource name
	TYPEDefaultRoleRegistry = "default"
)

var (
	buildRoleRegistryHandlers = map[string]func(*RoleRegistryConfig) (RoleRegistry, error){}
)

func init() {
	RegisterRoleRegistry(TYPEDefaultRoleRegistry, NewRoleRegistryDefault)
}

// RoleRegistry defines acl role registry
type RoleRegistry interface {
	All() []Role
	Count() int
	Add(role Role, parents []string) error
	Has(roleID string) bool
	Get(roleID string) (Role, error)
	Parents(roleID string) []Role
	Inherits(roleID string, inheritID string, onlyParents bool) bool
	Remove(roleID string) error
	RemoveAll()
}

// DefaultRoleRegistry is a default acl role registry
type DefaultRoleRegistry struct {
	Roles map[string]*RolePack
}

// All returns all registered roles
func (rr *DefaultRoleRegistry) All() []Role {
	s := make([]Role, len(rr.Roles))
	i := 0
	for _, role := range rr.Roles {
		s[i] = role.Instance
		i++
	}

	return s
}

// Count returns number of registered roles
func (rr *DefaultRoleRegistry) Count() int {
	return len(rr.Roles)
}

// Add adds a role to the registry inheriting from parent
func (rr *DefaultRoleRegistry) Add(role Role, parents []string) error {
	roleID := role.Alias()
	if rr.Has(roleID) {
		return errors.Errorf("Role id '%s' already exists in the registry", roleID)
	}

	roleParents := make([]Role, 0)
	if len(parents) > 0 {
		for _, parentID := range parents {
			roleParent, err := rr.Get(parentID)
			if err != nil {
				return errors.Errorf("Parent Role id '%s' does not exist", parentID)
			}

			roleParents = append(roleParents, roleParent)
		}
	}

	rr.Roles[roleID] = NewRolePack(role, roleParents)

	for _, parent := range roleParents {
		rr.Roles[parent.Alias()].AddChild(rr.Roles[roleID])
	}
	return nil
}

// Has return true if role registered within registry
func (rr *DefaultRoleRegistry) Has(roleID string) bool {
	if _, ok := rr.Roles[roleID]; ok {
		return true
	}

	return false
}

// Get returns a role from registry
func (rr *DefaultRoleRegistry) Get(roleID string) (Role, error) {
	if !rr.Has(roleID) {
		return nil, errors.Errorf("Role '%s' not found", roleID)
	}

	return rr.Roles[roleID].Instance, nil
}

// Parents returns registered role parents
func (rr *DefaultRoleRegistry) Parents(roleID string) []Role {
	if !rr.Has(roleID) {
		return []Role{}
	}

	s := make([]Role, len(rr.Roles[roleID].Parents))
	i := 0
	for _, parent := range rr.Roles[roleID].Parents {
		s[i] = parent
		i++
	}

	return s
}

// Inherits returns true if role inherits from inherit
func (rr *DefaultRoleRegistry) Inherits(roleID string, inheritID string, onlyParents bool) bool {
	if !rr.Has(roleID) || !rr.Has(inheritID) {
		return false
	}

	doesinherits := false
	if _, ok := rr.Roles[roleID].Parents[inheritID]; ok {
		doesinherits = true
	}

	if doesinherits || onlyParents {
		return doesinherits
	}

	for parentID := range rr.Roles[roleID].Parents {
		if rr.Inherits(parentID, inheritID, onlyParents) {
			return true
		}
	}

	return false
}

// Remove removes role from registry
func (rr *DefaultRoleRegistry) Remove(roleID string) error {
	if !rr.Has(roleID) {
		return errors.Errorf("Role '%s' not found", roleID)
	}

	for _, child := range rr.Roles[roleID].Children {
		child.UnsetParent(roleID)
	}

	for parentID := range rr.Roles[roleID].Parents {
		rr.Roles[parentID].UnsetChild(roleID)
	}

	delete(rr.Roles, roleID)
	return nil
}

// RemoveAll clears the registry
func (rr *DefaultRoleRegistry) RemoveAll() {
	rr.Roles = make(map[string]*RolePack)
}

// NewRoleRegistry creates a new acl role registry of type typ
func NewRoleRegistry(typ string, options *RoleRegistryConfig) (RoleRegistry, error) {
	if f, ok := buildRoleRegistryHandlers[typ]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized acl role registry type '%s'", typ)
}

// NewRoleRegistryFromConfig creates a new acl role registry from config
func NewRoleRegistryFromConfig(options config.Config) (RoleRegistry, error) {
	cfg := &RoleRegistryConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildRoleRegistryHandlers[cfg.Type]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized acl role registry type '%s'", cfg.Type)
}

// RegisterRoleRegistry registers a handler for acl role registry creation
func RegisterRoleRegistry(typ string, handler func(*RoleRegistryConfig) (RoleRegistry, error)) {
	buildRoleRegistryHandlers[typ] = handler
}

// NewRoleRegistryDefault creates a new default acl role registry
func NewRoleRegistryDefault(options *RoleRegistryConfig) (RoleRegistry, error) {
	return &DefaultRoleRegistry{
		Roles: make(map[string]*RolePack),
	}, nil
}
