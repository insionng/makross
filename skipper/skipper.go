package skipper

import "github.com/insionng/makross"

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(c *makross.Context) bool
)

// defaultSkipper returns false which processes the middleware.
func DefaultSkipper(c *makross.Context) bool {
	return false
}
