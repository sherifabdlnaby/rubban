package bosun

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
	"github.com/sherifabdlnaby/bosun/bosun/kibana"
	"github.com/sherifabdlnaby/bosun/config"
)

var replaceForPattern = strings.NewReplacer("?", "*")

func replacerForRegex(s string) string {
	s = strings.NewReplacer("*", "(.*)", "?", "(.*)").Replace(s)
	n := strings.Count(s, "(.*)")
	s = strings.Replace(s, "(.*)", "(.*?)", n-1)
	return s
}

type GeneralPattern struct {
	Pattern       string
	Regex         regexp.Regexp
	TimeFieldName string
	MatchGroups   []int
}

type AutoIndexPattern struct {
	Enabled         bool
	GeneralPatterns []GeneralPattern
	Schedule        cron.Schedule
	entry           cron.Entry
}

func NewAutoIndexPattern(config config.AutoIndexPattern) *AutoIndexPattern {

	generalPattern := make([]GeneralPattern, 0)

	for _, pattern := range config.GeneralPatterns {
		regex := regexp.MustCompile(replacerForRegex(pattern.Pattern))
		generalPattern = append(generalPattern, GeneralPattern{
			Pattern:       replaceForPattern.Replace(pattern.Pattern),
			Regex:         *regex,
			TimeFieldName: pattern.TimeFieldName,
			MatchGroups:   getMatchGroups(pattern.Pattern),
		})
	}

	schedule, err := cron.ParseStandard(config.Schedule)
	if err != nil {
		panic(err)
	}

	return &AutoIndexPattern{
		Enabled:         config.Enabled,
		GeneralPatterns: generalPattern,
		Schedule:        schedule,
	}
}

func (b *Bosun) AutoIndexPattern() {

	b.logger.Info("Running Auto Index Pattern...")

	//// Set for Found Patterns ( a set datastructes using Map )
	computedIndexPatterns := make(map[string]kibana.IndexPattern)

	// TODO make this concurrent
	for _, generalPattern := range b.autoIndexPattern.GeneralPatterns {

		// TODO send those two concurrently.

		// Get Current IndexPattern Matching Given General Patterns
		indexPatterns, err := b.api.IndexPatterns(generalPattern.Pattern)
		if err != nil {
			b.logger.Warnw("failed to get index patterns matching a general pattern. escaping this one...", "generalPattern", generalPattern.Pattern)
			continue
		}
		patternsList := make([]string, 0)
		for _, index := range indexPatterns {
			patternsList = append(patternsList, replacerForRegex(index.Title))
		}

		// Get Indices Matching Given General Pattern
		indices, err := b.api.Indices(generalPattern.Pattern)
		unmatchedIndices := make([]string, 0)

		// Get Indices That Hasn't Matched ANY IndexPattern
		//// Build Regex
		matchedIndicesRegx := regexp.MustCompile(strings.Join(patternsList, "|"))

		//// Filter Indices
		for _, index := range indices {
			if !matchedIndicesRegx.MatchString(index.Name) {
				unmatchedIndices = append(unmatchedIndices, index.Name)
			}
		}

		// Build Index Pattern for every unmatched Index
		for _, unmatchedIndex := range unmatchedIndices {
			matchGroups := generalPattern.Regex.FindStringSubmatch(unmatchedIndex)
			newIndexPattern := generalPattern.Pattern
			/// Start from 1 to escape first match group which is the whole string.
			for i := 1; i < len(matchGroups); i++ {
				for _, matchGroup := range generalPattern.MatchGroups {
					if matchGroup == i {
						// This is a match Group
						newIndexPattern = strings.Replace(newIndexPattern, "*", matchGroups[i], 1)
					} else {
						// This is a wildcard (make it ? for now) (yes there can be a more efficient logic for that.)
						newIndexPattern = strings.Replace(newIndexPattern, "*", "?", 1)
					}
				}
			}
			newIndexPattern = strings.Replace(newIndexPattern, "?", "*", -1)
			_, ok := computedIndexPatterns[newIndexPattern]
			if !ok {
				computedIndexPatterns[newIndexPattern] = kibana.IndexPattern{
					Title:         newIndexPattern,
					TimeFieldName: generalPattern.TimeFieldName,
				}
			}
		}

	}

	// Bulk Create Index Patterns
	/// Create List of Index Patterns
	var newIndexPatterns []kibana.IndexPattern
	for _, indexPattern := range computedIndexPatterns {
		newIndexPatterns = append(newIndexPatterns, indexPattern)
	}

	err := b.api.BulkCreateIndexPattern(newIndexPatterns)
	if err != nil {
		b.logger.Errorw("Failed to bulk create new index patterns", "error", err.Error())
	}

	b.logger.Infow(fmt.Sprintf("Successfully created %d Index Patterns.", len(newIndexPatterns)), "Index Patterns", newIndexPatterns)
	next := b.autoIndexPattern.entry.Schedule.Next(time.Now())
	b.logger.Infof("Next run at %s (%s)", next.String(), humanize.Time(next))

	return
}

func getMatchGroups(pattern string) []int {
	groups := make([]int, 0)
	group := 1
	for _, char := range pattern {
		if char == 42 {
			groups = append(groups, group)
			group++
		} else if char == 63 {
			group++
		}
	}
	return groups
}
