package filter

// RuleStack is a stack of rules
type RuleStack struct {
	stack   []interface{}
	ref     map[string]int
	backref map[int]string
}

// Set rules to stack
func (rs *RuleStack) Set(rules map[string]interface{}) *RuleStack {
	rs.stack = make([]interface{}, len(rules))
	i := 0
	for k, v := range rules {
		rs.stack[i] = v
		rs.ref[k] = i
		rs.backref[i] = k
		i++
	}

	return rs
}

// Add rules to stack
func (rs *RuleStack) Add(rules map[string]interface{}) *RuleStack {
	for k, v := range rules {
		rs.stack = append(rs.stack, v)
		rs.ref[k] = len(rs.stack) - 1
		rs.backref[rs.ref[k]] = k
	}

	return rs
}

// Stack returns all rules as slice
func (rs *RuleStack) Stack() []interface{} {
	return rs.stack
}

// StackSpec returns rule spec by stack index
func (rs *RuleStack) StackSpec(index int) string {
	if v, ok := rs.backref[index]; ok {
		return v
	}

	return ""
}

// Rules returns all rules as reference map
func (rs *RuleStack) Rules() map[string]interface{} {
	reles := make(map[string]interface{})
	for k, idx := range rs.ref {
		reles[k] = rs.stack[idx]
	}

	return reles
}

// SpecRules returns all rules as reference map for spec
func (rs *RuleStack) SpecRules(spec string) interface{} {
	if idx, ok := rs.ref[spec]; ok {
		return rs.stack[idx]
	}

	return nil
}

// Rule returns spec rule by its index
func (rs *RuleStack) Rule(spec string, index int) interface{} {
	if idx, ok := rs.ref[spec]; ok {
		switch rs.stack[idx].(type) {
		case []interface{}:
			if len(rs.stack[idx].([]interface{})) < index {
				return rs.stack[idx].([]interface{})[index]
			}

		default:
			return nil
		}
	}

	return nil
}

// SetRule sets rule for a spec
func (rs *RuleStack) SetRule(spec string, ruleSet []interface{}) (*RuleStack, error) {
	if _, ok := rs.ref[spec]; ok {
		rs.stack[rs.ref[spec]] = nil
	}

	return rs.AddRule(spec, ruleSet)
}

// AddRule adds a filter rule for spec
func (rs *RuleStack) AddRule(spec string, ruleSet []interface{}) (*RuleStack, error) {
	if _, ok := rs.ref[spec]; !ok {
		rs.stack = append(rs.stack, ruleSet)
		rs.ref[spec] = len(rs.stack) - 1
		rs.backref[rs.ref[spec]] = spec
	}

	rs.stack[rs.ref[spec]] = append(rs.stack[rs.ref[spec]].([]interface{}), ruleSet...)
	return rs, nil
}

// SetStaticRule sets a static rule for a spec. This is a single string value
func (rs *RuleStack) SetStaticRule(spec string, value string) *RuleStack {
	if _, ok := rs.ref[spec]; !ok {
		rs.stack = append(rs.stack, value)
		rs.ref[spec] = len(rs.stack) - 1
		rs.backref[rs.ref[spec]] = spec
	} else {
		rs.stack[rs.ref[spec]] = value
	}

	return rs
}

// ClearRules clears al existing rules
func (rs *RuleStack) ClearRules() *RuleStack {
	rs.ref = make(map[string]int)
	rs.stack = make([]interface{}, 0)
	rs.backref = make(map[int]string)
	return rs
}

// NewRuleStack creates a ne rule stack
func NewRuleStack() *RuleStack {
	return &RuleStack{
		ref:     make(map[string]int),
		stack:   make([]interface{}, 0),
		backref: make(map[int]string),
	}
}
