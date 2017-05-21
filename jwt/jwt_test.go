package jwt

import (
	"github.com/insionng/makross"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

// jwtCustomInfo defines some custom types we're going to use within our tokens.
type jwtCustomInfo struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// jwtCustomClaims are custom claims expanding default ones.
type jwtCustomClaims struct {
	*jwt.StandardClaims
	jwtCustomInfo
}

func TestJWT(t *testing.T) {

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	validKey := []byte("secret")
	invalidKey := []byte("invalid-key")
	validAuth := Bearer + " " + token

	for _, tc := range []struct {
		expPanic   bool
		expErrCode int // 0 for Success
		config     JWTConfig
		reqURL     string // "/" if empty
		hdrAuth    string
		hdrCookie  string // test.Request doesn't provide SetCookie(); use name=val
		info       string
	}{
		{
			expPanic: true,
			info:     "No signing key provided",
		},
		{
			expErrCode: makross.StatusBadRequest,
			config: JWTConfig{
				SigningKey:    validKey,
				SigningMethod: "RS256",
			},
			info: "Unexpected signing method",
		},
		{
			expErrCode: makross.StatusUnauthorized,
			hdrAuth:    validAuth,
			config:     JWTConfig{SigningKey: invalidKey},
			info:       "Invalid key",
		},
		{
			hdrAuth: validAuth,
			config:  JWTConfig{SigningKey: validKey},
			info:    "Valid JWT",
		},
		{
			hdrAuth: validAuth,
			config: JWTConfig{
				Claims:     &jwtCustomClaims{},
				SigningKey: []byte("secret"),
			},
			info: "Valid JWT with custom claims",
		},
		{
			hdrAuth:    "invalid-auth",
			expErrCode: makross.StatusBadRequest,
			config:     JWTConfig{SigningKey: validKey},
			info:       "Invalid Authorization header",
		},
		{
			config:     JWTConfig{SigningKey: validKey},
			expErrCode: makross.StatusBadRequest,
			info:       "Empty header auth field",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL: "/?a=b&jwt=" + token,
			info:   "Valid query method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwtxyz=" + token,
			expErrCode: makross.StatusBadRequest,
			info:       "Invalid query param name",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b&jwt=invalid-token",
			expErrCode: makross.StatusUnauthorized,
			info:       "Invalid query param value",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "query:jwt",
			},
			reqURL:     "/?a=b",
			expErrCode: makross.StatusBadRequest,
			info:       "Empty query",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			hdrCookie: "jwt=" + token,
			info:      "Valid cookie method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: makross.StatusUnauthorized,
			hdrCookie:  "jwt=invalid",
			info:       "Invalid token with cookie method",
		},
		{
			config: JWTConfig{
				SigningKey:  validKey,
				TokenLookup: "cookie:jwt",
			},
			expErrCode: makross.StatusBadRequest,
			info:       "Empty cookie",
		},
	} {
		if tc.reqURL == "" {
			tc.reqURL = "/"
		}

		if tc.expPanic {
			assert.Panics(t, func() {
				JWTWithConfig(tc.config)
			}, tc.info)
			continue
		}

	}
}
