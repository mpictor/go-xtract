# go-xtract
<a href="https://travis-ci.com/mpictor/go-xtract.svg?branch=master" alt="build status">
  <img src="https://travis-ci.com/mpictor/go-xtract.svg?branch=master" /></a>

An aid for translations.
* `xtract` tool: extracts strings seen in calls to a particular function, such as `xlate.T`, and write them as text or key-value json.
* `xlate` package: can load those translations and replace strings from the primary language with ones in the current language.
  * Note that xlate.Setup() requires a map equivalent to that output by go-bindata.

### Fork

This is a fork of github.com/chriskirkland/go-xtract, with additions to make it work in my translation workflow. Original description:
Library for extracting arbitrary strings from Go code.

### License

I see no references to copyright or license in the upstream repo. I am choosing to interpret this as "at least as permissive as BSD", and licensing my contributions under BSD 3-clause. Chris, if you have any issues with this, please reach out.

## Usage

### Installing the CLI
```
go get github.com/mpictor/go-xtract/cmd/xtract
```

### xtract examples
Help:
```console
~$ xtract -h
Usage of xtract:
  -c string
        Compare all json files in dir containing given file, verifying
        that all contain the keys this one contains. Only compares - run
        with -j first to create/update output file.
  -func string
        target func (default "github.com/mpictor/go-xtract/pkg/xlate.T")
  -j    output json - ignores template
  -o string
        output file (default "<stdout>")
  -template string
        output template (default "{{range .Strings}}{{print .}}\n{{end}}")
  -v    enable debug output

```

#### all files
Run for all Go files in the repo:
```sh
xtract **/*.go
```

#### specific files
Run over specific set of Go files:
```sh
xtract 'plugins/models/command_metadata.go' 'plugins/commands/cmdalb/*.go'
xtract 'pkg/*.go'
```

#### json
Write json output to a file:
```sh
xtract -j -o data/en_us.json **/*.go
```
When xtract runs with `-j`, it outputs key-value pairs to the file. The value is the string content, while the key is the string or constant's name. In the case of a literal, a key is created from the literal. Non-literals must be exported (capitalized) for xtract to be able to use them.

### xlate example
```go
const Ello = "Hello, World!"
// _bindata created by running go-bindata
err = xlate.Setup("US English",_bindata)
fmt.Println("Available languages: %s", strings.Join(xlate.AvailableLanguages,", ")
fmt.Println(xlate.T(Ello)) // output: "Hello, World!"

err = xlate.SetLanguage("Latin") //this string must match AA_NativeLangName in some *.json asset in _bindata
fmt.Println(xlate.T(Ello)) // output: the string translated to latin
```
Note that the asset names in _bindata must end in .json. Typically they'll identify the language and country (i.e. en-us.json) for the benefit of translators, developers, etc - but this is not a requirement.

### combined example

For an example of `xtract` and `xlate` used together, see _integration/xlate_example/src. This example depends on code generation at compile time (using cmd/xtract and go-bindata), but the code generation could be done earlier. 
