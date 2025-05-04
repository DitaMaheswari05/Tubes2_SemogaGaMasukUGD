package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
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

func BFSBuildMulti(target string, byPair map[Pair][]string, maxPaths int) map[string][]Info {
	graph := BuildGraph(byPair)

	resultChan := make(chan Info, 10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var pathCount int
	seenElements := make(map[string]bool)
	for _, base := range baseElements {
		seenElements[base] = true
	}

	bfsWorker := func(start string) {
		defer wg.Done()

		queue := list.New()
		queue.PushBack([][]string{{start}})

		seenPaths := make(map[string]bool)
		seenPaths[start] = true
		pathPrev := make(map[string][]Info)

		for queue.Len() > 0 {
			mu.Lock()
			if pathCount >= maxPaths && maxPaths > 0 {
				mu.Unlock()
				break
			}
			mu.Unlock()

			currentSteps := queue.Remove(queue.Front()).([][]string)
			var cur string
			if len(currentSteps) > 0 {
				cur = currentSteps[len(currentSteps)-1][2]
			} else {
				cur = start
			}

			neighbors := graph[cur]
			for _, neighbor := range neighbors {
				partner := neighbor.Partner
				prod := neighbor.Product

				mu.Lock()
				seenPartner := seenElements[partner]
				mu.Unlock()

				if seenPartner {
					newStep := []string{cur, partner, prod}
					sort.Strings(newStep)
					newSteps := append(append([][]string{}, currentSteps...), newStep)

					pathKey := fmt.Sprintf("%v", newSteps)

					if !seenPaths[pathKey] {
						seenPaths[pathKey] = true
						mu.Lock()
						seenElements[prod] = true
						mu.Unlock()
						pathPrev[prod] = append(pathPrev[prod], Info{Parent: cur, Partner: partner, Path: newSteps})

						queue.PushBack(newSteps)
						if prod == target {
							mu.Lock()
							if pathCount < maxPaths || maxPaths <= 0 {
								pathCount++
								fmt.Printf("Found path for %s: %v\n", target, newSteps)
							}
							mu.Unlock()
						}
					}
				}
			}
		}

		if infos, ok := pathPrev[target]; ok {
			for _, info := range infos {
				mu.Lock()
				if pathCount < maxPaths || maxPaths <= 0 {
					resultChan <- info
				}
				mu.Unlock()
			}
		}
	}

	for _, start := range baseElements {
		wg.Add(1)
		go bfsWorker(start)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	resultMap := make(map[string][]Info)
	for result := range resultChan {
		resultMap[target] = append(resultMap[target], result)
	}

	fmt.Printf("Final recipeMap for %s: %+v\n", target, resultMap)
	return resultMap
}
