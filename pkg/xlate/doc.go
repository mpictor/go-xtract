// Package xlate translates strings from one language to another, using data
// injected via the Setup function. The injected data is a map, as generated
// by go-bindata. We support two binary data generators:
// - github.com/jteeuwen/go-bindata (archived, not supported any more)
// - github.com/go-bindata/go-bindata.
// In either case, map keys are asset names, such as en-us.json, while values
// are asset access functions. The asset value (payload) is json from cmd/xtract.
//
// This package assumes there are no duplicate strings in the primary language.
// If two strings are the same in the primary language but differ in another,
// for example due to context, this package will fail to accurately translate
// one of those strings; the other translation will always be used. To avoid,
// ensure the input strings (not the const/var names) are unique.
package xlate
