# cache

Middleware cache provides cache management for Macross. It can use many cache adapters, including memory, file, Redis.


## import
```go
    "github.com/macross-contrib/cache"
	_ "github.com/macross-contrib/cache/redis"
```

## Documentation
```
package cache

import (
	"testing"
)

func Test_TagCache(t *testing.T) {

	c, err := New(Options{Adapter: "memory"})
	if err != nil {
		t.Fatal(err)
	}

	// base use
	err = c.Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}

	res := ""
	c.Get("da", &res)

	if res != "weisd" {
		t.Fatal("base put faield")
	}

	t.Log("ok")

	// use tags/namespace
	err = c.Tags([]string{"dd"}).Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}
	res = ""
	c.Tags([]string{"dd"}).Get("da", &res)

	if res != "weisd" {
		t.Fatal("tags put faield")
	}

	t.Log("ok")

	err = c.Tags([]string{"macross"}).Set("macross", "macross_contrib", 300)
	if err != nil {
		t.Fatal(err)
	}

	res = ""
	c.Tags([]string{"macross"}).Get("macross", &res)

	if res != "macross_contrib" {
		t.Fatal("not macross_contrib")
	}

	t.Log("ok")

	// flush namespace
	err = c.Tags([]string{"macross"}).Flush()
	if err != nil {
		t.Fatal(err)
	}

	res = ""
	c.Tags([]string{"macross"}).Get("macross", &res)
	if res != "" {
		t.Fatal("flush faield")
	}

	res = ""
	c.Tags([]string{"macross"}).Get("bb", &res)
	if res != "" {
		t.Fatal("flush faield")
	}

	// still store in
	res = ""
	c.Tags([]string{"dd"}).Get("da", &res)
	if res != "weisd" {
		t.Fatal("where")
	}

	t.Log("ok")

}
```


## Macross Middleware
```go
package main

import (
	"fmt"

	"github.com/insionng/macross"
	"github.com/macross-contrib/cache"
	_ "github.com/macross-contrib/cache/redis"
)

func main() {

	v := macross.New()
	v.Use(cache.Cacher(cache.Options{Adapter: "redis", AdapterConfig: `{"Addr":"127.0.0.1:6379"}`, Section: "test", Interval: 5}))

	v.Get("/cache/put/", func(self *macross.Context) error {
		err := cache.Store(self).Set("name", "macross", 60)
		if err != nil {
			return err
		}

		return self.String("store okay")
	})

	v.Get("/cache/get/", func(self *macross.Context) error {
		var name string
		cache.Store(self).Get("name", &name)

		return self.String(fmt.Sprintf("get name %s", name))
	})

	v.Listen(":7777")
}

```
