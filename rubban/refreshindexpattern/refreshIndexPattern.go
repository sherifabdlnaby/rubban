package refreshindexpattern

import (
	"context"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//RefreshIndexPattern hold attributes for a RunAutoIndexPattern loaded from config.
type RefreshIndexPattern struct {
	name        string
	concurrency int
	Patterns    []string
	kibana      kibana.API
	log         log.Logger
}

//NewRefreshIndexPattern Constructor
func NewRefreshIndexPattern(config config.RefreshIndexPattern, kibana kibana.API, log log.Logger) *RefreshIndexPattern {
	return &RefreshIndexPattern{
		name:        "Refresh Indices Patterns",
		concurrency: config.Concurrency,
		Patterns:    config.Patterns,
		kibana:      kibana,
		log:         log,
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
