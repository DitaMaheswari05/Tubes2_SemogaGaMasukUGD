package recipeFinder

import (
	"container/list"
	"sort"
	"strings"
)


type PathExploration struct {
	Path           map[string]RecipeStep  // Current path from target toward base
	IncompletePath map[string]bool        // Elements in path that aren't base elements yet
	Queue          *list.List             // BFS queue for this path
	SearchSteps    []SearchStep           // Visualization steps for this path
	NodeCount      int                    // Nodes visited in this path
}

// Multi-recipe BFS that works in reverse (target → base)
func ReversedMultiPathBFS(targetName string, graph IndexedGraph, maxPaths int) ([]ProductToIngredients, []SearchStep, int) {
    targetID := graph.NameToID[targetName]
    
    // Track complete paths (each is a separate recipe tree)
    var completePaths []ProductToIngredients
    pathHashes := make(map[string]bool)
    
    // Track all search steps for visualization
    allSearchSteps := []SearchStep{}
    totalNodes := 0
    
    // We'll need a reverse graph (product → ingredient combinations)
    // This gives us all ways to create an element
    reverseGraph := buildReverseGraph(graph)
    
    // Each path is a BFS exploration from target to base
    // We'll track multiple separate path explorations

    
    // Start with initial exploration from target
    initialExploration := PathExploration{
        Path:           make(map[string]RecipeStep),
        IncompletePath: map[string]bool{targetName: true},
        Queue:          list.New(),
        SearchSteps:    []SearchStep{},
        NodeCount:      0,
    }
    
    // Add target to queue
    initialExploration.Queue.PushBack(targetID)
    
    // Initialize search steps
    initialExploration.SearchSteps = append(initialExploration.SearchSteps, SearchStep{
        CurrentID:       -1,
        CurrentName:     "",
        QueueIDs:        []int{targetID},
        QueueNames:      []string{targetName},
        SeenIDs:         []int{},
        SeenNames:       []string{},
        DiscoveredEdges: make(map[int]struct{ ParentID, PartnerID int }),
        DiscoveredNames: make(map[string]struct{ A, B string }),
        StepNumber:      0,
        FoundTarget:     false,
    })
    
    // All active path explorations
    activeExplorations := []PathExploration{initialExploration}
    
    // We'll keep exploring paths until we have enough or run out of options
    for len(completePaths) < maxPaths && len(activeExplorations) > 0 {
        // Take the first active exploration
        currentExploration := activeExplorations[0]
        activeExplorations = activeExplorations[1:]
        
        // Continue BFS on this path until it's complete or we need to branch
        for currentExploration.Queue.Len() > 0 {
            // Get next element to explore
            front := currentExploration.Queue.Front() // FIX: Access queue directly
            curID := currentExploration.Queue.Remove(front).(int)
            curName := graph.IDToName[curID]
            
            currentExploration.NodeCount++
            totalNodes++
            
            // Record search step
            currentExploration.SearchSteps = append(currentExploration.SearchSteps, SearchStep{
                CurrentID:   curID,
                CurrentName: curName,
                QueueIDs:    queueToSlice(currentExploration.Queue),
                QueueNames:  queueToNameSlice(currentExploration.Queue, graph),
                // Other visualization fields...
                StepNumber:  currentExploration.NodeCount,
                FoundTarget: false,  // Not relevant in reverse search
            })
            
            // Skip if this is a base element
            if isBaseElement(curName) {
                // Remove from incomplete path
                delete(currentExploration.IncompletePath, curName)
                continue
            }
            
            // Get all ways to create this element
            recipes := reverseGraph[curID]
            
            // Check if we need to branch the path
            if len(recipes) > 1 {
                // First recipe continues in this exploration
                firstRecipe := recipes[0]
                
                // Add recipe to current path
                currentExploration.Path[curName] = RecipeStep{
                    Combo: IngredientCombo{
                        A: graph.IDToName[firstRecipe.InputA],
                        B: graph.IDToName[firstRecipe.InputB],
                    },
                }
                
                // Add ingredient elements to incomplete path
                ingredientA := graph.IDToName[firstRecipe.InputA]
                ingredientB := graph.IDToName[firstRecipe.InputB]
                
                // Remove current from incomplete, add ingredients if not base
                delete(currentExploration.IncompletePath, curName)
                if !isBaseElement(ingredientA) {
                    currentExploration.IncompletePath[ingredientA] = true
                    currentExploration.Queue.PushBack(firstRecipe.InputA)
                }
                
                if !isBaseElement(ingredientB) {
                    currentExploration.IncompletePath[ingredientB] = true
                    currentExploration.Queue.PushBack(firstRecipe.InputB)
                }
                
                // Create new explorations for remaining recipes (branch paths)
                for i := 1; i < len(recipes); i++ {
                    recipe := recipes[i]
                    
                    // Clone the current exploration
                    newExploration := cloneExploration(currentExploration)
                    
                    // Add this recipe variant
                    newExploration.Path[curName] = RecipeStep{
                        Combo: IngredientCombo{
                            A: graph.IDToName[recipe.InputA],
                            B: graph.IDToName[recipe.InputB],
                        },
                    }
                    
                    // Add ingredient elements to new path's incomplete list
                    newIngredientA := graph.IDToName[recipe.InputA]
                    newIngredientB := graph.IDToName[recipe.InputB]
                    
                    // Remove current, add ingredients if not base
                    delete(newExploration.IncompletePath, curName)
                    if !isBaseElement(newIngredientA) {
                        newExploration.IncompletePath[newIngredientA] = true
                        newExploration.Queue.PushBack(recipe.InputA)
                    }
                    
                    if !isBaseElement(newIngredientB) {
                        newExploration.IncompletePath[newIngredientB] = true
                        newExploration.Queue.PushBack(recipe.InputB)
                    }
                    
                    // Add to active explorations
                    activeExplorations = append(activeExplorations, newExploration)
                }
            } else if len(recipes) == 1 {
                // Just one recipe, continue current path
                recipe := recipes[0]
                
                // Add recipe to current path
                currentExploration.Path[curName] = RecipeStep{
                    Combo: IngredientCombo{
                        A: graph.IDToName[recipe.InputA],
                        B: graph.IDToName[recipe.InputB],
                    },
                }
                
                // Add ingredient elements to incomplete path
                ingredientA := graph.IDToName[recipe.InputA]
                ingredientB := graph.IDToName[recipe.InputB]
                
                // Remove current, add ingredients if not base
                delete(currentExploration.IncompletePath, curName)
                if !isBaseElement(ingredientA) {
                    currentExploration.IncompletePath[ingredientA] = true
                    currentExploration.Queue.PushBack(recipe.InputA)
                }
                
                if !isBaseElement(ingredientB) {
                    currentExploration.IncompletePath[ingredientB] = true
                    currentExploration.Queue.PushBack(recipe.InputB)
                }
            }
            
            // Check if this path is now complete (all elements resolved to base)
            if len(currentExploration.IncompletePath) == 0 {
                // Path is complete - check if it's unique
                pathHash := createPathHash(currentExploration.Path)
                
                if !pathHashes[pathHash] {
                    pathHashes[pathHash] = true
                    completePaths = append(completePaths, currentExploration.Path)
                    allSearchSteps = append(allSearchSteps, currentExploration.SearchSteps...)
                    
                    // If we have enough paths, break
                    if len(completePaths) >= maxPaths {
                        break
                    }
                }
                
                // Don't continue with this exploration
                break
            }
        }
        
        // Limit the number of active explorations to avoid memory issues
        if len(activeExplorations) > maxPaths*3 {
            activeExplorations = activeExplorations[:maxPaths*3]
        }
    }
    
    return completePaths, allSearchSteps, totalNodes
}

// Helper function to build reverse graph (product -> ingredient combinations)
func buildReverseGraph(graph IndexedGraph) map[int][]struct{ InputA, InputB int } {
    reverse := make(map[int][]struct{ InputA, InputB int })
    
    // Iterate through all elements
    for elementID, neighbors := range graph.Edges {
        for _, neighbor := range neighbors {
            partnerID := neighbor.PartnerID
            productID := neighbor.ProductID
            
            // Add recipe to reverse graph
            reverse[productID] = append(reverse[productID], struct{ InputA, InputB int }{
                InputA: elementID,
                InputB: partnerID,
            })
        }
    }
    
    return reverse
}

// Clone a path exploration
func cloneExploration(original PathExploration) PathExploration {
    clone := PathExploration{
        Path:           make(map[string]RecipeStep),
        IncompletePath: make(map[string]bool),
        Queue:          list.New(),
        SearchSteps:    make([]SearchStep, len(original.SearchSteps)),
        NodeCount:      original.NodeCount,
    }
    
    // Copy path
    for k, v := range original.Path {
        clone.Path[k] = v
    }
    
    // Copy incomplete path
    for k := range original.IncompletePath {
        clone.IncompletePath[k] = true
    }
    
    // Copy queue
    for e := original.Queue.Front(); e != nil; e = e.Next() {
        clone.Queue.PushBack(e.Value)
    }
    
    // Copy search steps
    copy(clone.SearchSteps, original.SearchSteps)
    
    return clone
}

// Create a unique hash for a path to detect duplicates
func createPathHash(path map[string]RecipeStep) string {
    // Sort keys for consistent hash
    keys := make([]string, 0, len(path))
    for k := range path {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    // Build hash string
    var hashBuilder strings.Builder
    for _, k := range keys {
        hashBuilder.WriteString(k)
        hashBuilder.WriteString(":")
        hashBuilder.WriteString(path[k].Combo.A)
        hashBuilder.WriteString("+")
        hashBuilder.WriteString(path[k].Combo.B)
        hashBuilder.WriteString(";")
    }
    
    return hashBuilder.String()
}