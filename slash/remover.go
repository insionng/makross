// Package makross is a high productive and modular web framework in Golang.

// Package slash provides a trailing slash remover handler for the makross.
package slash

import (
	"net/http"
	"strings"

	"github.com/insionng/makross"
)

// Remover returns a handler that removes the trailing slash (if any) from the requested URL.
// The handler will redirect the browser to the new URL without the trailing slash.
// The status parameter should be either http.StatusMovedPermanently (301) or http.StatusFound (302).
// If the original URL has no trailing slash, the handler will do nothing. For example,
//
//     import (
//         "net/http"
//         "github.com/insionng/makross"
//         "github.com/insionng/makross/slash"
//     )
//
//     r := makross.New()
//     r.Use(slash.Remover(http.StatusMovedPermanently))
func Remover(status int) makross.Handler {
	return func(c *makross.Context) error {
		if c.Request.URL.Path != "/" && strings.HasSuffix(c.Request.URL.Path, "/") {
			http.Redirect(c.Response, c.Request, strings.TrimRight(c.Request.URL.Path, "/"), status)
			c.Abort()
		}
		return nil
	}
}
