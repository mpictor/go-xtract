package xlate

import (
	"log"
)

// T looks up a translation. Input is in the primary language, while output is
// in the current language. If no match is found, a warning is logged and the
// string passes through as-is.
func T(in string) string {
	if curLang == defaultLanguage {
		return in
	}
	if translations == nil {
		log.Printf("T(%s) called before xlate.SetLanguage - translation impossible", in)
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
