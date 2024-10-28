package acl

// Rule defines a rule for acl
type Rule interface {
	Global() Rule
	ClearGlobal()
	UnsetGlobal()
	All() map[string]Rule
	Has(key string) bool
	Get(key string) Rule
	Clear(key string)
	ClearAll()
	Unset(key string)
	Create(key string) Rule
	SetType(typ string)
	Type() string
	SetAssert(assrt Assert)
	Assert() Assert
}

type ruleset struct {
	resources Rule
	resource  map[string]Rule
}

// Global returns global instructions
func (rs *ruleset) Global() Rule {
	return rs.resource["allResources"]
}

// ClearGlobal clears global instructions
func (rs *ruleset) ClearGlobal() {
	rs.resource["allResources"] = newResourceRule()
}

// UnsetGlobal removes global instructions
func (rs *ruleset) UnsetGlobal() {
	rs.resource["allResources"] = nil
}

// All returns all specified instructions
func (rs *ruleset) All() map[string]Rule {
	return rs.resource
}

// Has returns true if rule has specific instruction
func (rs *ruleset) Has(key string) bool {
	if _, ok := rs.resource[key]; ok {
		return true
	}

	return false
}

// Get specific instruction
func (rs *ruleset) Get(key string) Rule {
	if rs.Has(key) {
		return rs.resource[key]
	}

	return nil
}

// Clear the instruction
func (rs *ruleset) Clear(key string) {
	if rs.Has(key) {
		rs.resource[key] = newResourceRule()
	}
}

// ClearAll the instructions
func (rs *ruleset) ClearAll() {
	rs.resource = make(map[string]Rule)
}

// Unset the instruction
func (rs *ruleset) Unset(key string) {
	if rs.Has(key) {
		delete(rs.resource, key)
	}
}

// Create a specific instruction
func (rs *ruleset) Create(key string) Rule {
	rs.resource[key] = newResourceRule()
	return rs.resource[key]
}

// SetType sets a rule type
func (rs *ruleset) SetType(typ string) {
}

// Type returns a rule type
func (rs *ruleset) Type() string {
	return ""
}

// SetAssert registeres assert
func (rs *ruleset) SetAssert(assrt Assert) {
}

// Assert returns registered assert
func (rs *ruleset) Assert() Assert {
	return nil
}

type resourcerule struct {
	roles Rule
	role  map[string]Rule
}

// Global returns global instructions
func (rr *resourcerule) Global() Rule {
	return rr.role["allRoles"]
}

// ClearGlobal clears global instructions
func (rr *resourcerule) ClearGlobal() {
	rr.role["allRoles"] = newRoleRule()
}

// UnsetGlobal removes global instructions
func (rr *resourcerule) UnsetGlobal() {
	rr.role["allRoles"] = nil
}

// All returns all specified instructions
func (rr *resourcerule) All() map[string]Rule {
	return rr.role
}

// Has returns true if rule has specific instruction
func (rr *resourcerule) Has(key string) bool {
	if _, ok := rr.role[key]; ok {
		return true
	}

	return false
}

// Get specific instruction
func (rr *resourcerule) Get(key string) Rule {
	if rr.Has(key) {
		return rr.role[key]
	}

	return nil
}

// Clear the instruction
func (rr *resourcerule) Clear(key string) {
	if rr.Has(key) {
		rr.role[key] = newRoleRule()
	}
}

// ClearAll the instructions
func (rr *resourcerule) ClearAll() {
	rr.role = make(map[string]Rule)
}

// Unset the instruction
func (rr *resourcerule) Unset(key string) {
	if rr.Has(key) {
		delete(rr.role, key)
	}
}

// Create a specific instruction
func (rr *resourcerule) Create(key string) Rule {
	rr.role[key] = newRoleRule()
	return rr.role[key]
}

// SetType sets a rule type
func (rr *resourcerule) SetType(typ string) {
}

// Type returns a rule type
func (rr *resourcerule) Type() string {
	return ""
}

// SetAssert registeres assert
func (rr *resourcerule) SetAssert(assrt Assert) {
}

// Assert returns registered assert
func (rr *resourcerule) Assert() Assert {
	return nil
}

type rolerule struct {
	privileges Rule
	privilege  map[string]Rule
}

// Global returns global instructions
func (rlr *rolerule) Global() Rule {
	return rlr.privilege["allPrivileges"]
}

// ClearGlobal clears global instructions
func (rlr *rolerule) ClearGlobal() {
	rlr.privilege["allPrivileges"] = newPrivilegeRule()
}

// UnsetGlobal removes global instructions
func (rlr *rolerule) UnsetGlobal() {
	rlr.privilege["allPrivileges"] = nil
}

// All returns all specified instructions
func (rlr *rolerule) All() map[string]Rule {
	return rlr.privilege
}

// Has returns true if rule has specific instruction
func (rlr *rolerule) Has(key string) bool {
	if _, ok := rlr.privilege[key]; ok {
		return true
	}

	return false
}

// Get specific instruction
func (rlr *rolerule) Get(key string) Rule {
	if rlr.Has(key) {
		return rlr.privilege[key]
	}

	return nil
}

// Clear the instruction
func (rlr *rolerule) Clear(key string) {
	if rlr.Has(key) {
		rlr.privilege[key] = newPrivilegeRule()
	}
}

// ClearAll the instructions
func (rlr *rolerule) ClearAll() {
	rlr.privilege = make(map[string]Rule)
}

// Unset the instruction
func (rlr *rolerule) Unset(key string) {
	if rlr.Has(key) {
		delete(rlr.privilege, key)
	}
}

// Create a specific instruction
func (rlr *rolerule) Create(key string) Rule {
	rlr.privilege[key] = newPrivilegeRule()
	return rlr.privilege[key]
}

// SetType sets a rule type
func (rlr *rolerule) SetType(typ string) {
}

// Type returns a rule type
func (rlr *rolerule) Type() string {
	return ""
}

// SetAssert registeres assert
func (rlr *rolerule) SetAssert(assrt Assert) {
}

// Assert returns registered assert
func (rlr *rolerule) Assert() Assert {
	return nil
}

type privilegerule struct {
	typ    string
	assert Assert
}

// Global returns global instructions
func (pr *privilegerule) Global() Rule {
	return nil
}

// ClearGlobal clears global instructions
func (pr *privilegerule) ClearGlobal() {
}

// UnsetGlobal removes global instructions
func (pr *privilegerule) UnsetGlobal() {
}

// All returns all specified instructions
func (pr *privilegerule) All() map[string]Rule {
	return nil
}

// Has returns true if rule has specific instruction
func (pr *privilegerule) Has(key string) bool {
	return false
}

// Get specific instruction
func (pr *privilegerule) Get(key string) Rule {
	return nil
}

// Clear the instruction
func (pr *privilegerule) Clear(key string) {
	pr.typ = TYPEDeny
	pr.assert = nil
}

// ClearAll the instructions
func (pr *privilegerule) ClearAll() {
	pr.typ = TYPEDeny
	pr.assert = nil
}

// Unset the instruction
func (pr *privilegerule) Unset(key string) {
	pr.typ = TYPEDeny
	pr.assert = nil
}

// Create a specific instruction
func (pr *privilegerule) Create(key string) Rule {
	return nil
}

// SetType sets a rule type
func (pr *privilegerule) SetType(typ string) {
	pr.typ = typ
}

// Type returns a rule type
func (pr *privilegerule) Type() string {
	return pr.typ
}

// SetAssert registeres assert
func (pr *privilegerule) SetAssert(assrt Assert) {
	pr.assert = assrt
}

// Assert returns registered assert
func (pr *privilegerule) Assert() Assert {
	return pr.assert
}

// NewRule creates a new RuleSet
func NewRule() Rule {
	r := &ruleset{
		resources: newResourceRule(),
		resource:  make(map[string]Rule),
	}

	r.Create("allResources").Create("allRoles").Create("allPrivileges")
	return r
}

func newResourceRule() Rule {
	return &resourcerule{
		//roles: newRoleRule(),
		role: make(map[string]Rule),
	}
}

func newRoleRule() Rule {
	return &rolerule{
		//privileges: newPrivilegeRule(),
		privilege: make(map[string]Rule),
	}
}

func newPrivilegeRule() Rule {
	return &privilegerule{
		typ:    TYPEDeny,
		assert: nil,
	}
}

/*array(
'allResources' => array(
	'allRoles' => array(
		'allPrivileges' => array(
			'type'   => self::TYPE_DENY,
			'assert' => null
			),
		'byPrivilegeId' => array()
		),
	'byRoleId' => array()
	),
'byResourceId' => array()
);*/
