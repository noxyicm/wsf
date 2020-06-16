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
)

var (
	buildRouteHandlers = map[string]func(*RouteConfig, string, map[string]string, map[string]string) RouteInterface{}
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
	Options      *RouteConfig
	Method       string
	Path         string
	Action       string
	Module       string
	Controller   string
	Params       []string
	IsTranslated bool
	Vars         map[int]string
	Parts        []string
	Translatable []string
	Defs         map[string]string
	Requirements map[string]string
	Values       map[string]string
	WildcardData map[string]string
	DefaultRegex *regexp.Regexp
	StaticCount  int
	//translator
	Loc string
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *RouteMatch) {
	translateMessages := make(map[string]string)
	//if r.isTranslated {
	//	translateMessages := r.translator.GetMessages()
	//}

	path := req.PathInfo()
	if r.Options.ModulePrefix != "" {
		path = strings.Replace(path, r.Options.ModulePrefix, "", 1)
		path = strings.Trim(path, r.Options.URIDelimiter)
		path = r.Options.ModulePrefix + path
	}
	pathStaticCount := 0
	values := make(map[string]string)
	matchedPath := ""

	if !partial {
		path = strings.Trim(path, r.Options.URIDelimiter)
	}

	if path != "" {
		pathParts := strings.Split(path, r.Options.URIDelimiter)
		for pos, part := range pathParts {
			// Path is longer than a route, it's not a match
			if len(r.Parts) <= pos {
				if partial {
					break
				} else {
					return false, nil
				}
			}

			matchedPath = matchedPath + part + r.Options.URIDelimiter
			// If it's a wildcard, get the rest of URL as wildcard data and stop matching
			if r.Parts[pos] == "*" {
				count := len(pathParts)
				for i := pos; i < count; i += 2 {
					variable, _ := url.QueryUnescape(pathParts[i])
					if !utils.MapSSKeyExists(variable, r.WildcardData) && !utils.MapSSKeyExists(variable, r.Defs) && !utils.MapSSKeyExists(variable, values) {
						if count <= i+1 {
							r.WildcardData[variable], _ = url.QueryUnescape(pathParts[i+1])
						} else {
							r.WildcardData[variable] = ""
						}
					}
				}

				matchedPath = strings.Join(pathParts, r.Options.URIDelimiter)
				break
			}

			name := r.Vars[pos]
			part, _ = url.QueryUnescape(part)

			// Translate value if required
			routePart := r.Parts[pos]
			if r.IsTranslated && (routePart[0:1] == "@" && routePart[1:2] != "@" && name == "") || name != "" && utils.InSSlice(name, r.Translatable) {
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
	if r.StaticCount != pathStaticCount {
		return false, nil
	}

	values = utils.MapSSMerge(values, r.WildcardData)
	values = utils.MapSSMerge(values, r.Defs)

	for _, value := range r.Vars {
		if _, ok := values[value]; !ok {
			return false, nil
		} else if values[value] == "" {
			values[value] = r.Defs[value]
		}
	}

	r.Path = strings.TrimRight(matchedPath, r.Options.URIDelimiter)
	r.Values = values

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

	for key, part := range r.Parts {
		name := r.Vars[key]
		useDefault := false
		if name != "" && utils.MapSSKeyExists(name, data) && data[name] == "" {
			useDefault = true
		}

		if name != "" {
			if utils.MapSSKeyExists(name, data) && data[name] != "" && !useDefault {
				value = data[name]
				delete(data, name)
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, r.Values) && r.Values[name] != "" {
				value = r.Values[name]
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, r.WildcardData) && r.WildcardData[name] != "" {
				value = r.WildcardData[name]
			} else if utils.MapSSKeyExists(name, r.Defs) {
				value = r.Defs[name]
			} else {
				return "", errors.Errorf("Value %s is not specified", name)
			}

			if r.IsTranslated && utils.InSSlice(name, r.Translatable) {
				//urlParts[key] = r.translator.Translate(value, locale)
			} else {
				urlParts[key] = value
			}
		} else if part != "*" {
			if r.IsTranslated && part[0:1] == "@" {
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
				data = utils.MapSSMerge(data, r.WildcardData)
			}

			for variable, val := range data {
				if val != "" && ((utils.MapSSKeyExists(name, r.Defs) && r.Defs[name] != "") || val != r.Defs[variable]) {
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
		if len(r.Vars) > key && r.Vars[key] != "" {
			defaultValue = r.Default(r.Vars[key])

			if r.IsTranslated && defaultValue != "" && utils.InSSlice(r.Vars[key], r.Translatable) {
				//defaultValue = r.translator.Translate(defaultValue, locale)
			}
		}

		if flag || value != defaultValue || partial {
			v := value
			if encode {
				v = url.QueryEscape(value)
			}
			path = r.Options.URIDelimiter + v + path
			flag = true
		}
	}

	return strings.Trim(path, r.Options.URIDelimiter), nil
}

// Default returns default value if defined
func (r *Route) Default(key string) string {
	if utils.MapSSKeyExists(key, r.Defs) && r.Defs[key] != "" {
		return r.Defs[key]
	}

	return ""
}

// Defaults returns map of default values
func (r *Route) Defaults() map[string]string {
	return r.Defs
}

// Variables returns map of variables
func (r *Route) Variables() map[int]string {
	return r.Vars
}

// Locale returns route locale
func (r *Route) Locale() string {
	if r.Loc != "" {
		return r.Loc
	}

	return ""
}

// RouteMatch is a route matched values
type RouteMatch struct {
	Values map[string]string
	Match  bool
}

// NewRouteRoute creates a new route structure
func NewRouteRoute(options *RouteConfig, route string, defaults map[string]string, reqs map[string]string) RouteInterface {
	r := &Route{
		Options:      options,
		Vars:         make(map[int]string),
		Parts:        make([]string, 0),
		Translatable: make([]string, 0),
		Defs:         defaults,
		Requirements: reqs,
		Values:       make(map[string]string),
		WildcardData: make(map[string]string),
		DefaultRegex: nil,
	}

	route = strings.Trim(route, r.Options.URIDelimiter)
	if route != "" {
		routeParts := strings.Split(route, r.Options.URIDelimiter)
		r.Parts = make([]string, len(routeParts))
		for pos, part := range routeParts {
			if part[0:1] == r.Options.URIVariable && part[1:2] != r.Options.URIVariable {
				name := part[1:]
				if name[0:1] == "@" && name[1:2] != "@" {
					name = name[1:]
					r.Translatable = append(r.Translatable, name)
					r.IsTranslated = true
				}

				if v, ok := reqs[name]; ok {
					r.Parts[pos] = v
				} else {
					r.Parts[pos] = ""
				}
				r.Vars[pos] = name
			} else {
				if part[0:1] == r.Options.URIVariable {
					part = part[1:]
				}

				if part[0:1] == "@" && part[1:2] != "@" {
					r.IsTranslated = true
				}

				r.Parts[pos] = part
				if part != "*" {
					r.StaticCount++
				}
			}
		}
	}

	return r
}

// FromConfig creates a new route from config
func FromConfig(options config.Config) (*Route, error) {
	r := &Route{}

	cfg := &RouteConfig{}
	cfg.Defaults()
	cfg.Populate(options)
	r.Options = cfg
	return r, nil
}

// NewRoute creates a new router specified by type
func NewRoute(routeType string, options *RouteConfig, route string, defaults map[string]string, reqs map[string]string) (RouteInterface, error) {
	if f, ok := buildRouteHandlers[routeType]; ok {
		return f(options, route, defaults, reqs), nil
	}

	return nil, errors.Errorf("Unrecognized route type \"%v\"", routeType)
}

// RegisterRoute registers a route type for router creation
func RegisterRoute(routeType string, handler func(*RouteConfig, string, map[string]string, map[string]string) RouteInterface) {
	buildRouteHandlers[routeType] = handler
}
