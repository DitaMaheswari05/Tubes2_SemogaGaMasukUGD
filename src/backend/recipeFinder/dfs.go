package recipeFinder

import (
	"sort"
)

// Global cache to remember results
var canReachBaseCache map[int]bool

// DFSBuildTargetToBase finds a path from target to base elements
func DFSBuildTargetToBase(target string, graph IndexedGraph) ProductToIngredients {
    targetID := graph.NameToID[target]
    recipes := make(ProductToIngredients)
    visited := make(map[int]bool)
    
    // Initialize cache
    canReachBaseCache = make(map[int]bool)
    
    // Pre-populate cache with base elements
    for _, baseName := range BaseElements {
        baseID := graph.NameToID[baseName]
        canReachBaseCache[baseID] = true
    }
    
    // Use a very large depth limit - basically unlimited
    maxDepth := 1000 
    
    // Start DFS with depth limiting
    success := findPathToBase(targetID, 0, maxDepth, graph, recipes, visited)
    
    if !success {
        // If not found with very high depth, try one last time with unlimited depth
        visited = make(map[int]bool) // Reset visited map
        findPathToBase(targetID, 0, 10000, graph, recipes, visited)
    }
    
    return recipes
}

func findPathToBase(elementID int, depth int, maxDepth int, graph IndexedGraph, recipes ProductToIngredients, visited map[int]bool) bool {
    // Depth limit check (very high to ensure completeness)
    if depth > maxDepth {
        return false // Too deep, but this should rarely happen
    }
    
    // Check cache first
    if result, found := canReachBaseCache[elementID]; found {
        return result
    }
    
    // Avoid cycles
    if visited[elementID] {
        canReachBaseCache[elementID] = false
        return false
    }
    
    visited[elementID] = true
    defer func() { visited[elementID] = false }() // Clean up on return
    
    // Check if it's a base element
    elementName := graph.IDToName[elementID]
    for _, base := range BaseElements {
        if elementName == base {
            canReachBaseCache[elementID] = true
            return true
        }
    }
    
    // Find all recipes that produce this element
    ingredients := findIngredientsFor(elementID, graph)
    
    // Sort ingredients by tier (lower tiers first)
    sort.Slice(ingredients, func(i, j int) bool {
        iTier := getElementTier(graph.IDToName[ingredients[i][0]]) + 
                 getElementTier(graph.IDToName[ingredients[i][1]])
        jTier := getElementTier(graph.IDToName[ingredients[j][0]]) + 
                 getElementTier(graph.IDToName[ingredients[j][1]])
        return iTier < jTier
    })
    
    // Try each recipe
    for _, pair := range ingredients {
        ingredient1 := pair[0]
        ingredient2 := pair[1]
        
        // Try to find paths from both ingredients to base elements
        if findPathToBase(ingredient1, depth+1, maxDepth, graph, recipes, visited) && 
           findPathToBase(ingredient2, depth+1, maxDepth, graph, recipes, visited) {
            
            // Record the recipe
            recipes[elementName] = RecipeStep{
                Combo: IngredientCombo{
                    A: graph.IDToName[ingredient1],
                    B: graph.IDToName[ingredient2],
                },
            }
            
            canReachBaseCache[elementID] = true
            return true
        }
    }
    
    canReachBaseCache[elementID] = false
    return false
}

// RangeDFSPaths finds up to maxPaths distinct DFS paths from target → bases.
func RangeDFSPaths(target string, maxPaths int, graph IndexedGraph) []RecipeStep {
    targetID := graph.NameToID[target]
    var out []RecipeStep
    seenSig := make(map[string]bool)
    visit := make(map[int]bool)
    var path [][]int

    var dfs func(int)
    dfs = func(cur int) {
        if len(out) >= maxPaths { 
            return 
        }
        // If base reached, record this full path
        name := graph.IDToName[cur]
        for _, b := range BaseElements {
            if name == b {
                // build signature
                sig := canonicalHash(path)
                if !seenSig[sig] {
                    seenSig[sig] = true
                    out = append(out, buildRecipeStepFromPath(path, targetID, graph))
                }
                return
            }
        }
        // Prevent cycles
        if visit[cur] {
            return
        }
        visit[cur] = true
        defer func() { visit[cur] = false }()

        // Explore each recipe that produces cur
        // findIngredientsFor returns [][]int{{aID, bID},...}
        for _, pair := range findIngredientsFor(cur, graph) {
            a, b := pair[0], pair[1]
            // append this step
            path = append(path, []int{min(a, b), max(a, b), cur})
            // continue DFS on both ingredients
            dfs(a)
            dfs(b)
            // pop
            path = path[:len(path)-1]
        }
    }

    dfs(targetID)
    return out
}

// findIngredientsFor returns all [a,b] pairs that yield product.
func findIngredientsFor(productID int, graph IndexedGraph) [][]int {
    var res [][]int
    for aID, nbrs := range graph.Edges {
        for _, e := range nbrs {
            if e.ProductID == productID {
                res = append(res, []int{aID, e.PartnerID})
            }
        }
    }
    return res
}