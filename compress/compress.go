package compress

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
)

type (
	// GzipConfig defines the config for Gzip middleware.
	GzipConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Gzip compression level.
		// Optional. Default value -1.
		Level int `json:"level"`
	}

	gzipResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

const (
	gzipScheme = "gzip"
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	DefaultGzipConfig = GzipConfig{
		Skipper: skipper.DefaultSkipper,
		Level:   -1,
	}
)

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip() makross.Handler {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithConfig return Gzip middleware with config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultGzipConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		res := c.Response
		res.Header().Add(makross.HeaderVary, makross.HeaderAcceptEncoding)
		if strings.Contains(c.Request.Header.Get(makross.HeaderAcceptEncoding), gzipScheme) {
			res.Header().Add(makross.HeaderContentEncoding, gzipScheme) // Issue #806
			rw := res.Writer
			w, err := gzip.NewWriterLevel(rw, config.Level)
			if err != nil {
				return err
			}
			defer func() {
				if res.Size == 0 {
					if res.Header().Get(makross.HeaderContentEncoding) == gzipScheme {
						res.Header().Del(makross.HeaderContentEncoding)
					}
					// We have to reset response to it's pristine state when
					// nothing is written to body or error is returned.
					// See issue #424, #407.
					res.Writer = rw
					w.Reset(ioutil.Discard)
				}
				w.Close()
			}()
			grw := &gzipResponseWriter{Writer: w, ResponseWriter: rw}
			res.Writer = grw
		}
		return c.Next()
	}
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if code == makross.StatusNoContent { // Issue #489
		w.ResponseWriter.Header().Del(makross.HeaderContentEncoding)
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(makross.HeaderContentType) == "" {
		w.Header().Set(makross.HeaderContentType, http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	w.Writer.(*gzip.Writer).Flush()
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
