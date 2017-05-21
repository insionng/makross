package moverride_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/moverride"
	"testing"
)

func TestMethodOverride(t *testing.T) {
	m := makross.New()
	m.Use(moverride.MethodOverrideWithConfig(moverride.MethodOverrideConfig{
		Getter: moverride.MethodFromForm("_method"),
	}))
	go m.Listen(":9000")
}
