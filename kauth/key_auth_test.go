package kauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestKeyAuth(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	})
	config := KeyAuthConfig{
		Validator: func(key string, c *makross.Context) (error, bool) {
			return nil, key == "valid-key"
		},
	}
	h := KeyAuthWithConfig(config)

	// Valid key
	auth := DefaultKeyAuthConfig.AuthScheme + " " + "valid-key"
	req.Header.Set(makross.HeaderAuthorization, auth)
	assert.NoError(t, h(c))

	// Invalid key
	auth = DefaultKeyAuthConfig.AuthScheme + " " + "invalid-key"
	req.Header.Set(makross.HeaderAuthorization, auth)
	he := h(c).(makross.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.StatusCode())

	// Missing Authorization header
	req.Header.Del(makross.HeaderAuthorization)
	he = h(c).(makross.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.StatusCode())

	// Key from custom header
	config.KeyLookup = "header:API-Key"
	c = e.NewContext(req, res, func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	})
	h = KeyAuthWithConfig(config)
	req.Header.Set("API-Key", "valid-key")
	assert.NoError(t, h(c))

	// Key from query string
	config.KeyLookup = "query:key"
	c = e.NewContext(req, res, func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	})
	h = KeyAuthWithConfig(config)
	q := req.URL.Query()
	q.Add("key", "valid-key")
	req.URL.RawQuery = q.Encode()
	assert.NoError(t, h(c))
}
