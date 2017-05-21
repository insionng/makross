package compress

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
)

type (
	// GzipConfig defines the config for Gzip makross.
	GzipConfig struct {
		// Skipper defines a function to skip makross.
		Skipper skipper.Skipper

		// Gzip compression level.
		// Optional. Default value -1.
		Level int `json:"level"`
	}

	gzipResponseWriter struct {
		http.Response
		io.Writer
	}
)

var (
	// DefaultGzipConfig is the default Gzip makross config.
	DefaultGzipConfig = GzipConfig{
		Skipper: skipper.DefaultSkipper,
		Level:   -1,
	}
)

// Gzip returns a makross which compresses HTTP response using gzip compression
// scheme.
func Gzip() makross.Handler {
	return GzipWithConfig(DefaultGzipConfig)
}

// GzipWithConfig return Gzip makross with config.
// See: `Gzip()`.
func GzipWithConfig(config GzipConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultGzipConfig.Skipper
	}
	if config.Level == 0 {
		config.Level = DefaultGzipConfig.Level
	}

	pool := gzipPool(config)
	scheme := "gzip"

	return func(next makross.Handler) makross.Handler {
		return func(c makross.Context) error {
			if config.Skipper(c) {
				return c.Next()
			}

			res := c.Response
			res.Header().Add(makross.HeaderVary, makross.HeaderAcceptEncoding)
			if strings.Contains(c.Request.Header.Get(makross.HeaderAcceptEncoding), scheme) {
				rw := res.Writer()
				gw := pool.Get().(*gzip.Writer)
				gw.Reset(rw)
				defer func() {
					if res.Size() == 0 {
						// We have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						// See issue #424, #407.
						res.SetWriter(rw)
						res.Header().Del(makross.HeaderContentEncoding)
						gw.Reset(ioutil.Discard)
					}
					gw.Close()
					pool.Put(gw)
				}()
				g := gzipResponseWriter{Response: res, Writer: gw}
				res.Header().Set(makross.HeaderContentEncoding, scheme)
				res.SetWriter(g)
			}
			return c.Next()
		}
	}
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header.Get(makross.HeaderContentType) == "" {
		g.Header.Set(makross.HeaderContentType, http.DetectContentType(b))
	}
	return g.Writer.Write(b)
}

func gzipPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(ioutil.Discard, config.Level)
			return w
		},
	}
}
