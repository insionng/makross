package csrf_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/csrf"
	"testing"
)

func TestCSRF(t *testing.T) {
	e := makross.New()
	e.Use(csrf.CSRFWithConfig(csrf.CSRFConfig{
		TokenLookup: "header:X-XSRF-TOKEN",
	}))
	go e.Listen(9000)
}
