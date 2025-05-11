package recipeFinder

import (
	"hash/fnv"
	"sort"
)

/*
-------------------------------------------------------------------------

	Reverse-index (read-only)
	This index maps product IDs to the ingredient pairs that can create them.
	It enables efficient lookup of "what ingredients make this product?" which
	is essential for target→base DFS traversal.
*/
type pair struct{ a, b int } // Represents an ingredient pair (a,b)
type revIndex map[int][]pair // Maps product ID to all ingredient pairs

var revIdx revIndex // Global reverse index: productID → pairs

// BuildReverseIndex creates a reverse mapping from products to their ingredient pairs.
// This should be called once at startup before running DFS algorithms.
// The index is sorted by ingredient tier sum (lower tiers first) to optimize search.
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
	This algorithm finds a single path from target to base elements using
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

// RangeDFSPaths finds multiple unique recipe paths from a target element to base elements.
// Uses an iterative, stack-based DFS approach to avoid recursion stack overflow.
// Parameters:
//   - target: Name of the target element to find recipes for
//   - maxPaths: Maximum number of unique paths to find
//   - g: The indexed graph containing all element relationships
//
// Returns:
//   - []RecipeStep: Slice of recipe steps, one for each unique path found
//   - int: Count of nodes visited during the search
func RangeDFSPaths(target string, maxPaths int, g IndexedGraph) ([]RecipeStep, int) {
	// Stack element for iterative DFS
	type elem struct {
		id       int // Current element ID
		childPos int // Position in the children list
	}

	targetID := g.NameToID[target]
	stack := []elem{{id: targetID}}        // Start with target element
	path := make([][]int, 0, 64)           // Current path being built
	visited := make(map[int]bool)          // Track visited elements to prevent cycles
	seenSig := make(map[uint64]struct{})   // Track seen paths for deduplication
	out := make([]RecipeStep, 0, maxPaths) // Output collection
	nodes := 0                             // Node visit counter

	// Continue until stack is empty or we've found enough paths
	for len(stack) > 0 && len(out) < maxPaths {
		top := &stack[len(stack)-1] // Peek at top element
		id := top.id
		nodes++ // Count this node as visited

		// Check if current element is a base element
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
			if _, ok := seenSig[sig]; !ok {
				// This is a new unique path
				seenSig[sig] = struct{}{}
				out = append(out, buildRecipeStepFromPath(path, targetID, g))
			}
			stack = stack[:len(stack)-1] // Pop from stack
			continue
		}

		// Check for cycles in current path
		if visited[id] {
			stack = stack[:len(stack)-1] // Pop from stack
			continue
		}
		visited[id] = true

		// Check if we've exhausted all children for this element
		if top.childPos >= len(revIdx[id]) {
			visited[id] = false // No longer in path
			if len(path) > 0 {
				path = path[:len(path)-1] // Pop from path
			}
			stack = stack[:len(stack)-1] // Pop from stack
			continue
		}

		// Process next child
		p := revIdx[id][top.childPos] // Get next ingredient pair
		top.childPos++                // Move to next child for future

		// Push both ingredients onto stack for DFS
		stack = append(stack, elem{id: p.a})
		stack = append(stack, elem{id: p.b})

		// Add this step to the current path
		path = append(path, []int{p.a, p.b, id})
	}

	return out, nodes
}

/* -------------------------------------------------------------------------
   Helper functions
   These provide utility functionality used by the DFS algorithms.         */

// findIngredientsFor returns all ingredient pairs that can create a specific product.
// Parameters:
//   - productID: ID of the product element to find ingredients for
//   - g: The indexed graph containing all element relationships
//
// Returns:
//   - [][]int: Slice of [ingredientA, ingredientB] pairs that produce the target
func findIngredientsFor(productID int, g IndexedGraph) [][]int {
	var res [][]int

	// Scan through the entire graph looking for combinations that produce our target
	for aID, nbrs := range g.Edges {
		for _, e := range nbrs {
			if e.ProductID == productID {
				// Found a combination that produces our target
				res = append(res, []int{aID, e.PartnerID})
			}
		}
	}

	return res
}
