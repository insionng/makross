package redis

import (
	"strconv"
	"strings"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/insionng/makross"
	"github.com/insionng/makross/session"
)

var redispder = &Provider{}

// MaxPoolSize redis max pool size
var MaxPoolSize = 100

// SessionStore redis session store
type SessionStore struct {
	p           *redis.Pool
	sid         string
	lock        sync.RWMutex
	values      map[interface{}]interface{}
	maxLifetime int64
}

// Set value in redis session
func (rs *SessionStore) Set(key, value interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value

	return nil
}

// Get value in redis session
func (rs *SessionStore) Get(key interface{}) interface{} {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in redis session
func (rs *SessionStore) Delete(key interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

// Flush clear all values in redis session
func (rs *SessionStore) Flush() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[interface{}]interface{})
	return nil
}

// SessionID get redis session id
func (rs *SessionStore) ID() string {
	return rs.sid
}

// SessionRelease save session values to redis
func (rs *SessionStore) Release(ctx *makross.Context) (err error) {
	var b []byte
	b, err = session.EncodeGob(rs.values)
	if err != nil {
		return
	}
	c := rs.p.Get()
	defer c.Close()

	c.Do("SETEX", rs.sid, rs.maxLifetime, string(b))
	return
}

// Provider redis session provider
type Provider struct {
	maxLifetime int64
	savePath    string
	poolsize    int
	password    string
	dbNum       int
	poollist    *redis.Pool
}

// Init init redis session
// savepath like redis server addr,pool size,password,dbnum
// e.g. 127.0.0.1:6379,100,astaxie,0
func (rp *Provider) Init(maxLifetime int64, savePath string) error {
	rp.maxLifetime = maxLifetime
	configs := strings.Split(savePath, ",")
	if len(configs) > 0 {
		rp.savePath = configs[0]
	}
	if len(configs) > 1 {
		poolsize, err := strconv.Atoi(configs[1])
		if err != nil || poolsize <= 0 {
			rp.poolsize = MaxPoolSize
		} else {
			rp.poolsize = poolsize
		}
	} else {
		rp.poolsize = MaxPoolSize
	}
	if len(configs) > 2 {
		rp.password = configs[2]
	}
	if len(configs) > 3 {
		dbnum, err := strconv.Atoi(configs[3])
		if err != nil || dbnum < 0 {
			rp.dbNum = 0
		} else {
			rp.dbNum = dbnum
		}
	} else {
		rp.dbNum = 0
	}
	rp.poollist = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", rp.savePath)
		if err != nil {
			return nil, err
		}
		if rp.password != "" {
			if _, err := c.Do("AUTH", rp.password); err != nil {
				c.Close()
				return nil, err
			}
		}
		_, err = c.Do("SELECT", rp.dbNum)
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, err
	}, rp.poolsize)

	return rp.poollist.Get().Err()
}

// Read read redis session by sid
func (rp *Provider) Read(sid string) (makross.RawStore, error) {
	c := rp.poollist.Get()
	defer c.Close()

	kvs, err := redis.String(c.Do("GET", sid))
	var kv map[interface{}]interface{}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = session.DecodeGob([]byte(kvs))
		if err != nil {
			return nil, err
		}
	}

	rs := &SessionStore{p: rp.poollist, sid: sid, values: kv, maxLifetime: rp.maxLifetime}
	return rs, nil
}

// Exist check redis session exist by sid
func (rp *Provider) Exist(sid string) bool {
	c := rp.poollist.Get()
	defer c.Close()

	if existed, err := redis.Int(c.Do("EXISTS", sid)); err != nil || existed == 0 {
		return false
	}
	return true
}

// Regenerate generate new sid for redis session
func (rp *Provider) Regenerate(oldsid, sid string) (makross.RawStore, error) {
	c := rp.poollist.Get()
	defer c.Close()

	if existed, _ := redis.Int(c.Do("EXISTS", oldsid)); existed == 0 {
		// oldsid doesn't exists, set the new sid directly
		// ignore error here, since if it return error
		// the existed value will be 0
		c.Do("SET", sid, "", "EX", rp.maxLifetime)
	} else {
		c.Do("RENAME", oldsid, sid)
		c.Do("EXPIRE", sid, rp.maxLifetime)
	}

	kvs, err := redis.String(c.Do("GET", sid))
	var kv map[interface{}]interface{}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = session.DecodeGob([]byte(kvs))
		if err != nil {
			return nil, err
		}
	}

	rs := &SessionStore{p: rp.poollist, sid: sid, values: kv, maxLifetime: rp.maxLifetime}
	return rs, nil
}

// Destory delete redis session by id
func (rp *Provider) Destory(sid string) error {
	c := rp.poollist.Get()
	defer c.Close()

	c.Do("DEL", sid)
	return nil
}

// GC Impelment method, no used.
func (rp *Provider) GC() {
	return
}

// SessionAll return all activeSession
func (rp *Provider) Count() int {
	return 0
}

func init() {
	session.Register("redis", redispder)
}
