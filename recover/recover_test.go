package recover_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/recover"
	"testing"
)

func TestRecover(t *testing.T) {
	m := makross.New()
	m.Use(recover.RecoverWithConfig(recover.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	go m.Listen(":8888")
}
