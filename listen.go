// Package makross is a high productive and modular web framework in Golang.

package makross

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

func (m *Makross) Listen(args ...interface{}) {
	addr := GetAddress(args...)
	if runtime.NumCPU() > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	}
	m.Server.Addr = addr
	m.Server.ListenAndServe()
}

func (m *Makross) ListenTLS(certFile, keyFile string, args ...interface{}) {
	addr := GetAddress(args...)
	if runtime.NumCPU() > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	}
	m.Server.Addr = addr
	m.Server.ListenAndServeTLS(certFile, keyFile)
}

func GetAddress(args ...interface{}) string {

	var host string
	var port int

	if len(args) == 1 {
		switch arg := args[0].(type) {
		case string:
			addrs := strings.Split(args[0].(string), ":")
			if len(addrs) == 1 {
				host = addrs[0]
			} else if len(addrs) >= 2 {
				host = addrs[0]
				_port, _ := strconv.ParseInt(addrs[1], 10, 0)
				port = int(_port)
			}
		case int:
			port = arg
		case int64:
			port = int(arg)
		}
	} else if len(args) >= 2 {
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = arg
		}
	}

	if iHost := os.Getenv("HOST"); len(iHost) != 0 {
		host = iHost
	} else if len(host) == 0 {
		host = "0.0.0.0"
	}

	if iPort, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32); iPort != 0 {
		port = int(iPort)
	} else if port == 0 {
		port = 8000
	}

	addr := host + ":" + strconv.FormatInt(int64(port), 10)
	return addr

}
