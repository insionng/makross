// Package makross is a high productive and modular web framework in Golang.

package slazh

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestRemover(t *testing.T) {
	h := Remover(http.StatusMovedPermanently)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	m := makross.New()
	c := m.NewContext(req, res)
	err := h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "/users", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	c = m.NewContext(req, res)
	err = h(c)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users", nil)
	c = m.NewContext(req, res)
	err = h(c)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/users/", nil)
	c = m.NewContext(req, res)
	err = h(c)
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "/users", res.Header().Get("Location"))
}
