package main

import (
	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/chrome"
	_ "github.com/zellyn/kooky/firefox"
	_ "github.com/zellyn/kooky/safari"
)

func main() {
	_ = kooky.Filter(nil)
}
