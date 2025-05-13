package recipeFinder

import (
	"container/list"
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

type SearchStep struct {
	CurrentID       int                                       `json:"current_id"`
	CurrentName     string                                    `json:"current"`
	QueueIDs        []int                                     `json:"queue_ids"`
	QueueNames      []string                                  `json:"queue"`
	SeenIDs         []int                                     `json:"seen_ids"`
	SeenNames       []string                                  `json:"seen"`
	DiscoveredEdges map[int]struct{ ParentID, PartnerID int } `json:"discovered_ids"`
	DiscoveredNames map[string]struct{ A, B string }          `json:"discovered"`
	StepNumber      int                                       `json:"step"`
	FoundTarget     bool                                      `json:"found_target"`
}

/*
-------------------------------------------------------------------------
Single-recipe BFS
*/
func IndexedBFSBuild(targetName string, graph IndexedGraph) (ProductToIngredients, []SearchStep, int) {
	targetID := graph.NameToID[targetName]

	queue := list.New()
	seen := make(map[int]bool)

	searchSteps := []SearchStep{}

	for _, baseName := range BaseElements {
		baseID := graph.NameToID[baseName]
		queue.PushBack(baseID)
		seen[baseID] = true
	}

	searchSteps = append(searchSteps, SearchStep{
		CurrentID:       -1,
		CurrentName:     "",
		QueueIDs:        queueToSlice(queue),
		QueueNames:      queueToNameSlice(queue, graph),
		SeenIDs:         mapKeysToSlice(seen),
		SeenNames:       mapKeysToNameSlice(seen, graph),
		DiscoveredEdges: make(map[int]struct{ ParentID, PartnerID int }),
		DiscoveredNames: make(map[string]struct{ A, B string }),
		StepNumber:      0,
		FoundTarget:     false,
	})

	// Track parents using integer IDs
	prevIDs := make(map[int]struct{ ParentID, PartnerID int })
	nodes := 0
	for queue.Len() > 0 {
		curID := queue.Remove(queue.Front()).(int)
		nodes++

		searchSteps = append(searchSteps, SearchStep{
			CurrentID:       curID,
			CurrentName:     graph.IDToName[curID],
			QueueIDs:        queueToSlice(queue),
			QueueNames:      queueToNameSlice(queue, graph),
			SeenIDs:         mapKeysToSlice(seen),
			SeenNames:       mapKeysToNameSlice(seen, graph),
			DiscoveredEdges: copyMap(prevIDs), // Need a deep copy
			DiscoveredNames: prevIDsToNames(prevIDs, graph),
			StepNumber:      nodes,
			FoundTarget:     curID == targetID,
		})

		if curID == targetID {
			break
		}

		for _, neighbor := range graph.Edges[curID] {
			partnerID := neighbor.PartnerID
			productID := neighbor.ProductID
		
			if seen[partnerID] && !seen[productID] {
				// We found a new product - record path
				seen[productID] = true
				prevIDs[productID] = struct{ ParentID, PartnerID int }{
					ParentID:  curID,
					PartnerID: partnerID,
				}
				
				// If this is the target, we can stop immediately
				if productID == targetID {
					// Create a copy of prevIDs that includes the target we just found
					finalDiscoveredEdges := copyMap(prevIDs)
					finalDiscoveredEdges[productID] = struct{ ParentID, PartnerID int }{
						ParentID:  curID,
						PartnerID: partnerID,
					}
					
					// Convert to names
					finalDiscoveredNames := prevIDsToNames(finalDiscoveredEdges, graph)
					
					// Create final search step with target found
					searchSteps = append(searchSteps, SearchStep{
						CurrentID:       productID, // Use target ID as current
						CurrentName:     graph.IDToName[productID], // Show target as current element
						QueueIDs:        queueToSlice(queue),
						QueueNames:      queueToNameSlice(queue, graph),
						SeenIDs:         mapKeysToSlice(seen),
						SeenNames:       mapKeysToNameSlice(seen, graph),
						DiscoveredEdges: finalDiscoveredEdges, // Include target
						DiscoveredNames: finalDiscoveredNames, // Include target
						StepNumber:      nodes + 1,
						FoundTarget:     true,
					})
					
					// Break out of both loops
					goto TargetFound
				}
				
				queue.PushBack(productID)
			}
		}
	}

	TargetFound:
	// Convert integer results to ProductToIngredients
	recipes := make(ProductToIngredients)
	for productID, info := range prevIDs {
		productName := graph.IDToName[productID]
		parentName := graph.IDToName[info.ParentID]
		partnerName := graph.IDToName[info.PartnerID]

		recipes[productName] = RecipeStep{
			Combo: IngredientCombo{
				A: parentName,
				B: partnerName,
			},
		}
	}

	return recipes, searchSteps, nodes
}

/*
-------------------------------------------------------------------------
Path (skip+1) (distinct)
*/
// findKthPathIndexed finds the (skip+1)-th distinct path to targetID using a level-based parallel BFS,
// with deterministic ordering: sorted neighbors and stable bounding.
func findKthPathIndexed(targetID, skip int, g IndexedGraph) (RecipeStep, int) {
	type state struct {
		elem  int
		path  [][]int
		depth int
	}

	// pre-sort neighbor edges for deterministic order
	for u := range g.Edges {
		sort.Slice(g.Edges[u], func(i, j int) bool {
			a, b := g.Edges[u][i], g.Edges[u][j]
			if a.PartnerID != b.PartnerID {
				return a.PartnerID < b.PartnerID
			}
			return a.ProductID < b.ProductID
		})
	}

	// cache for seen paths (thread-safe)
	const cacheCapacity = 10000
	type entry struct{ key string }
	cacheList := make([]entry, 0, cacheCapacity)
	cacheMap := make(map[string]struct{})
	var cacheMu sync.Mutex
	evict := func() {
		if len(cacheList) > cacheCapacity {
			old := cacheList[0]
			cacheList = cacheList[1:]
			delete(cacheMap, old.key)
		}
	}
	checkAndAdd := func(sig string) bool {
		cacheMu.Lock()
		defer cacheMu.Unlock()
		if _, ok := cacheMap[sig]; ok {
			return false
		}
		cacheMap[sig] = struct{}{}
		cacheList = append(cacheList, entry{key: sig})
		evict()
		return true
	}

	// Track product revisits for diversity
	productVisits := make(map[int]int)
	const maxProductVisits = 5 // Allow more diverse paths for each product

	// Cache product-to-recipe mappings to ensure diversity
	productRecipes := make(map[int]map[string]bool)

	// helper to copy and extend a path
	appendCopyPath := func(cur [][]int, a, b, c int) [][]int {
		np := make([][]int, len(cur)+1)
		copy(np, cur)
		np[len(cur)] = []int{min(a, b), max(a, b), c}
		return np
	}

	// initialization
	reachable := make(map[int]bool)
	var currLevel []state
	for _, b := range BaseElements {
		id := g.NameToID[b]
		reachable[id] = true
		currLevel = append(currLevel, state{elem: id, depth: 0})
	}

	nodes := 0
	hits := 0
	const maxDepth = 30
	maxWorkers := runtime.NumCPU()
	const maxLevelSize = 10000

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout
	defer cancel()

	for depth := 0; depth <= maxDepth && len(currLevel) > 0; depth++ {
		select {
		case <-ctx.Done():
			return RecipeStep{}, nodes
		default:
		}
		var nextLevel []state
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, maxWorkers)

		for _, st := range currLevel {
			nodes++
			if st.elem == targetID {
				hits++
				if hits-1 == skip {
					return buildRecipeStepFromPath(st.path, targetID, g), nodes
				}
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(st state) {
				defer wg.Done()
				defer func() { <-sem }()
				for _, r := range g.Edges[st.elem] {
					select {
					case <-ctx.Done():
						return
					default:
					}

					partnerID := r.PartnerID
					productID := r.ProductID

					mu.Lock()
					reached := reachable[partnerID]
					revisitCount := productVisits[productID]

					// Initialize product recipe tracking if needed
					if productRecipes[productID] == nil {
						productRecipes[productID] = make(map[string]bool)
					}

					// Generate recipe signature
					recipeKey := fmt.Sprintf("%d+%d", min(st.elem, partnerID), max(st.elem, partnerID))
					seenRecipe := productRecipes[productID][recipeKey]
					mu.Unlock()

					// Skip if partner is not reachable
					if !reached {
						continue
					}

					// Allow exploring even if product was seen before, under conditions:
					// 1. If we haven't exceeded max visits for this product
					// 2. If we haven't seen this specific recipe for this product
					shouldExplore := !reachable[productID] ||
						(!seenRecipe && revisitCount < maxProductVisits)

					if !shouldExplore {
						continue
					}

					// Create extended path
					np := appendCopyPath(st.path, st.elem, partnerID, productID)

					// Check path canonicalization to avoid duplicates
					sig := canonicalHash(np)
					if !checkAndAdd(sig) {
						continue
					}

					mu.Lock()
					// Track this recipe for this product
					productRecipes[productID][recipeKey] = true

					// If product was already reachable, increment visit count
					if reachable[productID] {
						productVisits[productID]++
					}

					// Add to next level
					nextLevel = append(nextLevel, state{elem: productID, path: np, depth: st.depth + 1})
					mu.Unlock()
				}
			}(st)
		}
		wg.Wait()

		// sort nextLevel for stability before bounding
		sort.Slice(nextLevel, func(i, j int) bool {
			if nextLevel[i].elem != nextLevel[j].elem {
				return nextLevel[i].elem < nextLevel[j].elem
			}
			return len(nextLevel[i].path) < len(nextLevel[j].path)
		})

		if len(nextLevel) > maxLevelSize {
			nextLevel = nextLevel[:maxLevelSize]
		}

		// update reachable
		for _, st := range nextLevel {
			reachable[st.elem] = true
		}
		currLevel = nextLevel
	}
	return RecipeStep{}, nodes
}
