package controller

import (
	"encoding/json"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
)

const (
	// TYPEHelperJSON represents JSON action helper
	TYPEHelperJSON = "json"
	// JSONResponseKey is a string key in context data structure
	// that contains data to encode
	JSONResponseKey = "jsonresponse"
)

func init() {
	RegisterHelper(TYPEHelperJSON, NewJSONHelper)
}

// JSON is a action helper that handles sending json response
type JSON struct {
	name         string
	suppressExit bool
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
func (h *JSON) PreDispatch(ctx context.Context) error {
	return nil
}

// PostDispatch do dispatch aftermath
func (h *JSON) PostDispatch(ctx context.Context) error {
	if h.shouldRender(ctx) {
		encoded, err := h.Encode(ctx.DataValue(JSONResponseKey), true)
		if err != nil {
			return err
		}

		rsp := ctx.Response()
		if rsp == nil {
			return errors.Errorf("[%s] Response object is undefined", h.name)
		}
		rsp.SetBody(encoded)
		rsp.SetHeader("Content-Type", "application/json; charset=utf-8")

		ctx.SetParam("noViewRenderer", true)
	}

	return nil
}

// Should the JSON encode data and set it as response body
func (h *JSON) shouldRender(ctx context.Context) bool {
	return ctx.ParamBool("isJSONResponse") && ctx.Request().IsDispatched() && !ctx.Response().IsRedirect()
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
			return nil, errors.Errorf("[%s] With encodeData = false encoded data should be of type '[]byte'", h.name)
		}
	}

	return encoded, nil
}

// NewJSONHelper creates new JSON action helper
func NewJSONHelper() (HelperInterface, error) {
	js := &JSON{
		name:         "JSON",
		suppressExit: false,
	}

	return js, nil
}

// JSONResponse represents json response
type JSONResponse struct {
	Version   string      `json:"version"`
	BasePath  string      `json:"basePath"`
	ErrorCode int         `json:"errorCode"`
	Error     interface{} `json:"error"`
	Message   string      `json:"message"`
	URI       string      `json:"uri"`
	Data      interface{} `json:"data"`
}
