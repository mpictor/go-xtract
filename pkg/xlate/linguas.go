package xlate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	// AvailableLanguages is a list of supported languages for the UI. The
	// first language (english) will be the default. Otherwise unsorted as the
	// source is keys from a map.
	AvailableLanguages []string

	//the language T()'s input strings are in
	defaultLanguage string

	//current language for translations
	curLang string

	// map from language name to asset name, for example
	// {"English": "en-us.json",}
	langAssetMap map[string]string

	//maps from phrase in primary language to current
	translations map[string]string

	ErrNotFound       = errors.New("Lang asset not found")
	ErrMultiSetup     = errors.New("Setup called multiple times")
	ErrDefLangAbsent  = errors.New("Default language not loaded")
	ErrLangNameAbsent = errors.New("Missing key AA_NativeLangName")
)

// SetLanguage sets the current language, returning an error if the language
// is not found. To find the correct asset, lang is compared to the field
// AA_NativeLangName in each *.json asset, until a match is found. Once
// this is found, a map is constructed mapping from a phrase in the default
// language to a phrase in the new language. Subsequent calls to T() use
// this map to find the correct phrase to return.
func SetLanguage(lang string) (err error) {
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
func langMap(lang string) (m map[string]string, err error) {
	var data []byte
	datafn, ok := bindata[langAssetMap[lang]]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, lang)
	}
	data, err = datafn()
	if err == nil {
		err = json.Unmarshal(data, &m)
	}
	return
}

func GetLanguage() string { return curLang }

type Bindata map[string]func() ([]byte, error)

var bindata Bindata

var loaded = false

// Setup adds languages to AvailableLanguages based on assets found. This
// function must be called exactly once, and before any other funcs in the
// package.
func Setup(defaultLang string, bdata Bindata) error {
	if loaded {
		return ErrMultiSetup
	}
	defaultLanguage = defaultLang
	curLang = defaultLang
	bindata = bdata
	langAssetMap = make(map[string]string)
	AvailableLanguages = []string{defaultLanguage}
	for fname, loader := range bindata {
		if !strings.HasSuffix(fname, ".json") {
			continue
		}
		var err error
		var data []byte
		var lname string
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
		langAssetMap[lname] = fname
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
func getName(jdata []byte, fname string) (string, error) {
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
	return l.AA_NativeLangName, nil
}
