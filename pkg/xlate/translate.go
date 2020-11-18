package xlate

import (
	"log"
)

//look up translations
func T(in string) string {
	if curLang == defaultLanguage {
		return in
	}
	out, ok := translations[in]
	if ok {
		return out
	}
	//shouldn't get here, but just in case...
	log.Printf("T(%q): missing translation to %s", in, curLang)
	return in
}
