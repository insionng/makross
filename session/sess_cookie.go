package session

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/insionng/makross"
)

var cookiepder = &CookieProvider{}

// CookieSessionStore Cookie SessionStore
type CookieSessionStore struct {
	sid    string
	values map[interface{}]interface{} // session data
	lock   sync.RWMutex
}

// Set value to cookie session.
// the value are encoded as gob with hash block string.
func (st *CookieSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values[key] = value
	return nil
}

// Get value from cookie session
func (st *CookieSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	defer st.lock.RUnlock()
	if v, ok := st.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in cookie session
func (st *CookieSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.values, key)
	return nil
}

// Flush Clean all values in cookie session
func (st *CookieSessionStore) Flush() error {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.values = make(map[interface{}]interface{})
	return nil
}

// SessionID Return id of this cookie session
func (st *CookieSessionStore) ID() string {
	return st.sid
}

// SessionRelease Write cookie session to http response cookie
func (st *CookieSessionStore) Release(ctx *makross.Context) error {
	str, err := encodeCookie(cookiepder.block,
		cookiepder.config.SecurityKey,
		cookiepder.config.SecurityName,
		st.values)
	if err != nil {
		return err
	}

	cookie := ctx.NewCookie()
	cookie.Name = cookiepder.config.CookieName
	cookie.Value = url.QueryEscape(str)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = cookiepder.config.Secure
	cookie.Expires = time.Now().Add(time.Duration(cookiepder.config.MaxAge) * time.Second)
	ctx.SetCookie(cookie)
	return nil
}

type cookieConfig struct {
	SecurityKey  string `json:"securityKey"`
	BlockKey     string `json:"blockKey"`
	SecurityName string `json:"securityName"`
	CookieName   string `json:"cookieName"`
	Secure       bool   `json:"secure"`
	MaxAge       int    `json:"maxAge"`
}

// CookieProvider Cookie session provider
type CookieProvider struct {
	maxLifetime int64
	config      *cookieConfig
	block       cipher.Block
}

// Init Init cookie session provider with max lifetime and config json.
// maxLifetime is ignored.
// json config:
// 	securityKey - hash string
// 	blockKey - gob encode hash string. it's saved as aes crypto.
// 	securityName - recognized name in encoded cookie string
// 	cookieName - cookie name
// 	maxAge - cookie max life time.
func (pder *CookieProvider) Init(maxLifetime int64, config string) error {
	pder.config = &cookieConfig{}
	err := json.Unmarshal([]byte(config), pder.config)
	if err != nil {
		return err
	}
	if pder.config.BlockKey == "" {
		pder.config.BlockKey = string(generateRandomKey(16))
	}
	if pder.config.SecurityName == "" {
		pder.config.SecurityName = string(generateRandomKey(20))
	}
	pder.block, err = aes.NewCipher([]byte(pder.config.BlockKey))
	if err != nil {
		return err
	}
	pder.maxLifetime = maxLifetime
	return nil
}

// Read Get SessionStore in cooke.
// decode cooke string to map and put into SessionStore with sid.
func (pder *CookieProvider) Read(sid string) (makross.RawStore, error) {
	maps, _ := decodeCookie(pder.block,
		pder.config.SecurityKey,
		pder.config.SecurityName,
		sid, pder.maxLifetime)
	if maps == nil {
		maps = make(map[interface{}]interface{})
	}
	rs := &CookieSessionStore{sid: sid, values: maps}
	return rs, nil
}

// Exist Cookie session is always existed
func (pder *CookieProvider) Exist(sid string) bool {
	return true
}

// Regenerate Implement method, no used.
func (pder *CookieProvider) Regenerate(oldsid, sid string) (makross.RawStore, error) {
	return nil, nil
}

// Destory Implement method, no used.
func (pder *CookieProvider) Destory(sid string) error {
	return nil
}

// GC Implement method, no used.
func (pder *CookieProvider) GC() {
	return
}

// SessionCount Implement method, return 0.
func (pder *CookieProvider) Count() int {
	return 0
}

// SessionUpdate Implement method, no used.
func (pder *CookieProvider) SessionUpdate(sid string) error {
	return nil
}

func init() {
	Register("cookie", cookiepder)
}
