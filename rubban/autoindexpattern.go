package rubban

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
	"github.com/sherifabdlnaby/gpool"
	"github.com/sherifabdlnaby/rubban/config"
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
	Enabled         bool
	GeneralPatterns []GeneralPattern
	Schedule        cron.Schedule
	entry           cron.Entry
}

// string replacers
var replaceForPattern = strings.NewReplacer("?", "*")

func replacerForRegex(s string) string {
	s = regexp.QuoteMeta(s)
	s = strings.NewReplacer("\\*", "(.*)", "\\?", "(.*)").Replace(s)
	n := strings.Count(s, "(.*)")
	s = strings.Replace(s, "(.*)", "(.*?)", n-1)
	return s
}

// map will be used to agg results of multiple concurrent api requests
var mu sync.Mutex

//indexPatternMap A map with a concurrent-safe set operation.
type indexPatternMap map[string]kibana.IndexPattern

func (i indexPatternMap) set(indexPattern, timeFieldName string) {
	mu.Lock()
	_, ok := i[indexPattern]
	if !ok {
		i[indexPattern] = kibana.IndexPattern{
			Title:         indexPattern,
			TimeFieldName: timeFieldName,
		}
	}
	mu.Unlock()
}

//NewAutoIndexPattern Constructor
func NewAutoIndexPattern(config config.AutoIndexPattern) *AutoIndexPattern {

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

func (b *rubban) RunAutoIndexPattern() {

	b.logger.Info("Running Auto Index Pattern...")
	startTime := time.Now()

	//// Set for Found Patterns ( a set datastructes using Map )
	computedIndexPatterns := make(indexPatternMap)

	pool := gpool.NewPool(10)
	for _, generalPattern := range b.autoIndexPattern.GeneralPatterns {
		_ = pool.Enqueue(context.TODO(), func() {
			b.getIndexPattern(generalPattern, computedIndexPatterns)
		})
	}

	// Wait for Above to Return
	pool.Stop()

	// Bulk Create Index Patterns
	/// Create List of Index Patterns
	newIndexPatterns := make([]kibana.IndexPattern, 0)
	for _, indexPattern := range computedIndexPatterns {
		newIndexPatterns = append(newIndexPatterns, indexPattern)
	}

	err := b.api.BulkCreateIndexPattern(newIndexPatterns)
	if err != nil {
		b.logger.Errorw("Failed to bulk create new index patterns", "error", err.Error())
	}

	b.logger.Infow(fmt.Sprintf("Successfully created %d Index Patterns. (took ≅ %dms)", len(newIndexPatterns),
		time.Since(startTime).Milliseconds()), "Index Patterns", newIndexPatterns)

	next := b.autoIndexPattern.entry.Schedule.Next(time.Now())
	b.logger.Infof("Next run at %s (%s)", next.String(), humanize.Time(next))
}

func (b *rubban) getIndexPattern(generalPattern GeneralPattern, computedIndexPatterns indexPatternMap) {
	// Get Current IndexPattern Matching Given General Patterns
	indexPatterns, err := b.api.IndexPatterns(generalPattern.Pattern)
	if err != nil {
		b.logger.Warnw("failed to get index patterns matching general pattern. escaping this one...",
			"generalPattern", generalPattern.Pattern, "error", err.Error())
		return
	}

	patternsList := make([]string, 0)
	for _, index := range indexPatterns {
		patternsList = append(patternsList, replacerForRegex(index.Title))
	}

	// Get Indices Matching Given General Pattern
	indices, err := b.api.Indices(generalPattern.Pattern)
	if err != nil {
		b.logger.Warnw("failed to get indices matching a general pattern. escaping this one...",
			"generalPattern", generalPattern.Pattern, "error", err.Error())
		return
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
		computedIndexPatterns.set(newIndexPattern, generalPattern.TimeFieldName)
	}
}

func buildIndexPattern(generalPattern GeneralPattern, unmatchedIndex string) string {
	matchGroups := generalPattern.regex.FindStringSubmatch(unmatchedIndex)
	newIndexPattern := generalPattern.Pattern
	/// Start from 1 to escape first match group which is the whole string.
	for i := 1; i < len(matchGroups); i++ {
		for _, matchGroup := range generalPattern.matchGroups {
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
