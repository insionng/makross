package main

import (
	"github.com/insionng/makross"
	//"github.com/insionng/makross/access"
	//"github.com/insionng/makross/file"
	"github.com/insionng/makross/gonder"
	//"github.com/insionng/makross/logger"
	//"github.com/insionng/makross/recover"
	"github.com/insionng/makross/static"
	//"log"
)

func main() {
	v := makross.New()

	//v.Use(recover.Recover())
	//v.Use(logger.Logger())
	//v.Use(access.Logger(log.Printf))
	v.Use(static.Static("static"))
	//v.Use(file.Server(file.PathMap{"/static": "/static"}))

	v.SetRenderer(gonder.Renderor(gonder.Option{
		DelimLeft:  "{{",
		DelimRight: "}}",
	}))

	v.Get("/", func(self *makross.Context) error {
		var data = make(map[string]interface{})
		data["name"] = "Insion Ng"
		self.SetStore(data)

		self.SetStore(map[string]interface{}{
			"title": "你好，世界",
			"oh":    "no",
		})
		self.Set("oh", "yes") //覆盖前面指定KEY
		return self.Render("index")
	})

	v.Listen(":9000")
}
