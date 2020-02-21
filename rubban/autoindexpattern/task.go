package autoindexpattern

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/gpool"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//Run Run Auto Index Pattern creation task
func (a *AutoIndexPattern) Run(ctx context.Context) {
	//// Set for Found Patterns ( a set datastructes using Map )
	newIndexPatterns := make(map[string]kibana.IndexPattern)

	// Send Requests Concurrently // TODO Make concurrency configurable
	pool := gpool.NewPool(10)
	for _, generalPattern := range a.GeneralPatterns {
		mx := sync.Mutex{}
		_ = pool.Enqueue(ctx, func() {
			indexPatterns := a.getIndexPattern(ctx, generalPattern)

			// Add Result to global Result
			mx.Lock()
			for _, pattern := range indexPatterns {
				newIndexPatterns[pattern.Title] = pattern
			}
			mx.Unlock()
		})
	}

	// Wait for all above jobs to Return
	pool.Stop()

	err := a.kibana.BulkCreateIndexPattern(ctx, newIndexPatterns)
	if err != nil {
		a.log.Errorw("Failed to bulk create new index patterns", "error", err.Error())
	}

	a.log.Infow(fmt.Sprintf("Successfully created %d Index Patterns.", len(newIndexPatterns)), "Index Patterns", newIndexPatterns)

}

//Name Return Task Name
func (a *AutoIndexPattern) Name() string {
	return a.name
}
