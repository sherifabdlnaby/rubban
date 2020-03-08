package refreshindexpattern

import (
	"context"
	"fmt"
	"sync"

	"github.com/sherifabdlnaby/gpool"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
	"go.uber.org/atomic"
)

//Run Run Auto Index Pattern creation task
func (a *RefreshIndexPattern) Run(ctx context.Context) {

	// Send Requests Concurrently
	idxPatternPool := gpool.NewPool(a.concurrency)
	idxPatternChan := make(chan []kibana.IndexPattern, a.concurrency)

	wg := sync.WaitGroup{}

	// 1 - Get Index Patterns Matching The Give Patterns
	go func() {
		for _, pattern := range a.Patterns {
			shadowPattern := pattern

			wg.Add(1)
			err := idxPatternPool.Enqueue(ctx, func() {
				defer wg.Done()
				indexPatterns, err := a.getIndexPattern(ctx, shadowPattern)
				if err != nil {
					a.log.Warnw("Failed to get index pattern...",
						"pattern", shadowPattern, "error", err.Error())
					return
				}

				if len(indexPatterns) > 0 {
					idxPatternChan <- indexPatterns
				}
			})

			if err != nil {
				wg.Done()
			}
		}

		// Wait for all above jobs to Return and Close the Channel
		wg.Wait()
		close(idxPatternChan)
	}()

	// 2- Update Found Index Patterns
	count := atomic.Int32{}
	for patterns := range idxPatternChan {
		shadowedPatterns := patterns
		_ = idxPatternPool.Enqueue(ctx, func() {
			err := a.kibana.BulkCreateIndexPattern(ctx, shadowedPatterns)
			if err != nil {
				a.log.Warnw("Failed to update index patterns", "error", err.Error(), "patterns", shadowedPatterns)
			}
			count.Add(int32(len(shadowedPatterns)))
		})
	}
	idxPatternPool.Stop()
	a.log.Info(fmt.Sprintf("Finished Updating Index Pattern(s) Fields, Updated (%d) Index Pattern.", count.Load()))
}

//Name Return Task Name
func (a *RefreshIndexPattern) Name() string {
	return a.name
}
