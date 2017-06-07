package main

import (
	"fmt"

	"github.com/insionng/makross"
	"github.com/insionng/makross/cache"
	"github.com/insionng/makross/captcha"
	"github.com/insionng/makross/logger"
	"github.com/insionng/makross/pongor"
	"github.com/insionng/makross/recover"
)

func main() {
	v := makross.New()
	v.Use(logger.Logger())
	v.Use(recover.Recover())
	v.Use(cache.Cacher(cache.Options{Adapter: "memory"}))
	v.Use(captcha.Captchaer())
	v.SetRenderer(pongor.Renderor())

	v.Get("/", func(self *makross.Context) error {
		if cpt := self.Get("Captcha"); cpt != nil {
			fmt.Println("Got:", cpt)
		} else {
			fmt.Println("Captcha is nil!")
		}

		self.Set("title", "你好，世界")
		return self.Render("index")
	})

	v.Listen(":7891")
}
