package static_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/static"
	"testing"
)

func TestStatic(t *testing.T) {
	m := makross.New()
	m.Use(static.Static("public"))
	go m.Listen(":8888")

	n := makross.New()
	n.Use(static.StaticWithConfig(static.StaticConfig{
		Root:   "public",
		Browse: true,
	}))
	go n.Listen(9999)
}
