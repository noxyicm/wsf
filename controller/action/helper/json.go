package helper

import (
	"encoding/json"
	"wsf/context"
	"wsf/errors"
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
	Abstract

	suppressExit bool
}

// PostDispatch do dispatch aftermath
func (h *JSON) PostDispatch(ctx context.Context) error {
	if h.shouldRender(ctx) {
		encoded, err := h.Encode(ctx.Data(), true)
		if err != nil {
			return err
		}

		rsp := ctx.Response()
		if rsp == nil {
			return errors.New("[JSON] Response object is undefined")
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
			return nil, errors.New("[JSON] With encodeData = false encoded data should be of type '[]byte'")
		}
	}

	return encoded, nil
}

// NewJSON creates new JSON action helper
func NewJSON() (Interface, error) {
	js := &JSON{
		suppressExit: false,
	}

	js.name = "json"
	return js, nil
}

// JSONResponse represents json response
type JSONResponse struct {
	Version   string      `json:"version"`
	BasePath  string      `json:"basePath"`
	ErrorCode int         `json:"errorCode"`
	Message   string      `json:"message"`
	URI       string      `json:"uri"`
	Data      interface{} `json:"data"`
}
