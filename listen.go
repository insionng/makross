// Package makross is a high productive and modular web framework in Golang.

package makross

import (
	"net/http"
	"runtime"
)

func (m *Makross) Listen(args ...interface{}) {
	addr := GetAddress(args...)
	if runtime.NumCPU() > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	}
	http.ListenAndServe(addr, m)
}

func (m *Makross) ListenTLS(certFile, keyFile string, args ...interface{}) {
	addr := GetAddress(args...)
	if runtime.NumCPU() > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	}
	http.ListenAndServeTLS(addr, certFile, keyFile, m)
}
