package main

//set up shortcuts
//go:generate -command xtract go run github.com/mpictor/go-xtract/cmd/xtract
//go:generate -command bindata go run github.com/jteeuwen/go-bindata/go-bindata

//extract strings, overwriting the primary language's file (here, en-us.json)
//NOTE: cannot use relative path for files to scan
//go:generate xtract -func github.com/mpictor/go-xtract/pkg/xlate.T -j -o data/en-us.json **/*.go

//check all json files in data against the primary language's file (again, en-us.json)
//go:generate xtract -v -c data/en-us.json

//embed
//go:generate bindata -prefix=data -pkg=$GOPACKAGE data
