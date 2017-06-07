package main

import (
	"fmt"
	"github.com/insionng/macross"
	"github.com/macross-contrib/i18n"
)

func main() {
	m := macross.Classic()
	m.Use(i18n.I18n(i18n.Options{
		Directory:   "locale",
		DefaultLang: "zh-CN",
		Langs:       []string{"en-US", "zh-CN"},
		Names:       []string{"English", "简体中文"},
		Redirect:    true,
	}))

	m.Get("/", func(self *macross.Context) error {
		return self.String("current language is " + self.Language())
	})

	// Use in handler.
	m.Get("/trans", func(self *macross.Context) error {
		return self.String(fmt.Sprintf("hello %s", self.Tr("world")))
	})

	fmt.Println("Listen on 9999")
	m.Listen(9999)
}
