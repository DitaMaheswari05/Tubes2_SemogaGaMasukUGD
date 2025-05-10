package recipeFinder

import (
	"hash/fnv"
	"sort"
)

/* -------------------------------------------------------------------------
   Reverse index (read-only, dibuat sekali & di-share)                      */
type pair struct{ a, b int }

var revIdx revIndex            // productID → []pair, package-level read-only

type revIndex map[int][]pair

func BuildReverseIndex(g IndexedGraph) { // panggil sekali di init/server main
    idx := make(revIndex)
    for a, nbrs := range g.Edges {
        for _, e := range nbrs {
            p := pair{a: min(a, e.PartnerID), b: max(a, e.PartnerID)}
            idx[e.ProductID] = append(idx[e.ProductID], p)
        }
    }
    for prod, list := range idx {
        sort.Slice(list, func(i, j int) bool {
            ti := getElementTier(g.IDToName[list[i].a]) +
                getElementTier(g.IDToName[list[i].b])
            tj := getElementTier(g.IDToName[list[j].a]) +
                getElementTier(g.IDToName[list[j].b])
            return ti < tj
        })
        idx[prod] = list
    }
    revIdx = idx // publish
}

var canReachBaseCache map[int]bool

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

/* -------------------------------------------------------------------------
   Iteratif DFS — satu jalur (target → base)                                */
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

/* -------------------------------------------------------------------------
   Iteratif DFS — banyak jalur, berhenti tepat di maxPaths                  */
func RangeDFSPaths(target string, maxPaths int, g IndexedGraph) []RecipeStep {
    targetID := g.NameToID[target]
    var out []RecipeStep

    type elem struct {
        id       int
        childPos int
    }
    pathTriples := make([][]int, 0, 64) // re-used slice

    // signature set (uint64) untuk deduplikasi
    seen := make(map[uint64]struct{})

    stack := []elem{{id: targetID}}
    visited := make(map[int]bool)

    for len(stack) > 0 && len(out) < maxPaths {
        top := &stack[len(stack)-1]
        id := top.id

        // base?
        isBase := false
        for _, b := range BaseElements {
            if id == g.NameToID[b] {
                isBase = true
                break
            }
        }
        if isBase {
            sig := hashPath(pathTriples)
            if _, ok := seen[sig]; !ok {
                seen[sig] = struct{}{}
                out = append(out, buildRecipeStepFromPath(pathTriples, targetID, g))
            }
            stack = stack[:len(stack)-1]
            continue
        }

        // siklus path
        if visited[id] {
            stack = stack[:len(stack)-1]
            continue
        }
        visited[id] = true

        if top.childPos >= len(revIdx[id]) {
            // habis, backtrack
            visited[id] = false
            if len(pathTriples) > 0 {
                pathTriples = pathTriples[:len(pathTriples)-1]
            }
            stack = stack[:len(stack)-1]
            continue
        }

        // jelajahi anak berikut
        p := revIdx[id][top.childPos]
        top.childPos++
        stack = append(stack, elem{id: p.a})
        stack = append(stack, elem{id: p.b})
        pathTriples = append(pathTriples, []int{p.a, p.b, id})
    }
    return out
}

/* ------------------------------------------------------------------------- */
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
        put(t[0]); put(t[1]); put(t[2])
    }
    return h.Sum64()
}
