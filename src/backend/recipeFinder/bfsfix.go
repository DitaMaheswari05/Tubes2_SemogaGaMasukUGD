package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

func BFSBuild2(target string, byPair map[Pair][]string) map[string]Info {
	graph := BuildGraph(byPair)

	queue := list.New()
	for _, base := range baseElements {
		queue.PushBack(base)
	}

	seen := make(map[string]bool)
	for _, base := range baseElements {
		seen[base] = true
	}

	prev := make(map[string]Info)

	for queue.Len() > 0 {
		cur := queue.Remove(queue.Front()).(string)

		if cur == target {
			return prev
		}

		neighbors := graph[cur]
		for _, neighbor := range neighbors {
			partner := neighbor.Partner
			prod := neighbor.Product

			if seen[partner] && !seen[prod] {
				seen[prod] = true
				prev[prod] = Info{Parent: cur, Partner: partner}
				queue.PushBack(prod)
			}
		}
	}

	// target not found
	return prev
}

func BFSBuildMulti(target string, byPair map[Pair][]string, maxPaths int64) map[string][]Info {
	graph := BuildGraph(byPair)

	resultChan := make(chan Info, 1000)
	var wg sync.WaitGroup
	var mu sync.RWMutex
	var pathCount int64
	seenElements := make(map[string]bool)
	seenCombinations := make(map[string]bool) // Tracks unique (parent, partner) pairs for the target

	// Initialize seenElements with base elements
	for _, base := range baseElements {
		seenElements[base] = true
	}

	bfsWorker := func(start string) {
		defer wg.Done()

		queue := list.New()
		queue.PushBack(start)

		for queue.Len() > 0 {
			mu.RLock()
			if pathCount >= maxPaths && maxPaths > 0 {
				mu.RUnlock()
				break
			}
			mu.RUnlock()

			currentElement := queue.Remove(queue.Front()).(string)

			neighbors := graph[currentElement]
			for _, neighbor := range neighbors {
				partner := neighbor.Partner
				prod := neighbor.Product

				mu.RLock()
				seenPartner := seenElements[partner]
				mu.RUnlock()

				if seenPartner {
					if prod == target {
						// Normalize the pair to handle commutativity
						pair := []string{currentElement, partner}
						sort.Strings(pair)
						combinationKey := fmt.Sprintf("%s,%s", pair[0], pair[1]) // Target is fixed, so omitted

						mu.Lock()
						if !seenCombinations[combinationKey] {
							seenCombinations[combinationKey] = true
							mu.Unlock()

							// Create Info with normalized Parent/Partner and Path (pair1, pair2, product)
							parent, partner := pair[0], pair[1]
							path := [][]string{{parent, partner, target}}
							info := Info{Parent: parent, Partner: partner, Path: path}

							newCount := atomic.AddInt64(&pathCount, 1)
							if maxPaths <= 0 || newCount <= maxPaths {
								fmt.Printf("Found path for %s: %v\n", target, path)
								resultChan <- info
							}
						} else {
							mu.Unlock()
						}
					} else if !seenElements[prod] {
						// Continue exploration by enqueuing unseen products (except target)
						mu.Lock()
						seenElements[prod] = true
						mu.Unlock()
						queue.PushBack(prod)
					}
				}
			}
		}
	}

	// Start a worker for each base element
	for _, start := range baseElements {
		wg.Add(1)
		go bfsWorker(start)
	}

	// Close resultChan when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	resultMap := make(map[string][]Info)
	for result := range resultChan {
		resultMap[target] = append(resultMap[target], result)
	}

	fmt.Printf("Final recipeMap for %s: %+v\n", target, resultMap)
	return resultMap
}
