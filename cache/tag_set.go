// Copyright 2015 rance
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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type TagSet struct {
	store CacheStore
	names []string
}

func NewTagSet(store CacheStore, names []string) *TagSet {

	t := &TagSet{store, names}
	t.SetNames(names)
	return t
}

func (this *TagSet) SetNames(names []string) {
	this.names = names
}

func (this *TagSet) AddNames(names []string) {
	names = append(this.names, names...)
	m := make(map[string]bool)
	for i, l := 0, len(names); i < l; i++ {
		name := names[i]
		m[name] = true
	}

	filterNames := make([]string, len(m))
	i := 0
	for k, _ := range m {
		filterNames[i] = k
	}

	this.names = filterNames
}

// 刷新所有 tag key
func (this *TagSet) Reset() error {
	for _, name := range this.names {
		this.ResetTag(name)
	}
	return nil
}

// 取tag id
func (this *TagSet) TagId(name string) string {
	var idstr string
	this.store.Get(this.TagKey(name), &idstr)

	if len(idstr) == 0 {
		return this.ResetTag(name)
	}

	return idstr
}

// 取所有的tagid
func (this *TagSet) TagIds() []string {
	l := len(this.names)
	if l == 0 {
		return []string{}
	}

	ids := make([]string, l)
	for i, name := range this.names {
		id := this.TagId(name)
		ids[i] = id
	}

	return ids
}

// 取命名空间
func (this *TagSet) GetNamespace() string {
	ids := this.TagIds()
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(this.TagIds(), "|")
}

// 重置key
func (this *TagSet) ResetTag(name string) string {
	id := this.generateId()
	err := this.store.Set(this.TagKey(name), id, 3600)
	if err != nil {
		panic(fmt.Errorf("ResetTag store Forever err : %v", err))
	}
	return id
}

// id标识算法
func (this *TagSet) generateId() string {
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(9))
}

// Tag key
func (this *TagSet) TagKey(name string) string {
	return fmt.Sprintf("tag:%s:key", name)
}

func EncodeSha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
