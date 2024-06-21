package main

import (
	"regexp"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonLetters = regexp.MustCompile(`[^a-zA-Z0-9:<>= ]+`)

func clean(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	result = nonLetters.ReplaceAllString(result, "")
	return result
}
