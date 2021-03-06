package main

import (
	"flag"
	"log"
	"os"
	"sort"

	"github.com/mpictor/go-xtract/_integration/xlate_example/src/pkg"
	"github.com/mpictor/go-xtract/pkg/xlate"
)

func main() {
	lingua := flag.String("lingua", "Dagobah", "language to use (all -> loop over all found)")
	flag.Parse()

	log.SetOutput(os.Stderr)

	//set default language's name and load translations
	//_bindata is generated by go-bindata
	err := xlate.Setup("English", _bindata)
	if err != nil {
		log.Fatalf("xlate setup: %s", err)
	}

	if *lingua == "all" {
		//for demonstration purposes, print message in all languages

		//these languages come from a map, so order is not guaranteed
		sort.Sort(xlate.AvailableLanguages)
		for _, l := range xlate.AvailableLanguages {
			doSomething(l)
		}
	} else {
		doSomething(*lingua)
	}
}

func doSomething(lingua string) {
	if err := xlate.SetLanguage(lingua); err != nil {
		log.Fatalf("setting language: %s", err)
	}
	pkg.DoSomething()
}
