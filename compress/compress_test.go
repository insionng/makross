package compress

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	e := makross.New()
	req := httptest.NewRequest(makross.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Skip if no Accept-Encoding header
	h := Gzip()(func(c makross.Context) error {
		c.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(c)
	assert.Equal(t, "test", rec.Body.String())

	// Gzip
	req = httptest.NewRequest(makross.GET, "/", nil)
	req.Header.Set(makross.HeaderAcceptEncoding, gzipScheme)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h(c)
	assert.Equal(t, gzipScheme, rec.Header().Get(makross.HeaderContentEncoding))
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
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Header.Set(makross.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := Gzip()(func(c makross.Context) error {
		return c.NoContent(http.StatusNoContent)
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
		return makross.ErrNotFound
	})
	req := httptest.NewRequest(makross.GET, "/", nil)
	req.Header.Set(makross.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(makross.HeaderContentEncoding))
}

// Issue #806
func TestGzipWithStatic(t *testing.T) {
	e := makross.New()
	e.Use(Gzip())
	e.Static("/test", "../_fixture/images")
	req := httptest.NewRequest(makross.GET, "/test/walle.png", nil)
	req.Header.Set(makross.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	defer r.Close()
	want, err := ioutil.ReadFile("../_fixture/images/walle.png")
	if assert.NoError(t, err) {
		var buf bytes.Buffer
		buf.ReadFrom(r)
		assert.Equal(t, want, buf.Bytes())
	}
}
