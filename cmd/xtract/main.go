package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/chriskirkland/go-xtract/pkg/extractor"
	"github.com/chriskirkland/go-xtract/pkg/util"
)

const (
	stdoutSentinel = "<stdout>"
	compareHelp    = `Compare all json files in dir containing given file, verifying
that all contain the keys this one contains. Only compares - run
with -j first to create/update output file.`
)

var (
	targetFunc = flag.String("func", "fmt.Sprintf", "target func")
	//TODO(cmkirkla): fix character escaping in default template
	outputTemplate = flag.String("template", "{{range .Strings}}{{print .}}\n{{end}}", "output template")
	outputJson     = flag.Bool("j", false, "output json - ignores template")
	outputFile     = flag.String("o", stdoutSentinel, "output file")
	debug          = flag.Bool("v", false, "enable debug output")
	compare        = flag.String("c", "", compareHelp)
)

func main() {
	flag.Parse()

	if *debug {
		log.SetOutput(os.Stdout)
	}

	if len(*compare) > 0 {
		compareFiles(*compare)
		return
	}

	dot := strings.LastIndex(*targetFunc, ".")
	slash := strings.LastIndexByte(*targetFunc, os.PathSeparator)
	if dot < 0 || slash > dot {
		log.Println("'-func' allowed values: 'pkg.Func' or 'path.to/some/pkg.Func'")
		log.Fatalf("'-func' must be a valid qualified function name but found '%s'", *targetFunc)
	}
	tfPackage, tfName := (*targetFunc)[:dot], (*targetFunc)[dot+1:]

	if flag.NArg() == 0 {
		log.Fatalf("one or more file patterns must be provided")
	}
	globs := flag.Args()

	files, err := util.FilesFromPatterns(globs...)
	if err != nil {
		log.Fatalf("error resolving one more provide file pattern: %s", err.Error())
	}

	ext := extractor.New(tfPackage, tfName)
	extractor.ProcessFiles(ext, files...)

	var writer io.Writer = os.Stdout
	if *outputFile != stdoutSentinel {
		f, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("output file '%s' not found: %s", *outputFile, err)
		}
		defer f.Close()

		writer = f
	}

	// generate user output
	if *outputJson {
		vars := ext.Vars()
		m := make(map[string]string)
		for _, v := range vars {
			if len(v.Vars) == 1 {
				m[v.Vars[0]] = v.Val
			} else {
				//0 or multiple var names
				log.Printf("val %q: vars %v", v.Val, v.Vars)
				m[v.Val] = v.Val
			}
		}
		out, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Fprint(writer, string(out))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		//use template
		t, err := template.New("output").Parse(*outputTemplate)
		if err != nil {
			log.Fatalf("failed to parse provided output template: %s", err)
		}

		log.Println("writing extracted strings")
		if err := t.Execute(writer, struct {
			Strings []string
		}{
			Strings: ext.Strings(),
		}); err != nil {
			log.Fatalf("failed to execute template: %s", err)
		}
	}
}

// Read all files in dir containing fname, verify that all have the keys in
// that file. If other files are a superset of the given file, that is not
// treated as an error.
//
// Not deterministic - if multiple keys are missing, reported key may
// differ between runs. This is due to go's runtime map randomization.
func compareFiles(fname string) {
	if !strings.HasSuffix(fname, ".json") {
		log.Fatal("-c: name must end with .json")
	}
	inmap := mapFile(fname)
	if len(inmap) == 0 {
		log.Fatalf("no k-v pairs read from file %s", fname)
	}
	//get all json files in that dir
	dir := fp.Dir(fname)
	files, _ := fp.Glob(fp.Join(dir, "*.json"))
	for _, f := range files {
		if f == fname {
			continue
		}
		m := mapFile(f)
		for k := range inmap {
			_, ok := m[k]
			if !ok {
				log.Fatalf("file %s is missing key %s, which is present in %s", f, k, fname)
			}
		}
	}
}

func mapFile(fname string) map[string]string {
	f, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatalf("error reading %s: %s", fname, err)
	}
	fmap := make(map[string]string)
	err = json.Unmarshal(f, &fmap)
	if err != nil {
		log.Fatalf("json error in %s: %s", fname, err)
	}
	return fmap
}
