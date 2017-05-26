package requestid

import (
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	handler := func(c *makross.Context) error {
		return c.String("test", makross.StatusOK)
	}
	c := e.NewContext(req, rec, handler)
	h := RequestIDWithConfig(RequestIDConfig{})
	h(c)
	assert.Len(t, rec.Header().Get(makross.HeaderXRequestID), 32)

	// Custom generator
	c = e.NewContext(req, rec, handler)
	h = RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h(c)
	assert.Equal(t, rec.Header().Get(makross.HeaderXRequestID), "customGenerator")
}
