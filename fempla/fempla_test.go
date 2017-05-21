package fempla_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/fempla"
	"testing"
)

func TestRender(t *testing.T) {
	m := makross.New()
	m.SetRenderer(fempla.Renderor())
	m.Get("/", func() makross.Handler {
		return func(self *makross.Context) error {
			self.Set("title", "你好，世界")

			// render ./templates/index.html file.
			return self.Render("index")
		}
	}())
}
