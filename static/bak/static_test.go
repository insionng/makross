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
	c := e.NewContext(req, rec)
	config := StaticConfig{
		Root: "../public",
	}

	// Directory
	h := StaticWithConfig(config)
	if assert.NoError(t, h(c)) {
		//fmt.Println("rec.Body.String()>", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "Makross")
	}

	// File found
	req = httptest.NewRequest(makross.GET, "/images/makross.jpg", nil)
	rec = httptest.NewRecorder()
	m := makross.New()
	m.Use(Static("../public"))
	m.ServeHTTP(rec, req)
	c = e.NewContext(req, rec)
	err := h(c)
	if assert.NoError(t, err) {
		assert.Equal(t, makross.StatusOK, rec.Code)
		println(rec.Header().Get(makross.HeaderContentLength))
		assert.Equal(t, rec.Header().Get(makross.HeaderContentLength), "91808")
	}

	// File not found
	req = httptest.NewRequest(makross.GET, "/none", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	he := h(c).(makross.HTTPError)
	assert.Equal(t, makross.StatusNotFound, he.StatusCode)

	// HTML5
	req = httptest.NewRequest(makross.GET, "/random", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.HTML5 = true
	h = StaticWithConfig(config)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Makross")
	}

	// Browse
	req = httptest.NewRequest(makross.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	config.Root = "../public/certs"
	config.Browse = true
	h = StaticWithConfig(config)
	if assert.NoError(t, h(c)) {
		assert.Equal(t, makross.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "cert.pem")
	}
}
