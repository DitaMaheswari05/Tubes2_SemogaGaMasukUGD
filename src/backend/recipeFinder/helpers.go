package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
)

/*
-------------------------------------------------------------------------
Helper & util (GENERAL)
*/
func canonicalHash(p [][]int) string {
	var steps []string
	for _, t := range p {
		a, b := min(t[0], t[1]), max(t[0], t[1])
		steps = append(steps, fmt.Sprintf("%d-%d-%d", a, b, t[2]))
	}
	return strings.Join(steps, "|")
}

func buildRecipeStepFromPath(path [][]int, targetID int, g IndexedGraph) RecipeStep {
	if len(path) == 0 {
		return RecipeStep{}
	}
	strPath := make([][]string, len(path))
	for i, t := range path {
		strPath[i] = []string{
			g.IDToName[t[0]],
			g.IDToName[t[1]],
			g.IDToName[t[2]],
		}
	}
	last := path[len(path)-1]
	return RecipeStep{
		Combo: IngredientCombo{
			A: g.IDToName[last[0]],
			B: g.IDToName[last[1]],
		},
		Path: strPath,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


// Improved deduplication of recipe trees
func DeduplicateRecipeTrees(trees []*RecipeNode) []*RecipeNode {
	if len(trees) <= 1 {
		return trees
	}

	// Map to track signatures we've seen
	seen := make(map[string]bool)
	result := make([]*RecipeNode, 0, len(trees))

	for _, tree := range trees {
		sig := treeSignature(tree)
		if !seen[sig] {
			seen[sig] = true
			result = append(result, tree)
		}
	}

	return result
}

// Calculate similarity between two paths
func pathSimilarity(path1, path2 [][]string) float64 {
	if len(path1) == 0 || len(path2) == 0 {
		return 0
	}

	// Create sets of unique elements in each path
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, step := range path1 {
		for _, elem := range step {
			set1[elem] = true
		}
	}

	for _, step := range path2 {
		for _, elem := range step {
			set2[elem] = true
		}
	}

	// Count common elements
	common := 0
	for elem := range set1 {
		if set2[elem] {
			common++
		}
	}

	// Calculate Jaccard similarity: intersection/union
	return float64(common) / float64(len(set1)+len(set2)-common)
}


// TreeSignature creates a structural signature for deduplication
func treeSignature(tree *RecipeNode) string {
	if tree == nil {
		return ""
	}

	// If it's a leaf node
	if len(tree.Children) == 0 {
		return tree.Name
	}

	// Get signatures for children and sort them
	childSigs := make([]string, len(tree.Children))
	for i, child := range tree.Children {
		childSigs[i] = treeSignature(child)
	}
	sort.Strings(childSigs) // Sort to make order-independent

	// Combine into a unique signature
	return fmt.Sprintf("%s(%s)", tree.Name, strings.Join(childSigs, "|"))
}

/*
-------------------------------------------------------------------------
Helper & util (BFS)
*/
// Helper to extract a path from single-recipe BFS result
func extractPathFromRecipes(targetID int, recipes ProductToIngredients, g IndexedGraph) RecipeStep {
	// Build a path by following the recipe chain from target to base elements
	var path [][]string
	current := g.IDToName[targetID]

	// Track visited elements to avoid cycles
	visited := make(map[string]bool)

	// Walk the recipe chain
	for {
		if visited[current] {
			break // Avoid cycles
		}
		visited[current] = true

		recipe, exists := recipes[current]
		if !exists {
			break
		}

		a := recipe.Combo.A
		b := recipe.Combo.B

		// Add this step to the path
		path = append(path, []string{a, b, current})

		// If both ingredients are base elements, we're done
		aIsBase := isBaseElement(a)
		bIsBase := isBaseElement(b)

		if aIsBase && bIsBase {
			break
		}

		// Continue with a non-base ingredient
		if !aIsBase {
			current = a
		} else {
			current = b
		}
	}

	// Reverse the path since we built it backwards
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// Convert to RecipeStep format
	return RecipeStep{
		Combo: IngredientCombo{
			A: recipes[g.IDToName[targetID]].Combo.A,
			B: recipes[g.IDToName[targetID]].Combo.B,
		},
		Path: path,
	}
}

// Helper function to convert map keys to a slice of ints
func mapKeysToSlice(m map[int]bool) []int {
	result := make([]int, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Helper function to convert map keys to element names
func mapKeysToNameSlice(m map[int]bool, g IndexedGraph) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, g.IDToName[k])
	}
	return result
}

// Helper function to convert queue to a slice of ints
func queueToSlice(q *list.List) []int {
	result := make([]int, 0, q.Len())
	for e := q.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(int))
	}
	return result
}

// Helper function to convert queue to a slice of names
func queueToNameSlice(q *list.List, g IndexedGraph) []string {
	result := make([]string, 0, q.Len())
	for e := q.Front(); e != nil; e = e.Next() {
		id := e.Value.(int)
		result = append(result, g.IDToName[id])
	}
	return result
}

// copyMap creates a deep copy of a map of product IDs to parent/partner IDs
func copyMap(m map[int]struct{ ParentID, PartnerID int }) map[int]struct{ ParentID, PartnerID int } {
	result := make(map[int]struct{ ParentID, PartnerID int }, len(m))
	for k, v := range m {
		result[k] = v // struct is copied by value
	}
	return result
}

// prevIDsToNames converts the integer IDs in the prevIDs map to their string names
func prevIDsToNames(m map[int]struct{ ParentID, PartnerID int }, g IndexedGraph) map[string]struct{ A, B string } {
	result := make(map[string]struct{ A, B string }, len(m))
	for productID, info := range m {
		productName := g.IDToName[productID]
		result[productName] = struct{ A, B string }{
			A: g.IDToName[info.ParentID],
			B: g.IDToName[info.PartnerID],
		}
	}
	return result
}

/*
-------------------------------------------------------------------------
Helper & util (DFS)
*/
// findIngredientsFor returns all ingredient pairs that can create a specific product.

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
