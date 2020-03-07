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

	concurrency := 10

	// Send Requests Concurrently // TODO Make concurrency configurable
	idxPatternPool := gpool.NewPool(concurrency)
	idxPatternChan := make(chan []kibana.IndexPattern, concurrency)

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
	wg2 := sync.WaitGroup{}
	count := atomic.Int32{}

	for patterns := range idxPatternChan {
		for _, pattern := range patterns {
			shadowedPattern := pattern

			wg2.Add(1)

			err := idxPatternPool.Enqueue(ctx, func() {
				defer wg2.Done()
				err := a.updateIndexPattern(ctx, shadowedPattern)
				if err != nil {
					a.log.Warnw(fmt.Sprintf("Failed to update index pattern [%s]", shadowedPattern.Title), "error", err.Error())
				}
				a.log.Info(fmt.Sprintf("Updated index pattern [%s] fields", shadowedPattern.Title))
				count.Add(1)
			})

			if err != nil {
				wg2.Done()
			}

		}
	}

	wg2.Wait()
	idxPatternPool.Stop()
	a.log.Info(fmt.Sprintf("Finished Updating Index Pattern(s) Fields, Updated (%d) Index Pattern.", count.Load()))
}

//Name Return Task Name
func (a *RefreshIndexPattern) Name() string {
	return a.name
}
