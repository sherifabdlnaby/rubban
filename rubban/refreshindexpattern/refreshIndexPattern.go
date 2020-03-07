package refreshindexpattern

import (
	"context"
	"fmt"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//RefreshIndexPattern hold attributes for a RunAutoIndexPattern loaded from config.
type RefreshIndexPattern struct {
	name     string
	Patterns []string
	kibana   kibana.API
	log      log.Logger
}

//NewRefreshIndexPattern Constructor
func NewRefreshIndexPattern(config config.RefreshIndexPattern, kibana kibana.API, log log.Logger) *RefreshIndexPattern {
	return &RefreshIndexPattern{
		name:     "Refresh Indices Patterns",
		Patterns: config.Patterns,
		kibana:   kibana,
		log:      log,
	}
}

func (a *RefreshIndexPattern) getIndexPattern(ctx context.Context, pattern string) ([]kibana.IndexPattern, error) {
	// Get Current IndexPattern Matching Given General Patterns
	Patterns, err := a.kibana.IndexPatterns(ctx, pattern, []string{"version"})
	if err != nil {
		return nil, err
	}
	return Patterns, nil
}

func (a *RefreshIndexPattern) updateIndexPattern(ctx context.Context, indexPattern kibana.IndexPattern) error {

	fields, err := a.getIndexPatternFields(ctx, indexPattern.Title)
	if err != nil {
		return err
	}

	indexPattern.IndexPatternFields = *fields

	err = a.kibana.PutIndexPattern(ctx, indexPattern)
	if err != nil {
		return fmt.Errorf("failed to update Index Pattern. Index Pattern:[%s]", indexPattern.Title)
	}

	return nil
}

func (a *RefreshIndexPattern) getIndexPatternFields(ctx context.Context, indexPatternTitle string) (*kibana.IndexPatternFields, error) {
	return a.kibana.IndexPatternFields(ctx, indexPatternTitle)
}
