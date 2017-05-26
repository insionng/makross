package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHTTPSRedirect(t *testing.T) {
	e := makross.New()
	next := func(c *makross.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Host = "at3.net"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, next)
	HTTPSRedirect()(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://at3.net/", res.Header().Get(makross.HeaderLocation))
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	e := makross.New()
	next := func(c *makross.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Host = "at3.net"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, next)
	HTTPSWWWRedirect()(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://www.at3.net/", res.Header().Get(makross.HeaderLocation))
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	e := makross.New()
	next := func(c *makross.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Host = "www.at3.net"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, next)
	HTTPSNonWWWRedirect()(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://at3.net/", res.Header().Get(makross.HeaderLocation))
}

func TestRedirectWWWRedirect(t *testing.T) {
	e := makross.New()
	next := func(c *makross.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Host = "at3.net"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, next)
	WWWRedirect()(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://www.at3.net/", res.Header().Get(makross.HeaderLocation))
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	e := makross.New()
	next := func(c *makross.Context) (err error) {
		return c.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Host = "www.at3.net"
	res := httptest.NewRecorder()
	c := e.NewContext(req, res, next)
	NonWWWRedirect()(c)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://at3.net/", res.Header().Get(makross.HeaderLocation))
}
