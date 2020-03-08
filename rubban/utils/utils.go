package utils

import (
	"regexp"
	"strings"
)

//PatternToRegex Transform Index Pattern with Wildcards to a valid Regex.
func PatternToRegex(s string) string {
	// Escape Index Pattern name (to escape dots(.) and other regex special symbols
	s = regexp.QuoteMeta(s)

	// Unescape only \* and \? to actual Regex symbols
	s = strings.NewReplacer("\\*", "(.*)", "\\?", "(.*)").Replace(s)

	// Make Wildcards Lazy Except Last one (hence the n-1)
	n := strings.Count(s, "(.*)")
	s = strings.Replace(s, "(.*)", "(.*?)", n-1)

	return s
}
