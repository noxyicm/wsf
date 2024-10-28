package http

const (
	// ApplicationHTML states for "text/html" context
	ApplicationHTML = "text/html"

	// ApplicationJSON states for "application/json" context
	ApplicationJSON = "application/json"

	// ApplicationXML states for "application/xml" context
	ApplicationXML = "application/xml"

	// ApplicationYAML states for "application/x-yaml" context
	ApplicationYAML = "application/x-yaml"

	// TextXML states for "text/xml" context
	TextXML = "text/xml"
)

// NewContext return the Context with Input and Output
func NewContext() *Context {
	return &Context{
		Request:  NewWSFRequest(),
		Response: NewWSFResponse(),
	}
}

// Context Http request context struct including WSFRequest, WSFResponse.
type Context struct {
	Request    *WSFRequest
	Response   *WSFResponse
	_xsrfToken string
}
