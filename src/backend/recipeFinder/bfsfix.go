package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

func IndexedBFSBuild(targetName string, graph IndexedGraph) map[string]Info {
    targetID := graph.NameToID[targetName]
    
    queue := list.New()
    seen := make(map[int]bool)
    
    for _, baseName := range baseElements {
        baseID := graph.NameToID[baseName]
        queue.PushBack(baseID)
        seen[baseID] = true
    }
    
    // Track parents using integer IDs
    prevIDs := make(map[int]struct{ParentID, PartnerID int})
    
    for queue.Len() > 0 {
        curID := queue.Remove(queue.Front()).(int)
        
        if curID == targetID {
            break
        }
        
        for _, neighbor := range graph.Edges[curID] {
            partnerID := neighbor.PartnerID
            productID := neighbor.ProductID
            
            if seen[partnerID] && !seen[productID] {
                seen[productID] = true
                prevIDs[productID] = struct{ParentID, PartnerID int}{
                    ParentID:  curID,
                    PartnerID: partnerID,
                }
                queue.PushBack(productID)
            }
        }
    }
    
    // Convert integer results back to strings
    prev := make(map[string]Info)
    for productID, info := range prevIDs {
        productName := graph.IDToName[productID]
        parentName := graph.IDToName[info.ParentID]
        partnerName := graph.IDToName[info.PartnerID]
        
        prev[productName] = Info{
            Parent:  parentName,
            Partner: partnerName,
        }
    }
    
    return prev
}

// func BFSBuild2(target string, recipeGraph Graph) map[string]Info {
// 	queue := list.New()
// 	for _, base := range baseElements {
// 		queue.PushBack(base)
// 	}

// 	seen := make(map[string]bool)
// 	for _, base := range baseElements {
// 		seen[base] = true
// 	}

// 	prev := make(map[string]Info)

// 	for queue.Len() > 0 {
// 		cur := queue.Remove(queue.Front()).(string)

// 		if cur == target {
// 			return prev
// 		}

// 		neighbors := recipeGraph[cur]
// 		for _, neighbor := range neighbors {
// 			partner := neighbor.Partner
// 			prod := neighbor.Product

// 			if seen[partner] && !seen[prod] {
// 				seen[prod] = true
// 				prev[prod] = Info{Parent: cur, Partner: partner}
// 				queue.PushBack(prod)
// 			}
// 		}
// 	}

// 	// target not found
// 	return prev
// }

func BFSBuildMulti(target string, recipeGraph Graph, maxPaths int64) map[string][]Info {
	var wg sync.WaitGroup
	var pathCount int64

	// Protects seenElements
	seenElements := make(map[string]bool)
	var muElem sync.RWMutex

	// Protects seenCombinations & pathPrev
	seenCombinations := make(map[string]bool) // "parent,partner,product"
	pathPrev := make(map[string][]Info)       // records every discovered combo
	var muCombo sync.Mutex

	// Debug: track activation sequence
	var activationCount int64
	const maxDebugActivations = 100

	// Seed base elements
	muElem.Lock()
	for _, b := range baseElements {
		seenElements[b] = true
	}
	muElem.Unlock()

	bfsWorker := func(start string) {
		defer wg.Done()

		// Each queue entry is the steps to current product.
		queue := list.New()
		queue.PushBack([][]string{})

		for queue.Len() > 0 {
			// Stop if we've hit the target cap
			if maxPaths > 0 && atomic.LoadInt64(&pathCount) >= maxPaths {
				return
			}

			currentSteps := queue.Remove(queue.Front()).([][]string)

			// Determine current node
			var cur string
			if len(currentSteps) == 0 {
				cur = start
			} else {
				cur = currentSteps[len(currentSteps)-1][2]
			}

			// Explore all combinations from cur
			for _, edge := range recipeGraph[cur] {
				partner := edge.Partner
				prod := edge.Product

				// Only combine if partner is available
				muElem.RLock()
				partnerSeen := seenElements[partner]
				muElem.RUnlock()
				if !partnerSeen {
					continue
				}

				// Normalize the two inputs, keep prod separate
				inputs := []string{cur, partner}
				sort.Strings(inputs)
				parent, partnerNorm := inputs[0], inputs[1]
				product := prod

				// Build the full path up to this combination
				newPath := append(append([][]string{}, currentSteps...), []string{parent, partnerNorm, product})
				comboKey := fmt.Sprintf("%s,%s,%s", parent, partnerNorm, product)

				// Dedup and record this combination
				muCombo.Lock()
				already := seenCombinations[comboKey]
				if !already {
					seenCombinations[comboKey] = true
					pathPrev[product] = append(pathPrev[product], Info{
						Parent:  parent,
						Partner: partnerNorm,
						Path:    newPath,
					})
				}
				muCombo.Unlock()

				if already {
					continue
				}

				// If this produced the target, count and possibly stop
				if product == target {
					newCount := atomic.AddInt64(&pathCount, 1)
					fmt.Printf("Found path for %s: %v\n", target, newPath)
					if maxPaths > 0 && newCount >= maxPaths {
						return
					}
				} else {
					// Otherwise enqueue for further exploration
					muElem.Lock()
					seen := seenElements[product]
					if !seen {
						seenElements[product] = true
						queue.PushBack(newPath)

						// Log activation for first few nodes
						count := atomic.AddInt64(&activationCount, 1)
						if count <= maxDebugActivations {
							fmt.Printf("[Activation #%d] %s discovered by combining %s + %s\n", count, product, parent, partnerNorm)
						}
					}
					muElem.Unlock()
				}
			}
		}
	}

	// Launch one worker per base element: "Air", "Earth", "Fire", "Water"
	for _, start := range baseElements {
		wg.Add(1)
		go bfsWorker(start)
	}

	wg.Wait()
	return pathPrev
}
