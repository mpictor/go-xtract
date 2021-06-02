package xlate

import (
	"fmt"
	"log"
)

// T looks up a translation. Input is in the primary language, while output is
// in the current language. If no match is found, a warning is logged and the
// string passes through as-is.
func T(in string) string {
	out, err := TErr(in)
	if err != nil {
		log.Print(err)
	}
	return out
}

//Like T, but returns an error rather than logging.
func TErr(in string) (string, error) {
	if curLang == defaultLanguage {
		return in, nil
	}
	if translations == nil {
		return in, fmt.Errorf("T(%s) called before xlate.SetLanguage - translation impossible", in)
	}
	out, ok := translations[in]
	if ok {
		return out, nil
	}
	//shouldn't get here, but just in case...
	return in, fmt.Errorf("T(%q): missing translation to %s", in, curLang)
}
