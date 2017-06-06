package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
	//"log"

	"github.com/insionng/makross"
)

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	Init(gcLifetime int64, config string) error
	Read(sid string) (makross.RawStore, error)
	Exist(sid string) bool
	Regenerate(oldsid, sid string) (makross.RawStore, error)
	Destory(sid string) error
	Count() int //get all active session
	GC()
}

var provides = make(map[string]Provider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

type managerConfig struct {
	CookieName      string `json:"cookieName"`
	EnableSetCookie bool   `json:"enableSetCookie,omitempty"`
	GcLifetime      int64  `json:"gcLifetime"`
	MaxLifetime     int64  `json:"maxLifetime"`
	Secure          bool   `json:"secure"`
	CookieLifetime  int    `json:"cookieLifetime"`
	ProviderConfig  string `json:"providerConfig"`
	Domain          string `json:"domain"`
	SessionIDLength int64  `json:"sessionIDLength"`
}

// Manager contains Provider and its configuration.
type Manager struct {
	provider Provider
	config   *managerConfig
}

// NewManager Create new Manager with provider name and json config string.
// provider name:
// 1. cookie
// 2. file
// 3. memory
// 4. redis
// 5. mysql
// json config:
// 1. is https  default false
// 2. hashfunc  default sha1
// 3. hashkey default beegosessionkey
// 4. maxage default is none
func NewManager(provideName, config string) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	cf := new(managerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(config), cf)
	if err != nil {
		return nil, err
	}
	if cf.MaxLifetime == 0 {
		cf.MaxLifetime = cf.GcLifetime
	}
	err = provider.Init(cf.MaxLifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 16
	}

	return &Manager{
		provider,
		cf,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *Manager) getSid(ctx *makross.Context) (string, error) {
	//log.Println("get cookie name", manager.config.CookieName)
	cookie, err := ctx.GetCookie(manager.config.CookieName)

	if err != nil || cookie.Value == "" {
		//log.Println("read from query")
		sid := ctx.Form(manager.config.CookieName)
		return sid, nil
	}

	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// Start generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) Start(ctx *makross.Context) (session makross.RawStore, err error) {
	sid, errs := manager.getSid(ctx)
	if errs != nil {
		return nil, errs
	}

	//log.Println("start sid", sid)

	if sid != "" && manager.provider.Exist(sid) {
		//log.Println("sid exists")
		return manager.provider.Read(sid)
	}

	//log.Println("sid not exists")

	// Generate a new session
	sid, errs = manager.sessionID()
	if errs != nil {
		return nil, errs
	}

	session, err = manager.provider.Read(sid)
	cookie := ctx.NewCookie()
	cookie.Name = manager.config.CookieName
	cookie.Value = url.QueryEscape(sid)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = manager.isSecure(ctx)
	cookie.Domain = manager.config.Domain

	if manager.config.CookieLifetime > 0 {
		// cookie.MaxAge = manager.config.CookieLifetime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifetime))
	}
	if manager.config.EnableSetCookie {
		ctx.SetCookie(cookie)
	}

	// r.AddCookie(cookie)

	return
}

// Read returns raw session store by session ID.
func (manager *Manager) Read(sid string) (rawStore makross.RawStore, err error) {
	rawStore, err = manager.provider.Read(sid)
	return
}

// Count counts and returns number of sessions.
func (m *Manager) Count() int {
	return m.provider.Count()
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *Manager) GC() {
	manager.provider.GC()
	time.AfterFunc(time.Duration(manager.config.GcLifetime)*time.Second, func() { manager.GC() })
}

// RegenerateId Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *Manager) RegenerateId(ctx *makross.Context) (session makross.RawStore, err error) {
	sid, err := manager.sessionID()
	if err != nil {
		return
	}
	var c = ctx.NewCookie()
	cookie, err := ctx.GetCookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session, _ = manager.provider.Read(sid)

		c.Name = manager.config.CookieName
		c.Value = url.QueryEscape(sid)
		c.Path = "/"
		c.HttpOnly = true
		c.Secure = manager.isSecure(ctx)
		c.Domain = manager.config.Domain

	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.Regenerate(oldsid, sid)

		c.Name = cookie.Name
		c.Value = url.QueryEscape(sid)
		c.Path = "/"
		c.HttpOnly = true
		c.Secure = cookie.Secure
		c.Domain = cookie.Domain
	}
	if manager.config.CookieLifetime > 0 {
		// cookie.MaxAge = manager.config.CookieLifetime
		c.Expires = time.Now().Add(time.Duration(manager.config.CookieLifetime))
	}
	if manager.config.EnableSetCookie {
		ctx.SetCookie(c)
	}
	// r.AddCookie(c)
	return
}

// Destory deletes a session by given ID.
func (m *Manager) Destory(self *makross.Context) error {

	var sid string
	c, e := self.GetCookie(m.config.CookieName)
	if !(e != nil) {
		sid = c.Value
	}

	if len(sid) == 0 {
		return nil
	}

	if err := m.provider.Destory(sid); err != nil {
		return err
	}

	cookie := self.NewCookie()
	cookie.Name = m.config.CookieName
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Expires = time.Now()
	self.SetCookie(cookie)
	return nil
}

// SetSecure Set cookie with https.
func (manager *Manager) SetSecure(secure bool) {
	manager.config.Secure = secure
}

func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}

// Set cookie with https.
func (manager *Manager) isSecure(ctx *makross.Context) bool {
	if !manager.config.Secure {
		return false
	}
	if ctx.Scheme() != "" {
		return ctx.Scheme() == "https"
	}

	return false
	// if req.TLS == nil {
	// 	return false
	// }
	// return true
}
