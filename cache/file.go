// Copyright 2013 Beego Authors
// Copyright 2014 Unknwon
// Copyright 2015  iseejun
// Copyright 2016~2017  Insion Ng
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package cache

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

// Item represents a cache item.
type Item struct {
	Val     interface{}
	Created int64
	Expire  int64
}

func (item *Item) hasExpired() bool {
	return item.Expire > 0 &&
		(time.Now().Unix()-item.Created) >= item.Expire
}

// FileCacher represents a file cache adapter implementation.
type FileCacher struct {
	rootPath string
	interval int // GC interval.
}

// NewFileCacher creates and returns a new file cacher.
func NewFileCacher() *FileCacher {
	return &FileCacher{}
}

func (c *FileCacher) filepath(key string) string {
	m := md5.Sum([]byte(key))
	hash := hex.EncodeToString(m[:])
	return filepath.Join(c.rootPath, string(hash[0]), string(hash[1]), hash)
}

// Set puts value into cache with key and expire time.
// If expired is 0, it will be deleted by next GC operation.
func (c *FileCacher) Set(key string, val interface{}, expire int64) error {
	filename := c.filepath(key)
	item := &Item{val, time.Now().Unix(), expire}
	data, err := EncodeGob(item)
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	return ioutil.WriteFile(filename, data, os.ModePerm)
}

func (c *FileCacher) read(key string) (*Item, error) {
	filename := c.filepath(key)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	item := new(Item)
	return item, DecodeGob(data, item)
}

// Get gets cached value by given key.
func (c *FileCacher) Get(key string, _val interface{}) error {
	item, err := c.read(key)
	if err != nil {
		return err
	}

	if item.hasExpired() {
		return os.Remove(c.filepath(key))
	}
	b, _ := item.Val.([]byte)
	return msgpack.Unmarshal(b, _val)

}

// Delete deletes cached value by given key.
func (c *FileCacher) Delete(key string) error {
	return os.Remove(c.filepath(key))
}

// // Incr increases cached int-type value by given key as a counter.
// func (c *FileCacher) Incr(key string) error {
// 	item, err := c.read(key)
// 	if err != nil {
// 		return err
// 	}
//
// 	item.Val, err = Incr(item.Val)
// 	if err != nil {
// 		return err
// 	}
//
// 	return c.Put(key, item.Val, item.Expire)
// }

// Decrease cached int value.
// func (c *FileCacher) Decr(key string) error {
// 	item, err := c.read(key)
// 	if err != nil {
// 		return err
// 	}
//
// 	item.Val, err = Decr(item.Val)
// 	if err != nil {
// 		return err
// 	}
//
// 	return c.Put(key, item.Val, item.Expire)
// }

// IsExist returns true if cached value exists.
func (c *FileCacher) IsExist(key string) bool {
	return IsExist(c.filepath(key))
}

// Flush deletes all cached data.
func (c *FileCacher) Flush() error {
	return os.RemoveAll(c.rootPath)
}

func (c *FileCacher) startGC() {
	if c.interval < 1 {
		return
	}

	if err := filepath.Walk(c.rootPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Walk: %v", err)
		}

		if fi.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("ReadFile: %v", err)
		}

		item := new(Item)
		if err = DecodeGob(data, item); err != nil {
			return err
		}
		if item.hasExpired() {
			if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("Remove: %v", err)
			}
		}
		return nil
	}); err != nil {
		log.Printf("error garbage collecting cache files: %v", err)
	}

	time.AfterFunc(time.Duration(c.interval)*time.Second, func() { c.startGC() })
}

// StartAndGC starts GC routine based on config string settings.
func (c *FileCacher) StartAndGC(opt Options) error {
	c.rootPath = opt.AdapterConfig
	c.interval = opt.Interval

	if err := os.MkdirAll(c.rootPath, os.ModePerm); err != nil {
		return err
	}

	go c.startGC()
	return nil
}

// Incr increases cached int-type value by given key as a counter.
func (c *FileCacher) Incr(key string) (int64, error) {

	item, err := c.read(key)

	if err != nil {
		return 0, err
	}
	i, okay := item.Val.(int64)
	//i, errParse := strconv.ParseInt(item.Val, 10, 32)
	if !okay {
		return 0, errors.New("item value is not int64 type")
	}
	item.Val = strconv.FormatInt(i+1, 10)
	c.Set(key, item.Val, item.Expire)
	return i + 1, nil
}

// Decr decreases cached int-type value by given key as a counter.
func (c *FileCacher) Decr(key string) (int64, error) {
	item, err := c.read(key)

	if err != nil {
		return 0, err
	}
	i, okay := item.Val.(int64)
	//i, errParse := strconv.ParseInt(item.Val, 10, 32)
	if !okay {
		return 0, errors.New("item value is not int64 type")
	}
	item.Val = strconv.FormatInt(i-1, 10)
	c.Set(key, item.Val, item.Expire)
	return i - 1, nil
}

// update expire time
func (c *FileCacher) Touch(key string, expire int64) error {
	item, err := c.read(key)

	if err != nil {
		return err
	}

	c.Set(key, item.Val, item.Expire)

	return nil

}
func init() {
	Register("file", NewFileCacher())
}
