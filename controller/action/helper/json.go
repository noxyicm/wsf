package helper

import (
	"encoding/json"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/session"
	"wsf/view"
)

const (
	// TYPEJson represents JSON action helper
	TYPEJson = "json"
)

func init() {
	Register(TYPEJson, NewJSON)
}

// JSON is a action helper that handles sending json response
type JSON struct {
	name             string
	View             view.Interface
	actionController ControllerInterface
	suppressExit     bool
}

// Name returns helper name
func (h *JSON) Name() string {
	return h.name
}

// Init the helper
func (h *JSON) Init(options map[string]interface{}) error {
	return nil
}

// PreDispatch do dispatch preparations
func (h *JSON) PreDispatch() error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *JSON) PostDispatch() error {
	return nil
}

// SetController sets action controller
func (h *JSON) SetController(ctrl ControllerInterface) error {
	h.actionController = ctrl
	return nil
}

// Controller returns action controller
func (h *JSON) Controller() ControllerInterface {
	return h.actionController
}

// Request returns request object
func (h *JSON) Request() request.Interface {
	return h.Controller().Request()
}

// Response return response object
func (h *JSON) Response() response.Interface {
	return h.Controller().Response()
}

// Session return session object
func (h *JSON) Session() session.Interface {
	return h.Controller().Session()
}

// Send writes encoded data into to response
func (h *JSON) Send(data interface{}, keepLayouts bool, encodeData bool) error {
	encoded, err := h.Encode(data, encodeData)
	if err != nil {
		return err
	}

	response := h.Response()
	if response == nil {
		return errors.New("[JSON] Response object is undefined")
	}
	response.SetBody(encoded)
	response.SetHeader("Content-Type", "application/json; charset=utf-8")

	if !keepLayouts && h.Controller().HasHelper("viewRenderer") {
		if vr := h.Controller().Helper("viewRenderer"); vr != nil {
			vr.(*ViewRenderer).SetNoRender(true)
		}
	}

	if !h.suppressExit {
		//response.SetResponseCode(200)
		//response.Write()
	}

	return nil
}

// Encode encodes data into json
func (h *JSON) Encode(data interface{}, encodeData bool) (encoded []byte, err error) {
	var ok bool
	if encodeData {
		encoded, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	} else {
		if encoded, ok = data.([]byte); !ok {
			return nil, errors.New("[JSON] With encodeData = false encoded data should be of type '[]byte'")
		}
	}

	return encoded, nil
}

// NewJSON creates new JSON action helper
func NewJSON() (Interface, error) {
	return &JSON{
		name:         "json",
		suppressExit: false,
	}, nil
}

// JSONResponse represents json response
type JSONResponse struct {
	StatusCode int
	Version    string
	BasePath   string
	Status     int
	Message    string
	URL        string
	Data       interface{}
}
