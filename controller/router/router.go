package router

import (
	"strings"
	"wsf/controller/request"
	"wsf/errors"
	"wsf/locale"
	"wsf/utils"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface is an interface for controllers
type Interface interface {
	AddDefaultRoutes()
	AddRouteSpec(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) error
	AddRoute(route RouteInterface, name string) error
	Route(request.Interface) (bool, error)
}

type router struct {
	Type                     string
	routes                   map[string]RouteInterface
	currentRoute             string
	useDefaultRoutes         bool
	useCurrentParamsAsGlobal bool
	globalParams             map[string]interface{}
	defaultLanguage          *locale.Language
	language                 *locale.Language
	languages                map[string]*locale.Language
}

// AddDefaultRoutes creates default routes
func (r *router) AddDefaultRoutes() {

}

// AddRouteSpec creates and adds route to stack from params
func (r *router) AddRouteSpec(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) (err error) {
	r.routes[name], err = NewRoute(routeType, route, defaults, reqs)
	if err != nil {
		return err
	}

	return nil
}

// AddRoute creates and adds route to stack
func (r *router) AddRoute(route RouteInterface, name string) (err error) {
	r.routes[name] = route
	return nil
}

// Route matches the route
func (r *router) Route(req request.Interface) (bool, error) {
	if r.useDefaultRoutes {
		r.AddDefaultRoutes()
	}

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

	for name, route := range reverseRoutes(r.routes) {
		//match := req.PathInfo()
		if ok, params := route.Match(req, false); ok {
			r.setRequestParams(req, params)
			r.currentRoute = name
			routeMatched = true
			break
		}
	}

	if !routeMatched {
		return false, errors.New("No route matched the request")
	}

	if r.useCurrentParamsAsGlobal {
		params := req.Params()
		for param, value := range params {
			r.SetGlobalParam(param, value)
		}
	}

	return true, nil
}

/*
public function assemble($userParams, $name = null, $reset = false, $encode = true)
    {
        if (!is_array($userParams)) {
            // require_once 'Zend/Controller/Router/Exception.php';
            throw new Zend_Controller_Router_Exception('userParams must be an array');
        }

        if ($name == null) {
            try {
                $name = $this->getCurrentRouteName();
            } catch (Zend_Controller_Router_Exception $e) {
                $name = 'default';
            }
        }

        // Use UNION (+) in order to preserve numeric keys
        $params = $userParams + $this->_globalParams;

        $route = $this->getRoute($name);
        $url   = $route->assemble($params, $reset, $encode);

        if (!preg_match('|^[a-z]+://|', $url)) {
            $url = rtrim($this->getFrontController()->getBaseUrl(), self::URI_DELIMITER) . self::URI_DELIMITER . $url;
        }

        return $url;
    }*/

// SetGlobalParam sets
func (r *router) SetGlobalParam(param string, value interface{}) error {
	r.globalParams[param] = value
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

func (r *router) stripLanguage(uri string, lang string) string {
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

func (r *router) setRequestParams(req request.Interface, params *RouteMatch) error {
	for param, value := range params.Values {
		req.SetParam(param, value)

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

	return nil
}

func reverseRoutes(m map[string]RouteInterface) map[string]RouteInterface {
	s := make(map[string]RouteInterface)
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	keys = utils.ReverseSliceS(keys)
	for _, k := range keys {
		s[k] = m[k]
	}

	return s
}

// NewRouter creates a new router specified by type
func NewRouter(routerType string, options *Config) (Interface, error) {
	if f, ok := buildHandlers[routerType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized router type \"%v\"", routerType)
}

// Register registers a handler for router creation
func Register(routerType string, handler func(*Config) (Interface, error)) {
	buildHandlers[routerType] = handler
}
