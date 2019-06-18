package filter

import (
	"reflect"
	"regexp"

	"github.com/pkg/errors"
)

func init() {
	Register("RegexpReplace", reflect.TypeOf((*RegexpReplace)(nil)).Elem())
}

// RegexpReplace filter
type RegexpReplace struct {
	matchPatterns         []*regexp.Regexp
	replacements          []string
	unicodeSupportEnabled bool
}

// Filter applyes filter
func (rr *RegexpReplace) Filter(value interface{}) (interface{}, error) {
	if len(rr.matchPatterns) == 0 {
		return nil, errors.New("Match pattern is not set")
	}

	resultString := value.(string)
	for k, v := range rr.matchPatterns {
		var rpls string
		if len(rr.replacements) > k {
			rpls = rr.replacements[k]
		} else {
			rpls = rr.replacements[len(rr.replacements)-1]
		}

		resultString = v.ReplaceAllString(resultString, rpls)
	}

	return resultString, nil
}

// Defaults sets default properties
func (rr *RegexpReplace) Defaults() error {
	return nil
}

// SetMatchPatterns sets the matching patterns stack
func (rr *RegexpReplace) SetMatchPatterns(matchs []string) error {
	rr.matchPatterns = make([]*regexp.Regexp, len(matchs))
	for k, v := range matchs {
		matchPattern, err := regexp.Compile(v)
		if err != nil {
			return err
		}

		rr.matchPatterns[k] = matchPattern
	}

	return nil
}

// AddMatchPatterns adds the matching patterns to stack
func (rr *RegexpReplace) AddMatchPatterns(matchs []string) error {
	for _, v := range matchs {
		matchPattern, err := regexp.Compile(v)
		if err != nil {
			return err
		}

		rr.matchPatterns = append(rr.matchPatterns, matchPattern)
	}

	return nil
}

// SetMatchPattern sets the matching pattern
func (rr *RegexpReplace) SetMatchPattern(match string) error {
	matchPattern, err := regexp.Compile(match)
	if err != nil {
		return err
	}

	rr.matchPatterns = make([]*regexp.Regexp, 0)
	rr.matchPatterns = append(rr.matchPatterns, matchPattern)
	return nil
}

// AddMatchPattern adds the matching pattern to stack
func (rr *RegexpReplace) AddMatchPattern(match string) error {
	matchPattern, err := regexp.Compile(match)
	if err != nil {
		return err
	}

	rr.matchPatterns = append(rr.matchPatterns, matchPattern)
	return nil
}

// SetReplacements sets replacement strings
func (rr *RegexpReplace) SetReplacements(replacements []string) error {
	rr.replacements = replacements
	return nil
}

// AddReplacements adds replacement strings to stack
func (rr *RegexpReplace) AddReplacements(replacements []string) error {
	rr.replacements = append(rr.replacements, replacements...)
	return nil
}

// SetReplacement sets replacement string
func (rr *RegexpReplace) SetReplacement(replacement string) error {
	rr.replacements = make([]string, 0)
	rr.replacements = append(rr.replacements, replacement)
	return nil
}

// AddReplacement adds replacement string to stack
func (rr *RegexpReplace) AddReplacement(replacement string) error {
	rr.replacements = append(rr.replacements, replacement)
	return nil
}

// IsUnicodeSupportEnabled returns true if unicode is enabled
func (rr *RegexpReplace) IsUnicodeSupportEnabled() bool {
	return rr.unicodeSupportEnabled
}

// NewRegexpReplace creates new regexp replace filter
func NewRegexpReplace(match string, replace string) (Interface, error) {
	rr := &RegexpReplace{}
	rr.SetMatchPattern(match)
	rr.SetReplacement(replace)
	return rr, nil
}
