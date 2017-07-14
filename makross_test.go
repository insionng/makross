// Package makross is a high productive and modular web framework in Golang.

package makross

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name"`
	}
)

const (
	userJSON       = `{"id":1,"name":"Jon Snow"}`
	userXML        = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm       = `id=1&name=Jon Snow`
	invalidContent = "invalid content"
)

func TestRouterNotFound(t *testing.T) {
	r := New()
	r.Get("/users", func(c *Context) error {
		return c.String("ok")
	})
	r.Post("/users", func(c *Context) error {
		return c.String("ok")
	})
	r.NotFound(MethodNotAllowedHandler, NotFoundHandler)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, StatusMethodNotAllowed, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET, OPTIONS, POST", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, StatusOK, res.Code, "HTTP status code")

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/posts", nil)
	r.ServeHTTP(res, req)
	assert.Equal(t, "", res.Header().Get("Allow"), "Allow header")
	assert.Equal(t, StatusNotFound, res.Code, "HTTP status code")
}

func TestRouterUse(t *testing.T) {
	m := New()
	assert.Equal(t, 2, len(m.notFoundHandlers))
	m.Use(NotFoundHandler)
	assert.Equal(t, 3, len(m.notFoundHandlers))
}

func TestRouterRoute(t *testing.T) {
	r := New()
	r.Get("/users").Name("users")
	assert.NotNil(t, r.Route("users"))
	assert.Nil(t, r.Route("users2"))
}

func TestRouterAdd(t *testing.T) {
	m := New()
	assert.Equal(t, 0, m.maxParams)
	m.add("GET", "/users/<id>", nil)
	assert.Equal(t, 1, m.maxParams)
}

func TestRouterFind(t *testing.T) {
	r := New()
	r.add("GET", "/users/<id>", []Handler{NotFoundHandler})
	pvalues := make([]string, 10)
	handlers, pnames := r.find("GET", "/users/1", pvalues)
	assert.Equal(t, 1, len(handlers))
	if assert.Equal(t, 1, len(pnames)) {
		assert.Equal(t, "id", pnames[0])
	}
	assert.Equal(t, "1", pvalues[0])
}

func TestRouterHandleError(t *testing.T) {
	m := New()
	res := httptest.NewRecorder()
	c := m.NewContext(nil, res)
	m.HandleError(c, errors.New("abc"))
	assert.Equal(t, StatusInternalServerError, res.Code)

	res = httptest.NewRecorder()
	c = m.NewContext(nil, res)
	m.HandleError(c, NewHTTPError(http.StatusNotFound))
	assert.Equal(t, StatusNotFound, res.Code)
}

func TestHTTPHandler(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)

	m := New()
	c := m.NewContext(req, res)

	h1 := HTTPHandlerFunc(http.NotFound)
	assert.Nil(t, h1(c))
	assert.Equal(t, StatusNotFound, res.Code)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)

	c = m.NewContext(req, res)
	h2 := HTTPHandler(http.NotFoundHandler())
	assert.Nil(t, h2(c))
	assert.Equal(t, StatusNotFound, res.Code)
}
