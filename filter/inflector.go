package filter

import (
	"reflect"
	"regexp"
	"strings"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/registry"
)

func init() {
	Register("Inflector", reflect.TypeOf((*Inflector)(nil)).Elem())
}

// Inflector is a filter chain for string inflection
type Inflector struct {
	target                      string
	targetReplacementIdentifier string
	throwTargetExceptionsOn     bool
	rules                       *RuleStack
}

// Filter applyes filtering
func (i *Inflector) Filter(source interface{}) (interface{}, error) {
	// clean source
	value := make(map[string]string)
	switch source.(type) {
	case map[string]string:
		for sourceName, sourceValue := range source.(map[string]string) {
			value[strings.TrimLeft(sourceName, ":")] = sourceValue
		}

	default:
		return source, errors.Errorf("Bad source %v", source)
	}

	rxQuotedTargetReplacementIdentifier := regexp.QuoteMeta(i.targetReplacementIdentifier)
	processedParts := make([]filterPart, 0)

	for ruleIndex, ruleValue := range i.rules.Stack() {
		ruleName := i.rules.StackSpec(ruleIndex)
		if ruleName == "" {
			continue
		}

		if v, ok := value[ruleName]; ok {
			switch ruleValue.(type) {
			case string:
				// overriding the set rule
				processedParts = append(processedParts, filterPart{Pattern: rxQuotedTargetReplacementIdentifier + ruleName, Part: strings.Replace(v, "\\", "\\\\", -1)})

			case []interface{}:
				processedPart := v
				for _, ruleFilter := range ruleValue.([]interface{}) {
					filtered, err := ruleFilter.(Interface).Filter(processedPart)
					if err != nil {
						return nil, err
					}

					processedPart = filtered.(string)
				}

				processedParts = append(processedParts, filterPart{Pattern: rxQuotedTargetReplacementIdentifier + ruleName, Part: strings.Replace(processedPart, "\\", "\\\\", -1)})
			}
		} else if v, ok := ruleValue.(string); ok {
			processedParts = append(processedParts, filterPart{Pattern: rxQuotedTargetReplacementIdentifier + ruleName, Part: strings.Replace(v, "\\", "\\\\", -1)})
		}
	}

	// all of the values of processedParts would have been strings.Replace(..., '\\', '\\\\')'d to disable regexp replace backreferences
	inflectedTarget := i.target
	for _, processedPart := range processedParts {
		rx, err := regexp.Compile(processedPart.Pattern)
		if err != nil && i.throwTargetExceptionsOn {
			return nil, errors.Errorf("A replacement identifier %s was found inside the inflected target, perhaps a rule was not satisfied with a target source? Unsatisfied inflected target: %s", i.targetReplacementIdentifier, inflectedTarget)
		}

		inflectedTarget = rx.ReplaceAllString(inflectedTarget, processedPart.Part)
	}

	return inflectedTarget, nil
}

// Defaults sets default properties
func (i *Inflector) Defaults() error {
	return nil
}

// SetTarget sets a target
func (i *Inflector) SetTarget(target string) error {
	i.target = target
	return nil
}

// GetTarget returns a target
func (i *Inflector) GetTarget() string {
	return i.target
}

// SetTargetReplacementIdentifier sets target replacement identifier
func (i *Inflector) SetTargetReplacementIdentifier(value string) error {
	i.targetReplacementIdentifier = value
	return nil
}

// GetTargetReplacementIdentifier returns target replacement identifier
func (i *Inflector) GetTargetReplacementIdentifier() string {
	return i.targetReplacementIdentifier
}

// SetRules clears existing rules and adds new ones
func (i *Inflector) SetRules(rules map[string]interface{}) *Inflector {
	i.ClearRules()
	i.AddRules(rules)
	return i
}

// AddRules adds new rules
func (i *Inflector) AddRules(rules map[string]interface{}) *Inflector {
	for spec, rule := range rules {
		if spec[0:1] == ":" {
			i.AddFilterRule(spec, rule)
		} else {
			i.SetStaticRule(spec, rule.(string))
		}
	}

	return i
}

// Rules returns all rules
func (i *Inflector) Rules() map[string]interface{} {
	return i.rules.Rules()
}

// SpecRules returns all specific rules
func (i *Inflector) SpecRules(spec string) interface{} {
	if spec != "" {
		spec = i.normalizeSpec(spec)
		return i.rules.SpecRules(spec)
	}

	return nil
}

// Rule returns spec rule by its index
func (i *Inflector) Rule(spec string, index int) interface{} {
	spec = i.normalizeSpec(spec)
	return i.rules.Rule(spec, index)
}

// SetFilterRule sets a filtering rule for a spec. ruleSet can be a string, Filter object
// or a slice of strings or filter objects
func (i *Inflector) SetFilterRule(spec string, ruleSet interface{}) (*Inflector, error) {
	spec = i.normalizeSpec(spec)
	_, err := i.rules.SetRule(spec, []interface{}{ruleSet})
	if err != nil {
		return i, err
	}

	return i, nil
}

// AddFilterRule adds a filter rule for spec
func (i *Inflector) AddFilterRule(spec string, ruleSet interface{}) (*Inflector, error) {
	spec = i.normalizeSpec(spec)
	var rSet []interface{}
	switch ruleSet.(type) {
	case string, Interface:
		rSet = append([]interface{}{}, ruleSet)

	case []interface{}:
		rSet = ruleSet.([]interface{})

	default:
		return i, nil
	}

	ruleSpec := i.rules.SpecRules(spec)
	switch ruleSpec.(type) {
	case string, Interface:
		tmp := ruleSpec
		ruleSpec = append([]interface{}{}, tmp)
	}

	rules := make([]interface{}, 0)
	for _, rule := range rSet {
		ruleFilter, err := i.getRule(rule)
		if err != nil {
			return nil, err
		}

		rules = append(rules, ruleFilter)
	}

	_, err := i.rules.AddRule(spec, rules)
	if err != nil {
		return i, err
	}

	return i, nil
}

// SetStaticRule sets a static rule for a spec. This is a single string value
func (i *Inflector) SetStaticRule(spec string, value string) *Inflector {
	spec = i.normalizeSpec(spec)
	i.rules.SetStaticRule(spec, value)
	return i
}

// ClearRules clears al existing rules
func (i *Inflector) ClearRules() *Inflector {
	i.rules.ClearRules()
	return i
}

func (i *Inflector) normalizeSpec(spec string) string {
	return strings.TrimLeft(spec, ":&")
}

func (i *Inflector) getRule(rule interface{}) (Interface, error) {
	switch t := rule.(type) {
	case Interface:
		return rule.(Interface), nil

	case string:
		filterTypes := registry.Get("filterTypes")
		if filterTypes == nil {
			return nil, errors.New("There are no filter types")
		}

		filters := filterTypes.(map[string]reflect.Type)
		if v, ok := filters[rule.(string)]; ok {
			flt := reflect.New(v).Interface().(Interface)
			err := flt.Defaults()
			if err != nil {
				return nil, err
			}
			return flt, nil
		}

	default:
		return nil, errors.Errorf("Invalid rule type '%t'", t)
	}

	return nil, nil
}

// NewInflector creates new default inflector filter
func NewInflector() (*Inflector, error) {
	inf := &Inflector{
		rules:                       NewRuleStack(),
		targetReplacementIdentifier: ":",
		throwTargetExceptionsOn:     true,
	}

	return inf, nil
}

// NewInflectorConfig creates new inflector filter from config
func NewInflectorConfig(options config.Config) (*Inflector, error) {
	inf := &Inflector{
		target:                      options.GetString("target"),
		rules:                       NewRuleStack(),
		targetReplacementIdentifier: options.GetStringDefault("targetReplacementIdentifier", ":"),
		throwTargetExceptionsOn:     options.GetBoolDefault("throwTargetExceptionsOn", true),
	}

	return inf, nil
}

type filterPart struct {
	Pattern string
	Part    string
}
