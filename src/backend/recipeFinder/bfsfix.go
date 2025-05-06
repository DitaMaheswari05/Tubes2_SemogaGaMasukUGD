package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
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

// IndexedBFSBuildMulti finds up to maxPaths unique complete paths (recipes) 
// to the target element by running the BFS repeatedly. Instead of using usedCombos
// to prune immediately during the search, we record the full recipe signature
// and check for duplicates after each BFS run.
func IndexedBFSBuildMulti(targetName string, graph IndexedGraph, maxPaths int64) map[string][]Info {
    targetID, exists := graph.NameToID[targetName]
    if !exists {
        return make(map[string][]Info)
    }
    
    result := make(map[string][]Info)
    // usedCombos is used here only to mark combinations if desired; in our approach
    // we will generate a full signature and compare.
    usedCombos := make(map[string]bool)
    
    // Run BFS up to maxPaths times to find unique complete paths.
    for i := int64(0); i < maxPaths; i++ {
        // Run one BFS iteration to attempt to get a complete recipe path.
        path := findUniquePathIndexed(targetID, targetName, graph, usedCombos)
        if path.Path == nil {
            // No complete path was found.
            break
        }
        
        // Generate a signature for the complete recipe path.
        sig := generatePathSignature(path)
        // If we already have this complete recipe, repeat the iteration.
        if alreadyFound(sig, result[targetName]) {
            i--
            continue
        }
        
        // Add this unique path to the result.
        result[targetName] = append(result[targetName], path)
        
        // (Optional) Mark all combinations in this path as used.
        for _, step := range path.Path {
            if len(step) == 3 {
                combo := fmt.Sprintf("%s,%s,%s", step[0], step[1], step[2])
                usedCombos[combo] = true
            }
        }
    }
    
    return result
}

// findUniquePathIndexed performs a BFS (using the IndexedGraph) to find a complete
// recipe path from any base element to the target. The usedCombos parameter is not
// used for pruning in this primitive approach.
func findUniquePathIndexed(targetID int, targetName string, graph IndexedGraph, usedCombos map[string]bool) Info {
    queue := list.New()
    // seen tracks visited nodes (by their integer ID) in this BFS run.
    seen := make(map[int]bool)
    
    // Add all base elements into the queue.
    for _, baseName := range baseElements {
        baseID := graph.NameToID[baseName]
        seen[baseID] = true
        queue.PushBack(struct {
            elemID int
            path   [][]int
        }{
            elemID: baseID,
            path:   [][]int{}, // empty path at start
        })
    }
    
    // BFS loop.
    for queue.Len() > 0 {
        curr := queue.Remove(queue.Front()).(struct {
            elemID int
            path   [][]int
        })
        
        // If reached the target, convert the integer path to a string path.
        if curr.elemID == targetID {
            stringPath := make([][]string, len(curr.path))
            for i, step := range curr.path {
                stringPath[i] = []string{
                    graph.IDToName[step[0]],
                    graph.IDToName[step[1]],
                    graph.IDToName[step[2]],
                }
            }
            info := Info{Path: stringPath}
            if len(curr.path) > 0 {
                lastStep := curr.path[len(curr.path)-1]
                info.Parent = graph.IDToName[lastStep[0]]
                info.Partner = graph.IDToName[lastStep[1]]
            }
            return info
        }
        
        // Try all neighbors (possible ingredient combinations) from the current element.
        for _, neighbor := range graph.Edges[curr.elemID] {
            partnerID := neighbor.PartnerID
            productID := neighbor.ProductID
            
            // Skip if partner is not seen or product already seen.
            if !seen[partnerID] || seen[productID] {
                continue
            }
            
            // For consistent deduplication, sort the two ingredient IDs.
            var a, b int
            if curr.elemID < partnerID {
                a, b = curr.elemID, partnerID
            } else {
                a, b = partnerID, curr.elemID
            }
            
            // Build a combo string from the names.
            aName, bName, productName := graph.IDToName[a], graph.IDToName[b], graph.IDToName[productID]
            combo := fmt.Sprintf("%s,%s,%s", aName, bName, productName)
            if usedCombos != nil && usedCombos[combo] {
                continue
            }
            
            // Create a new path that appends the current combination step.
            newPath := make([][]int, len(curr.path)+1)
            copy(newPath, curr.path)
            newPath[len(curr.path)] = []int{a, b, productID}
            
            // Mark the product as visited (to avoid cycles).
            seen[productID] = true
            
            // Push the new state into the BFS queue.
            queue.PushBack(struct {
                elemID int
                path   [][]int
            }{
                elemID: productID,
                path:   newPath,
            })
        }
    }
    
    // No path was found.
    return Info{}
}

// generatePathSignature builds a signature string from an Info object representing
// a complete recipe path. Two complete paths that are identical will generate the same signature.
func generatePathSignature(info Info) string {
    var sb strings.Builder
    for _, step := range info.Path {
        sb.WriteString(strings.Join(step, ","))
        sb.WriteString("|")
    }
    return sb.String()
}

// alreadyFound checks if a signature already exists in the given slice of Info.
func alreadyFound(sig string, infos []Info) bool {
    for _, i := range infos {
        if generatePathSignature(i) == sig {
            return true
        }
    }
    return false
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
