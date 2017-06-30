// Copyright 2013 Beego Authors
// Copyright 2014 The Macaron Authors
// Copyright 2016~2017 Insion Ng
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

// Package captcha a middleware that provides captcha service for Macross.
package captcha

import (
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/insionng/makross"
	"github.com/insionng/makross/cache"
	"github.com/insionng/makross/libraries/com"
)

const _VERSION = "0.1.0"

func Version() string {
	return _VERSION
}

var (
	defaultChars = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
)

// Captcha represents a captcha service.
type Captcha struct {
	store            cache.Cache
	SubURL           string
	URLPrefix        string
	FieldIDName      string
	FieldCaptchaName string
	StdWidth         int
	StdHeight        int
	ChallengeNums    int
	Expiration       int64
	CachePrefix      string
}

// generate key string
func (c *Captcha) key(id string) string {
	return c.CachePrefix + id
}

// generate rand chars with default chars
func (c *Captcha) genRandChars() string {
	return string(com.RandomCreateBytes(c.ChallengeNums, defaultChars...))
}

// CreateHTML tempalte func for output html
func (c *Captcha) CreateHTML() template.HTML {
	value, err := c.CreateCaptcha()
	if err != nil {
		panic(fmt.Errorf("fail to create captcha: %v", err))
	}
	return template.HTML(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">
	<a class="captcha" href="javascript:">
		<img onclick="this.src=('%s%s%s.png?reload='+(new Date()).getTime())" class="captcha-img" src="%s%s%s.png">
	</a>`, c.FieldIDName, value, c.SubURL, c.URLPrefix, value, c.SubURL, c.URLPrefix, value))
}

// CreateCaptcha create a new captcha id
func (c *Captcha) CreateCaptcha() (string, error) {
	id := string(com.RandomCreateBytes(15))
	if err := c.store.Set(c.key(id), c.genRandChars(), c.Expiration); err != nil {
		return "", err
	}
	return id, nil
}

// VerifyReq verify from a request
func (c *Captcha) VerifyReq(self *makross.Context) bool {
	return c.Verify(self.Args(c.FieldIDName).String(), self.Args(c.FieldCaptchaName).String())
}

// Verify direct verify id and challenge string
func (c *Captcha) Verify(id string, challenge string) bool {
	if len(challenge) == 0 || len(id) == 0 {
		return false
	}

	var chars string

	key := c.key(id)

	if c.store.Get(key, &chars); len(chars) == 0 {
		return false
	}

	defer c.store.Delete(key)

	if len(chars) != len(challenge) {
		return false
	}

	// verify challenge
	for i, c := range []byte(chars) {
		if c != challenge[i]-48 {
			return false
		}
	}

	return true
}

// Options a captcha's options
type Options struct {
	// Suburl path. Default is empty.
	SubURL string

	// URL prefix of getting captcha pictures. Default is "/captcha/".
	URLPrefix string

	// Hidden input element ID. Default is "captcha_id".
	FieldIDName string

	// User input value element name in request form. Default is "captcha".
	FieldCaptchaName string

	// Challenge number. Default is 6.
	ChallengeNums int

	// Captcha image width. Default is 240.
	Width int

	// Captcha image height. Default is 80.
	Height int

	// Captcha expiration time in seconds. Default is 600.
	Expiration int64

	// Cache key prefix captcha characters. Default is "captcha_".
	CachePrefix string
}

func prepareOptions(options ...Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

	opt.SubURL = strings.TrimSuffix(opt.SubURL, "/")

	// Defaults.
	if len(opt.URLPrefix) == 0 {
		opt.URLPrefix = "/captcha/"
	} else if opt.URLPrefix[len(opt.URLPrefix)-1] != '/' {
		opt.URLPrefix += "/"
	}
	if len(opt.FieldIDName) == 0 {
		opt.FieldIDName = "captcha_id"
	}
	if len(opt.FieldCaptchaName) == 0 {
		opt.FieldCaptchaName = "captcha"
	}
	if opt.ChallengeNums == 0 {
		opt.ChallengeNums = 6
	}
	if opt.Width == 0 {
		opt.Width = stdWidth
	}
	if opt.Height == 0 {
		opt.Height = stdHeight
	}
	if opt.Expiration == 0 {
		opt.Expiration = 600
	}
	if len(opt.CachePrefix) == 0 {
		opt.CachePrefix = "captcha_"
	}

	return opt
}

// NewCaptcha initializes and returns a captcha with given options.
func NewCaptcha(opts ...Options) *Captcha {
	opt := prepareOptions(opts...)
	return &Captcha{
		SubURL:           opt.SubURL,
		URLPrefix:        opt.URLPrefix,
		FieldIDName:      opt.FieldIDName,
		FieldCaptchaName: opt.FieldCaptchaName,
		StdWidth:         opt.Width,
		StdHeight:        opt.Height,
		ChallengeNums:    opt.ChallengeNums,
		Expiration:       opt.Expiration,
		CachePrefix:      opt.CachePrefix,
	}
}

// Captchaer is a middleware that maps a captcha.Captcha service into the Macross handler chain.
// An single variadic captcha.Options struct can be optionally provided to configure.
// This should be register after cache.Cacher.
func Captchaer(options ...Options) makross.Handler {
	return func(self *makross.Context) error {
		cpt := NewCaptcha(options...)
		cpt.store = cache.Store(self)
		if strings.HasPrefix(string(self.Request.URL.Path), cpt.URLPrefix) {
			var chars string
			id := path.Base(string(self.Request.URL.Path))
			if i := strings.Index(id, "."); i > -1 {
				id = id[:i]
			}
			key := cpt.key(id)

			// Reload captcha.
			if len(self.Query("reload")) > 0 {
				chars = cpt.genRandChars()
				if err := cpt.store.Set(key, chars, cpt.Expiration); err != nil {
					self.Response.WriteHeader(makross.StatusInternalServerError)
					self.Write([]byte("captcha reload error"))
					panic(fmt.Errorf("reload captcha: %v", err))
				}
			} else {
				if cpt.store.Get(key, &chars); len(chars) == 0 {
					self.Response.WriteHeader(makross.StatusNotFound)
					self.Write([]byte("captcha not found"))
					return self.Abort()
				}
			}

			self.Response.Header().Set(makross.HeaderContentType, "image/png")
			if _, err := NewImage([]byte(chars), cpt.StdWidth, cpt.StdHeight).WriteTo(self.Response); err != nil {
				panic(fmt.Errorf("fail to write captcha: %v", err))
			}
			return self.Abort()
		}

		self.Set("Captcha", cpt)
		return self.Next()
	}
}

func Store(self *makross.Context) (cpt *Captcha) {
	if cpta, okay := self.Get("Captcha").(*Captcha); okay {
		cpt = cpta
	}
	return
}
