package slash

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
)

type (
	// TrailingSlashConfig defines the config for TrailingSlash middleware.
	TrailingSlashConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Status code to be used when redirecting the request.
		// Optional, but when provided the request is redirected using this code.
		RedirectCode int `json:"redirect_code"`
	}
)

var (
	// DefaultTrailingSlashConfig is the default TrailingSlash middleware config.
	DefaultTrailingSlashConfig = TrailingSlashConfig{
		Skipper: skipper.DefaultSkipper,
	}
)

// AddTrailingSlash returns a root level (before router) middleware which adds a
// trailing slash to the request `URL#Path`.
//
// Usage `makross#Pre(AddTrailingSlash())`
func AddTrailingSlash() makross.Handler {
	return AddTrailingSlashWithConfig(DefaultTrailingSlashConfig)
}

// AddTrailingSlashWithConfig returns a AddTrailingSlash middleware with config.
// See `AddTrailingSlash()`.
func AddTrailingSlashWithConfig(config TrailingSlashConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		url := req.URL
		path := url.Path
		qs := c.QueryString()
		if path != "/" && path[len(path)-1] != '/' {
			path += "/"
			uri := path
			if qs != "" {
				uri += "?" + qs
			}

			// Redirect
			if config.RedirectCode != 0 {
				return c.Redirect(uri, config.RedirectCode)
			}

			// Forward
			req.RequestURI = uri
			url.Path = path
		}
		return c.Next()
	}
}

// RemoveTrailingSlash returns a root level (before router) middleware which removes
// a trailing slash from the request URI.
//
// Usage `makross#Pre(RemoveTrailingSlash())`
func RemoveTrailingSlash() makross.Handler {
	return RemoveTrailingSlashWithConfig(TrailingSlashConfig{})
}

// RemoveTrailingSlashWithConfig returns a RemoveTrailingSlash middleware with config.
// See `RemoveTrailingSlash()`.
func RemoveTrailingSlashWithConfig(config TrailingSlashConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		url := req.URL
		path := url.Path
		qs := c.QueryString()
		l := len(path) - 1
		if l >= 0 && path != "/" && path[l] == '/' {
			path = path[:l]
			uri := path
			if qs != "" {
				uri += "?" + qs
			}

			// Redirect
			if config.RedirectCode != 0 {
				return c.Redirect(uri, config.RedirectCode)
			}

			// Forward
			req.RequestURI = uri
			url.Path = path
		}
		return c.Next()
	}
}
