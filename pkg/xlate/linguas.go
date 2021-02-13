package xlate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type (
	// Lingua is a native name for a language. Must be unique.
	Lingua string

	// Linguas is a list of languages. A named type is needed for sorting.
	Linguas []Lingua

	// Locale is a shorter name for a lingua, such as 'de' or 'en-us'. Lowercase.
	// For our purposes, corresponds to the bindata asset name.
	//
	// Note that Locale cannot be normalized to lowercase without additional work in matching with bindata assets.
	Locale string
)

var (
	// AvailableLanguages is a list of supported languages for the UI. The
	// first language (english) will be the default. Otherwise unsorted as the
	// source is keys from a map.
	AvailableLanguages Linguas

	//the language T()'s input strings are in
	defaultLanguage Lingua

	//current language for translations
	curLang Lingua

	// map from language name to locale, which must match asset name - for example
	// {"English": "en-us",}
	//      ==>   en-us.json
	langAssetMap map[Lingua]Locale

	//maps from phrase in primary language to current
	translations map[string]string

	ErrNotFound       = errors.New("Lang asset not found")
	ErrMultiSetup     = errors.New("Setup called multiple times")
	ErrDefLangAbsent  = errors.New("Default language not loaded")
	ErrLangNameAbsent = errors.New("Missing key AA_NativeLangName")
)

func (l Linguas) Len() int           { return len(l) }
func (l Linguas) Less(i, j int) bool { return strings.Compare(string(l[i]), string(l[j])) < 0 }
func (l Linguas) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// Locale converts Lingua to Locale.
func (l Lingua) Locale() Locale {
	return Locale(langAssetMap[l])
}

// Lingua converts Locale to Lingua.
func (l Locale) Lingua() Lingua {
	fl := l.FuzzyMatch(GetLocales())
	if len(fl) == 0 {
		return ""
	}
	for lin, loc := range langAssetMap {
		if fl == loc {
			return lin
		}
	}
	return ""
}

// Equal performs a case-insensitive comparison.
func (l Locale) Equal(r Locale) bool {
	return strings.EqualFold(string(l), string(r))
	// return strings.ToLower(string(l)) == strings.ToLower(string(r))
}

// FuzzyMatch finds a locale we have that's close to what is requested (i.e. en-gb will match en or en-us).
func (l Locale) FuzzyMatch(set []Locale) Locale {
	//first try for exact match (exact except for case)
	for _, loc := range set {
		if l.Equal(loc) {
			return l
		}
	}
	//no exact match, try without a prefix
	//also assumes the first match is good enough
	elems := strings.Split(string(l), "-")
	reqPfx := Locale(elems[0])
	if len(elems) > 1 {
		for _, loc := range set {
			if reqPfx.Equal(loc) {
				log.Printf("warning: using inexact locale %s when %s was requested", reqPfx, l)
				return reqPfx
			}
		}
	}
	for _, loc := range set {
		elements := strings.Split(string(loc), "-")
		pfx := Locale(elements[0])
		if reqPfx.Equal(pfx) {
			log.Printf("warning: using inexact locale %s when %s was requested", reqPfx, l)
			return reqPfx
		}
	}
	log.Printf("warning: no match, exact or approximate, found for locale %s", l)
	return ""
}

// SetLanguage sets the current language, returning an error if the language
// is not found. To find the correct asset, lang is compared to the field
// AA_NativeLangName in each *.json asset, until a match is found. Once
// this is found, a map is constructed mapping from a phrase in the default
// language to a phrase in the new language. Subsequent calls to T() use
// this map to find the correct phrase to return.
func SetLanguage(lang Lingua) (err error) {
	if langAssetMap == nil {
		return fmt.Errorf("Must call xlate.Setup() first. %s: %w", lang, ErrNotFound)
	}
	log.Printf("Setting language to %s", lang)
	_, ok := langAssetMap[lang]
	if !ok {
		return fmt.Errorf("%s: %w", lang, ErrNotFound)
	}
	//load default and target lang, use keys to map def val to target val
	var defLang, tgtLang map[string]string
	defLang, err = langMap(defaultLanguage)
	if err != nil {
		return err
	}
	tgtLang, err = langMap(lang)
	if err != nil {
		return err
	}
	//clear out any existing translation
	for k := range translations {
		delete(translations, k)
	}
	translations = make(map[string]string)

	for varname, phrase := range defLang {
		translations[phrase] = tgtLang[varname]
	}
	curLang = lang
	return nil
}

//loads lang asset; asset maps from var name to phrase
func langMap(lang Lingua) (m map[string]string, err error) {
	var data []byte
	assetName := string(langAssetMap[lang]) + ".json"
	datafn, ok := bindata[assetName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, lang)
	}
	data, err = datafn()
	if err == nil {
		err = json.Unmarshal(data, &m)
	}
	return
}

// GetLanguage returns the current lingua.
func GetLanguage() Lingua { return curLang }

// GetLocale returns the current locale.
func GetLocale() Locale { return langAssetMap[curLang] }

// GetLocales returns all available locales.
func GetLocales() []Locale {
	var locs []Locale
	for _, l := range langAssetMap {
		locs = append(locs, l)
	}
	return locs
}

// Bindata matches the type used for go-bindata assets.
type Bindata map[string]func() ([]byte, error)

var bindata Bindata

var loaded = false

// Setup adds languages to AvailableLanguages based on assets found. This
// function must be called exactly once, and before any other funcs in the
// package.
func Setup(defaultLang Lingua, bdata Bindata) error {
	if loaded {
		return ErrMultiSetup
	}
	defaultLanguage = defaultLang
	curLang = defaultLang
	bindata = bdata
	langAssetMap = make(map[Lingua]Locale)
	AvailableLanguages = []Lingua{defaultLanguage}
	for fname, loader := range bindata {
		if !strings.HasSuffix(fname, ".json") {
			continue
		}
		var err error
		var data []byte
		var lname Lingua
		//if name == "yo-da.json" {continue}
		if data, err = loader(); err != nil {
			return err
		}
		if lname, err = getName(data, fname); err != nil {
			return err
		}
		if lname != defaultLanguage {
			AvailableLanguages = append(AvailableLanguages, lname)
		}
		langAssetMap[lname] = Locale(strings.TrimSuffix(fname, ".json"))
	}
	if _, present := langAssetMap[defaultLanguage]; !present {
		langAssetMap = nil
		AvailableLanguages = nil
		return ErrDefLangAbsent
	}
	loaded = true
	return nil
}

// looks for AA_NativeLangName in json, returns if present
func getName(jdata []byte, fname string) (Lingua, error) {
	var l struct{ AA_NativeLangName string } //we only care about AA_NativeLangName at the moment
	err := json.Unmarshal(jdata, &l)
	if err != nil {
		return "", err
	}
	//default (english) is already present, to ensure it's first item
	if len(l.AA_NativeLangName) == 0 {
		err = fmt.Errorf("%s: %w", fname, ErrLangNameAbsent)
		return "", err
	}
	return Lingua(l.AA_NativeLangName), nil
}
