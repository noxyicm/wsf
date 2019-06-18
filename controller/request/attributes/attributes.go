package attributes

import (
	"context"
	"errors"
	"net/http"
)

const contextKey attributeKey = iota

type attributeKey int

type attributes map[string]interface{}

// Delete deletes values associated with attribute key
func (v attributes) Delete(key string) {
	if v == nil {
		return
	}

	v.del(key)
}

func (v attributes) get(key string) interface{} {
	if v == nil {
		return ""
	}

	return v[key]
}

func (v attributes) set(key string, value interface{}) {
	v[key] = value
}

func (v attributes) del(key string) {
	delete(v, key)
}

// Init returns request with new context and attribute bag
func Init(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey, attributes{}))
}

// All returns all context attributes
func All(r *http.Request) map[string]interface{} {
	v := r.Context().Value(contextKey)
	if v == nil {
		return attributes{}
	}

	return v.(attributes)
}

// Get gets the value from request context
func Get(r *http.Request, key string) interface{} {
	v := r.Context().Value(contextKey)
	if v == nil {
		return nil
	}

	return v.(attributes).get(key)
}

// Set sets the key to value
func Set(r *http.Request, key string, value interface{}) error {
	v := r.Context().Value(contextKey)
	if v == nil {
		return errors.New("Unable to find `psr:attributes` context key")
	}

	v.(attributes).set(key, value)
	return nil
}
