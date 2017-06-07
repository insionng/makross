// Copyright 2013 Beego Authors
// Copyright 2014 The Macaron Authors
// Copyright 2016 The Vodka Authors
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

package captcha

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Siphash(t *testing.T) {
	Convey("Siphash", t, func() {
		good := uint64(0xe849e8bb6ffe2567)
		cur := siphash(0, 0, 0)
		So(cur, ShouldEqual, good)
	})
}

func BenchmarkSiprng(b *testing.B) {
	b.SetBytes(8)
	p := &siprng{}
	for i := 0; i < b.N; i++ {
		p.Uint64()
	}
}
