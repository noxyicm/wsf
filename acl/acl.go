package acl

import (
	"strings"
	"wsf/config"
	"wsf/errors"
)

const (
	// TYPEAllow defines rule type: allow
	TYPEAllow = "TYPE_ALLOW"

	// TYPEDeny defines rule type: deny
	TYPEDeny = "TYPE_DENY"

	// OPAdd defines operation: add
	OPAdd = "OP_ADD"

	// OPRemove defines operation: remove
	OPRemove = "OP_REMOVE"

	// TYPEDefault resource name
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(config.Config) (Interface, error){}

	inst Interface
)

func init() {
	Register(TYPEDefault, NewDefault)
}

// Interface defines the acl resource
type Interface interface {
	Init(options *Config) (bool, error)
	Priority() int
	Enabled() bool
	AddRole(role Role, parents []string) error
	HasRole(roleID string) bool
	Role(roleID string) (Role, error)
	InheritsRole(roleID string, inheritID string, onlyParents bool) bool
	RemoveRole(roleID string) error
	RemoveRoleAll()
	AddResource(res Resource, parentID string) error
	Has(resID string) bool
	Get(resID string) (Resource, error)
	Inherits(resID string, inheritID string, onlyParent bool) bool
	Remove(resID string) error
	RemoveAll()
	Allow(roleID string, resourceID string, privileges []string, assert Assert) error
	Deny(roleID string, resourceID string, privileges []string, assert Assert) error
	RemoveAllow(roleID string, resourceID string, privileges []string) error
	RemoveDeny(roleID string, resourceID string, privileges []string) error
	SetRule(operation string, typ string, roleID string, resourceID string, privileges []string, assert Assert) error
	IsAllowed(roleID string, resourceID string, privilege string) bool
	RoleRegistry() RoleRegistry
	Roles() []string
	Resources() []string
}

// Default is a default acl
type Default struct {
	Options     *Config
	RoleReg     RoleRegistry
	ResourceReg map[string]*ResourcePack
	Rules       Rule
}

// Priority returns resource initialization priority
func (a *Default) Priority() int {
	return a.Options.Priority
}

// Init resource
func (a *Default) Init(options *Config) (bool, error) {
	a.Options = options
	return true, nil
}

// Enabled returns true if ACL is enabled
func (a *Default) Enabled() bool {
	return a.Options.Enable
}

// AddRole adds a Role having an identifier unique to the registry
//
// The parents parameter is a slice of role identifiers, indicating the Roles
// from which the newly added Role will directly inherit.
//
// In order to resolve potential ambiguities with conflicting rules inherited
// from different parents, the most recently added parent takes precedence over
// parents that were previously added. In other words, the first parent added
// will have the least priority, and the last parent added will have the
// highest priority.
func (a *Default) AddRole(role Role, parents []string) error {
	return a.RoleReg.Add(role, parents)
}

// Role returns the identified Role
func (a *Default) Role(roleID string) (Role, error) {
	return a.RoleReg.Get(roleID)
}

// HasRole returns true if and only if the Role exists in the registry
func (a *Default) HasRole(roleID string) bool {
	return a.RoleReg.Has(roleID)
}

// InheritsRole returns true if and only if role inherits from inherit
// If onlyParents is true, then role must inherit directly from
// inherit in order to return true. By default, this method looks
// through the entire inheritance DAG to determine whether role
// inherits from inherit through its ancestor Roles.
func (a *Default) InheritsRole(roleID string, inheritID string, onlyParents bool) bool {
	return a.RoleReg.Inherits(roleID, inheritID, onlyParents)
}

// RemoveRole removes the Role from the registry
func (a *Default) RemoveRole(roleID string) error {
	if err := a.RoleReg.Remove(roleID); err != nil {
		return err
	}

	for roleIDCurrent, rules := range a.Rules.Global().All() {
		if roleID == roleIDCurrent {
			rules.Unset(roleIDCurrent)
		}
	}

	for _, visitor := range a.Rules.All() {
		for roleIDCurrent, rules := range visitor.All() {
			if roleID == roleIDCurrent {
				rules.Unset(roleIDCurrent)
			}
		}
	}

	return nil
}

// RemoveRoleAll removes all Roles from the registry
func (a *Default) RemoveRoleAll() {
	a.RoleReg.RemoveAll()
	a.Rules.ClearGlobal()
	a.Rules.ClearAll()
}

// AddResource adds a Resource having an identifier unique to the ACL
// The parent parameter may be a reference to,
// the existing Resource from which the newly added Resource will inherit
func (a *Default) AddResource(res Resource, parentID string) error {
	resourceID := res.ID()
	if a.Has(resourceID) {
		return errors.Errorf("Resource id '%s' already exists in the ACL", resourceID)
	}

	var resourceParent Resource
	var err error
	if parentID != "" {
		resourceParent, err = a.Get(parentID)
		if err != nil {
			return errors.Errorf("Parent Resource id '%s' does not exist", parentID)
		}
	}

	a.ResourceReg[resourceID] = NewResourcePack(res, resourceParent)
	if resourceParent != nil {
		a.ResourceReg[resourceParent.ID()].AddChild(a.ResourceReg[resourceID])
	}
	return nil
}

// Get returns the identified Resource
func (a *Default) Get(resID string) (Resource, error) {
	if !a.Has(resID) {
		return nil, errors.Errorf("Resource '%s' not found", resID)
	}

	return a.ResourceReg[resID].Instance, nil
}

// Has returns true if and only if the Resource exists in the ACL
func (a *Default) Has(resID string) bool {
	if _, ok := a.ResourceReg[resID]; ok {
		return true
	}

	return false
}

// Inherits returns true if and only if res inherits from inherit
func (a *Default) Inherits(resID string, inheritID string, onlyParent bool) bool {
	if !a.Has(resID) {
		return false
	}

	if !a.Has(inheritID) {
		return false
	}

	var parentID string
	if resp, ok := a.ResourceReg[resID]; ok && resp.Parent != nil {
		parentID = resp.Parent.ID()
		if inheritID == parentID {
			return true
		} else if onlyParent {
			return false
		}
	} else {
		return false
	}

	for a.ResourceReg[parentID].Parent != nil {
		parentID = a.ResourceReg[parentID].Parent.ID()
		if inheritID == parentID {
			return true
		}
	}

	return false
}

// Remove removes a Resource and all of its children
func (a *Default) Remove(resID string) error {
	if !a.Has(resID) {
		return errors.Errorf("Resource '%s' not found", resID)
	}

	resourcesRemoved := []string{resID}
	if a.ResourceReg[resID].Parent != nil {
		a.ResourceReg[a.ResourceReg[resID].Parent.ID()].UnsetChild(resID)
	}

	for childID := range a.ResourceReg[resID].Children {
		a.Remove(childID)
		resourcesRemoved = append(resourcesRemoved, childID)
	}

	for _, resourceIDRemoved := range resourcesRemoved {
		for resourceIDCurrent := range a.Rules.All() {
			if resourceIDRemoved == resourceIDCurrent {
				a.Rules.Unset(resourceIDCurrent)
			}
		}
	}

	delete(a.ResourceReg, resID)
	return nil
}

// RemoveAll removes all Resources
func (a *Default) RemoveAll() {
	for resourceID := range a.ResourceReg {
		for resourceIDCurrent := range a.Rules.All() {
			if resourceID == resourceIDCurrent {
				a.Rules.Unset(resourceIDCurrent)
			}
		}
	}

	a.ResourceReg = make(map[string]*ResourcePack)
}

// Allow adds an "allow" rule to the ACL
func (a *Default) Allow(roleID string, resourceID string, privileges []string, assert Assert) error {
	return a.SetRule(OPAdd, TYPEAllow, roleID, resourceID, privileges, assert)
}

// Deny adds a "deny" rule to the ACL
func (a *Default) Deny(roleID string, resourceID string, privileges []string, assert Assert) error {
	return a.SetRule(OPAdd, TYPEDeny, roleID, resourceID, privileges, assert)
}

// RemoveAllow removes "allow" permissions from the ACL
func (a *Default) RemoveAllow(roleID string, resourceID string, privileges []string) error {
	return a.SetRule(OPRemove, TYPEAllow, roleID, resourceID, privileges, nil)
}

// RemoveDeny removes "deny" restrictions from the ACL
func (a *Default) RemoveDeny(roleID string, resourceID string, privileges []string) error {
	return a.SetRule(OPRemove, TYPEDeny, roleID, resourceID, privileges, nil)
}

// SetRule performs operations on ACL rules
// The operation parameter may be either OPAdd or OPRemove, depending on whether the
// user wants to add or remove a rule, respectively:
//
// OPAdd specifics:
//
//     A rule is added that would allow one or more Roles access to [certain $privileges
//     upon] the specified Resource(s).
//
// OPRemove specifics:
//
//     The rule is removed only in the context of the given Roles, Resources, and privileges.
//     Existing rules to which the remove operation does not apply would remain in the
//     ACL.
//
// The typ parameter may be either TYPEAllow or TYPEDeny, depending on whether the
// rule is intended to allow or deny permission, respectively.
//
// The roleID and resourceID parameters is the string identifiers for
// existing Resource/Role indicating the Resource and Role to which the rule applies. If either
// roleID or resourceID is empty, then the rule applies to all Roles or all Resources, respectively.
// Both may be empty in order to work with the default rule of the ACL.
//
// The privileges parameter may be used to further specify that the rule applies only
// to certain privileges upon the Resource(s) in question. This may be specified to be a single
// privilege with a string, and multiple privileges may be specified as an array of strings.
//
// If assert is provided, then its Assert() method must return true in order for
// the rule to apply. If assert is provided with roleID, resourceID, and privileges all
// empty, then a rule having a type of:
//
//     TYPEAllow will imply a type of TYPEDeny, and
//
//     TYPEDeny will imply a type of TYPEAllow
//
// when the rule's assertion fails. This is because the ACL needs to provide expected
// behavior when an assertion upon the default ACL rule fails.
func (a *Default) SetRule(operation string, typ string, roleID string, resourceID string, privileges []string, assert Assert) error {
	typ = strings.ToUpper(typ)
	if typ != TYPEAllow && typ != TYPEDeny {
		return errors.Errorf("Unable to set rule: Unsupported rule type, must be either '%s' or '%s'", TYPEAllow, TYPEDeny)
	}

	var err error
	var role Role
	if roleID != "" {
		if role, err = a.RoleReg.Get(roleID); err != nil {
			return errors.Wrap(err, "Unable to set rule")
		}
	}

	var res Resource
	if resourceID != "" {
		if res, err = a.Get(resourceID); err != nil {
			return errors.Wrap(err, "Unable to set rule")
		}
	}

	switch operation {
	case OPAdd:
		rules := a.getRules(res, role, true)
		if len(privileges) == 0 {
			rules.Global().SetType(typ)
			rules.Global().SetAssert(assert)
		} else {
			for i := range privileges {
				if rules.Get(privileges[i]) == nil {
					rl := rules.Create(privileges[i])
					rl.SetType(typ)
					rl.SetAssert(assert)
				} else {
					rules.Get(privileges[i]).SetType(typ)
					rules.Get(privileges[i]).SetAssert(assert)
				}
			}
		}

	case OPRemove:
		if res != nil {
			rules := a.getRules(res, role, false)
			if rules == nil {
				return nil
			}

			if len(privileges) == 0 {
				if rules.Global() != nil && typ == rules.Global().Type() {
					rules.ClearGlobal()
				}
			} else {
				for i := range privileges {
					if rules.Get(privileges[i]) != nil && typ == rules.Get(privileges[i]).Type() {
						rules.Unset(privileges[i])
					}
				}
			}
		} else {
			rules := a.getRules(res, role, true)
			if len(privileges) == 0 {
				if role == nil {
					if typ == rules.Global().Type() {
						rules.ClearGlobal()
					}
				} else if rules.Global() != nil && typ == rules.Global().Type() {
					rules.UnsetGlobal()
				}
			} else {
				for i := range privileges {
					if rules.Get(privileges[i]) != nil && typ == rules.Get(privileges[i]).Type() {
						rules.Unset(privileges[i])
					}
				}
			}

			for _, allres := range a.ResourceReg {
				rules := a.getRules(allres.Instance, role, true)
				if rules == nil {
					continue
				}

				if len(privileges) == 0 {
					if rules.Global() != nil && typ == rules.Global().Type() {
						rules.UnsetGlobal()
					}
				} else {
					for i := range privileges {
						if rules.Get(privileges[i]) != nil && typ == rules.Get(privileges[i]).Type() {
							rules.Unset(privileges[i])
						}
					}
				}
			}
		}

	default:
		return errors.Errorf("Unable to set rule: Unsupported operation, must be either '%s' or '%s'", OPAdd, OPRemove)
	}

	return nil
}

// IsAllowed returns true if and only if the Role has access to the Resource
//
// The roleID and resourceID parameters is the string identifiers for
// an existing Resource and Role combination.
//
// If either roleID or resourceID is empty, then the query applies to all Roles or all Resources,
// respectively. Both may be empty to query whether the ACL has a "blacklist" rule
// (allow everything to all). By default, Acl creates a "whitelist" rule (deny
// everything to all), and this method would return false unless this default has
// been overridden (i.e., by executing acl.Allow("", "", []string{}, nil)).
//
// If a privilege is not provided, then this method returns false if and only if the
// Role is denied access to at least one privilege upon the Resource. In other words, this
// method returns true if and only if the Role is allowed all privileges on the Resource.
//
// This method checks Role inheritance using a depth-first traversal of the Role registry.
// The highest priority parent (i.e., the parent most recently added) is checked first,
// and its respective parents are checked similarly before the lower-priority parents of
// the Role are checked.
func (a *Default) IsAllowed(roleID string, resourceID string, privilege string) bool {
	//var isAllowedRole string
	//var isAllowedResource string
	//var isAllowedPrivilege string
	var role Role
	var res Resource
	var err error

	if roleID != "" {
		//isAllowedRole = roleID
		role, err = a.RoleReg.Get(roleID)
		if err != nil {
			//	isAllowedRole = ""
		}
	}

	if resourceID != "" {
		//isAllowedResource = resourceID
		res, err = a.Get(resourceID)
		if err != nil {
			//	isAllowedResource = ""
		}
	}

	if privilege == "" {
		for {
			// depth-first search on $role if it is not 'allRoles' pseudo-parent
			if role != nil {
				if allowed, err := a.roleDFSAllPrivileges(role, res, privilege); err == nil {
					return allowed
				}
			}

			// look for rule on 'allRoles' psuedo-parent
			rules := a.getRules(res, nil, false)
			if rules != nil {
				for privilege := range rules.All() {
					if ruleTypeOnePrivilege, err := a.getRuleType(res, nil, privilege); err == nil && ruleTypeOnePrivilege == TYPEDeny {
						return false
					}
				}

				if ruleTypeAllPrivileges, err := a.getRuleType(res, nil, ""); err == nil {
					return ruleTypeAllPrivileges == TYPEAllow
				}
			}

			// try next Resource
			res = a.ResourceReg[res.ID()].Parent
		}
	} else {
		//isAllowedPrivilege = privilege
		// query on one privilege
		for {
			// depth-first search on $role if it is not 'allRoles' pseudo-parent
			if role != nil {
				if allowed, err := a.roleDFSOnePrivilege(role, res, privilege); err == nil {
					return allowed
				}
			}

			// look for rule on 'allRoles' pseudo-parent
			if ruleType, err := a.getRuleType(res, nil, privilege); err == nil {
				return ruleType == TYPEAllow
			}

			if ruleTypeAllPrivileges, err := a.getRuleType(res, nil, ""); err == nil {
				return ruleTypeAllPrivileges == TYPEAllow
			}

			// try next Resource
			res = a.ResourceReg[res.ID()].Parent
		}
	}
}

// RoleRegistry returns the Role registry for this ACL
//
// If no Role registry has been created yet, a new default Role registry
// is created and returned.
func (a *Default) RoleRegistry() RoleRegistry {
	if a.RoleReg == nil {
		a.RoleReg, _ = NewRoleRegistryFromConfig(a.Options.Role)
	}

	return a.RoleReg
}

// roleDFSAllPrivileges performs a depth-first search of the Role DAG, starting at role, in order to find a rule
// allowing/denying role access to all privileges upon resource
//
// This method returns true if a rule is found and allows access. If a rule exists and denies access,
// then this method returns false. If no applicable rule is found, then this method returns error.
func (a *Default) roleDFSAllPrivileges(role Role, res Resource, privilege string) (bool, error) {
	dfs := NewDFS()
	if allowed, err := a.roleDFSVisitAllPrivileges(role, res, dfs); err == nil {
		return allowed, nil
	}

	for {
		role := dfs.Pop()
		if role == nil {
			break
		}

		if !dfs.IsVisited(role.ID()) {
			if allowed, err := a.roleDFSVisitAllPrivileges(role, res, dfs); err == nil {
				return allowed, nil
			}
		}
	}

	return false, errors.New("Not found")
}

// roleDFSVisitAllPrivileges visits an role in order to look for a rule allowing/denying role access to all privileges upon resource
//
// This method returns true if a rule is found and allows access. If a rule exists and denies access,
// then this method returns false. If no applicable rule is found, then this method returns error.
//
// This method is used by the internal depth-first search algorithm and may modify the DFS data structure.
func (a *Default) roleDFSVisitAllPrivileges(role Role, res Resource, dfs *DFS) (bool, error) {
	if dfs == nil {
		return false, errors.New("dfs parameter may not be nil")
	}

	rules := a.getRules(res, role, false)
	if rules != nil {
		for privilege := range rules.All() {
			if ruleTypeOnePrivilege, err := a.getRuleType(res, role, privilege); err == nil && ruleTypeOnePrivilege == TYPEDeny {
				return false, nil
			}
		}

		if ruleTypeAllPrivileges, err := a.getRuleType(res, role, ""); err == nil {
			return ruleTypeAllPrivileges == TYPEAllow, nil
		}
	}

	dfs.Visit(role.ID())
	for _, roleParent := range a.RoleReg.Parents(role.ID()) {
		dfs.Push(roleParent)
	}

	return false, errors.New("Not found")
}

// roleDFSOnePrivilege performs a depth-first search of the Role DAG, starting at role, in order to find a rule
// allowing/denying role access to a privilege upon resource
//
// This method returns true if a rule is found and allows access. If a rule exists and denies access,
// then this method returns false. If no applicable rule is found, then this method returns error.
func (a *Default) roleDFSOnePrivilege(role Role, res Resource, privilege string) (bool, error) {
	if privilege == "" {
		return false, errors.New("privilege parameter may not be empty")
	}

	dfs := NewDFS()
	if allowed, err := a.roleDFSVisitOnePrivilege(role, res, privilege, dfs); err == nil {
		return allowed, nil
	}

	for {
		role := dfs.Pop()
		if role == nil {
			break
		}

		if !dfs.IsVisited(role.ID()) {
			if allowed, err := a.roleDFSVisitOnePrivilege(role, res, privilege, dfs); err == nil {
				return allowed, nil
			}
		}
	}

	return false, errors.New("Not found")
}

// roleDFSVisitOnePrivilege visits an role in order to look for a rule allowing/denying role access to a privilege upon resource
//
// This method returns true if a rule is found and allows access. If a rule exists and denies access,
// then this method returns false. If no applicable rule is found, then this method returns error.
//
// This method is used by the internal depth-first search algorithm and may modify the DFS data structure.
func (a *Default) roleDFSVisitOnePrivilege(role Role, res Resource, privilege string, dfs *DFS) (bool, error) {
	if privilege == "" {
		return false, errors.New("privilege parameter may not be empty")
	}

	if dfs == nil {
		return false, errors.New("dfs parameter may not be nil")
	}

	if ruleTypeOnePrivilege, err := a.getRuleType(res, role, privilege); err == nil {
		return ruleTypeOnePrivilege == TYPEAllow, nil
	}

	if ruleTypeAllPrivileges, err := a.getRuleType(res, role, ""); err == nil {
		return ruleTypeAllPrivileges == TYPEAllow, nil
	}

	dfs.Visit(role.ID())
	for _, roleParent := range a.RoleReg.Parents(role.ID()) {
		dfs.Push(roleParent)
	}

	return false, errors.New("Not found")
}

// getRuleType returns the rule type associated with the specified Resource, Role, and privilege
// combination.
//
// If a rule does not exist or its attached assertion fails, which means that
// the rule is not applicable, then this method returns error. Otherwise, the
// rule type applies and is returned as either TYPEAllow or TYPEDeny.
//
// If res or role is null, then this means that the rule must apply to
// all Resources or Roles, respectively.
//
// If privilege is empty, then the rule must apply to all privileges.
//
// If all three parameters are null, then the default ACL rule type is returned,
// based on whether its assertion method passes.
func (a *Default) getRuleType(res Resource, role Role, privilege string) (string, error) {
	rules := a.getRules(res, role, false)
	if rules == nil {
		return "", errors.New("Not applicable")
	}

	var rule Rule
	if privilege == "" {
		if rules.Global() != nil {
			rule = rules.Global()
		} else {
			return "", errors.New("Not applicable")
		}
	} else if rules.Get(privilege) == nil {
		return "", errors.New("Not applicable")
	} else {
		rule = rules.Get(privilege)
	}

	var assertionValue bool
	if rule.Assert() != nil {
		panic("not implemented")
	}

	if rule.Assert() == nil || assertionValue {
		return rule.Type(), nil
	} else if res != nil || role != nil || privilege != "" {
		return "", errors.New("Not applicable")
	} else if rule.Type() == TYPEAllow {
		return TYPEAllow, nil
	}

	return TYPEDeny, nil
}

// getRules returns the rules associated with a Resource and a Role, or null if no such rules exist
//
// If either $resource or role is null, this means that the rules returned are for all Resources or all Roles,
// respectively. Both can be null to return the default rule set for all Resources and all Roles.
//
// If the create parameter is true, then a rule set is first created and then returned to the caller.
func (a *Default) getRules(res Resource, role Role, create bool) Rule {
	var visitor Rule

	if res == nil {
		visitor = a.Rules.Global()
		if visitor == nil {
			if !create {
				return nil
			}

			a.Rules.ClearGlobal()
			visitor = a.Rules.Global()
		}
	} else {
		resourceID := res.ID()
		if !a.Rules.Has(resourceID) {
			if !create {
				return nil
			}

			a.Rules.Create(resourceID)
		}

		visitor = a.Rules.Get(resourceID)
	}

	if role == nil {
		if visitor.Global() == nil {
			if !create {
				return nil
			}

			visitor.ClearGlobal()
		}

		return visitor.Global()
	}

	roleID := role.ID()
	if !visitor.Has(roleID) {
		if !create {
			return nil
		}

		visitor.Create(roleID)
	}

	return visitor.Get(roleID)
}

// Roles returns a slice of registered roles.
//
// Note that this method does not return instances of registered roles,
// but only the role identifiers.
func (a *Default) Roles() []string {
	s := make([]string, a.RoleReg.Count())
	i := 0
	for _, role := range a.RoleReg.All() {
		s[i] = role.ID()
		i++
	}

	return s
}

// Resources returns a slice of registered resources.
//
// Note that this method does not return instances of registered resources,
// but only the resource identifiers.
func (a *Default) Resources() []string {
	s := make([]string, len(a.ResourceReg))
	i := 0
	for resourceID := range a.ResourceReg {
		s[i] = resourceID
		i++
	}

	return s
}

// NewDefault creates a new acl of type default
func NewDefault(options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	a := &Default{
		Options:     cfg,
		ResourceReg: make(map[string]*ResourcePack),
		Rules:       NewRule(),
	}

	rr, err := NewRoleRegistryFromConfig(cfg.Role)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create ACL instance")
	}
	a.RoleReg = rr

	return a, nil
}

// DFS struct
type DFS struct {
	Visited map[string]bool
	Stack   []Role
}

// Pop element from the top of the stack
func (d *DFS) Pop() Role {
	if len(d.Stack) == 0 {
		return nil
	}

	role := d.Stack[0]
	d.Stack[0] = nil
	d.Stack = d.Stack[1:len(d.Stack)]

	return role
}

// Push element to the bottom of the stack
func (d *DFS) Push(role Role) {
	d.Stack = append(d.Stack, role)
}

// IsVisited return true if dfs has record of roleID
func (d *DFS) IsVisited(roleID string) bool {
	if _, ok := d.Visited[roleID]; ok {
		return true
	}

	return false
}

// Visit sets a record for visiting roleID
func (d *DFS) Visit(roleID string) {
	d.Visited[roleID] = true
}

// NewDFS creates and initializes a new DFS struct
func NewDFS() *DFS {
	return &DFS{
		Visited: make(map[string]bool),
		Stack:   make([]Role, 0),
	}
}

// NewACL creates a new ACL resource of type typ
func NewACL(typ string, options config.Config) (Interface, error) {
	if f, ok := buildHandlers[typ]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized ACL type '%s'", typ)
}

// Register registers a handler for acl creation
func Register(typ string, handler func(config.Config) (Interface, error)) {
	buildHandlers[typ] = handler
}

// SetInstance sets global instance
func SetInstance(a Interface) {
	inst = a
}

// Instance returns global instance
func Instance() Interface {
	return inst
}
