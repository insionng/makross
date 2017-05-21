package moverride

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
)

type (
	// MethodOverrideConfig defines the config for MethodOverride middleware.
	MethodOverrideConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Getter is a function that gets overridden method from the request.
		// Optional. Default values MethodFromHeader(makross.HeaderXHTTPMethodOverride).
		Getter MethodOverrideGetter
	}

	// MethodOverrideGetter is a function that gets overridden method from the request
	MethodOverrideGetter func(*makross.Context) string
)

var (
	// DefaultMethodOverrideConfig is the default MethodOverride middleware config.
	DefaultMethodOverrideConfig = MethodOverrideConfig{
		Skipper: skipper.DefaultSkipper,
		Getter:  MethodFromHeader(makross.HeaderXHTTPMethodOverride),
	}
)

// MethodOverride returns a MethodOverride middleware.
// MethodOverride  middleware checks for the overridden method from the request and
// uses it instead of the original method.
//
// For security reasons, only `POST` method can be overridden.
func MethodOverride() makross.Handler {
	return MethodOverrideWithConfig(DefaultMethodOverrideConfig)
}

// MethodOverrideWithConfig returns a MethodOverride middleware with config.
// See: `MethodOverride()`.
func MethodOverrideWithConfig(config MethodOverrideConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultMethodOverrideConfig.Skipper
	}
	if config.Getter == nil {
		config.Getter = DefaultMethodOverrideConfig.Getter
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		if req.Method == makross.POST {
			m := config.Getter(c)
			if len(m) != 0 {
				req.Method = m
			}
		}
		return c.Next()
	}
}

// MethodFromHeader is a `MethodOverrideGetter` that gets overridden method from
// the request header.
func MethodFromHeader(header string) MethodOverrideGetter {
	return func(c *makross.Context) string {
		return c.Request.Header.Get(header)
	}
}

// MethodFromForm is a `MethodOverrideGetter` that gets overridden method from the
// form parameter.
func MethodFromForm(param string) MethodOverrideGetter {
	return func(c *makross.Context) string {
		return c.Form(param)
	}
}

// MethodFromQuery is a `MethodOverrideGetter` that gets overridden method from
// the query parameter.
func MethodFromQuery(param string) MethodOverrideGetter {
	return func(c *makross.Context) string {
		return c.Query(param)
	}
}
