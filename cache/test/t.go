package main

import (
	"fmt"

	"github.com/insionng/makross"
	"github.com/insionng/makross/cache"
	_ "github.com/insionng/makross/cache/redis"
)

func main() {

	v := makross.New()
	v.Use(cache.Cacher(cache.Options{Adapter: "redis", AdapterConfig: `{"Addr":"127.0.0.1:6379"}`, Section: "test", Interval: 5}))

	v.Get("/cache/put/", func(self *makross.Context) error {
		err := cache.Store(self).Set("name", "makross", 60)
		if err != nil {
			return err
		}

		return self.String("store okay")
	})

	v.Get("/cache/get/", func(self *makross.Context) error {
		var name string
		cache.Store(self).Get("name", &name)

		return self.String(fmt.Sprintf("get name %s", name))
	})

	v.Listen(":7891")
}
