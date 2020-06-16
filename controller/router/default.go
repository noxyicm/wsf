package router

import (
	"strings"
	"wsf/config"
	"wsf/controller/request"
	"wsf/errors"
	"wsf/locale"
)

const (
	// TYPEDefault represents default router
	TYPEDefault = "default"
)

func init() {
	Register(TYPEDefault, NewDefaultRouter)
}

// Default is a default router
type Default struct {
	Options                  *Config
	Routes                   map[string]RouteInterface
	CurrentRoute             string
	UseDefaultRoutes         bool
	UseCurrentParamsAsGlobal bool
	GlobalParams             map[string]interface{}
	DefaultLanguage          *locale.Language
	Language                 *locale.Language
	Languages                map[string]*locale.Language
}

// AddRoute creates and adds route to stack from params
func (r *Default) AddRoute(routeType string, route string, defaults map[string]string, reqs map[string]string, name string) (err error) {
	options := &RouteConfig{
		Type:              routeType,
		URIDelimiter:      r.Options.URIDelimiter,
		URIVariable:       r.Options.URIVariable,
		URIRegexDelimiter: r.Options.URIRegexDelimiter,
	}

	r.Routes[name], err = NewRoute(routeType, options, route, defaults, reqs)
	if err != nil {
		return err
	}

	return nil
}

// Route matches the route
func (r *Default) Route(req request.Interface) (bool, error) {
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

	for name, route := range ReverseRoutes(r.Routes) {
		//match := req.PathInfo()
		if ok, params := route.Match(req, false); ok {
			r.SetRequestParams(req, params)
			r.CurrentRoute = name
			routeMatched = true
			break
		}
	}

	if !routeMatched {
		return false, errors.New("No route matched the request")
	}

	if r.UseCurrentParamsAsGlobal {
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
func (r *Default) SetGlobalParam(param string, value interface{}) error {
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
func (r *Default) StripLanguage(uri string, lang string) string {
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
func (r *Default) SetRequestParams(req request.Interface, params *RouteMatch) error {
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

// NewDefaultRouter creates new default router
func NewDefaultRouter(options config.Config) (ri Interface, err error) {
	r := &Default{
		Routes:       make(map[string]RouteInterface),
		GlobalParams: make(map[string]interface{}),
		Languages:    make(map[string]*locale.Language),
	}

	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)
	r.Options = cfg
	return r, nil
}
