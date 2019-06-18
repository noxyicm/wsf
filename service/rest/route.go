package rest

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"wsf/controller/request"
	"wsf/controller/router"
	"wsf/errors"
	"wsf/service"
	"wsf/utils"
)

const (
	// TYPERestRoute represents rest route
	TYPERestRoute = "rest"
)

func init() {
	router.RegisterRoute(TYPERestRoute, NewRestRoute)
}

// Route is a restfull route
type Route struct {
	URIDelimiter   string
	URIVariable    string
	RegexDelimiter string
	prefix         string
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
	service        *Service
	matchedPath    string
	staticCount    int
	//translator
	locale string
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *router.RouteMatch) {
	var rqs *request.HTTP
	var ok bool
	if rqs, ok = req.(*request.HTTP); !ok {
		fmt.Println("Not a *request.HTTP")
		return false, nil
	}

	if r.service == nil {
		return false, nil
	}

	path := rqs.PathInfo()
	path = strings.Replace(path, r.service.RoutePrefix(), "", 1)
	path = strings.Trim(path, r.URIDelimiter)
	params := rqs.Params()
	values := make(map[string]string)

	if path != "" {
		parts := strings.Split(path, r.URIDelimiter)
		if len(parts) < 2 {
			return false, nil
		}

		if !r.service.IsRestfull(parts[0] + "." + parts[1]) {
			return false, nil
		}
		values[rqs.ModuleKey()], _ = utils.ShiftSSlice(&parts)
		values[rqs.ControllerKey()], _ = utils.ShiftSSlice(&parts)
		values[rqs.ActionKey()] = "get"

		pathElementCount := len(parts)

		// Check for "special get" URI's
		var specialGetTarget string
		if pathElementCount > 0 && (parts[0] == "index" || parts[0] == "new") {
			specialGetTarget, _ = utils.ShiftSSlice(&parts)
		} else if pathElementCount > 1 && parts[pathElementCount-1] == "edit" {
			specialGetTarget = "edit"
			params["id"], _ = url.QueryUnescape(parts[pathElementCount-2])
		} else if pathElementCount > 0 {
			v, _ := utils.ShiftSSlice(&parts)
			params["id"], _ = url.QueryUnescape(v)
		} else if _, ok := params["id"]; !ok && pathElementCount == 0 {
			specialGetTarget = "index"
		}

		// Digest URI params
		if pathElementCount = len(parts); pathElementCount > 0 {
			for i := 0; i < pathElementCount; i = i + 2 {
				val := ""
				if pathElementCount > i+1 {
					val = parts[i+1]
				}

				key, _ := url.QueryUnescape(parts[i])
				val, _ = url.QueryUnescape(val)
				params[key] = val
			}
		}

		// Determine Action
		requestMethod := strings.ToLower(rqs.Method)
		if requestMethod != "get" {
			if mtd := rqs.Param("_method"); mtd != nil {
				values[rqs.ActionKey()] = strings.ToLower(mtd.(string))
			} else if hdr := rqs.Header("X-HTTP-Method-Override"); hdr != "" {
				values[rqs.ActionKey()] = strings.ToLower(hdr)
			} else {
				values[rqs.ActionKey()] = requestMethod
			}

			// Map PUT and POST to actual create/update actions
			// based on parameter count (posting to resource or collection)
			switch values[rqs.ActionKey()] {
			case "post":
				if pathElementCount > 0 {
					values[rqs.ActionKey()] = "put"
				} else {
					values[rqs.ActionKey()] = "post"
				}

			case "put":
				values[rqs.ActionKey()] = "put"
			}
		} else if specialGetTarget != "" {
			values[rqs.ActionKey()] = specialGetTarget
		}
	}

	r.values = utils.MapSSMerge(values, params)
	result := utils.MapSSMerge(r.defaults, r.values)

	if partial && len(result) > 0 {
		r.matchedPath = rqs.PathInfo()
	}

	return true, &router.RouteMatch{Values: result, Match: true}
}

// Assemble assembles user submitted parameters forming a URL path defined by this route
func (r *Route) Assemble(data map[string]string, args ...bool) (string, error) {
	return "", errors.New("Not implemented")
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

// SetService sets a route service
func (r *Route) SetService(svc service.Interface) error {
	if svcs, ok := svc.(*Service); ok {
		r.service = svcs
	}

	return errors.Errorf("Not a valid service provided")
}

// NewRestRoute creates a new route structure
func NewRestRoute(route string, defaults map[string]string, reqs map[string]string) (router.RouteInterface, error) {
	r := &Route{
		URIDelimiter:   router.URIDelimiter,
		URIVariable:    router.URIVariable,
		RegexDelimiter: router.URIRegexDelimiter,
		variables:      make(map[int]string),
		parts:          make([]string, 0),
		translatable:   make([]string, 0),
		defaults:       defaults,
		requirements:   reqs,
		values:         make(map[string]string),
		wildcardData:   make(map[string]string),
		defaultRegex:   nil,
	}

	//route = strings.Replace(route, r.service.RoutePrefix(), "", 1)
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
