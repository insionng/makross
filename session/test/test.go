package main

import (
	"log"

	"github.com/insionng/makross"
	"github.com/insionng/makross/recover"
	"github.com/insionng/makross/session"
	_ "github.com/insionng/makross/session/redis"
)

func main() {

	v := makross.New()
	v.Use(recover.Recover())
	//v.Use(session.Sessioner(session.Options{"file", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"./data/session"}`}))
	v.Use(session.Sessioner(session.Options{"redis", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"127.0.0.1:6379"}`}))

	v.Get("/get", func(self *makross.Context) error {
		value := "nil"
		valueIf := self.Session.Get("key")
		if valueIf != nil {
			value = valueIf.(string)
		}

		return self.String(value)

	})

	v.Get("/set", func(self *makross.Context) error {

		val := self.Query("v")
		if len(val) == 0 {
			val = "value"
		}

		err := self.Session.Set("key", val)
		if err != nil {
			log.Printf("sess.set %v \n", err)
		}
		return self.String("ok")
	})

	v.Listen(":8080")
}
