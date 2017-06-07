package redis

import (
	"testing"

	"github.com/insionng/makross/cache"
)

func TestRedisCache(t *testing.T) {
	var err error
	c, err := cache.New(cache.Options{Adapter: "redis", AdapterConfig: `{"Addr":"127.0.0.1:6379"}`, Section: "test"})
	if err != nil {
		t.Fatal(err)
	}

	err = c.Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}

	res := ""
	err = c.Get("da", &res)
	if err != nil {
		t.Fatal(err)
	}

	if res != "weisd" {
		t.Fatal(res)
	}

	t.Log("ok")
	t.Log("test", res)

	err = c.Tags([]string{"dd"}).Set("da", "weisd", 300)
	if err != nil {
		t.Fatal(err)
	}
	res = ""
	err = c.Tags([]string{"dd"}).Get("da", &res)
	if err != nil {
		t.Fatal(err)
	}

	if res != "weisd" {
		t.Fatal("not weisd")
	}

	t.Log("ok")
	t.Log("dd", res)

	err = c.Tags([]string{"makross"}).Set("makross", "makross_contrib", 300)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Tags([]string{"makross"}).Set("insion", "insionng", 300)
	if err != nil {
		t.Fatal(err)
	}

	res = ""
	err = c.Tags([]string{"makross"}).Get("makross", &res)
	if err != nil {
		t.Fatal(err)
	}

	if res != "makross_contrib" {
		t.Fatal("not makross_contrib")
	}

	t.Log("ok")
	t.Log("makross", res)

	err = c.Tags([]string{"makross"}).Flush()
	if err != nil {
		t.Fatal(err)
	}

	res = ""
	c.Tags([]string{"makross"}).Get("makross", &res)
	if res != "" {
		t.Fatal("flush faield")
	}

	res = ""
	c.Tags([]string{"makross"}).Get("insion", &res)
	if res != "" {
		t.Fatal("flush faield")
	}

	res = ""
	err = c.Tags([]string{"dd"}).Get("da", &res)
	if err != nil {
		t.Fatal(err)
	}

	if res != "weisd" {
		t.Fatal("not weisd")
	}

	t.Log("ok")

	err = c.Flush()
	if err != nil {
		t.Fatal(err)
	}

	res = ""
	c.Get("da", &res)
	if res != "" {
		t.Fatal("flush failed")
	}

	t.Log("get dd da", res)

}
