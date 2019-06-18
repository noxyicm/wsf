package router

import (
	"net/url"
	"regexp"
	"strings"
	"wsf/config"
	"wsf/controller/request"
	"wsf/errors"
	"wsf/utils"
)

const (
	// TYPERoute represents default route
	TYPERoute = "route"

	// URIDelimiter represents uri separator
	URIDelimiter = "/"

	// URIVariable represents prefix poiting to variable
	URIVariable = ":"

	// URIRegexDelimiter represents pointer to regex expresion
	URIRegexDelimiter = ""
)

var (
	buildRouteHandlers = map[string]func(string, map[string]string, map[string]string) (RouteInterface, error){}
)

func init() {
	RegisterRoute(TYPERoute, NewRouteRoute)
}

// RouteInterface is a route interface
type RouteInterface interface {
	Match(req request.Interface, partial bool) (bool, *RouteMatch)
	Assemble(data map[string]string, args ...bool) (string, error)
}

// Route is
type Route struct {
	URIDelimiter   string
	URIVariable    string
	RegexDelimiter string
	method         string
	path           string
	action         string
	module         string
	controller     string
	params         []string
	isTranslated   bool
	variables      map[int]string
	parts          []string
	translatable   []string
	defaults       map[string]string
	requirements   map[string]string
	values         map[string]string
	wildcardData   map[string]string
	defaultRegex   *regexp.Regexp
	staticCount    int
	//translator
	locale string
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *RouteMatch) {
	translateMessages := make(map[string]string)
	//if r.isTranslated {
	//	translateMessages := r.translator.GetMessages()
	//}

	path := req.PathInfo()
	pathStaticCount := 0
	values := make(map[string]string)
	matchedPath := ""

	if !partial {
		path = strings.Trim(path, r.URIDelimiter)
	}

	if path != "" {
		pathParts := strings.Split(path, r.URIDelimiter)
		for pos, part := range pathParts {
			// Path is longer than a route, it's not a match
			if len(r.parts) <= pos {
				if partial {
					break
				} else {
					return false, nil
				}
			}

			matchedPath = matchedPath + part + r.URIDelimiter
			// If it's a wildcard, get the rest of URL as wildcard data and stop matching
			if r.parts[pos] == "*" {
				count := len(pathParts)
				for i := pos; i < count; i += 2 {
					variable, _ := url.QueryUnescape(pathParts[i])
					if !utils.MapSSKeyExists(variable, r.wildcardData) && !utils.MapSSKeyExists(variable, r.defaults) && !utils.MapSSKeyExists(variable, values) {
						if count <= i+1 {
							r.wildcardData[variable], _ = url.QueryUnescape(pathParts[i+1])
						} else {
							r.wildcardData[variable] = ""
						}
					}
				}

				matchedPath = strings.Join(pathParts, r.URIDelimiter)
				break
			}

			name := r.variables[pos]
			part, _ = url.QueryUnescape(part)

			// Translate value if required
			routePart := r.parts[pos]
			if r.isTranslated && (routePart[0:1] == "@" && routePart[1:2] != "@" && name == "") || name != "" && utils.InSSlice(name, r.translatable) {
				if routePart[0:1] == "@" {
					routePart = routePart[1:]
				}

				if v, ok := translateMessages[path]; ok {
					part = v
				}
			}

			if routePart[0:2] == "@@" {
				routePart = routePart[1:]
			}

			// If it's a static part, match directly
			if name == "" && routePart != part {
				return false, nil
			}

			// If it's a variable with requirement, match a regex. If not - everything matches
			rx, err := regexp.Compile(`^` + routePart + `$`)
			if err != nil {
				return false, nil
			}

			if routePart != "" && !rx.MatchString(part) {
				return false, nil
			}

			// If it's a variable, store it's value for later
			if name != "" {
				values[name] = part
			} else {
				pathStaticCount++
			}
		}
	}

	// Check if all static mappings have been matched
	if r.staticCount != pathStaticCount {
		return false, nil
	}

	values = utils.MapSSMerge(values, r.wildcardData)
	values = utils.MapSSMerge(values, r.defaults)

	for _, value := range r.variables {
		if _, ok := values[value]; !ok {
			return false, nil
		} else if values[value] == "" {
			values[value] = r.defaults[value]
		}
	}

	r.path = strings.TrimRight(matchedPath, r.URIDelimiter)
	r.values = values

	return true, &RouteMatch{Values: values, Match: true}
}

// Assemble assembles user submitted parameters forming a URL path defined by this route
func (r *Route) Assemble(data map[string]string, args ...bool) (string, error) {
	reset := true
	encode := true
	partial := false
	for k, v := range args {
		if k == 0 {
			reset = v
		} else if k == 1 {
			encode = v
		} else if k == 2 {
			partial = v
		}
	}
	//var locale string
	/*if r.isTranslated {
		if v, ok := data["@locale"]; ok {
			locale = data["@locale"]
			delete(data["@locale"])
		} else {
			locale = r.locale
		}
	}*/

	value := ""
	urlParts := make(map[int]string)
	flag := false

	for key, part := range r.parts {
		name := r.variables[key]
		useDefault := false
		if name != "" && utils.MapSSKeyExists(name, data) && data[name] == "" {
			useDefault = true
		}

		if name != "" {
			if utils.MapSSKeyExists(name, data) && data[name] != "" && !useDefault {
				value = data[name]
				delete(data, name)
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, r.values) && r.values[name] != "" {
				value = r.values[name]
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, r.wildcardData) && r.wildcardData[name] != "" {
				value = r.wildcardData[name]
			} else if utils.MapSSKeyExists(name, r.defaults) {
				value = r.defaults[name]
			} else {
				return "", errors.Errorf("Value %s is not specified", name)
			}

			if r.isTranslated && utils.InSSlice(name, r.translatable) {
				//urlParts[key] = r.translator.Translate(value, locale)
			} else {
				urlParts[key] = value
			}
		} else if part != "*" {
			if r.isTranslated && part[0:1] == "@" {
				if part[1:2] != "@" {
					//urlParts[key] = r.translator.Translate(part[1:], locale)
				} else {
					urlParts[key] = part[1:]
				}
			} else {
				if part[0:2] == "@@" {
					part = part[1:]
				}

				urlParts[key] = part
			}
		} else {
			if !reset {
				data = utils.MapSSMerge(data, r.wildcardData)
			}

			for variable, val := range data {
				if val != "" && ((utils.MapSSKeyExists(name, r.defaults) && r.defaults[name] != "") || val != r.defaults[variable]) {
					key++
					urlParts[key] = variable
					key++
					urlParts[key] = val
					flag = true
				}
			}
		}
	}

	path := ""
	for key, value := range utils.ReverseMapIS(urlParts) {
		defaultValue := ""
		if len(r.variables) > key && r.variables[key] != "" {
			defaultValue = r.Default(r.variables[key])

			if r.isTranslated && defaultValue != "" && utils.InSSlice(r.variables[key], r.translatable) {
				//defaultValue = r.translator.Translate(defaultValue, locale)
			}
		}

		if flag || value != defaultValue || partial {
			v := value
			if encode {
				v = url.QueryEscape(value)
			}
			path = r.URIDelimiter + v + path
			flag = true
		}
	}

	return strings.Trim(path, r.URIDelimiter), nil
}

// Default returns default value if defined
func (r *Route) Default(key string) string {
	if utils.MapSSKeyExists(key, r.defaults) && r.defaults[key] != "" {
		return r.defaults[key]
	}

	return ""
}

// Defaults returns map of default values
func (r *Route) Defaults() map[string]string {
	return r.defaults
}

// Variables returns map of variables
func (r *Route) Variables() map[int]string {
	return r.variables
}

// Locale returns route locale
func (r *Route) Locale() string {
	if r.locale != "" {
		return r.locale
	}

	return ""
}

// RouteMatch is a route matched values
type RouteMatch struct {
	Values map[string]string
	Match  bool
}

// NewRouteRoute creates a new route structure
func NewRouteRoute(route string, defaults map[string]string, reqs map[string]string) (RouteInterface, error) {
	r := &Route{
		URIDelimiter:   URIDelimiter,
		URIVariable:    URIVariable,
		RegexDelimiter: URIRegexDelimiter,
		variables:      make(map[int]string),
		parts:          make([]string, 0),
		translatable:   make([]string, 0),
		defaults:       defaults,
		requirements:   reqs,
		values:         make(map[string]string),
		wildcardData:   make(map[string]string),
		defaultRegex:   nil,
	}

	route = strings.Trim(route, r.URIDelimiter)
	if route != "" {
		routeParts := strings.Split(route, r.URIDelimiter)
		r.parts = make([]string, len(routeParts))
		r.variables = make(map[int]string)
		for pos, part := range routeParts {
			if part[0:1] == r.URIVariable && part[1:2] != r.URIVariable {
				name := part[1:]
				if name[0:1] == "@" && name[1:2] != "@" {
					name = name[1:]
					r.translatable = append(r.translatable, name)
					r.isTranslated = true
				}

				if v, ok := reqs[name]; ok {
					r.parts[pos] = v
				} else {
					r.parts[pos] = ""
				}
				r.variables[pos] = name
			} else {
				if part[0:1] == r.URIVariable {
					part = part[1:]
				}

				if part[0:1] == "@" && part[1:2] != "@" {
					r.isTranslated = true
				}

				r.parts[pos] = part
				if part != "*" {
					r.staticCount++
				}
			}
		}
	}

	return r, nil
}

// FromConfig creates a new route from config
func FromConfig(cfg config.Config) (*Route, error) {
	r := &Route{}
	return r, nil
}

// NewRoute creates a new router specified by type
func NewRoute(routeType string, route string, defaults map[string]string, reqs map[string]string) (RouteInterface, error) {
	if f, ok := buildRouteHandlers[routeType]; ok {
		return f(route, defaults, reqs)
	}

	return nil, errors.Errorf("Unrecognized route type \"%v\"", routeType)
}

// RegisterRoute registers a route type for router creation
func RegisterRoute(routeType string, handler func(string, map[string]string, map[string]string) (RouteInterface, error)) {
	buildRouteHandlers[routeType] = handler
}
