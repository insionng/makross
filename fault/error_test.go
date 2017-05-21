// Package makross is a high productive and modular web framework in Golang.

package fault

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	var buf bytes.Buffer
	h := ErrorHandler(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := makross.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, makross.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc\n", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = makross.NewContext(res, req, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test", res.Body.String())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	h = ErrorHandler(getLogger(&buf), convertError)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = makross.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "123\n", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	h = ErrorHandler(nil)
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = makross.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc\n", res.Body.String())
	assert.Equal(t, "", buf.String())
}

func Test_writeError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := makross.NewContext(res, req)
	writeError(c, errors.New("abc"))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = makross.NewContext(res, req)
	writeError(c, makross.NewHTTPError(http.StatusNotFound, "xyz"))
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
}

func convertError(c *makross.Context, err error) error {
	return errors.New("123")
}
