package recipeFinder

import (
	"context"
	"hash/fnv"
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

// hashPath generates a hash signature for a specific recipe path.
// This is used to deduplicate paths that are functionally equivalent.
// Parameters:
//   - p: Slice of [ingredientA, ingredientB, product] triples representing a path
//
// Returns:
//   - uint64: Hash value representing the path signature
func hashPath(p [][]int) uint64 {
	h := fnv.New64a()
	var buf [4]byte
	put := func(v int) {
		buf[0] = byte(v)
		buf[1] = byte(v >> 8)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 24)
		_, _ = h.Write(buf[:])
	}
	for _, t := range p {
		put(t[0])
		put(t[1])
		put(t[2])
	}
	return h.Sum64()
}

/*
-------------------------------------------------------------------------
Multi-path DFS (iterative/parallel)
*/
// RangeDFSPaths finds up to maxPaths unique DFS recipes, exploring each top-level
// ingredient-pair for `target` in parallel and cancelling early once we hit the limit.
// RangeDFSPaths runs a concurrent, stack‐based DFS across root pairs.
// RangeDFSPaths runs a concurrent, stack‐based DFS across root pairs,
// but never more than NumCPU() workers in flight at a time.
// Uses an iterative, stack-based DFS approach to avoid recursion stack overflow.
func RangeDFSPaths(target string, maxPaths int, g IndexedGraph) ([]RecipeStep, int) {
	// Stack element for iterative DFS
	type elem struct{ id, childPos int }
	targetID := g.NameToID[target]
	roots := revIdx[targetID]

	var (
		out     []RecipeStep
		seenSig = make(map[uint64]struct{})
		mu      sync.Mutex
		nodes   int64
	)

	internalMaxPaths := maxPaths * 4

	// Create a bounded semaphore channel of size = number of CPUs
	maxWorkers := runtime.NumCPU()
	sem := make(chan struct{}, maxWorkers)

	// Context to cancel all workers when we hit maxPaths
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(roots))

	// Worker launches one stack-based DFS for a single root pair
	worker := func(pr pair) {
		defer wg.Done()
		// Release our "token" when this goroutine exits
		defer func() { <-sem }()

		// Each goroutine has its own stack, path, visited
		stack := []elem{{id: targetID}}
		path := make([][]int, 0, 64)
		visited := make(map[int]bool)
		revisitCounts := make(map[int]int) // Track element revisits
		const maxRevisits = 2              // Allow revisiting non-base elements up to 2 times

		// Seed the initial step
		path = append(path, []int{pr.a, pr.b, targetID})
		stack = append(stack, elem{id: pr.b}, elem{id: pr.a})

		for len(stack) > 0 {
			// Early cancel
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Pop
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			id := f.id
			atomic.AddInt64(&nodes, 1)

			// Base check
			isBase := false
			for _, b := range BaseElements {
				if id == g.NameToID[b] {
					isBase = true
					break
				}
			}
			if isBase {
				// We've reached a base element, potentially completing a path
				sig := hashPath(path)
				mu.Lock()
				if len(out) < internalMaxPaths {
					if _, seen := seenSig[sig]; !seen {
						seenSig[sig] = struct{}{}
						out = append(out, buildRecipeStepFromPath(path, targetID, g))
						if len(out) == internalMaxPaths {
							cancel()
						}
					}
				}
				mu.Unlock()
				// backtrack path (pop this step from path)
				if len(path) > 0 {
					path = path[:len(path)-1]
				}
				continue
			}

			// Modified cycle guard - allow some revisits for diversity
			if visited[id] {
				// Base elements should never be revisited
				isBaseElement := false
				for _, b := range BaseElements {
					if id == g.NameToID[b] {
						isBaseElement = true
						break
					}
				}

				// Never revisit base elements or elements we've revisited too much
				if isBaseElement || revisitCounts[id] >= maxRevisits {
					if len(path) > 0 {
						path = path[:len(path)-1]
					}
					continue
				}

				// Allow revisit but count it
				revisitCounts[id]++
			}
			visited[id] = true

			children := revIdx[id]
			if f.childPos >= len(children) {
				visited[id] = false // no longer in path
				if len(path) > 0 {
					path = path[:len(path)-1] // pop from path
				}
				continue
			}

			// Process next child
			pr2 := children[f.childPos]
			f.childPos++
			stack = append(stack, f)

			// Extend path, push b then a
			path = append(path, []int{pr2.a, pr2.b, id})
			stack = append(stack, elem{id: pr2.b}, elem{id: pr2.a})
		}
	}

	// 3) Launch each worker, but block if we've hit maxWorkers in flight
	for _, pr := range roots {
		sem <- struct{}{} // acquire a "slot" (blocks if full)
		go worker(pr)
	}
	wg.Wait()

	if len(out) > maxPaths {
		out = deduplicateRecipesDFS(out)

		// Limit to maxPaths
		if len(out) > maxPaths {
			out = out[:maxPaths]
		}
	}

	return out, int(atomic.LoadInt64(&nodes))
}

// Update the deduplication function as well
func deduplicateRecipesDFS(recipes []RecipeStep) []RecipeStep {
	if len(recipes) <= 1 {
		return recipes
	}

	result := []RecipeStep{recipes[0]}

	for i := 1; i < len(recipes); i++ {
		unique := true

		// Compare with all previously accepted recipes
		for j := 0; j < len(result); j++ {
			// Compare based on ingredients used
			// Lower threshold to keep more diverse paths
			if recipePathsAreSimilar(recipes[i], result[j]) > 0.75 { // Lower from 0.85 to 0.75
				unique = false
				break
			}
		}

		if unique {
			result = append(result, recipes[i])
		}
	}

	return result
}

func recipePathsAreSimilar(r1, r2 RecipeStep) float64 {
	// Extract all elements from both recipes
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	// Add elements from first recipe
	set1[r1.Combo.A] = true
	set1[r1.Combo.B] = true

	// Add elements from second recipe
	set2[r2.Combo.A] = true
	set2[r2.Combo.B] = true

	// Count common elements
	common := 0
	for elem := range set1 {
		if set2[elem] {
			common++
		}
	}

	// Calculate Jaccard similarity
	totalUnique := len(set1) + len(set2) - common
	if totalUnique == 0 {
		return 1.0
	}
	return float64(common) / float64(totalUnique)
}
