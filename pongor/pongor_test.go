package pongor_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/pongor"
	"testing"
)

func TestRender(t *testing.T) {
	e := makross.New()
	e.SetRenderer(pongor.Renderor())
	e.Get("/", func() makross.Handler {
		return func(ctx *makross.Context) error {
			ctx.Set("title", "你好，世界")
			// render ./templates/index file.
			return ctx.Render("index")
		}
	}())
}
