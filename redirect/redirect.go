package redirect

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
	"github.com/insionng/makross/slash"
)

type (
	// RedirectConfig defines the config for Redirect middleware.
	RedirectConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Status code to be used when redirecting the request.
		// Optional. Default value http.StatusMovedPermanently.
		Code int `json:"code"`
	}
)

const (
	www = "www"
)

var (
	// DefaultRedirectConfig is the default Redirect middleware config.
	DefaultRedirectConfig = RedirectConfig{
		Skipper: skipper.DefaultSkipper,
		Code:    makross.StatusMovedPermanently,
	}
)

// HTTPSRedirect redirects http requests to https.
// For example, http://labstack.com will be redirect to https://labstack.com.
//
// Usage `makross#Pre(HTTPSRedirect())`
func HTTPSRedirect() makross.Handler {
	return HTTPSRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSRedirect()`.
func HTTPSRedirectWithConfig(config RedirectConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = slash.DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		host := req.Host
		uri := req.RequestURI
		if !c.IsTLS() {
			return c.Redirect("https://"+host+uri, config.Code)
		}
		return c.Next()
	}
}

// HTTPSWWWRedirect redirects http requests to https www.
// For example, http://labstack.com will be redirect to https://www.labstack.com.
//
// Usage `makross#Pre(HTTPSWWWRedirect())`
func HTTPSWWWRedirect() makross.Handler {
	return HTTPSWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSWWWRedirect()`.
func HTTPSWWWRedirectWithConfig(config RedirectConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = slash.DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		host := req.Host
		uri := req.RequestURI
		if !c.IsTLS() && host[:3] != www {
			return c.Redirect("https://www."+host+uri, config.Code)
		}
		return c.Next()
	}

}

// HTTPSNonWWWRedirect redirects http requests to https non www.
// For example, http://www.labstack.com will be redirect to https://labstack.com.
//
// Usage `makross#Pre(HTTPSNonWWWRedirect())`
func HTTPSNonWWWRedirect() makross.Handler {
	return HTTPSNonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSNonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSNonWWWRedirect()`.
func HTTPSNonWWWRedirectWithConfig(config RedirectConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = slash.DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		host := req.Host
		uri := req.RequestURI
		if !c.IsTLS() {
			if host[:3] == www {
				return c.Redirect("https://"+host[4:]+uri, config.Code)
			}
			return c.Redirect("https://"+host+uri, config.Code)
		}
		return c.Next()
	}

}

// WWWRedirect redirects non www requests to www.
// For example, http://labstack.com will be redirect to http://www.labstack.com.
//
// Usage `makross#Pre(WWWRedirect())`
func WWWRedirect() makross.Handler {
	return WWWRedirectWithConfig(DefaultRedirectConfig)
}

// WWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `WWWRedirect()`.
func WWWRedirectWithConfig(config RedirectConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = slash.DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		scheme := c.Scheme()
		host := req.Host
		if host[:3] != www {
			uri := req.RequestURI
			return c.Redirect(scheme+"://www."+host+uri, config.Code)
		}
		return c.Next()
	}

}

// NonWWWRedirect redirects www requests to non www.
// For example, http://www.labstack.com will be redirect to http://labstack.com.
//
// Usage `makross#Pre(NonWWWRedirect())`
func NonWWWRedirect() makross.Handler {
	return NonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// NonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `NonWWWRedirect()`.
func NonWWWRedirectWithConfig(config RedirectConfig) makross.Handler {
	if config.Skipper == nil {
		config.Skipper = slash.DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		scheme := c.Scheme()
		host := req.Host
		if host[:3] == www {
			uri := req.RequestURI
			return c.Redirect(scheme+"://"+host[4:]+uri, config.Code)
		}
		return c.Next()
	}

}
