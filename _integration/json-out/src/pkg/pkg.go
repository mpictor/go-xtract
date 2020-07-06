package pkg

import (
	"fmt"
	"os"
)

var SomeStr = "var SomeStr in package pkg"

func Fn(s string) {
	fmt.Fprintf(os.Stderr, "Fn(%q)\n", s)
}
