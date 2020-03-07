package autoindexpattern

import (
	"context"
	"regexp"
	"strings"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//GeneralPattern hold attributes for a GeneralPattern loaded from config.
type GeneralPattern struct {
	Pattern       string
	regex         regexp.Regexp
	TimeFieldName string
	matchGroups   []int
}

//AutoIndexPattern hold attributes for a RunAutoIndexPattern loaded from config.
type AutoIndexPattern struct {
	name            string
	concurrency     int
	GeneralPatterns []GeneralPattern
	kibana          kibana.API
	log             log.Logger
}

// string replacers
var replaceForPattern = strings.NewReplacer("?", "*")

//NewAutoIndexPattern Constructor
func NewAutoIndexPattern(config config.AutoIndexPattern, kibana kibana.API, log log.Logger) *AutoIndexPattern {

	generalPattern := make([]GeneralPattern, 0)

	for _, pattern := range config.GeneralPatterns {
		regex := regexp.MustCompile(replacerForRegex(pattern.Pattern))
		generalPattern = append(generalPattern, GeneralPattern{
			Pattern:       replaceForPattern.Replace(pattern.Pattern),
			regex:         *regex,
			TimeFieldName: pattern.TimeFieldName,
			matchGroups:   getMatchGroups(pattern.Pattern),
		})
	}

	return &AutoIndexPattern{
		name:            "Auto Index Pattern",
		concurrency:     config.Concurrency,
		GeneralPatterns: generalPattern,
		kibana:          kibana,
		log:             log,
	}
}

func (a *AutoIndexPattern) getIndexPattern(ctx context.Context, generalPattern GeneralPattern) map[string]kibana.IndexPattern {

	newIndexPatterns := make(map[string]kibana.IndexPattern)

	// Get Current IndexPattern Matching Given General Patterns
	indexPatterns, err := a.kibana.IndexPatterns(ctx, generalPattern.Pattern, nil)
	if err != nil {
		a.log.Warnw("failed to get index patterns matching general pattern. escaping this one...",
			"generalPattern", generalPattern.Pattern, "error", err.Error())
		return newIndexPatterns
	}

	patternsList := make([]string, 0)
	for _, index := range indexPatterns {
		patternsList = append(patternsList, replacerForRegex(index.Title))
	}

	// Get Indices Matching Given General Pattern
	indices, err := a.kibana.Indices(ctx, generalPattern.Pattern)
	if err != nil {
		a.log.Warnw("failed to get indices matching a general pattern. escaping this one...",
			"generalPattern", generalPattern.Pattern, "error", err.Error())
		return newIndexPatterns
	}

	// Get Indices That Hasn't Matched ANY IndexPattern

	//// Build regex
	var matchedIndicesRegx *regexp.Regexp

	if len(patternsList) > 0 {
		matchedIndicesRegx = regexp.MustCompile(strings.Join(patternsList, "|"))
	} else {
		// If no PatternList that means that the first time to encounter this pattern. So we won't match anything.
		matchedIndicesRegx = regexp.MustCompile("$.")
	}

	unmatchedIndices := make([]string, 0)
	for _, index := range indices {
		if !matchedIndicesRegx.MatchString(index.Name) {
			unmatchedIndices = append(unmatchedIndices, index.Name)
		}
	}

	// Build Index Pattern for every unmatched Index
	for _, unmatchedIndex := range unmatchedIndices {
		newIndexPattern := buildIndexPattern(generalPattern, unmatchedIndex)
		newIndexPatterns[newIndexPattern] = kibana.IndexPattern{
			Title:         newIndexPattern,
			TimeFieldName: generalPattern.TimeFieldName,
		}
	}

	return newIndexPatterns
}

func buildIndexPattern(generalPattern GeneralPattern, unmatchedIndex string) string {
	matchGroups := generalPattern.regex.FindStringSubmatch(unmatchedIndex)
	newIndexPattern := generalPattern.Pattern
	/// Start from 1 to escape first match group which is the whole string.
	for i := 1; i < len(matchGroups); i++ {
		var match bool
		for _, matchGroup := range generalPattern.matchGroups {
			if matchGroup == i {
				match = true
			}
		}

		if match {
			// This is a match Group
			newIndexPattern = strings.Replace(newIndexPattern, "*", matchGroups[i], 1)
		} else {
			// This is a wildcard (make it ? for now) (yes there can be a more efficient logic for that.)
			newIndexPattern = strings.Replace(newIndexPattern, "*", "?", 1)
		}

	}
	newIndexPattern = strings.Replace(newIndexPattern, "?", "*", -1)
	return newIndexPattern
}

func getMatchGroups(pattern string) []int {
	groups := make([]int, 0)
	group := 1
	for _, char := range pattern {
		if char == '?' {
			groups = append(groups, group)
			group++
		} else if char == '*' {
			group++
		}
	}
	return groups
}

func replacerForRegex(s string) string {
	// Escape Index Pattern name (to escape dots(.) and other regex special symbols
	s = regexp.QuoteMeta(s)

	// Unescape only \* and \? to actual Regex symbols
	s = strings.NewReplacer("\\*", "(.*)", "\\?", "(.*)").Replace(s)

	// Make Wildcards Lazy Except Last one (hence the n-1)
	n := strings.Count(s, "(.*)")
	s = strings.Replace(s, "(.*)", "(.*?)", n-1)

	return s
}
