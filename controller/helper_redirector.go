package controller

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
)

const (
	// TYPEHelperRedirector represents Redirector action helper
	TYPEHelperRedirector = "redirector"
)

var (
	protox   = regexp.MustCompile("^(https?|ftp)://")
	nlx      = regexp.MustCompile("\\n\\r")
	prependx = regexp.MustCompile("|^[a-z]+://|")
)

func init() {
	RegisterHelper(TYPEHelperRedirector, NewRedirectorHelper)
}

// Redirector is a action helper that handles redirects
type Redirector struct {
	name           string
	Code           int
	UseAbsoluteURI bool
	View           view.Interface
}

// Name returns helper name
func (h *Redirector) Name() string {
	return h.name
}

// Init the helper
func (h *Redirector) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *Redirector) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *Redirector) PostDispatch(ctx context.Context) error {
	return nil
}

// GotoSimple performs a redirect to an action/controller/module with params
func (h *Redirector) GotoSimple(ctx context.Context, module string, ctrl string, action string, params map[string]interface{}) error {
	dispatcher := Dispatcher()
	request := ctx.Request()
	useDefaultController := false

	if module == "" && ctrl == "" {
		useDefaultController = true
	}

	if module == "" {
		module = request.ModuleName()
	}

	if module == dispatcher.DefaultModule() {
		module = ""
	}

	if ctrl == "" && !useDefaultController {
		ctrl = request.ControllerName()
		if ctrl == "" {
			ctrl = dispatcher.DefaultController()
		}
	}

	params[request.ModuleKey()] = module
	params[request.ControllerKey()] = ctrl
	params[request.ActionKey()] = action

	url, err := Router().Assemble(ctx, params, "default", true, true)
	if err != nil {
		return errors.Wrap(err, "Unable to set simple redirect")
	}

	if err := h.redirect(ctx, url, h.Code); err != nil {
		return errors.Wrap(err, "Unable to set simple redirect")
	}

	return nil
}

// GotoURL performs a redirect to a url
func (h *Redirector) GotoURL(ctx context.Context, url string, params map[string]interface{}) error {
	url = nlx.ReplaceAllString(url, "")

	code := h.Code
	prependBase := false
	if params != nil {
		if v, ok := params["prependBase"].(bool); ok {
			prependBase = v
		}

		if v, ok := params["code"]; ok {
			code, _ = utils.InterfaceToInt(v)
		}
	}

	if prependx.MatchString(url) && prependBase {
		rqst := ctx.Request()
		if rqshttp, ok := rqst.(*request.HTTP); ok {
			base := strings.TrimRight(rqshttp.RequestURI, "/")
			if base != "" && base != "/" {
				url = base + "/" + strings.TrimLeft(url, "/")
			} else {
				url = "/" + strings.TrimLeft(url, "/")
			}
		}
	}

	if err := h.redirect(ctx, url, code); err != nil {
		return errors.Wrap(err, "Unable to set simple redirect")
	}

	return nil
}

// GotoRoute redirects to a route-based URL
func (h *Redirector) GotoRoute(ctx context.Context, params map[string]interface{}, name string, reset bool, encode bool) error {
	url, err := Router().Assemble(ctx, params, name, reset, encode)
	if err != nil {
		return errors.Wrapf(err, "Unable to set redirect to route '%s'", name)
	}

	if err := h.redirect(ctx, url, h.Code); err != nil {
		return errors.Wrapf(err, "Unable to set redirect to route '%s'", name)
	}

	return nil
}

// sets redirect in response object
func (h *Redirector) redirect(ctx context.Context, url string, code int) error {
	if h.UseAbsoluteURI && !protox.MatchString(url) {
		host := ""
		proto := "https"
		port := 443
		uri := proto + "://" + host
		if (proto == "http" && port != 80) || (proto == "https" && port != 443) {
			// do not append if HTTP_HOST already contains port
			if !strings.Contains(host, ":") {
				uri += ":" + strconv.Itoa(port)
			}
		}

		url = uri + "/" + strings.TrimLeft(url, "/")
	}

	return ctx.Response().SetRedirect(url, code)
}

func (h *Redirector) checkCode(code int) bool {
	if 300 > code || 307 < code || 304 == code || 306 == code {
		return false
	}

	return true
}

// NewRedirectorHelper creates new Redirector action helper
func NewRedirectorHelper() (HelperInterface, error) {
	return &Redirector{
		name:           "Redirector",
		Code:           302,
		UseAbsoluteURI: false,
	}, nil
}
