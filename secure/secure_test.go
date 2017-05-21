package secure_test

import (
	"github.com/insionng/macross"
	"github.com/insionng/macross/secure"
	"testing"
)

func TestSecure(t *testing.T) {
	m := macross.New()
	m.Use(secure.Secure())
	go m.Listen(":8000")

	m = macross.New()
	m.Use(secure.SecureWithConfig(secure.SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	go m.Listen(":9000")

}
