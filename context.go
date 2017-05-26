// Package makross is a high productive and modular web framework in Golang.

package makross

import (
	"bytes"
	ktx "context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	indexPage = "index.html"
)

// Context represents the contextual data and environment while processing an incoming HTTP request.
type Context struct {
	Request  *http.Request // the current request
	Response *Response     // the response writer
	ktx      ktx.Context   // standard context
	makross  *Makross
	pnames   []string               // list of route parameter names
	pvalues  []string               // list of parameter values corresponding to pnames
	data     map[string]interface{} // data items managed by Get and Set
	index    int                    // the index of the currently executing handler in handlers
	handlers []Handler              // the handlers associated with the current route
	writer   DataWriter
}

// Reset sets the request and response of the context and resets all other properties.
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.Response.reset(w)
	c.Request = r
	c.ktx = ktx.Background()
	c.data = nil
	c.index = -1
	c.writer = DefaultDataWriter
}

// NewContext creates a new Context object with the given response, request, and the handlers.
// This method is primarily provided for writing unit tests for handlers.
/*
func NewContext(w http.ResponseWriter, r *http.Request, handlers ...Handler) *Context {
	//c := &Context{handlers: handlers}
	m := New()
	c := &Context{
		Request:  r,
		Response: NewResponse(w, m),
		makross:  m,
		pvalues:  make([]string, m.maxParams),
		handlers: handlers,
	}

	c.Reset(w, r)
	return c
}
*/

// Makross returns the Makross that is handling the incoming HTTP request.
func (c *Context) Makross() *Makross {
	return c.makross
}

// Stop 优雅停止HTTP服务 不超过特定时长
func (c *Context) Stop(times ...int64) error {
	return c.makross.Stop(times...)
}

// Close 立即关闭HTTP服务
func (c *Context) Close() error {
	return c.makross.Server.Close()
}

func (c *Context) Kontext() ktx.Context {
	return c.ktx
}

func (c *Context) SetKontext(ktx ktx.Context) {
	c.ktx = ktx
}

func (c *Context) Handler() Handler {
	return c.handlers[c.index]
}

func (c *Context) SetHandler(h Handler) {
	c.handlers[c.index] = h
}

func (c *Context) NewCookie() *http.Cookie {
	return new(http.Cookie)
}

func (c *Context) GetCookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

func (c *Context) GetCookies() []*http.Cookie {
	return c.Request.Cookies()
}

func (c *Context) Error(status int, message ...interface{}) {
	herr := NewHTTPError(status, message...)
	c.HandleError(herr)
}

func (c *Context) HandleError(err error) {
	c.makross.HandleError(c, err)
}

// RealIP implements `Context#RealIP` function.
func (c *Context) RealIP() string {
	ra := c.Request.RemoteAddr
	if ip := c.Request.Header.Get(HeaderXForwardedFor); len(ip) > 0 {
		ra = ip
	} else if ip := c.Request.Header.Get(HeaderXRealIP); len(ip) > 0 {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// Param returns the named parameter value that is found in the URL path matching the current route.
// If the named parameter cannot be found, an empty string will be returned.
/*
func (c *Context) Param(name string) string {
	for i, n := range c.pnames {
		if n == name {
			return c.pvalues[i]
		}
	}
	return ""
}
*/

// Get returns the named data item previously registered with the context by calling Set.
// If the named data item cannot be found, nil will be returned.
func (c *Context) Get(name string) interface{} {
	return c.data[name]
}

// Set stores the named data item in the context so that it can be retrieved later.
func (c *Context) Set(name string, value interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[name] = value
}

func (c *Context) SetStore(data map[string]interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	for k, v := range data {
		c.data[k] = v
	}
}

func (c *Context) GetStore() map[string]interface{} {
	return c.data
}

func (c *Context) Pull(key string) interface{} {
	return c.makross.data[key]
}

func (c *Context) Push(key string, value interface{}) {
	if c.makross.data == nil {
		c.makross.data = make(map[string]interface{})
	}
	c.makross.data[key] = value
}

func (c *Context) PullStore() map[string]interface{} {
	return c.makross.data
}

func (c *Context) PushStore(data map[string]interface{}) {
	if c.makross.data == nil {
		c.makross.data = make(map[string]interface{})
	}
	for k, v := range data {
		c.makross.data[k] = v
	}
}

func (c *Context) QueryString() string {
	return c.Request.URL.RawQuery
}

// Query returns the first value for the named component of the URL query parameters.
// If key is not present, it returns the specified default value or an empty string.
func (c *Context) Query(name string, defaultValue ...string) string {
	if vs, _ := c.Request.URL.Query()[name]; len(vs) > 0 {
		return vs[0]
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// Form returns the first value for the named component of the query.
// Form reads the value from POST and PUT body parameters as well as URL query parameters.
// The form takes precedence over the latter.
// If key is not present, it returns the specified default value or an empty string.
func (c *Context) Form(key string, defaultValue ...string) string {
	r := c.Request
	r.ParseMultipartForm(32 << 20)
	if vs := r.Form[key]; len(vs) > 0 {
		return vs[0]
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// PostForm returns the first value for the named component from POST and PUT body parameters.
// If key is not present, it returns the specified default value or an empty string.
func (c *Context) PostForm(key string, defaultValue ...string) string {
	r := c.Request
	r.ParseMultipartForm(32 << 20)
	if vs := r.PostForm[key]; len(vs) > 0 {
		return vs[0]
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// Next calls the rest of the handlers associated with the current route.
// If any of these handlers returns an error, Next will return the error and skip the following handlers.
// Next is normally used when a handler needs to do some postprocessing after the rest of the handlers
// are executed.
func (c *Context) Next() error {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

// Abort skips the rest of the handlers associated with the current route.
// Abort is normally used when a handler handles the request normally and wants to skip the rest of the handlers.
// If a handler wants to indicate an error condition, it should simply return the error without calling Abort.
func (c *Context) Abort() error {
	c.index = len(c.handlers)
	return nil
}

// Break 中断继续执行后续动作，返回指定状态及错误，不设置错误亦可.
func (c *Context) Break(status int, err ...error) error {
	var e error
	if len(err) > 0 {
		e = err[0]
	}
	c.Response.WriteHeader(status)
	c.HandleError(e)
	return c.Abort()
}

// URL creates a URL using the named route and the parameter values.
// The parameters should be given in the sequence of name1, value1, name2, value2, and so on.
// If a parameter in the route is not provided a value, the parameter token will remain in the resulting URL.
// Parameter values will be properly URL encoded.
// The method returns an empty string if the URL creation fails.
func (c *Context) URL(route string, pairs ...interface{}) string {
	if r := c.makross.namedRoutes[route]; r != nil {
		return r.URL(pairs...)
	}
	return ""
}

// Read populates the given struct variable with the data from the current request.
// If the request is NOT a GET request, it will check the "Content-Type" header
// and find a matching reader from DataReaders to read the request data.
// If there is no match or if the request is a GET request, it will use DefaultFormDataReader
// to read the request data.
func (c *Context) Read(data interface{}) error {
	if c.Request.Method != "GET" {
		t := getContentType(c.Request)
		if reader, ok := DataReaders[t]; ok {
			return reader.Read(c.Request, data)
		}
	}

	return DefaultFormDataReader.Read(c.Request, data)
}

// Write writes the given data of arbitrary type to the response.
// The method calls the data writer set via SetDataWriter() to do the actual writing.
// By default, the DefaultDataWriter will be used.
func (c *Context) Write(data interface{}) error {
	return c.writer.Write(c.Response, data)
}

func (c *Context) Redirect(url string, status ...int) error {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusFound
	}
	if code < StatusMultipleChoices || code > StatusPermanentRedirect {
		return ErrInvalidRedirectCode
	}

	c.Response.Header().Set(HeaderLocation, url)
	c.Response.WriteHeader(code)
	return nil
}

func (c *Context) Render(name string, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	if c.makross.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.makross.renderer.Render(buf, name, c); err != nil {
		return
	}
	c.Response.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	c.Response.WriteHeader(code)
	err = c.Write(buf.Bytes())
	c.Abort()
	return
}

func (c *Context) String(s string, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	c.Response.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	c.Response.WriteHeader(code)
	err = c.Write([]byte(s))
	c.Abort()
	return
}

func (c *Context) JSON(i interface{}, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONBlob(b, code)
}

func (c *Context) JSONPretty(i interface{}, indent string, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return c.JSONBlob(b, code)
}

func (c *Context) JSONBlob(b []byte, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	return c.Blob(MIMEApplicationJSONCharsetUTF8, b, code)
}

func (c *Context) JSONP(callback string, i interface{}, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(callback, b, code)
}

func (c *Context) JSONPBlob(callback string, b []byte, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	c.Response.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	c.Response.WriteHeader(code)
	if err = c.Write([]byte(callback + "(")); err != nil {
		return
	}
	if err = c.Write(b); err != nil {
		return
	}
	err = c.Write([]byte(");"))
	c.Abort()
	return
}

func (c *Context) XML(i interface{}, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	return c.XMLBlob(b, code)
}

func (c *Context) XMLPretty(i interface{}, indent string, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return c.XMLBlob(b, code)
}

func (c *Context) XMLBlob(b []byte, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	c.Response.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	c.Response.WriteHeader(code)
	if err = c.Write([]byte(xml.Header)); err != nil {
		return
	}
	err = c.Write(b)
	c.Abort()
	return
}

func (c *Context) Blob(contentType string, b []byte, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}

	c.Response.Header().Set(HeaderContentType, contentType)
	c.Response.WriteHeader(code)
	err = c.Write(b)
	c.Abort()
	return
}

func (c *Context) Stream(contentType string, r io.Reader, status ...int) (err error) {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	c.Response.Header().Set(HeaderContentType, contentType)
	c.Response.WriteHeader(code)
	_, err = io.Copy(c.Response, r)
	c.Abort()
	return
}

// ServeFile serves a view file, to send a file ( zip for example) to the client
// you should use the SendFile(serverfilename,clientfilename)
//
// You can define your own "Content-Type" header also, after this function call
// This function doesn't implement resuming (by range), use ctx.SendFile/fasthttp.ServeFileUncompressed(ctx.RequestCtx,path)/fasthttpServeFile(ctx.RequestCtx,path) instead
//
// Use it when you want to serve css/js/... files to the client, for bigger files and 'force-download' use the SendFile
func (c *Context) ServeFile(file string) (err error) {
	file, err = url.QueryUnescape(file) // Issue #839
	if err != nil {
		return
	}

	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()
	fi, _ := f.Stat()
	if fi.IsDir() {
		file = path.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		fi, _ = f.Stat()
	}
	http.ServeContent(c.Response, c.Request, fi.Name(), fi.ModTime(), f)
	return c.Abort() //c.ServeContent(f, fi.Name(), fi.ModTime())
}

// SendFile sends file for force-download to the client
//
// Use this instead of ServeFile to 'force-download' bigger files to the client
func (c *Context) SendFile(filename string, destinationName string) error {
	f, err := os.Open(filename)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	c.Response.Header().Set(HeaderContentDisposition, "attachment;filename="+destinationName)
	_, err = io.Copy(c.Response, f)
	c.Abort()
	return err
}

func (c *Context) Attachment(file, name string) (err error) {
	return c.contentDisposition(file, name, "attachment")
}

func (c *Context) Inline(file, name string) (err error) {
	return c.contentDisposition(file, name, "inline")
}

func (c *Context) contentDisposition(file, name, dispositionType string) (err error) {
	c.Response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", dispositionType, name))
	c.ServeFile(file)
	return
}

// NoContent Only header
func (c *Context) NoContent(status ...int) error {
	var code int
	if len(status) > 0 {
		code = status[0]
	} else {
		code = StatusOK
	}
	c.Response.WriteHeader(code)
	return nil
}

// SetDataWriter sets the data writer that will be used by Write().
func (c *Context) SetDataWriter(writer DataWriter) {
	c.writer = writer
}

func getContentType(req *http.Request) string {
	t := req.Header.Get("Content-Type")
	for i, c := range t {
		if c == ' ' || c == ';' {
			return t[:i]
		}
	}
	return t
}

// ContentTypeByExtension returns the MIME type associated with the file based on
// its extension. It returns `application/octet-stream` incase MIME type is not
// found.
func (c *Context) ContentTypeByExtension(name string) (t string) {
	ext := filepath.Ext(name)
	//these should be found by the windows(registry) and unix(apache) but on windows some machines have problems on this part.
	if t = mime.TypeByExtension(ext); t == "" {
		// no use of map here because we will have to lock/unlock it, by hand is better, no problem:
		if ext == ".json" {
			t = MIMEApplicationJSON
		} else if ext == ".zip" {
			t = "application/zip"
		} else if ext == ".3gp" {
			t = "video/3gpp"
		} else if ext == ".7z" {
			t = "application/x-7z-compressed"
		} else if ext == ".ace" {
			t = "application/x-ace-compressed"
		} else if ext == ".aac" {
			t = "audio/x-aac"
		} else if ext == ".ico" { // for any case
			t = "image/x-icon"
		} else if ext == ".png" {
			t = "image/png"
		} else {
			t = MIMEOctetStream
		}
	}
	return
}

// TimeFormat is the time format to use when generating times in HTTP
// headers. It is like time.RFC1123 but hard-codes GMT as the time
// zone. The time being formatted must be in UTC for Format to
// generate the correct format.
//
// For parsing this time format, see ParseTime.
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// RequestHeader returns the request header's value
// accepts one parameter, the key of the header (string)
// returns string
func (c *Context) RequestHeader(key string) string {
	return c.Request.Header.Get(key)
}

// ServeContent serves content, headers are autoset
// receives three parameters, it's low-level function, instead you can use .ServeFile(string,bool)/SendFile(string,string)
//
// You can define your own "Content-Type" header also, after this function call
// Doesn't implements resuming (by range), use ctx.SendFile instead
func (c *Context) ServeContent(content io.ReadSeeker, filename string, modtime time.Time) error {
	if t, err := time.Parse(TimeFormat, c.RequestHeader(HeaderIfModifiedSince)); err == nil && modtime.Before(t.Add(1*time.Second)) {
		c.Response.Header().Del(HeaderContentType)
		c.Response.Header().Del(HeaderContentLength)
		c.Response.WriteHeader(StatusNotModified)
		return nil
	}

	c.Response.Header().Set(HeaderContentType, c.ContentTypeByExtension(filename))
	c.Response.Header().Set(HeaderLastModified, modtime.UTC().Format(TimeFormat))

	size := func() int64 {
		size, err := content.Seek(0, io.SeekEnd)
		if err != nil {
			return 0
		}
		_, err = content.Seek(0, io.SeekStart)
		if err != nil {
			return 0
		}
		return size
	}()

	c.Response.Header().Set(HeaderContentLength, fmt.Sprintf("%v", size))
	c.Response.WriteHeader(StatusOK)
	_, err := io.Copy(c.Response, content)
	c.Abort()
	return err
}

// IsTLS implements `Context#TLS` function.
func (c *Context) IsTLS() bool {
	return c.Request.TLS != nil
}

func (c *Context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	return "http"
}
