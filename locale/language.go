package locale

import (
	"context"
)

type contextKey int

var (
	langCodeContextKey contextKey
	langContextKey     contextKey
)

// Language represents lenguage
type Language struct {
	ID     int
	Label  string
	Code   string
	Locale string
	Icon   string
}

// LangToContext returns a new context with stored language
func LangToContext(ctx context.Context, langCode string, lang *Language) context.Context {
	ctx = context.WithValue(ctx, langCodeContextKey, langCode)
	ctx = context.WithValue(ctx, langContextKey, lang)
	return ctx
}

// LangFromContext returns a language stored in context
func LangFromContext(ctx context.Context) (*Language, bool) {
	v, ok := ctx.Value(langContextKey).(*Language)
	return v, ok
}

// LangCodeFromContext returns a language code stored in context
func LangCodeFromContext(ctx context.Context) string {
	v, ok := ctx.Value(langCodeContextKey).(string)
	if !ok {
		return "uk"
		//return config.App.GetString("application.lang")
	}

	return v
}
