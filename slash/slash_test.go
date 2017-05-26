package slash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestAddTrailingSlash(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/add-slash", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec, func(c *makross.Context) error {
		return nil
	})
	h := AddTrailingSlash()
	h(c)
	assert.Equal(t, "/add-slash/", req.URL.Path)
	assert.Equal(t, "/add-slash/", req.RequestURI)

	// With config
	req = httptest.NewRequest(makross.GET, "/add-slash?key=value", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return nil
	})
	h = AddTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/add-slash/?key=value", rec.Header().Get(makross.HeaderLocation))
}

func TestRemoveTrailingSlash(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/remove-slash/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec, func(c *makross.Context) error {
		return nil
	})
	h := RemoveTrailingSlash()
	h(c)
	assert.Equal(t, "/remove-slash", req.URL.Path)
	assert.Equal(t, "/remove-slash", req.RequestURI)

	// With config
	req = httptest.NewRequest(makross.GET, "/remove-slash/?key=value", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return nil
	})
	h = RemoveTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})
	h(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/remove-slash?key=value", rec.Header().Get(makross.HeaderLocation))

	// With bare URL
	req = httptest.NewRequest(makross.GET, "http://localhost", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, func(c *makross.Context) error {
		return nil
	})
	h = RemoveTrailingSlash()
	h(c)
	assert.Equal(t, "", req.URL.Path)
}
