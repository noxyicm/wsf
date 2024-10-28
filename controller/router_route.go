package controller

import (
	"net/url"
	"regexp"
	"strings"
	"wsf/config"
	"wsf/context"
	"wsf/controller/request"
	"wsf/errors"
	"wsf/utils"
)

const (
	// TYPERouteRoute represents default route
	TYPERouteRoute = "route"
)

var (
	buildRouteHandlers = map[string]func(*RouteConfig, string, string, map[string]string, map[string]string) RouteInterface{}
)

func init() {
	RegisterRoute(TYPERouteRoute, NewRouteRoute)
}

// RouteInterface is a route interface
type RouteInterface interface {
	Name() string
	Match(req request.Interface, partial bool) (bool, *context.RouteMatch)
	Assemble(data map[string]interface{}, reset bool, encode bool) (string, error)
}

// Route is
type Route struct {
	Options      *RouteConfig
	RouteName    string
	Method       string
	Path         string
	Action       string
	Module       string
	Controller   string
	IsTranslated bool
	Vars         map[int]string
	Parts        []string
	Translatable []string
	Defs         map[string]string
	Requirements map[string]string
	Values       map[string]string
	DefaultRegex *regexp.Regexp
	StaticCount  int
	Loc          string
}

// Name return route name
func (r *Route) Name() string {
	return r.RouteName
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *context.RouteMatch) {
	translateMessages := make(map[string]string)
	//if r.isTranslated {
	//	translateMessages := r.translator.GetMessages()
	//}

	path := req.PathInfo()
	if r.Options.ModulePrefix != "" {
		if strings.Index(path, r.Options.ModulePrefix) >= 0 {
			path = strings.Replace(path, r.Options.ModulePrefix, "", 1)
			//path = r.Options.ModulePrefix + path
		} else {
			return false, nil
		}
	}
	path = strings.Trim(path, r.Options.URIDelimiter)
	pathStaticCount := 0
	values := make(map[string]string)
	matchedPath := ""
	wildcardData := make(map[string]string)

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
					if !utils.MapSSKeyExists(variable, wildcardData) && !utils.MapSSKeyExists(variable, r.Defs) && !utils.MapSSKeyExists(variable, values) {
						if count > i+1 {
							wildcardData[variable], _ = url.QueryUnescape(pathParts[i+1])
						} else {
							wildcardData[variable] = ""
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
			if r.IsTranslated && (strings.Index(routePart, "@") == 0 && strings.Index(routePart, "@@") != 0 && name == "") || name != "" && utils.InSSlice(name, r.Translatable) {
				if strings.Index(routePart, "@") == 0 {
					routePart = routePart[1:]
				}

				if v, ok := translateMessages[path]; ok {
					part = v
				}
			}

			if strings.Index(routePart, "@@") == 0 {
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

	//values = utils.MapSSMerge(values, wildcardData)
	values = utils.MapSSMerge(r.Defs, values)

	for _, value := range r.Vars {
		if _, ok := values[value]; !ok {
			return false, nil
		} else if values[value] == "" {
			values[value] = r.Defs[value]
		}
	}

	return true, &context.RouteMatch{Values: values, WildcardData: wildcardData, Name: r.Name(), Match: true}
}

// Assemble assembles user submitted parameters forming a URL path defined by this route
func (r *Route) Assemble(data map[string]interface{}, reset bool, encode bool) (string, error) {
	partial := false
	//var locale string
	/*if r.isTranslated {
		if v, ok := data["@locale"]; ok {
			locale = data["@locale"]
			delete(data["@locale"])
		} else {
			locale = r.locale
		}
	}*/

	var err error
	value := ""
	urlParts := make([]string, 0)
	wildcardData := make(map[string]string)
	flag := false

	for key, part := range r.Parts {
		name := r.Vars[key]
		useDefault := false
		if name != "" && utils.MapSKeyExists(name, data) && (data[name] == "" || data[name] == nil) {
			useDefault = true
		}

		if name != "" {
			if !reset && !useDefault && utils.MapSKeyExists(name, data) && data[name] != "" && data[name] != nil {
				value, err = utils.InterfaceToString(data[name])
				if err != nil {
					return "", errors.Wrapf(err, "Unable to assemble route '%s': Value '%s' can not be converted to string", r.RouteName, name)
				}

				delete(data, name)
			} else if utils.MapSSKeyExists(name, r.Defs) {
				value = r.Defs[name]
			} else {
				return "", errors.Errorf("Value %s is not specified", name)
			}

			if r.IsTranslated && utils.InSSlice(name, r.Translatable) {
				//urlParts[key] = r.translator.Translate(value, locale)
			} else {
				//urlParts[key] = value
				urlParts = append(urlParts, value)
			}
		} else if part != "*" {
			if r.IsTranslated && part[0:1] == "@" {
				if part[1:2] != "@" {
					//urlParts[key] = r.translator.Translate(part[1:], locale)
				} else {
					//urlParts[key] = part[1:]
					urlParts = append(urlParts, part[1:])
				}
			} else {
				if part[0:2] == "@@" {
					part = part[1:]
				}

				//urlParts[key] = part
				urlParts = append(urlParts, part)
			}
		} else {
			if !reset {
				data = utils.MapSMerge(data, wildcardData)
			}

			for variable, val := range data {
				if val != "" && val != nil && ((utils.MapSSKeyExists(variable, r.Defs) && r.Defs[variable] != "") || val != r.Defs[variable]) {
					urlParts = append(urlParts, variable)

					v, err := utils.InterfaceToString(val)
					if err != nil {
						return "", errors.Wrapf(err, "Value '%s' can not be converted to string", variable)
					}
					urlParts = append(urlParts, v)
					flag = true
				}
			}
		}
	}

	path := ""
	for key, value := range utils.ReverseSliceS(urlParts) {
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

	return strings.TrimRight(path, r.Options.URIDelimiter), nil
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

// NewRouteRoute creates a new route structure
func NewRouteRoute(options *RouteConfig, name string, route string, defaults map[string]string, reqs map[string]string) RouteInterface {
	r := &Route{
		Options:      options,
		RouteName:    name,
		Vars:         make(map[int]string),
		Parts:        make([]string, 0),
		Translatable: make([]string, 0),
		Defs:         defaults,
		Requirements: reqs,
		Values:       make(map[string]string),
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
func NewRoute(routeType string, options *RouteConfig, name string, route string, defaults map[string]string, reqs map[string]string) (RouteInterface, error) {
	if f, ok := buildRouteHandlers[routeType]; ok {
		return f(options, name, route, defaults, reqs), nil
	}

	return nil, errors.Errorf("Unrecognized route type \"%v\"", routeType)
}

// RegisterRoute registers a route type for router creation
func RegisterRoute(routeType string, handler func(*RouteConfig, string, string, map[string]string, map[string]string) RouteInterface) {
	buildRouteHandlers[routeType] = handler
}
