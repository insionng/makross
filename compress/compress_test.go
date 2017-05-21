package compress

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"testing"

	"github.com/insionng/makross"
	"github.com/insionng/makross/test"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	e := makross.New()
	req := test.NewRequest(makross.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Accept-Encoding header
	h := Gzip()(func(c makross.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)
	assert.Equal(t, "test", rec.Body.String())

	req = test.NewRequest(makross.GET, "/", nil)
	req.Header().Set(makross.HeaderAcceptEncoding, "gzip")
	rec = test.NewResponseRecorder()
	c = e.NewContext(req, rec)

	// Gzip
	h(c)
	assert.Equal(t, "gzip", rec.Header().Get(makross.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(makross.HeaderContentType), makross.MIMETextPlain)
	r, err := gzip.NewReader(rec.Body)
	defer r.Close()
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		assert.Equal(t, "test", buf.String())
	}
}

func TestGzipNoContent(t *testing.T) {
	e := makross.New()
	req := test.NewRequest(makross.GET, "/", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c makross.Context) error {
		return c.NoContent(http.StatusOK)
	})
	if assert.NoError(t, h(c)) {
		assert.Empty(t, rec.Header().Get(makross.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(makross.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzipErrorReturned(t *testing.T) {
	e := makross.New()
	e.Use(Gzip())
	e.GET("/", func(c makross.Context) error {
		return makross.NewHTTPError(http.StatusInternalServerError, "error")
	})
	req := test.NewRequest(makross.GET, "/", nil)
	rec := test.NewResponseRecorder()
	e.ServeHTTP(req, rec)
	assert.Empty(t, rec.Header().Get(makross.HeaderContentEncoding))
	assert.Equal(t, "error", rec.Body.String())
}
