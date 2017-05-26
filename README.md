# Makross

Package makross is a high productive and modular web framework in Golang.

## Description

Makross is a Go package that provides high performance and powerful HTTP makross capabilities for Web applications.
Makross is very fast, thanks to the radix tree data structure and the usage of `sync.Pool` 

It has the following features:

* middleware pipeline architecture, similar to that of the [Express framework](http://expressjs.com).
* extremely fast request makross with zero dynamic memory allocation (the performance is comparable to that of [httprouter](https://github.com/julienschmidt/httprouter) and
[gin](https://github.com/gin-gonic/gin), see the [performance comparison below](#benchmarks))
* modular code organization through route grouping
* flexible URL path matching, supporting URL parameters and regular expressions
* URL creation according to the predefined routes
* compatible with `http.Handler` and `http.HandlerFunc`
* ready-to-use handlers sufficient for building RESTful APIs

If you are using [fasthttp](https://github.com/valyala/fasthttp), you may use a similar makross package [macross](https://github.com/insionng/macross) which is adapted from Makross.

## Requirements

Go 1.8 or above.

## Installation

Run the following command to install the package:

```
go get github.com/insionng/makross
```

## Getting Started

Below we describe how to create a simple REST API using Makross.

Create a `server.go` file with the following content:

```go
package main

import (
	"log"
	"net/http"
	"github.com/insionng/makross"
	"github.com/insionng/makross/access"
	"github.com/insionng/makross/slash"
	"github.com/insionng/makross/content"
	"github.com/insionng/makross/fault"
	"github.com/insionng/makross/file"
)

func main() {
	m := makross.New()

	m.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.RemoveTrailingSlash(),
		fault.Recovery(log.Printf),
	)

	// serve RESTful APIs
	api := m.Group("/api")
	api.Use(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON, content.XML),
	)
	api.Get("/users", func(c *makross.Context) error {
		return c.Write("user list")
	})
	api.Post("/users", func(c *makross.Context) error {
		return c.Write("create a new user")
	})
	api.Put(`/users/<id:\d+>`, func(c *makross.Context) error {
		return c.Write("update user " + c.Param("id"))
	})

	// serve index file
	m.Get("/", file.Content("ui/index.html"))
	// serve files under the "ui" subdirectory
	m.Get("/*", file.Server(file.PathMap{
		"/": "/ui/",
	}))

	m.Listen(8888)
}
```

Create an HTML file `ui/index.html` with any content.

Now run the following command to start the Web server:

```
go run server.go
```

You should be able to access URLs such as `http://localhost:8080`, `http://localhost:8080/api/users`.


### Routes

Makross works by building a makross table in a router and then dispatching HTTP requests to the matching handlers 
found in the makross table. An intuitive illustration of a makross table is as follows:


Routes              |  Handlers
--------------------|-----------------
`GET /users`        |  m1, m2, h1, ...
`POST /users`       |  m1, m2, h2, ...
`PUT /users/<id>`   |  m1, m2, h3, ...
`DELETE /users/<id>`|  m1, m2, h4, ...


For an incoming request `GET /users`, the first route would match and the handlers m1, m2, and h1 would be executed.
If the request is `PUT /users/123`, the third route would match and the corresponding handlers would be executed.
Note that the token `<id>` can match any number of non-slash characters and the matching part can be accessed as 
a path parameter value in the handlers.

**If an incoming request matches multiple routes in the table, the route added first to the table will take precedence.
All other matching routes will be ignored.**

The actual implementation of the makross table uses a variant of the radix tree data structure, which makes the makross
process as fast as working with a hash table, thanks to the inspiration from [httprouter](https://github.com/julienschmidt/httprouter).

To add a new route and its handlers to the makross table, call the `To` method like the following:
  
```go
m := makross.New()
m.To("GET", "/users", m1, m2, h1)
m.To("POST", "/users", m1, m2, h2)
```

You can also use shortcut methods, such as `Get`, `Post`, `Put`, etc., which are named after the HTTP method names:
 
```go
m.Get("/users", m1, m2, h1)
m.Post("/users", m1, m2, h2)
```

If you have multiple routes with the same URL path but different HTTP methods, like the above example, you can 
chain them together as follows,

```go
m.Get("/users", m1, m2, h1).Post(m1, m2, h2)
```

If you want to use the same set of handlers to handle the same URL path but different HTTP methods, you can take
the following shortcut:

```go
m.To("GET,POST", "/users", m1, m2, h)
```

A route may contain parameter tokens which are in the format of `<name:pattern>`, where `name` stands for the parameter
name, and `pattern` is a regular expression which the parameter value should match. A token `<name>` is equivalent
to `<name:[^/]*>`, i.e., it matches any number of non-slash characters. At the end of a route, an asterisk character
can be used to match any number of arbitrary characters. Below are some examples:

* `/users/<username>`: matches `/users/admin`
* `/users/accnt-<id:\d+>`: matches `/users/accnt-123`, but not `/users/accnt-admin`
* `/users/<username>/*`: matches `/users/admin/profile/address`

When a URL path matches a route, the matching parameters on the URL path can be accessed via `Context.Param()`:

```go
m := makross.New()

m.Get("/users/<username>", func (c *makross.Context) error {
	fmt.Fprintf(c.Response, "Name: %v", c.Param("username"))
	return nil
})
```


### Route Groups

Route group is a way of grouping together the routes which have the same route prefix. The routes in a group also
share the same handlers that are registered with the group via its `Use` method. For example,

```go
m := makross.New()
api := m.Group("/api")
api.Use(m1, m2)
api.Get("/users", h1).Post(h2)
api.Put("/users/<id>", h3).Delete(h4)
```

The above `/api` route group establishes the following makross table:


Routes                  |  Handlers
------------------------|-------------
`GET /api/users`        |  m1, m2, h1, ...
`POST /api/users`       |  m1, m2, h2, ...
`PUT /api/users/<id>`   |  m1, m2, h3, ...
`DELETE /api/users/<id>`|  m1, m2, h4, ...


As you can see, all these routes have the same route prefix `/api` and the handlers `m1` and `m2`. In other similar
makross frameworks, the handlers registered with a route group are also called *middlewares*.

Route groups can be nested. That is, a route group can create a child group by calling the `Group()` method. The router
serves as the top level route group. A child group inherits the handlers registered with its parent group. For example, 

```go
m := makross.New()
m.Use(m1)

api := m.Group("/api")
api.Use(m2)

users := api.Group("/users")
users.Use(m3)
users.Put("/<id>", h1)
```

Because the makross serves as the parent of the `api` group which is the parent of the `users` group, 
the `PUT /api/users/<id>` route is associated with the handlers `m1`, `m2`, `m3`, and `h1`.


### Router

Router manages the makross table and dispatches incoming requests to appropriate handlers. A router instance is created
by calling the `makross.New()` method.

Because `Router` implements the `http.Handler` interface, it can be readily used to serve subtrees on existing Go servers.
For example,

```go
m := makross.New()
m.Listen(9999)
```


### Handlers

A handler is a function with the signature `func(*makross.Context) error`. A handler is executed by the router if
the incoming request URL path matches the route that the handler is associated with. Through the `makross.Context` 
parameter, you can access the request information in handlers.

A route may be associated with multiple handlers. These handlers will be executed in the order that they are registered
to the route. The execution sequence can be terminated in the middle using one of the following two methods:

* A handler returns an error: the router will skip the rest of the handlers and handle the returned error.
* A handler calls `Context.Abort()`: the router will simply skip the rest of the handlers. There is no error to be handled.
 
A handler can call `Context.Next()` to explicitly execute the rest of the unexecuted handlers and take actions after
they finish execution. For example, a response compression handler may start the output buffer, call `Context.Next()`,
and then compress and send the output to response.


### Context

For each incoming request, a `makross.Context` object is populated with the request information and passed through
the handlers that need to handle the request. Handlers can get the request information via `Context.Request` and
send a response back via `Context.Response`. The `Context.Param()` method allows handlers to access the URL path
parameters that match the current route.

Using `Context.Get()` and `Context.Set()`, handlers can share data between each other. For example, an authentication
handler can store the authenticated user identity by calling `Context.Set()`, and other handlers can retrieve back
the identity information by calling `Context.Get()`.


### Reading Request Data

Context provides a few shortcut methods to read query parameters. The `Context.Query()`  method returns
the named URL query parameter value; the `Context.PostForm()` method returns the named parameter value in the POST or
PUT body parameters; and the `Context.Form()` method returns the value from either POST/PUT or URL query parameters.

The `Context.Read()` method supports reading data from the request body and populating it into an object.
The method will check the `Content-Type` HTTP header and parse the body data as the corresponding format.
For example, if `Content-Type` is `application/json`, the request body will be parsed as JSON data.
The public fields in the object being populated will receive the parsed data if the data contains the same named fields.
For example,

```go
func foo(c *makross.Context) error {
    data := &struct{
        A string
        B bool
    }{}

    // assume the body data is: {"A":"abc", "B":true}
    // data will be populated as: {A: "abc", B: true}
    if err := c.Read(&data); err != nil {
        return err
    }
}
```

By default, `Context` supports reading data that are in JSON, XML, form, and multipart-form data.
You may modify `makross.DataReaders` to add support for other data formats.

Note that when the data is read as form data, you may use struct tag named `form` to customize
the name of the corresponding field in the form data. The form data reader also supports populating
data into embedded objects which are either named or anonymous.

### Writing Response Data

The `Context.Write()` method can be used to write data of arbitrary type to the response.
By default, if the data being written is neither a string nor a byte array, the method will
will call `fmt.Fprint()` to write the data into the response.

You can call `Context.SetWriter()` to replace the default data writer with a customized one.
For example, the `content.TypeNegotiator` will negotiate the content response type and set the data
writer with an appropriate one.

### Error Handling

A handler may return an error indicating some erroneous condition. Sometimes, a handler or the code it calls may cause
a panic. Both should be handled properly to ensure best user experience. It is recommended that you use 
the `fault.Recover` handler or a similar error handler to handle these errors.

If an error is not handled by any handler, the router will handle it by calling its `handleError()` method which
simply sets an appropriate HTTP status code and writes the error message to the response.

When an incoming request has no matching route, the router will call the handlers registered via the `Router.NotFound()`
method. All the handlers registered via `Router.Use()` will also be called in advance. By default, the following two
handlers are registered with `Router.NotFound()`:

* `makross.MethodNotAllowedHandler`: a handler that sends an `Allow` HTTP header indicating the allowed HTTP methods for a requested URL
* `makross.NotFoundHandler`: a handler triggering 404 HTTP error

## Serving Static Files

Static files can be served with the help of `file.Server` and `file.Content` handlers. The former serves files
under the specified directories, while the latter serves the content of a single file. For example,

```go
import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/file"
)

m := makross.New()

// serve index file
m.Get("/", file.Content("ui/index.html"))
// serve files under the "ui" subdirectory
m.Get("/*", file.Server(file.PathMap{
	"/": "/ui/",
}))
```

## Handlers

Makross comes with a few commonly used handlers in its subpackages:

Handler name 					| Description
--------------------------------|--------------------------------------------
[access.Logger](https://godoc.org/github.com/insionng/makross/access) | records an entry for every incoming request
[auth.Basic](https://godoc.org/github.com/insionng/makross/auth) | provides authentication via HTTP Basic
[auth.Bearer](https://godoc.org/github.com/insionng/makross/auth) | provides authentication via HTTP Bearer
[auth.Query](https://godoc.org/github.com/insionng/makross/auth) | provides authentication via token-based query parameter
[auth.JWT](https://godoc.org/github.com/insionng/makross/auth) | provides JWT-based authentication
[content.TypeNegotiator](https://godoc.org/github.com/insionng/makross/content) | supports content negotiation by response types
[content.LanguageNegotiator](https://godoc.org/github.com/insionng/makross/content) | supports content negotiation by accepted languages
[cors.Handler](https://godoc.org/github.com/insionng/makross/cors) | implements the CORS (Cross Origin Resource Sharing) specification from the W3C
[fault.Recovery](https://godoc.org/github.com/insionng/makross/fault) | recovers from panics and handles errors returned by handlers
[fault.PanicHandler](https://godoc.org/github.com/insionng/makross/fault) | recovers from panics happened in the handlers
[fault.ErrorHandler](https://godoc.org/github.com/insionng/makross/fault) | handles errors returned by handlers by writing them in an appropriate format to the response
[file.Server](https://godoc.org/github.com/insionng/makross/file) | serves the files under the specified folder as response content
[file.Content](https://godoc.org/github.com/insionng/makross/file) | serves the content of the specified file as the response
[slash.Remover](https://godoc.org/github.com/insionng/makross/slash) | removes the trailing slashes from the request URL and redirects to the proper URL

The following code shows how these handlers may be used:

```go
import (
	"log"
	"net/http"
	"github.com/insionng/makross"
	"github.com/insionng/makross/access"
	"github.com/insionng/makross/slash"
	"github.com/insionng/makross/fault"
)

m := makross.New()

m.Use(
	access.Logger(log.Printf),
	slash.Remover(http.StatusMovedPermanently),
	fault.Recovery(log.Printf),
)

...
```

### Third-party Handlers


The following third-party handlers are specifically designed for Makross:

Handler name 					| Description
--------------------------------|--------------------------------------------
[jwt.JWT](https://github.com/vvv-v13/ozzo-jwt) | supports JWT Authorization


Makross also provides adapters to support using third-party `http.HandlerFunc` or `http.Handler` handlers. 
For example,

```go
m := makross.New()

// using http.HandlerFunc
m.Use(makross.HTTPHandlerFunc(http.NotFound))

// using http.Handler
m.Use(makross.HTTPHandler(http.NotFoundHandler))
```

### Contributes

Thanks to the macross, com, echo/vodka, iris, gin, beego, ozzo-routing, FastTemplate, Pongo2, Jwt-go. And all other Go package dependencies projects
