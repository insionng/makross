package static

import (
	"fmt"
	"github.com/insionng/makross"
	"github.com/insionng/makross/skipper"
	"net/http"
	"path"
	"strings"
)

type (
	// StaticConfig defines the config for Static middleware.
	StaticConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Root directory from where the static content is served.
		// Required.
		Root string `json:"root"`

		// Index file for serving a directory.
		// Optional. Default value "index.html".
		Index string `json:"index"`

		// Enable HTML5 mode by forwarding all not-found requests to root so that
		// SPA (single-page application) can handle the routing.
		// Optional. Default value false.
		HTML5 bool `json:"html5"`

		// Enable directory browsing.
		// Optional. Default value false.
		Browse bool `json:"browse"`
	}
)

var (
	// DefaultStaticConfig is the default Static middleware config.
	DefaultStaticConfig = StaticConfig{
		Skipper: skipper.DefaultSkipper,
		Index:   "index.html",
	}
)

// Static returns a Static middleware to serves static content from the provided
// root directory.
func Static(root string) makross.Handler {
	c := DefaultStaticConfig
	c.Root = root
	return StaticWithConfig(c)
}

// StaticWithConfig returns a Static middleware with config.
// See `Static()`.
func StaticWithConfig(config StaticConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultStaticConfig.Skipper
	}
	if len(config.Index) == 0 {
		config.Index = DefaultStaticConfig.Index
	}

	return func(c *makross.Context) error {
		if config.Skipper(c) {
			return c.Next()
		}

		fs := http.Dir(config.Root)
		p := c.Request.URL.Path
		if strings.Contains(c.Request.URL.Path, "*") { // If serving from a group, e.g. `/static*`.
			p = c.Parameter(0)

		}
		file := path.Clean(p)
		f, err := fs.Open(file)
		if err != nil {
			// HTML5 mode
			err = c.Next()
			if herr, ok := err.(makross.HTTPError); ok {
				if config.HTML5 && herr.StatusCode() == http.StatusNotFound {
					file = ""
					f, err = fs.Open(file)
				} else {
					return err
				}
			} else {
				return err
			}
		}
		defer f.Close()
		fStat, err := f.Stat()
		if err != nil {
			return err
		}

		if fStat.IsDir() {
			/* NOTE:
			Not checking the Last-Modified header as it caches the response `304` when
			changing different directories for the same path.
			*/
			d := f

			// Index file
			file = path.Join(file, config.Index)
			f, err = fs.Open(file)
			if err == nil {
				// Index file
				if fStat, err = f.Stat(); err != nil {
					return err
				}
			} else if err != nil && config.Browse {
				dirs, err := d.Readdir(-1)
				if err != nil {
					return err
				}

				// Create a directory index
				c.Response.Header().Set(makross.HeaderContentType, makross.MIMETextHTMLCharsetUTF8)
				if _, err = fmt.Fprintf(c.Response, "<pre>\n"); err != nil {
					return err
				}
				for _, d := range dirs {
					name := d.Name()
					color := "#212121"
					if d.IsDir() {
						color = "#e91e63"
						name += "/"
					}
					if _, err = fmt.Fprintf(c.Response, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
						return err
					}
				}
				_, err = fmt.Fprintf(c.Response, "</pre>\n")
				return err
			} else {
				return c.Next()
			}
		}
		//err = c.ServeContent(f, fStat.Name(), fStat.ModTime())
		http.ServeContent(c.Response, c.Request, fStat.Name(), fStat.ModTime(), f)
		c.Abort()
		return nil
	}

}
