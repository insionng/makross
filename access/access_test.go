// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/insionng/makross"
	"github.com/stretchr/testify/assert"
)

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var customFunc = func(req *http.Request, rw *LogResponseWriter, elapsed float64) {
		var logWriter = getLogger(&buf)
		clientIP := GetClientIP(req)
		requestLine := fmt.Sprintf("%s %s %s", req.Method, req.URL.String(), req.Proto)
		logWriter(`[%s] [%.3fms] %s %d %d`, clientIP, elapsed, requestLine, rw.Status, rw.BytesWritten)
	}
	h := CustomLogger(customFunc)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	m := makross.New()
	c := m.NewContext(req, res, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Logger(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	m := makross.New()
	c := m.NewContext(req, res, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET http://127.0.0.1/users")
}

func TestGetClientIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("X-Real-IP", "192.168.100.1")
	req.Header.Set("X-Forwarded-For", "192.168.100.2")
	req.RemoteAddr = "192.168.100.3"

	assert.Equal(t, "192.168.100.1", GetClientIP(req))
	req.Header.Del("X-Real-IP")
	assert.Equal(t, "192.168.100.2", GetClientIP(req))
	req.Header.Del("X-Forwarded-For")
	assert.Equal(t, "192.168.100.3", GetClientIP(req))

	req.RemoteAddr = "192.168.100.3:8080"
	assert.Equal(t, "192.168.100.3", GetClientIP(req))
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *makross.Context) error {
	return errors.New("abc")
}
