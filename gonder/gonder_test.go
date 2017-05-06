package gonder_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/gonder"
	"testing"
)

func TestRender(t *testing.T) {
	e := makross.New()
	e.SetRenderer(gonder.Renderor())
	e.Get("/", func() makross.Handler {
		return func(self *makross.Context) error {
			self.Set("title", "你好，世界")
			// render ./templates/index file.
			return self.Render("index")
		}
	}())
}
