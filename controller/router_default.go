package controller

import (
	"strings"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/locale"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPERouterDefault represents default router
	TYPERouterDefault = "default"
)

func init() {
	RegisterRouter(TYPERouterDefault, NewDefaultRouter)
}

// DefaultRouter is a default router
type DefaultRouter struct {
	Options                  *RouterConfig
	Routes                   *RoutesList
	CurrentRoute             string
	UseDefaultRoutes         bool
	UseCurrentParamsAsGlobal bool
	GlobalParams             map[string]interface{}
	DefaultLanguage          *locale.Language
	Language                 *locale.Language
	Languages                map[string]*locale.Language
}

// AddRoute creates and adds route to stack from params
func (r *DefaultRouter) AddRoute(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) (err error) {
	options := &RouteConfig{
		Type:              routeType,
		URIDelimiter:      r.Options.URIDelimiter,
		URIVariable:       r.Options.URIVariable,
		URIRegexDelimiter: r.Options.URIRegexDelimiter,
	}

	rt, err := NewRoute(routeType, options, name, route, defaults, reqs)
	if err != nil {
		return err
	}

	return r.Routes.Append(name, rt)
}

// RouteByName return route by its name
func (r *DefaultRouter) RouteByName(name string) (RouteInterface, bool) {
	if r.Routes.Has(name) {
		return r.Routes.Value(name), true
	}

	return nil, false
}

// Match matches the routes against request
func (r *DefaultRouter) Match(ctx context.Context, req request.Interface) (bool, error) {
	/*path := req.PathInfo()
	parts := strings.Split(path, r.URIDelimiter)
	language := r.defaultLanguage.Code
	if len(parts) <= 1 {
		r.language = r.defaultLanguage
	} else {
		l := strings.ToLower(parts[1])
		if v, ok := r.languages[l]; ok {
			req.SetPathInfo(r.stripLanguage(path, l))
			r.language = v
			language = v.Code
		}
	}

	req.SetParam("lang", language)*/

	// Find the matching route
	routeMatched := false
	for _, route := range ReverseRoutesList(r.Routes).Stack() {
		//match := req.PathInfo()
		if ok, params := route.Match(req, false); ok {
			r.SetRequestParams(req, params)
			ctx.SetCurrentRoute(params)
			routeMatched = true
			break
		}
	}

	if !routeMatched {
		return true, errors.New("No route matched the request")
	}

	if r.UseCurrentParamsAsGlobal {
		params := req.Params()
		for param, value := range params {
			r.SetGlobalParam(param, value)
		}
	}

	return true, nil
}

// Assemble assembles a route parameters into url
func (r *DefaultRouter) Assemble(ctx context.Context, params map[string]interface{}, name string, reset bool, encode bool) (string, error) {
	if name == "" {
		name = ctx.CurrentRouteName()
		//params = utils.MapSMerge(ctx.CurrentRoute().Values, params)
		//if !reset {
		//	params = utils.MapSMerge(params, ctx.CurrentRoute().WildcardData)
		//}
	}

	params = utils.MapSMerge(params, r.GlobalParams)
	if v, ok := r.RouteByName(name); ok {
		url, err := v.Assemble(params, reset, encode)
		if err != nil {
			return "", errors.Wrapf(err, "Unable to assemble route '%s'", name)
		}

		// if (!preg_match('|^[a-z]+://|', $url)) {
		// 	$url = rtrim($this->getFrontController()->getBaseUrl(), self::URI_DELIMITER) . self::URI_DELIMITER . $url;
		// }
		return url, nil
	}

	return "", errors.Errorf("Route by name '%s' does not exists", name)
}

// SetGlobalParam sets
func (r *DefaultRouter) SetGlobalParam(param string, value interface{}) error {
	r.GlobalParams[param] = value
	return nil
}

/*func (r *router) PrependRoutes(routes)
{
	$this->_routes = $routes + $this->_routes;
	return $this;
}

public function setDefaultLanguage($lang = NULL)
{
	if(!empty($lang))
		$this->_defaultLanguage = $lang;
}

public function getDefaultLanguage()
{
	return $this->_defaultLanguage;
}

public function setLanguages($langs = NULL)
{
	if(!empty($langs))
		$this->_languages = $langs;
}

public function getLanguages()
{
	return $this->_languages;
}

public function setLanguage($lang = NULL)
{
	if(!empty($lang))
		$this->_language = $lang;
}

public function getLanguage()
{
	return $this->_language;
}*/

// StripLanguage strips language spec from URI
func (r *DefaultRouter) StripLanguage(uri string, lang string) string {
	if uri == "" {
		return ""
	}

	if lang == "" {
		return uri
	}

	//reg := regexp.MustCompile(`\/` + lang + `\/`)
	replaced := strings.Replace(uri, "", "/"+lang, 1)
	if replaced == "" {
		replaced = "/"
	}

	return replaced
}

// SetRequestParams sets matched parameters to request
func (r *DefaultRouter) SetRequestParams(req request.Interface, params *context.RouteMatch) error {
	for param, value := range params.Values {
		if !req.HasParam(param) {
			req.SetParam(param, value)
		}

		if param == req.ModuleKey() {
			req.SetModuleName(value)
		}

		if param == req.ControllerKey() {
			req.SetControllerName(value)
		}

		if param == req.ActionKey() {
			req.SetActionName(value)
		}
	}

	for param, value := range params.WildcardData {
		req.SetParam(param, value)
	}

	return nil
}

// NewDefaultRouter creates new default router
func NewDefaultRouter(options config.Config) (ri RouterInterface, err error) {
	r := &DefaultRouter{
		Routes:       NewRoutesList(),
		GlobalParams: make(map[string]interface{}),
		Languages:    make(map[string]*locale.Language),
	}

	cfg := &RouterConfig{}
	cfg.Defaults()
	cfg.Populate(options)
	r.Options = cfg
	return r, nil
}
