package session

import (
	"encoding/gob"
	"errors"
	"log"
	"net/url"

	"github.com/insionng/makross"
)

var GlobalManager *Manager

var defaultOtions = Options{"memory", `{"cookieName":"makrossSessionId","gcLifetime":3600}`}

//var defaultOtions = Options{"file", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"./data/session"}`}

//var defaultOtions = Options{"redis", `{"cookieName":"makrossSessionId","gcLifetime":3600,"providerConfig":"127.0.0.1:6379"}`}

const (
	CONTEXT_SESSION_KEY = "_SESSION_STORE"
	COOKIE_FLASH_KEY    = "_COOKIE_FLASH"
	CONTEXT_FLASH_KEY   = "Flash"
	SESSION_FLASH_KEY   = "_SESSION_FLASH"
	SESSION_INPUT_KEY   = "_SESSION_INPUT"
)

// Store is the interface that contains all data for one session process with specific ID.
type Store interface {
	makross.RawStore
	// Read returns raw session store by session ID.
	Read(string) (makross.RawStore, error)
	// Destory deletes a session.
	Destory(*makross.Context) error
	// RegenerateId regenerates a session store from old session ID to new one.
	RegenerateId(*makross.Context) (makross.RawStore, error)
	// Count counts and returns number of sessions.
	Count() int
	// GC calls GC to clean expired sessions.
	GC()
}

type store struct {
	makross.RawStore
	*Manager
}

var _ Store = &store{}

type Options struct {
	Provider string
	Config   string
}

func init() {
	gob.Register(url.Values{})
}

// setup 初始化并设置session配置
func setup(op ...Options) error {
	option := defaultOtions
	if len(op) > 0 {
		option = op[0]
	}

	if len(option.Provider) == 0 {
		option.Provider = defaultOtions.Provider
		option.Config = defaultOtions.Config
	}

	log.Println("makross session config:", option)

	var err error
	GlobalManager, err = NewManager(option.Provider, option.Config)
	if err != nil {
		return err
	}
	go GlobalManager.GC()

	return nil
}

// Sessioner makross session 中间件
func Sessioner(op ...Options) makross.Handler {
	if GlobalManager == nil {
		if err := setup(op...); err != nil {
			log.Fatalln("Sessioner() setup() errors:", err)
		}
	}
	return func(c *makross.Context) error {
		if GlobalManager == nil {
			return errors.New("session manager not found, use session middleware but not init ?")
		}

		sess, err := GlobalManager.Start(c)
		if err != nil {
			return err
		}

		c.Session = store{
			RawStore: sess,
			Manager:  GlobalManager,
		}

		var has bool
		flashVals := url.Values{}
		flashIf := c.Session.Get(SESSION_FLASH_KEY)
		if flashIf != nil {
			//vals, _ := url.QueryUnescape(flashIf.(string))
			if flasho, okay := flashIf.(*makross.Flash); okay {
				if flashVals, _ = url.ParseQuery(flasho.Encode()); len(flashVals) > 0 {
					flash := makross.Flash{Values: url.Values{}}
					flash.ErrorMsg = flashVals.Get("error")
					flash.WarningMsg = flashVals.Get("warning")
					flash.InfoMsg = flashVals.Get("info")
					flash.SuccessMsg = flashVals.Get("success")

					flash.Ctx = c
					if flasho.FlashNow {
						flash.FlashNow = true
						c.Set(CONTEXT_FLASH_KEY, flash)
					} else {
						flash.FlashNow = false
						flash.Ctx.Set(CONTEXT_FLASH_KEY, flash)
					}
					c.Flash = &flash
					has = true

				}
			}

		}

		if !has {
			c.Flash = NewFlash(new(makross.Context))
			c.Set(CONTEXT_FLASH_KEY, c.Flash)
		}

		c.Set(CONTEXT_SESSION_KEY, c.Session)

		defer func() {
			//log.Println("save session", sess)
			//sess.Set(SESSION_FLASH_KEY, url.QueryEscape(f.Encode()))
			c.Session.Set(SESSION_FLASH_KEY, c.Flash)
			c.Session.Release(c)
		}()
		return c.Next()
	}
}

func GetStore(c *makross.Context) Store {
	store := c.Get(CONTEXT_SESSION_KEY)
	if store != nil {
		if s, okay := store.(Store); okay {
			return s
		}
	}
	return nil
}

func GetFlash(c *makross.Context) *makross.Flash {
	if store := GetStore(c); store != nil {
		if tmp := store.Get(SESSION_FLASH_KEY); tmp != nil {
			if flash, okay := tmp.(*makross.Flash); okay {
				return flash
			}
		}
	}
	return NewFlash(c)
}

func FlashValue(c *makross.Context) makross.Flash {
	if tmp := c.Get(CONTEXT_FLASH_KEY); tmp != nil {
		return tmp.(makross.Flash)
	}
	return makross.Flash{}
}

func NewFlash(ctx *makross.Context) *makross.Flash {
	return &makross.Flash{makross.FlashNow, ctx, url.Values{}, "", "", "", ""}
}
