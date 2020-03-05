package refreshindexpattern

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/gpool"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

//Run Run Auto Index Pattern creation task
func (a *RefreshIndexPattern) Run(ctx context.Context) {

	// Send Requests Concurrently // TODO Make concurrency configurable
	pool := gpool.NewPool(10)
	for _, pattern := range a.Patterns {
		shadowPattern := pattern
		_ = pool.Enqueue(ctx, func() {
			indexPatterns := a.getIndexPattern(ctx, shadowPattern)

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
func (a *RefreshIndexPattern) Name() string {
	return a.name
}
