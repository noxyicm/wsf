package rest

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPERouteRest represents rest route
	TYPERouteRest = "rest"
)

func init() {
	controller.RegisterRoute(TYPERouteRest, NewRestRoute)
}

// Route is a restfull route
type Route struct {
	controller.Route

	Prefix string
	Path   string
}

// Match matches provided path against this route
func (r *Route) Match(req request.Interface, partial bool) (bool, *context.RouteMatch) {
	var rqs *request.HTTP
	var ok bool
	if rqs, ok = req.(*request.HTTP); !ok {
		fmt.Println("Not a *request.HTTP")
		return false, nil
	}

	path := rqs.PathInfo()
	if strings.Index(path, r.Options.ModulePrefix) >= 0 {
		path = strings.Replace(path, r.Options.ModulePrefix, "", 1)
	} else {
		return false, nil
	}

	path = strings.Trim(path, r.Options.URIDelimiter)
	params := rqs.Params()
	values := make(map[string]string)
	values[rqs.ModuleKey()] = strings.Replace(r.Options.ModulePrefix, r.Options.URIDelimiter, "", -1)

	if path != "" {
		parts := strings.Split(path, r.Options.URIDelimiter)
		if len(parts) < 1 {
			return false, nil
		}

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

	return true, &context.RouteMatch{Values: result, Name: r.Name(), Match: true}
}

// Assemble assembles user submitted parameters forming a URL path defined by this route
func (r *Route) Assemble(data map[string]interface{}, reset bool, encode bool) (string, error) {
	return "", errors.New("Not implemented")
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
	urlParts := make(map[int]string)
	wildcardData := make(map[string]string)
	flag := false

	for key, part := range r.Parts {
		name := r.Vars[key]
		useDefault := false
		if name != "" && utils.MapSKeyExists(name, data) && data[name] == "" {
			useDefault = true
		}

		if name != "" {
			if utils.MapSKeyExists(name, data) && data[name] != "" && !useDefault {
				value, err = utils.InterfaceToString(data[name])
				if err != nil {
					return "", errors.Wrapf(err, "Unable to assemble route '%s': Value '%s' can not be converted to string", r.RouteName, name)
				}

				delete(data, name)
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, r.Values) && r.Values[name] != "" {
				value = r.Values[name]
			} else if !reset && !useDefault && utils.MapSSKeyExists(name, wildcardData) && wildcardData[name] != "" {
				value = wildcardData[name]
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
				data = utils.MapSMerge(data, wildcardData)
			}

			for variable, val := range data {
				if val != "" && ((utils.MapSSKeyExists(name, r.Defs) && r.Defs[name] != "") || val != r.Defs[variable]) {
					key++
					urlParts[key] = variable
					key++

					v, err := utils.InterfaceToString(val)
					if err != nil {
						return "", errors.Wrapf(err, "Value '%s' can not be converted to string", name)
					}
					urlParts[key] = v
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
func NewRestRoute(options *controller.RouteConfig, name string, route string, defaults map[string]string, reqs map[string]string) controller.RouteInterface {
	r := &Route{
		Route: *controller.NewRouteRoute(options, name, route, defaults, reqs).(*controller.Route),
	}

	r.Options.ModulePrefix = route

	return r
}
