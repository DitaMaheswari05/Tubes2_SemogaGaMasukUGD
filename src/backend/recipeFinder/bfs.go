package recipeFinder

import (
	"container/list"
	"fmt"
	"strings"
)

func IndexedBFSBuild(targetName string, graph IndexedGraph) (ProductToIngredients, int) {
	targetID := graph.NameToID[targetName]

	queue := list.New()
	seen := make(map[int]bool)

	for _, baseName := range BaseElements {
		baseID := graph.NameToID[baseName]
		queue.PushBack(baseID)
		seen[baseID] = true
	}

	// Track parents using integer IDs
	prevIDs := make(map[int]struct{ParentID, PartnerID int})
	nodes := 0
	for queue.Len() > 0 {
		curID := queue.Remove(queue.Front()).(int)
		nodes++
		
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
			// Path is nil here since we're not tracking full paths in this function
		}
	}

	return recipes, nodes
}

/* -------------------------------------------------------------------------
   2.  Jalur ke‑(skip+1) (distinct)                                          */
   func findKthPathIndexed(targetID, skip int, g IndexedGraph) (RecipeStep, int) {
    type state struct {
        elem  int
        path  [][]int
        depth int
    }
    type edge struct{ PartnerID, ProductID int }

    reachable := make(map[int]bool)
    waiting := make(map[int][]edge)
    seenPath := make(map[string]bool)

    q := list.New()
    for _, b := range BaseElements {
        id := g.NameToID[b]
        reachable[id] = true
        q.PushBack(state{elem: id, depth: 0})
    }

    nodes, hits := 0, 0
    const maxDepth = 20       // Reduced from 40 to 20
    const maxQueueSize = 5000 // Limit queue size

    enqueue := func(a, b, c int, cur [][]int, d int) {
        // Skip if queue is getting too large
        if q.Len() >= maxQueueSize {
            return
        }
        
        // Skip if depth is already high
        if d >= maxDepth {
            return
        }
        
        // Create new path with minimal copying
        np := make([][]int, len(cur)+1)
        copy(np, cur)
        np[len(cur)] = []int{min(a, b), max(a, b), c}
        
        // Check for duplicates
        sig := canonicalHash(np)
        if seenPath[sig] {
            return
        }
        seenPath[sig] = true
        
        q.PushBack(state{elem: c, path: np, depth: d + 1})
    }

    for q.Len() > 0 {
        st := q.Remove(q.Front()).(state)
        nodes++
        
        if st.depth > maxDepth {
            continue
        }
        
        if st.elem == targetID {
            if hits == skip {
                // Free memory before return
                seenPath = nil
                waiting = nil
                return buildRecipeStepFromPath(st.path, targetID, g), nodes
            }
            hits++
            continue
        }

        // Memory safety check - if too many paths are being processed
        // don't add more unless they're very promising
        highMemoryPressure := q.Len() > maxQueueSize/2
        
        for _, r := range g.Edges[st.elem] {
            // During high memory pressure, prioritize only certain paths
            if highMemoryPressure {
                // Skip if this would create a very long path
                if st.depth > 10 {
                    continue
                }
            }
            
            if reachable[r.PartnerID] {
                enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
            } else {
                // Only track waiting reactions if we're not under pressure
                if !highMemoryPressure {
                    waiting[r.PartnerID] = append(waiting[r.PartnerID], edge{
                        PartnerID: st.elem, ProductID: r.ProductID})
                }
            }
        }
        
        // Free memory by dropping the path once processed
        st.path = nil
        
        // Process waiting reactions only if memory pressure isn't high
        if !highMemoryPressure {
            if lst, ok := waiting[st.elem]; ok {
                for _, r := range lst {
                    if reachable[r.PartnerID] {
                        enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
                    }
                }
                delete(waiting, st.elem)
            }
        }
        
        reachable[st.elem] = true
    }
    
    return RecipeStep{}, nodes
}

/* -------------------------------------------------------------------------
   3.  Beberapa jalur berurutan                                              */
// RangePathsIndexed returns up to `limit` distinct paths to `targetID`
func RangePathsIndexed(targetID int, start, limit int, g IndexedGraph) ([]RecipeStep, int) {
    // If we're looking for the first path (start=0), use the efficient single-path algorithm
    if start == 0 && limit > 0 {
        // Get single path efficiently first
        singleRecipes, nodes := IndexedBFSBuild(g.IDToName[targetID], g)
        
        // If we found a path, convert it to RecipeStep format
        if len(singleRecipes) > 0 {
            // Extract the first recipe path
            path := extractPathFromRecipes(targetID, singleRecipes, g)
            
            // Check if the extracted path has valid content
            if path.Path != nil && len(path.Path) > 0 {
                firstPath := []RecipeStep{path}
                
                // If we only need one path, return it immediately
                if limit == 1 {
                    return firstPath, nodes
                }
                
                // Otherwise, get additional paths starting from the second path (limit-1 more)
                additionalPaths, additionalNodes := findAdditionalPaths(targetID, 0, limit-1, g)
                
                // Combine the first path with additional paths
                result := append(firstPath, additionalPaths...)
                return result, nodes + additionalNodes
            }
        }
    }
    
    // Fall back to standard multi-path search
    return findAdditionalPaths(targetID, start, limit, g)
}

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

// findAdditionalPaths implements the standard multi-path search
func findAdditionalPaths(targetID, start, limit int, g IndexedGraph) ([]RecipeStep, int) {
    // This is your current implementation of finding multiple paths
    out := make([]RecipeStep, 0, limit)
    totalNodesVisited := 0
    
    for k := 0; k < limit; k++ {
        step, nodesVisited := findKthPathIndexed(targetID, start+k, g)
        totalNodesVisited += nodesVisited
        
        if step.Path == nil {
            break // BFS exhausted
        }
        out = append(out, step)
    }
    
    return out, totalNodesVisited
}

/* -------------------------------------------------------------------------
   Helper & util                                                            */
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

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }
