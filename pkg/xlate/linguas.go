package xlate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

var defaultLanguage string

var (
	// AvailableLanguages is a list of supported languages for the UI.
	// The first language (english) will be the default. Otherwise unsorted as the source is keys from a map.
	AvailableLanguages []string

	//current language for translations
	curLang = defaultLanguage

	//map from language name to asset name
	langAssetMap = make(map[string]string) //{"English": "en-us.json",}

	translations = make(map[string]string)

	ErrNotFound       = errors.New("Lang asset not found")
	ErrMultiSetup     = errors.New("Setup called multiple times")
	ErrDefLangAbsent  = errors.New("Default language not loaded")
	ErrLangNameAbsent = errors.New("Missing key AA_NativeLangName")
)

func SetLanguage(lang string) (err error) {
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
	for varname, phrase := range defLang {
		translations[phrase] = tgtLang[varname]
	}
	curLang = lang
	return nil
}

//loads lang asset; asset maps from var name to phrase
func langMap(lang string) (m map[string]string, err error) {
	var data []byte
	//data, err = Asset(langAssetMap[lang])
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

//adds languages to AvailableLanguages based on assets found
func Setup(defaultLang string, bdata Bindata) error {
	if loaded {
		return ErrMultiSetup
	}
	defaultLanguage = defaultLang
	bindata = bdata
	AvailableLanguages = []string{defaultLanguage}
	for name, loader := range bindata {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		//if name == "yo-da.json" {continue}
		data, err := loader()
		if err != nil {
			return err
		}
		var l struct{ AA_NativeLangName string } //we only care about AA_NativeLangName at the moment
		err = json.Unmarshal(data, &l)
		if err != nil {
			return err
		}
		//default (english) is already present, to ensure it's first item
		if len(l.AA_NativeLangName) == 0 {
			return fmt.Errorf("%s: %w", name, ErrLangNameAbsent)
		}
		if l.AA_NativeLangName != defaultLanguage {
			AvailableLanguages = append(AvailableLanguages, l.AA_NativeLangName)
		}
		langAssetMap[l.AA_NativeLangName] = name
	}
	if _, present := langAssetMap[defaultLanguage]; !present {
		langAssetMap = nil
		AvailableLanguages = nil
		return ErrDefLangAbsent
	}
	loaded = true
	return nil
}
