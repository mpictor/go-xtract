package main

import (
	"github.com/chriskirkland/go-xtract/_integration/json-out/src/pkg"
)

func main() {
	pkg.Fn("string passed to function in nested pkg")
	pkg.Fn(pkg.SomeStr)
	pkg.Fn("some very very long string blah blah blah blah blah blah blah blah blah blah blah blah")
}
