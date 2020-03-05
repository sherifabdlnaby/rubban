package refreshindexpattern

import (
	"context"
	"fmt"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//GeneralPattern hold attributes for a GeneralPattern loaded from config.
type Pattern struct {
	string
}

//RefreshIndexPattern hold attributes for a RunAutoIndexPattern loaded from config.
type RefreshIndexPattern struct {
	name     string
	Patterns []Pattern
	kibana   kibana.API
	log      log.Logger
}

//NewRefreshIndexPattern Constructor
func NewRefreshIndexPattern(config config.RefreshIndexPattern, kibana kibana.API, log log.Logger) *RefreshIndexPattern {

	Patterns := make([]Pattern, 0)

	for _, pattern := range config.Patterns {
		Patterns = append(Patterns, Pattern{
			pattern,
		})
	}

	return &RefreshIndexPattern{
		name:     "Refresh Indices Patterns",
		Patterns: Patterns,
		kibana:   kibana,
		log:      log,
	}
}

func (a *RefreshIndexPattern) getIndexPattern(ctx context.Context, pattern Pattern) []kibana.IndexPattern {
	// Get Current IndexPattern Matching Given General Patterns
	Patterns, err := a.kibana.IndexPatterns(ctx, pattern.string, []string{"version"})
	if err != nil {
		a.log.Warnw("failed to get index patterns matching general pattern...",
			"pattern", pattern.string, "error", err.Error())
		return nil
	}
	return Patterns
}

func (a *RefreshIndexPattern) updateIndexPattern(ctx context.Context, indexPattern kibana.IndexPattern) error {

	indexPatternFields, err := a.kibana.IndexPatternFields(ctx, indexPattern.Title)
	if err != nil {
		return fmt.Errorf("failed to get Index Pattern Fields. Index Pattern:[%s]", indexPattern.Title)
	}
	indexPattern.IndexPatternFields = *indexPatternFields

	err = a.kibana.PutIndexPattern(ctx, indexPattern)
	if err != nil {
		return fmt.Errorf("failed to update Index Pattern. Index Pattern:[%s]", indexPattern.Title)
	}

	return nil
}
