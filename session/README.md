Session
==============

The session package is a makross session manager. It can use many session providers.

## How to install?

	go get github.com/insionng/makross/session


## What providers are supported?

As of now this session manager support memory, file and Redis .


## How to use it?

First you must import it

	import (
		"github.com/insionng/makross/session"
	)


* Use **memory** as provider:

        session.Options{"memory", `{"cookieName":"makrossSessionId","gcLifetime":3600}`}

* Use **file** as provider, the last param is the path where you want file to be stored:

	    session.Options{"file", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"./data/session"}`}

* Use **Redis** as provider, the last param is the Redis conn address,poolsize,password:

		session.Options{"redis", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"127.0.0.1:6379,100,makross"}`}

* Use **Cookie** as provider:

		session.Options{"cookie", `{"cookieName":"makrossSessionId","enableSetCookie":false,"gcLifetime":3600,"providerConfig":"{\"cookieName\":\"makrossSessionId\",\"securityKey\":\"makrosscookiehashkey\"}"}`}


Finally in the code you can use it like this

```go
package main

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/recover"
	"github.com/insionng/makross/session"
	//_ "github.com/insionng/makross/session/redis"
	"log"
)

func main() {

	v := makross.New()
	v.Use(recover.Recover())
	v.Use(session.Sessioner(session.Options{"file", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"./data/session"}`}))
	//v.Use(session.Sessioner(session.Options{"redis", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"127.0.0.1:6379"}`}))

	v.Get("/get", func(self *makross.Context) error {
		value := "nil"
		valueIf := self.Session.Get("key")
		if valueIf != nil {
			value = valueIf.(string)
		}

		return self.String(value)

	})

	v.Get("/set", func(self *makross.Context) error {

		val := self.QueryParam("v")
		if len(val) == 0 {
			val = "value"
		}

		err := self.Session.Set("key", val)
		if err != nil {
			log.Printf("sess.set %v \n", err)
		}
		return self.String("ok")
	})

	v.Listen(":9000")
}

```


## How to write own provider?

When you develop a web app, maybe you want to write own provider because you must meet the requirements.

Writing a provider is easy. You only need to define two struct types
(Session and Provider), which satisfy the interface definition.
Maybe you will find the **memory** provider is a good example.

	type SessionStore interface {
		Set(key, value interface{}) error     //set session value
		Get(key interface{}) interface{}      //get session value
		Delete(key interface{}) error         //delete session value
		ID() string                    //back current sessionID
		Release(ctx *makross.Context) error // release the resource & save data to provider & return the data
		Flush() error                         //delete all data
	}

	type Provider interface {
		Init(gcLifetime int64, config string) error
		Read(sid string) (makross.RawStore, error)
		Exist(sid string) bool
		Regenerate(oldsid, sid string) (makross.RawStore, error)
		Destroy(sid string) error
		Count() int //get all active session
		GC()
	}


## LICENSE

MIT License
