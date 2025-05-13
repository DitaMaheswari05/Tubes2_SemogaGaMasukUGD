package recipeFinder

import (
	"context"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

/*
Reverse-index (read-only)

This index maps product IDs to the ingredient pairs that can create them.
It enables efficient lookup of "what ingredients make this product?" and
is essential for target-to-base DFS traversal.
*/
type pair struct{ a, b int } // Represents an ingredient pair (a,b)
type revIndex map[int][]pair // Maps product ID to all ingredient pairs


// BuildReverseIndex creates a reverse mapping from products to their ingredient pairs.
func BuildReverseIndex(g IndexedGraph) {
	idx := make(revIndex)

	// First pass: collect all ingredient pairs for each product
	for a, nbrs := range g.Edges {
		for _, e := range nbrs {
			// Store pair in normalized order (smaller ID first)
			p := pair{a: min(a, e.PartnerID), b: max(a, e.PartnerID)}
			idx[e.ProductID] = append(idx[e.ProductID], p)
		}
	}

	// Second pass: sort pairs by total tier (complexity) of ingredients
	// This prioritizes simpler ingredients during search
	for _, list := range idx {
		sort.Slice(list, func(i, j int) bool {
			ti := getElementTier(g.IDToName[list[i].a]) +
				getElementTier(g.IDToName[list[i].b])
			tj := getElementTier(g.IDToName[list[j].a]) +
				getElementTier(g.IDToName[list[j].b])
			return ti < tj
		})
	}
	revIdx = idx // Set the global variable
}

/*
-------------------------------------------------------------------------
Single-path DFS (recursive)

This algorithm finds a single path from a target element to base elements using
depth-first search with caching and pruning optimizations.
*/
var canReachBaseCache map[int]bool // Memoization cache: elementID → can reach base?

// findPathToBaseCnt is a recursive DFS function that finds a path from an element to base elements.
// Parameters:
//   - id: Current element ID being processed
//   - depth: Current recursion depth
//   - maxDepth: Maximum recursion depth limit to prevent stack overflow
//   - g: The indexed graph containing all element relationships
//   - recipes: Output map to store the found recipe steps
//   - visit: Map to track visited elements (prevents cycles)
//   - counter: Pointer to count nodes visited (for statistics)
//
// Returns:
//   - bool: True if a path to base elements was found, false otherwise
func findPathToBaseCnt(
	id, depth, maxDepth int,
	g IndexedGraph,
	recipes ProductToIngredients,
	visit map[int]bool,
	counter *int,
) bool {
	// Stop if we've gone too deep (prevents stack overflow)
	if depth > maxDepth {
		return false
	}

	// Check the cache for previous results (memoization)
	if res, ok := canReachBaseCache[id]; ok {
		return res
	}

	// Detect cycles in the current path
	if visit[id] {
		canReachBaseCache[id] = false
		return false
	}

	// Mark as visited temporarily for this path
	visit[id] = true
	defer func() { visit[id] = false }() // Clean up before returning
	*counter++                           // Count this node as visited

	// Check if current element is a base element (success case)
	name := g.IDToName[id]
	for _, b := range BaseElements {
		if name == b {
			canReachBaseCache[id] = true
			return true
		}
	}

	// Find all ingredient pairs that can make this element
	ing := findIngredientsFor(id, g)

	// Sort ingredients by total tier (simpler ingredients first)
	sort.Slice(ing, func(i, j int) bool {
		ti := getElementTier(g.IDToName[ing[i][0]]) + getElementTier(g.IDToName[ing[i][1]])
		tj := getElementTier(g.IDToName[ing[j][0]]) + getElementTier(g.IDToName[ing[j][1]])
		return ti < tj
	})

	// Recursively try each ingredient pair
	for _, pr := range ing {
		a, b := pr[0], pr[1]

		// Try to find paths from both ingredients to base elements
		if findPathToBaseCnt(a, depth+1, maxDepth, g, recipes, visit, counter) &&
			findPathToBaseCnt(b, depth+1, maxDepth, g, recipes, visit, counter) {

			// Record the successful recipe step
			recipes[name] = RecipeStep{
				Combo: IngredientCombo{
					A: g.IDToName[a],
					B: g.IDToName[b],
				},
			}
			canReachBaseCache[id] = true
			return true
		}
	}

	// No valid path found
	canReachBaseCache[id] = false
	return false
}

// DFSBuildTargetToBase performs a target-to-base DFS search to find a single valid recipe path.
// It starts from the target element and works backward to find constituent ingredients
// until reaching base elements.
// Parameters:
//   - target: Name of the target element to find a recipe for
//   - g: The indexed graph containing all element relationships
//
// Returns:
//   - ProductToIngredients: Map of products to their ingredient recipes
//   - int: Count of nodes visited during the search
func DFSBuildTargetToBase(target string, g IndexedGraph) (ProductToIngredients, int) {
	targetID := g.NameToID[target]
	recipes := make(ProductToIngredients)
	visited := make(map[int]bool)

	// Initialize cache with base elements (they can reach themselves)
	canReachBaseCache = map[int]bool{}
	for _, b := range BaseElements {
		canReachBaseCache[g.NameToID[b]] = true
	}

	// Start DFS with nodes counter
	nodes := 0

	// First try with reasonable depth limit
	if !findPathToBaseCnt(targetID, 0, 1000, g, recipes, visited, &nodes) {
		// If that fails, try again with much higher limit
		visited = map[int]bool{}
		findPathToBaseCnt(targetID, 0, 10000, g, recipes, visited, &nodes)
	}

	return recipes, nodes
}

/*
-------------------------------------------------------------------------
Multi-path DFS (iterative/parallel)
*/
// RangeDFSPaths finds up to maxPaths unique DFS recipes, exploring each top-level
//
// For each reverse-combination (root pair) that produces the target, a separate
// goroutine is launched to recursively explore all possible ingredient paths back
// to base elements. A bounded worker pool (limited to NumCPU()) ensures controlled
// parallelism.
//
// Each path is deduplicated using a hash signature to guarantee uniqueness.
// Once maxPaths unique results are found, all active searches are cancelled early.
func RangeDFSPaths(target string, maxPaths int, g IndexedGraph) ([]RecipeStep, int) {
	targetID := g.NameToID[target]
	roots := revIdx[targetID]

	var (
		out     []RecipeStep
		seenSig = make(map[uint64]struct{})
		mu      sync.Mutex
		nodes   int64
	)

	// bound concurrent workers
	sem := make(chan struct{}, runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// recursive DFS
	var dfs func(id int, path [][]int, visited map[int]bool)
	dfs = func(id int, path [][]int, visited map[int]bool) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		atomic.AddInt64(&nodes, 1)

		if isBaseID(id, g) {
			sig := hashPath(path)
			mu.Lock()
			if len(out) < maxPaths {
				if _, dup := seenSig[sig]; !dup {
					seenSig[sig] = struct{}{}
					out = append(out, buildRecipeStepFromPath(path, targetID, g))
					if len(out) == maxPaths {
						cancel()
					}
				}
			}
			mu.Unlock()
			return
		}

		if visited[id] {
			return
		}
		visited[id] = true
		defer func() { visited[id] = false }()

		for _, pr := range revIdx[id] {
			newPath := append(path, []int{pr.a, pr.b, id})
			dfs(pr.a, newPath, visited)
			dfs(pr.b, newPath, visited)
		}
	}

	// spawn one worker per root pair
	for _, pr := range roots {
		sem <- struct{}{} // acquire slot
		wg.Add(1)         // one Add per goroutine
		go func(pr pair) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			visited := make(map[int]bool)
			initial := [][]int{{pr.a, pr.b, targetID}}
			dfs(pr.a, initial, visited)
			dfs(pr.b, initial, visited)
		}(pr)
	}

	wg.Wait()
	return out, int(atomic.LoadInt64(&nodes))
}

// isBaseID returns true if id corresponds to one of the base elements.
func isBaseID(id int, g IndexedGraph) bool {
	for _, b := range BaseElements {
		if id == g.NameToID[b] {
			return true
		}
	}
	return false
}
