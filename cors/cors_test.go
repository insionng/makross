package cors

import (
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	e := makross.New()

	// Wildcard origin
	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec, makross.NotFoundHandler)
	h := CORS()
	h(c)
	assert.Equal(t, "*", rec.Header().Get(makross.HeaderAccessControlAllowOrigin))

	// Allow origins
	req = httptest.NewRequest(makross.GET, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, makross.NotFoundHandler)
	h = CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"localhost"},
	})
	req.Header.Set(makross.HeaderOrigin, "localhost")
	h(c)
	assert.Equal(t, "localhost", rec.Header().Get(makross.HeaderAccessControlAllowOrigin))

	// Preflight request
	req = httptest.NewRequest(makross.OPTIONS, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec, makross.NotFoundHandler)
	req.Header.Set(makross.HeaderOrigin, "localhost")
	req.Header.Set(makross.HeaderContentType, makross.MIMEApplicationJSON)
	cors := CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	cors(c)
	assert.Equal(t, "localhost", rec.Header().Get(makross.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(makross.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(makross.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(makross.HeaderAccessControlMaxAge))
}
