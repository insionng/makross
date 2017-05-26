// Package makross is a high productive and modular web framework in Golang.

package slazh

import (
	"net/http"
	"strings"

	"github.com/insionng/makross"
)

// Remover returns a handler that removes the trailing slazh (if any) from the requested URL.
// The handler will redirect the browser to the new URL without the trailing slazh.
// The status parameter should be either http.StatusMovedPermanently (301) or http.StatusFound (302), which is to
// be used for redirecting GET requests. For other requests, the status code will be http.StatusTemporaryRedirect (307).
// If the original URL has no trailing slazh, the handler will do nothing. For example,
//
//     import (
//         "net/http"
//         "github.com/insionng/makross"
//         "github.com/insionng/makross/slazh"
//     )
//
//     r := makross.New()
//     r.Use(slazh.Remover(http.StatusMovedPermanently))
//
// Note that Remover relies on HTTP redirection to remove the trailing slazhes.
// If you do not want redirection, please set `Router.IgnoreTrailingslazh` to be true without using Remover.
func Remover(status int) makross.Handler {
	return func(c *makross.Context) error {
		if c.Request.URL.Path != "/" && strings.HasSuffix(c.Request.URL.Path, "/") {
			if c.Request.Method != "GET" {
				status = http.StatusTemporaryRedirect
			}
			http.Redirect(c.Response, c.Request, strings.TrimRight(c.Request.URL.Path, "/"), status)
			c.Abort()
		}
		return nil
	}
}
