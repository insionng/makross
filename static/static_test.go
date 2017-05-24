package static

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := makross.NewContext(rec, req)
	config := StaticConfig{
		Root: "../public",
	}

	// Directory
	h := StaticWithConfig(config)(makross.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// File found
	req = httptest.NewRequest(makross.GET, "/images/walle.png", nil)
	rec = httptest.NewRecorder()
	c = makross.NewContext(rec, req)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, makross.StatusOK, rec.Code)
		assert.Equal(t, rec.Header().Get(makross.HeaderContentLength), "219885")
	}

	// File not found
	req = httptest.NewRequest(makross.GET, "/none", nil)
	rec = httptest.NewRecorder()
	c = makross.NewContext(rec, req)
	he := h(c).(*makross.HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)

	// HTML5
	req = httptest.NewRequest(makross.GET, "/random", nil)
	rec = httptest.NewRecorder()
	c = makross.NewContext(rec, req)
	config.HTML5 = true
	static := StaticWithConfig(config)
	h = static(makross.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, makross.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Echo")
	}

	// Browse
	req = httptest.NewRequest(makross.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = makross.NewContext(rec, req)
	config.Root = "../public/certs"
	config.Browse = true
	static = StaticWithConfig(config)
	h = static(makross.NotFoundHandler)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, makross.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "cert.pem")
	}
}
