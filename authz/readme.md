# Authz 

Authz is an authorization middleware for [Makross](https://github.com/insionng/makross), it's based on [https://github.com/casbin/casbin](https://github.com/casbin/casbin).

## Installation

```
go get github.com/insionng/makross/authz
```

## Simple Example

```Go
package main

import (
	"github.com/insionng/makross/authz"
	"github.com/insionng/makross"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	enf := authz.NewEnforcer("auth_model.conf", "auth_policy.csv")

	// define your makross, and use the Casbin authz middleware.
	// the access that is denied by authz will return HTTP 403 error.
    m := makross.New()
    m.Use(authz.Auth(enf))
}
```

## Documentation

The authorization determines a request based on ``{subject, object, action}``, which means what ``subject`` can perform what ``action`` on what ``object``. In this plumakross, the meanings are:

1. ``subject``: the logged-on user name
2. ``object``: the URL path for the web resource like "dataset1/item1"
3. ``action``: HTTP method like GET, POST, PUT, DELETE, or the high-level actions you defined like "read-file", "write-blog"


For how to write authorization policy and other details, please refer to [the Casbin's documentation](https://github.com/casbin/casbin).

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
