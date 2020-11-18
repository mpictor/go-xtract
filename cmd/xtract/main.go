package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/mpictor/go-xtract/pkg/extractor"
	"github.com/mpictor/go-xtract/pkg/util"
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
	globs := fixupGlobs()

	files, err := util.FilesFromPatterns(globs...)
	if err != nil {
		log.Fatalf("error resolving one more provide file pattern: %s", err.Error())
	}
	if len(files) == 0 {
		log.Fatalf("found 0 files in globs %v", globs)
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
		jsonOut(ext, writer)
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

//reads from a json file into a map
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

func jsonOut(ext extractor.Extractor, writer io.Writer) {
	vars := ext.Vars()
	m := make(map[string]string)
	for _, v := range vars {
		if len(v.Vars) == 1 {
			m[v.Vars[0]] = v.Val
		} else {
			//0 or multiple var names - use a sanitized copy of val as key
			sanitize := func(r rune) rune {
				//replace all but letters with underscores
				switch {
				case r < 65, r > 122, r > 90 && r < 97:
					return '_'
				default:
					return r
				}
			}
			k := strings.Map(sanitize, v.Val)
			if len(k) > 40 {
				sha := sha1.Sum([]byte(v.Val))
				enc := base64.RawStdEncoding.EncodeToString(sha[:])
				if len(enc) > 10 {
					enc = enc[:10]
				}
				k = k[:40-len(enc)] + string(enc)
			}
			log.Printf("val %q: vars %v - using %s as key", v.Val, v.Vars, k)
			m[k] = v.Val
		}
	}
	err := util.NewJSONEncoder(writer).Encode(m)
	if err != nil {
		log.Fatal(err)
	}
}

//get globs; if any are not absolute, fix.
func fixupGlobs() []string {
	globs := flag.Args()
	sep := string(os.PathSeparator)
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getting working dir: %s", err)
	}
	wd += sep
	for i := range globs {
		if !strings.HasPrefix(globs[i], sep) {
			log.Printf("fix glob %s add %s", globs[i], wd)
			globs[i] = wd + globs[i]
		}
	}
	return globs
}
