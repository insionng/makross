# jwt for makross

The jwt middleware for Macross Web Framework

## Requirements

makross


## Getting Started

Create a `server.go` file with the following content:

```go
package main

import (
	"makross"
	"makross/jwt"
	"net/http"
)

func main() {
	m := makross.New()


	m.Get("/", func(self *makross.Context) error {
		var data = map[string]interface{}{}
		data["version"] = "1.0.0"
		return self.JSON(http.StatusOK, data)
	})

	// 给用户返回token之前请先密码验证用户身份
	m.Post("/signin/", func(self *makross.Context) error {
		username := string(self.FormValue("username"))
		password := string(self.FormValue("password"))
		if (username == "insion") && (password == "PaSsworD") {
			claims := jwt.NewMapClaims()
			claims["address"] = "GD.GZ"
			tk, _ := jwt.NewToken("secret", "SigningMethodHS256", claims)

			return self.WriteData(tk)
		}
		return makross.ErrUnauthorized

	})

	g := m.Group("/jwt", jwt.JWT("secret"))
	g.Get("/say/", func(self *makross.Context) error {
		return self.WriteData("Hello, Macross")
	})

	m.Listen(":9000")
}

```

Now run the following command to start the Web server:

```
go run server.go
```

You should be able to access URLs such as `http://localhost:9000`.


