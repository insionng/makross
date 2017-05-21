package blimit_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/blimit"
	"github.com/insionng/makross/skipper"
	"testing"
)

func TestBodyLimit(t *testing.T) {
	m := makross.New()
	m.Use(blimit.BodyLimit("2M"))
	go m.Listen(":6666")

	m = makross.New()
	m.Use(blimit.BodyLimitWithConfig(blimit.BodyLimitConfig{Skipper: skipper.DefaultSkipper, Limit: "4M"}))
	go m.Listen(":7777")
}
