package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// WSFResponse does work for sending response header.
type WSFResponse struct {
	http.ResponseWriter
	Context    *Context
	Status     int
	EnableGzip bool
}

// NewWSFResponse returns new WSFResponse.
// it contains nothing now.
func NewWSFResponse() *WSFResponse {
	return &WSFResponse{}
}

// JSON writes json to response body.
// if encoding is true, it converts utf-8 to \u0000 type.
func (r *WSFResponse) JSON(data interface{}, hasIndent bool, encoding bool) error {
	r.Header().Set("Content-Type", "application/json; charset=utf-8")
	var content []byte
	var err error
	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}
	if err != nil {
		http.Error(r, err.Error(), http.StatusInternalServerError)
		return err
	}
	if encoding {
		content = []byte(stringsToJSON(string(content)))
	}
	return r.Body(content)
}

// YAML writes yaml to response body.
func (r *WSFResponse) YAML(data interface{}) error {
	r.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
	var content []byte
	var err error
	content, err = yaml.Marshal(data)
	if err != nil {
		http.Error(r, err.Error(), http.StatusInternalServerError)
		return err
	}
	return r.Body(content)
}

// JSONP writes jsonp to response body.
func (r *WSFResponse) JSONP(data interface{}, hasIndent bool) error {
	r.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	var content []byte
	var err error
	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}

	if err != nil {
		http.Error(r, err.Error(), http.StatusInternalServerError)
		return err
	}

	callback := r.Context.Request.Query("callback")
	if callback == "" {
		return errors.New(`"callback" parameter required`)
	}

	callback = template.JSEscapeString(callback)
	callbackContent := bytes.NewBufferString(" if(window." + callback + ")" + callback)
	callbackContent.WriteString("(")
	callbackContent.Write(content)
	callbackContent.WriteString(");\r\n")
	return r.Body(callbackContent.Bytes())
}

// XML writes xml string to response body.
func (r *WSFResponse) XML(data interface{}, hasIndent bool) error {
	r.Header().Set("Content-Type", "application/xml; charset=utf-8")
	var content []byte
	var err error
	if hasIndent {
		content, err = xml.MarshalIndent(data, "", "  ")
	} else {
		content, err = xml.Marshal(data)
	}

	if err != nil {
		http.Error(r, err.Error(), http.StatusInternalServerError)
		return err
	}
	return r.Body(content)
}

// ServeFormatted serve YAML, XML OR JSON, depending on the value of the Accept header
func (r *WSFResponse) ServeFormatted(data interface{}, hasIndent bool, hasEncode ...bool) {
	accept := r.Context.Request.Header.Set("Accept")
	switch accept {
	case ApplicationYAML:
		r.YAML(data)
	case ApplicationXML, TextXML:
		r.XML(data, hasIndent)
	default:
		r.JSON(data, hasIndent, len(hasEncode) > 0 && hasEncode[0])
	}
}

// Download forces response for download file.
// it prepares the download response header automatically.
func (r *WSFResponse) Download(file string, filename ...string) {
	// check get file error, file not found or other error.
	if _, err := os.Stat(file); err != nil {
		http.ServeFile(r, r.Context.Request, file)
		return
	}

	var fName string
	if len(filename) > 0 && filename[0] != "" {
		fName = filename[0]
	} else {
		fName = filepath.Base(file)
	}

	fn := url.PathEscape(fName)
	if fName == fn {
		fn = "filename=" + fn
	} else {
		/**
		  The parameters "filename" and "filename*" differ only in that
		  "filename*" uses the encoding defined in [RFC5987], allowing the use
		  of characters not present in the ISO-8859-1 character set
		  ([ISO-8859-1]).
		*/
		fn = "filename=" + fName + "; filename*=utf-8''" + fn
	}

	r.Header().Set("Content-Disposition", "attachment; "+fn)
	r.Header().Set("Content-Description", "File Transfer")
	r.Header().Set("Content-Type", "application/octet-stream")
	r.Header().Set("Content-Transfer-Encoding", "binary")
	r.Header().Set("Expires", "0")
	r.Header().Set("Cache-Control", "must-revalidate")
	r.Header().Set("Pragma", "public")
	http.ServeFile(r, r.Context.Request, file)
}

// ContentType sets the content type from ext string.
// MIME type is given in mime package.
func (r *WSFResponse) ContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		r.Header().Set("Content-Type", ctype)
	}
}
