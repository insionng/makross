// Package makross is a high productive and modular web framework in Golang.

package makross_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/access"
	"github.com/insionng/makross/content"
	"github.com/insionng/makross/fault"
	"github.com/insionng/makross/file"
	"github.com/insionng/makross/slash"
	"log"
	"net/http"
)

func Example() {
	m := makross.New()

	m.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
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
		return c.Write("update user " + c.Param("id").String())
	})

	// serve index file
	m.Get("/", file.Content("ui/index.html"))
	// serve files under the "ui" subdirectory
	m.Get("/*", file.Server(file.PathMap{
		"/": "/ui/",
	}))

	m.Listen(8888)

}
