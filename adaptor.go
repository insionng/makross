package makross

import (
	"net/http"
)

// WrapHTTPHandler wraps `http.Handler` into `makross.Handler`.
func WrapHTTPHandler(handler http.Handler) Handler {
	return func(c *Context) error {
		handler.ServeHTTP(c.Response, c.Request)
		return nil
	}
}
