// Copyright 2017 The Macross Authors
// Copyright 2017 The Makross Authors
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

package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Version(t *testing.T) {
	Convey("Check package version", t, func() {
		So(Version(), ShouldEqual, _VERSION)
	})
}

func Test_I18n(t *testing.T) {
	Convey("Use i18n middleware", t, func() {
		Convey("No langauge", func() {
			defer func() {
				So(recover(), ShouldNotBeNil)
			}()

			m := makross.New()
			m.Use(I18n(Options{}))
		})

		Convey("Languages and names not match", func() {
			defer func() {
				So(recover(), ShouldNotBeNil)
			}()

			m := makross.New()
			m.Use(I18n(Options{
				Langs: []string{"en-US"},
			}))
		})

		Convey("Invalid directory", func() {
			defer func() {
				So(recover(), ShouldNotBeNil)
			}()

			m := makross.New()
			m.Use(I18n(Options{
				Directory: "404",
				Langs:     []string{"en-US"},
				Names:     []string{"English"},
			}))
		})

		Convey("With correct options", func() {
			m := makross.New()
			m.Use(I18n(Options{
				Files: map[string][]byte{"locale_en-US.ini": []byte("")},
				Langs: []string{"en-US"},
				Names: []string{"English"},
			}))
			m.Get("/", func(*makross.Context) error { return nil })

			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)
			m.ServeHTTP(resp, req)

		})

		Convey("Set by redirect of URL parameter", func() {
			m := makross.New()
			m.Use(I18n(Options{
				Langs:    []string{"en-US"},
				Names:    []string{"English"},
				Redirect: true,
			}))
			m.Get("/", func(*makross.Context) error { return nil })

			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/?lang=en-US", nil)
			So(err, ShouldBeNil)
			req.RequestURI = "/?lang=en-US"
			m.ServeHTTP(resp, req)

		})

		Convey("Set by Accept-Language", func() {
			m := makross.New()
			m.Use(I18n(Options{
				Langs: []string{"en-US", "zh-CN", "it-IT"},
				Names: []string{"English", "简体中文", "Italiano"},
			}))
			m.Get("/", func(self *makross.Context) error {
				So(self.Localer.Language(), ShouldEqual, "it-IT")
				return nil
			})

			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			_, err = http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)
			req.Header.Set("Accept-Language", "it")
			m.ServeHTTP(resp, req)

		})

		Convey("Set to default language", func() {
			m := makross.New()
			m.Use(I18n(Options{
				Langs: []string{"en-US", "zh-CN", "it-IT"},
				Names: []string{"English", "简体中文", "Italiano"},
			}))
			m.Get("/", func(self *makross.Context) error {
				So(self.Localer.Language(), ShouldEqual, "en-US")
				return nil
			})

			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			So(err, ShouldBeNil)
			req.Header.Set("Accept-Language", "ru")
			m.ServeHTTP(resp, req)

		})
	})
}
