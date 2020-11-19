package pkg

import (
	"fmt"

	"github.com/mpictor/go-xtract/pkg/xlate"
)

var _ = xlate.T(AA_NativeLangName) //ensure name for default language is in json

func DoSomething() {
	fmt.Println(xlate.T(HelloWorld))
}
