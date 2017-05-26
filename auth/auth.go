// Package makross is a high productive and modular web framework in Golang.

// Package auth provides a set of user authentication handlers for the makross.
package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/insionng/makross"
)

// User is the key used to store and retrieve the user identity information in makross.Context
const User = "User"

// Identity represents an authenticated user. If a user is successfully authenticated by
// an auth handler (Basic, Bearer, or Query), an Identity object will be made available for injection.
type Identity interface{}

// DefaultRealm is the default realm name for HTTP authentication. It is used by HTTP authentication based on
// Basic and Bearer.
var DefaultRealm = "API"

// BasicAuthFunc is the function that does the actual user authentication according to the given username and password.
type BasicAuthFunc func(c *makross.Context, username, password string) (Identity, error)

// Basic returns a makross.Handler that performs HTTP basic authentication.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/insionng/makross"
//     "github.com/insionng/makross/auth"
//   )
//   func main() {
//     r := makross.New()
//     r.Use(auth.Basic(func(c *makross.Context, username, password string) (auth.Identity, error) {
//       if username == "demo" && password == "foo" {
//         return auth.Identity(username), nil
//       }
//       return nil, errors.New("invalid credential")
//     }))
//     r.Get("/demo", func(c *makross.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// By default, the auth realm is named as "API". You may customize it by specifying the realm parameter.
//
// When authentication fails, a "WWW-Authenticate" header will be sent, and an http.StatusUnauthorized
// error will be returned.
func Basic(fn BasicAuthFunc, realm ...string) makross.Handler {
	name := DefaultRealm
	if len(realm) > 0 {
		name = realm[0]
	}
	return func(c *makross.Context) error {
		username, password := parseBasicAuth(c.Request.Header.Get("Authorization"))
		identity, e := fn(c, username, password)
		if e == nil {
			c.Set(User, identity)
			return nil
		}
		c.Response.Header().Set("WWW-Authenticate", `Basic realm="`+name+`"`)
		return makross.NewHTTPError(http.StatusUnauthorized, e.Error())
	}
}

func parseBasicAuth(auth string) (username, password string) {
	if strings.HasPrefix(auth, "Basic ") {
		if bytes, err := base64.StdEncoding.DecodeString(auth[6:]); err == nil {
			str := string(bytes)
			if i := strings.IndexByte(str, ':'); i >= 0 {
				return str[:i], str[i+1:]
			}
		}
	}
	return
}

// TokenAuthFunc is the function for authenticating a user based on a secret token.
type TokenAuthFunc func(c *makross.Context, token string) (Identity, error)

// Bearer returns a makross.Handler that performs HTTP authentication based on bearer token.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/insionng/makross"
//     "github.com/insionng/makross/auth"
//   )
//   func main() {
//     r := makross.New()
//     r.Use(auth.Bearer(func(c *makross.Context, token string) (auth.Identity, error) {
//       if token == "secret" {
//         return auth.Identity("demo"), nil
//       }
//       return nil, errors.New("invalid credential")
//     }))
//     r.Get("/demo", func(c *makross.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// By default, the auth realm is named as "API". You may customize it by specifying the realm parameter.
//
// When authentication fails, a "WWW-Authenticate" header will be sent, and an http.StatusUnauthorized
// error will be returned.
func Bearer(fn TokenAuthFunc, realm ...string) makross.Handler {
	name := DefaultRealm
	if len(realm) > 0 {
		name = realm[0]
	}
	return func(c *makross.Context) error {
		token := parseBearerAuth(c.Request.Header.Get("Authorization"))
		identity, e := fn(c, token)
		if e == nil {
			c.Set(User, identity)
			return nil
		}
		c.Response.Header().Set("WWW-Authenticate", `Bearer realm="`+name+`"`)
		return makross.NewHTTPError(http.StatusUnauthorized, e.Error())
	}
}

func parseBearerAuth(auth string) string {
	if strings.HasPrefix(auth, "Bearer ") {
		if bearer, err := base64.StdEncoding.DecodeString(auth[7:]); err == nil {
			return string(bearer)
		}
	}
	return ""
}

// TokenName is the query parameter name for auth token.
var TokenName = "access-token"

// Query returns a makross.Handler that performs authentication based on a token passed via a query parameter.
// It can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/insionng/makross"
//     "github.com/insionng/makross/auth"
//   )
//   func main() {
//     r := makross.New()
//     r.Use(auth.Query(func(token string) (auth.Identity, error) {
//       if token == "secret" {
//         return auth.Identity("demo"), nil
//       }
//       return nil, errors.New("invalid credential")
//     }))
//     r.Get("/demo", func(c *makross.Context) error {
//       fmt.Fprintf(res, "Hello, %v", c.Get(auth.User))
//       return nil
//     })
//   }
//
// When authentication fails, an http.StatusUnauthorized error will be returned.
func Query(fn TokenAuthFunc, tokenName ...string) makross.Handler {
	name := TokenName
	if len(tokenName) > 0 {
		name = tokenName[0]
	}
	return func(c *makross.Context) error {
		token := c.Request.URL.Query().Get(name)
		identity, err := fn(c, token)
		if err != nil {
			return makross.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		c.Set(User, identity)
		return nil
	}
}

// JWTTokenHandler handles the parsed JWT token.
type JWTTokenHandler func(*makross.Context, *jwt.Token) error

//Get a dynamic VerificationKey
type VerificationKeyHandler func(*makross.Context) string

// JWTOptions represents the options that can be used with the JWT handler.
type JWTOptions struct {
	// auth realm. Defaults to "API".
	Realm string
	// the allowed signing method. This is required and should be the actual method that you use to create JWT token. It defaults to "HS256".
	SigningMethod string
	// a function that handles the parsed JWT token. Defaults to DefaultJWTTokenHandler, which stores the token in the makross.context with the key "JWT".
	TokenHandler JWTTokenHandler
	// a function to get a dynamic VerificationKey
	GetVerificationKey VerificationKeyHandler
}

// DefaultJWTTokenHandler stores the parsed JWT token in the makross.context with the key named "JWT".
func DefaultJWTTokenHandler(c *makross.Context, token *jwt.Token) error {
	c.Set("JWT", token)
	return nil
}

// JWT returns a JWT (JSON Web Token) handler which attempts to parse the Bearer header into a JWT token and validate it.
// If both are successful, it will call a JWTTokenHandler to further handle the token. By default, the token
// will be stored in the makross.context with the key named "JWT". Other handlers can retrieve this token to obtain
// the user identity information.
// If the parsing or validation fails, a "WWW-Authenticate" header will be sent, and an http.StatusUnauthorized
// error will be returned.
//
// JWT can be used like the following:
//
//   import (
//     "errors"
//     "fmt"
//     "net/http"
//     "github.com/dgrijalva/jwt-go"
//     "github.com/insionng/makross"
//     "github.com/insionng/makross/auth"
//   )
//   func main() {
//     signingKey := "secret-key"
//     r := makross.New()
//
//     r.Get("/login", func(c *makross.Context) error {
//       id, err := authenticate(c)
//       if err != nil {
//         return err
//       }
//       token, err := auth.NewJWT(jwt.MapClaims{
//         "id": id
//       }, signingKey)
//       if err != nil {
//         return err
//       }
//       return c.Write(token)
//     })
//
//     r.Use(auth.JWT(signingKey))
//     r.Get("/restricted", func(c *makross.Context) error {
//       claims := c.Get("JWT").(*jwt.Token).Claims.(jwt.MapClaims)
//       return c.Write(fmt.Sprint("Welcome, %v!", claims["id"]))
//     })
//   }
func JWT(verificationKey string, options ...JWTOptions) makross.Handler {
	var opt JWTOptions
	if len(options) > 0 {
		opt = options[0]
	}
	if opt.Realm == "" {
		opt.Realm = DefaultRealm
	}
	if opt.SigningMethod == "" {
		opt.SigningMethod = "HS256"
	}
	if opt.TokenHandler == nil {
		opt.TokenHandler = DefaultJWTTokenHandler
	}
	parser := &jwt.Parser{
		ValidMethods: []string{opt.SigningMethod},
	}
	return func(c *makross.Context) error {
		header := c.Request.Header.Get("Authorization")
		message := ""
		if opt.GetVerificationKey != nil {
			verificationKey = opt.GetVerificationKey(c)
		}
		if strings.HasPrefix(header, "Bearer ") {
			token, err := parser.Parse(header[7:], func(t *jwt.Token) (interface{}, error) { return []byte(verificationKey), nil })
			if err == nil && token.Valid {
				err = opt.TokenHandler(c, token)
			}
			if err == nil {
				return nil
			}
			message = err.Error()
		}

		c.Response.Header().Set("WWW-Authenticate", `Bearer realm="`+opt.Realm+`"`)
		if message != "" {
			return makross.NewHTTPError(http.StatusUnauthorized, message)
		}
		return makross.NewHTTPError(http.StatusUnauthorized)
	}
}

// NewJWT creates a new JWT token and returns it as a signed string that may be sent to the client side.
// The signingMethod parameter is optional. It defaults to the HS256 algorithm.
func NewJWT(claims jwt.MapClaims, signingKey string, signingMethod ...jwt.SigningMethod) (string, error) {
	var sm jwt.SigningMethod = jwt.SigningMethodHS256
	if len(signingMethod) > 0 {
		sm = signingMethod[0]
	}
	return jwt.NewWithClaims(sm, claims).SignedString([]byte(signingKey))
}
