package logger

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/insionng/makross"
	"github.com/insionng/makross/libraries/gommon/color"
	"github.com/insionng/makross/skipper"
	"github.com/valyala/fasttemplate"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper skipper.Skipper

		// Tags to constructed the logger format.
		//
		// - time_unix
		// - time_unix_nano
		// - time_rfc3339
		// - time_rfc3339_nano
		// - id (Request ID)
		// - remote_ip
		// - uri
		// - host
		// - method
		// - path
		// - referer
		// - user_agent
		// - status
		// - latency (In nanoseconds)
		// - latency_human (Human readable)
		// - bytes_in (Bytes received)
		// - bytes_out (Bytes sent)
		// - header:<NAME>
		// - query:<NAME>
		// - form:<NAME>
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		Format string `json:"format"`

		// Output is a writer where logs in JSON format are written.
		// Optional. Default value os.Stdout.
		Output io.Writer

		template *fasttemplate.Template
		colorer  *color.Color
		pool     *sync.Pool
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper: skipper.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
		Output:  os.Stdout,
		colorer: color.New(),
	}
)

// Logger returns a middleware that logs HTTP requests.
func Logger() makross.Handler {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) makross.Handler {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.colorer = color.New()
	config.colorer.SetOutput(config.Output)
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(c *makross.Context) (err error) {
		if config.Skipper(c) {
			return c.Next()
		}

		req := c.Request
		res := c.Response
		start := time.Now()
		if err = c.Next(); err != nil {
			c.HandleError(err)
		}
		stop := time.Now()
		buf := config.pool.Get().(*bytes.Buffer)
		buf.Reset()
		defer config.pool.Put(buf)

		if _, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "time_unix":
				return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
			case "time_unix_nano":
				return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
			case "time_rfc3339":
				return buf.WriteString(time.Now().Format(time.RFC3339))
			case "time_rfc3339_nano":
				return buf.WriteString(time.Now().Format(time.RFC3339Nano))
			case "id":
				id := req.Header.Get(makross.HeaderXRequestID)
				if id == "" {
					id = res.Header().Get(makross.HeaderXRequestID)
				}
				return buf.WriteString(id)
			case "remote_ip":
				return buf.WriteString(c.RealIP())
			case "host":
				return buf.WriteString(req.Host)
			case "uri":
				return buf.WriteString(req.RequestURI)
			case "method":
				return buf.WriteString(req.Method)
			case "path":
				p := req.URL.Path
				if p == "" {
					p = "/"
				}
				return buf.WriteString(p)
			case "referer":
				return buf.WriteString(req.Referer())
			case "user_agent":
				return buf.WriteString(req.UserAgent())
			case "status":
				n := res.Status
				s := config.colorer.Green(n)
				switch {
				case n >= 500:
					s = config.colorer.Red(n)
				case n >= 400:
					s = config.colorer.Yellow(n)
				case n >= 300:
					s = config.colorer.Cyan(n)
				}
				return buf.WriteString(s)
			case "latency":
				l := stop.Sub(start)
				return buf.WriteString(strconv.FormatInt(int64(l), 10))
			case "latency_human":
				return buf.WriteString(stop.Sub(start).String())
			case "bytes_in":
				cl := req.Header.Get(makross.HeaderContentLength)
				if cl == "" {
					cl = "0"
				}
				return buf.WriteString(cl)
			case "bytes_out":
				return buf.WriteString(strconv.FormatInt(res.Size, 10))
			default:
				switch {
				case strings.HasPrefix(tag, "header:"):
					return buf.Write([]byte(c.Request.Header.Get(tag[7:])))
				case strings.HasPrefix(tag, "query:"):
					return buf.Write([]byte(c.Query(tag[6:])))
				case strings.HasPrefix(tag, "form:"):
					return buf.Write([]byte(c.Form(tag[5:])))
				case strings.HasPrefix(tag, "cookie:"):
					cookie, err := c.GetCookie(tag[7:])
					if err == nil {
						return buf.Write([]byte(cookie.Value))
					}
				}
			}
			return 0, nil
		}); err != nil {
			return
		}

		_, err = config.Output.Write(buf.Bytes())
		return
	}
}
