package rest

import (
	"fmt"
	"net/url"
	"strings"
	"wsf/controller/request"
	"wsf/controller/router"
	"wsf/errors"
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
	router.Route

	Prefix string
	Path   string
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *router.RouteMatch) {
	var rqs *request.HTTP
	var ok bool
	if rqs, ok = req.(*request.HTTP); !ok {
		fmt.Println("Not a *request.HTTP")
		return false, nil
	}

	path := rqs.PathInfo()
	path = strings.Replace(path, r.Options.ModulePrefix, "", 1)
	path = strings.Trim(path, r.Options.URIDelimiter)
	path = r.Options.ModulePrefix + path
	params := rqs.Params()
	values := make(map[string]string)

	if path != "" {
		parts := strings.Split(path, r.Options.URIDelimiter)
		if len(parts) < 2 {
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

	r.Values = utils.MapSSMerge(values, params)
	result := utils.MapSSMerge(r.Defs, r.Values)

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

// NewRestRoute creates a new route structure
func NewRestRoute(options *router.RouteConfig, route string, defaults map[string]string, reqs map[string]string) router.RouteInterface {
	r := &Route{
		Route: *router.NewRouteRoute(options, route, defaults, reqs).(*router.Route),
	}

	r.Options.ModulePrefix = route

	return r
}
