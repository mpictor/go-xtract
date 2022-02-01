package xlate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	Str       = "This Is Only A Test"
	StrOther  = "other other other"
	tsjson    = `{"AA_NativeLangName":"test","Str":"` + Str + `"}`
	otherjson = `{"AA_NativeLangName":"other","Str":"` + StrOther + `"}`
)

func TestSetup(t *testing.T) {
	err := Setup("test", nil)
	require.Error(t, err, "expect error for nil bindata")

	bd := Bindata{"": func() ([]byte, error) { return nil, nil }}
	err = Setup("test", bd)
	require.Error(t, err, "expect error when no usable assets")

	bd["ts.json"] = func() ([]byte, error) { return []byte(tsjson), nil }
	err = Setup("test", bd)
	require.NoError(t, err, "bindata is valid")

	err = Setup("test", bd)
	require.Error(t, err, "multiple calls to setup")
}

func TestTranslations(t *testing.T) {
	bd := Bindata{
		"te-st.json": func() ([]byte, error) { return []byte(tsjson), nil },
		"ot-hr.json": func() ([]byte, error) { return []byte(otherjson), nil },
	}
	loaded = false
	err := Setup("test", bd)
	require.NoError(t, err, "bindata is valid")
	out := T(Str)
	require.Equal(t, out, Str, "same language - must match")

	err = SetLanguage("other")
	require.NoError(t, err, "set lang to valid choice")
	out = T(Str)
	require.Equal(t, out, StrOther, "translation available - must translate")

	unxlated := "something untranslated"
	out = T(unxlated)
	require.Equal(t, unxlated, out, "no translation - must pass through verbatim")
}

func TestMissingLang(t *testing.T) {
	bd := Bindata{
		"te-st.json": func() ([]byte, error) { return []byte(tsjson), nil },
		"ot-hr.json": func() ([]byte, error) { return []byte(otherjson), nil },
	}
	loaded = false
	err := Setup("test", bd)
	require.NoError(t, err, "bindata is valid")

	err = SetLanguage("missing")
	require.Error(t, err, "expect error for missing language")
	out := T(Str)
	require.Equal(t, out, Str, "no translation - must pass through verbatim")
}

func TestAdoptBindata(t *testing.T) {
	type someAsset struct {
		d []byte
		o string
	}
	type someBindata map[string]func() (*someAsset, error)
	bdIn := someBindata{
		"a": func() (*someAsset, error) {
			return &someAsset{[]byte("aa"), "bb"}, nil
		},
		"b": func() (*someAsset, error) {
			return &someAsset{[]byte("bb"), "bb"}, nil
		},
	}
	bdOut, err := AdoptBindata(bdIn)
	require.NotNil(t, bdOut)
	require.Nil(t, err)
	assert.Equal(t, 2, len(bdOut))
	{
		a, e := bdOut["a"]()
		assert.Nil(t, e)
		assert.Equal(t, []byte("aa"), a)
	}
	{
		a, e := bdOut["b"]()
		assert.Nil(t, e)
		assert.Equal(t, []byte("bb"), a)
	}
}
