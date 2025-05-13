package recipeFinder

import (
	"container/list"
	"context"
	"sort"
	"strings"
	"sync"
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

func shallowCloneExploration(orig PathExploration) PathExploration {
    // clone Path
    newPath := make(map[string]RecipeStep, len(orig.Path))
    for k, v := range orig.Path {
        newPath[k] = v
    }
    // clone IncompletePath
    newInc := make(map[string]bool, len(orig.IncompletePath))
    for k := range orig.IncompletePath {
        newInc[k] = true
    }
    // clone Queue
    newQ := list.New()
    for e := orig.Queue.Front(); e != nil; e = e.Next() {
        newQ.PushBack(e.Value)
    }
    return PathExploration{
        Path:           newPath,
        IncompletePath: newInc,
        Queue:          newQ,
        SearchSteps:    nil,               // drop history
        NodeCount:      orig.NodeCount,
    }
}

func ReversedMultiPathBFSParallel(targetName string, graph IndexedGraph, maxPaths int) ([]ProductToIngredients, []SearchStep, int) {
    targetID := graph.NameToID[targetName]
    var (
        completePaths  []ProductToIngredients
        pathHashes     = make(map[string]bool)
        allSearchSteps []SearchStep
        totalNodes     int
        mu             sync.Mutex
        wg             sync.WaitGroup
    )

    reverseGraph := buildReverseGraph(graph)

    initial := PathExploration{
        Path:           make(map[string]RecipeStep),
        IncompletePath: map[string]bool{targetName: true},
        Queue:          list.New(),
        SearchSteps:    []SearchStep{{ // keep just the very first step
            CurrentID:  -1, CurrentName: "", QueueIDs: []int{targetID}, QueueNames: []string{targetName},
            SeenIDs: nil, SeenNames: nil, DiscoveredEdges: nil, DiscoveredNames: nil,
            StepNumber: 0, FoundTarget: false,
        }},
        NodeCount: 0,
    }
    initial.Queue.PushBack(targetID)

    active := []PathExploration{initial}
    limiter := make(chan struct{}, 4) // max 4 goroutines at once

    for len(completePaths) < maxPaths && len(active) > 0 {
        nextBatch := make([]PathExploration, 0, len(active))

        for _, pe := range active {
            wg.Add(1)
            limiter <- struct{}{}

            go func(pe PathExploration) {
                defer wg.Done()
                defer func() { <-limiter }()

                const maxSteps = 1000
                for pe.Queue.Len() > 0 {
                    front := pe.Queue.Front()
                    curID := pe.Queue.Remove(front).(int)
                    name := graph.IDToName[curID]
                    pe.NodeCount++
                    totalNodes++

                    // only record up to maxSteps
                    if len(pe.SearchSteps) < maxSteps {
                        pe.SearchSteps = append(pe.SearchSteps, SearchStep{
                            CurrentID:   curID,
                            CurrentName: name,
                            QueueIDs:    queueToSlice(pe.Queue),
                            QueueNames:  queueToNameSlice(pe.Queue, graph),
                            StepNumber:  pe.NodeCount,
                            FoundTarget: false,
                        })
                    }

                    if isBaseElement(name) {
                        delete(pe.IncompletePath, name)
                        continue
                    }

                    recs := reverseGraph[curID]
                    if len(recs) > 1 {
                        // do first recipe in this goroutine
                        first := recs[0]
                        pe.Path[name] = RecipeStep{Combo: IngredientCombo{
                            A: graph.IDToName[first.InputA],
                            B: graph.IDToName[first.InputB],
                        }}
                        delete(pe.IncompletePath, name)
                        addIngredientToExploration(first.InputA, graph, &pe)
                        addIngredientToExploration(first.InputB, graph, &pe)

                        for j := 1; j < len(recs); j++ {
                            r := recs[j]
                            branch := shallowCloneExploration(pe)
                            branch.Path[name] = RecipeStep{Combo: IngredientCombo{
                                A: graph.IDToName[r.InputA],
                                B: graph.IDToName[r.InputB],
                            }}
                            delete(branch.IncompletePath, name)
                            addIngredientToExploration(r.InputA, graph, &branch)
                            addIngredientToExploration(r.InputB, graph, &branch)

                            mu.Lock()
                            nextBatch = append(nextBatch, branch)
                            mu.Unlock()
                        }
                    } else if len(recs) == 1 {
                        r := recs[0]
                        pe.Path[name] = RecipeStep{Combo: IngredientCombo{
                            A: graph.IDToName[r.InputA],
                            B: graph.IDToName[r.InputB],
                        }}
                        delete(pe.IncompletePath, name)
                        addIngredientToExploration(r.InputA, graph, &pe)
                        addIngredientToExploration(r.InputB, graph, &pe)
                    }

                    if len(pe.IncompletePath) == 0 {
                        hash := createPathHash(pe.Path)
                        mu.Lock()
                        if !pathHashes[hash] && len(completePaths) < maxPaths {
                            pathHashes[hash] = true
                            completePaths = append(completePaths, pe.Path)
                            allSearchSteps = append(allSearchSteps, pe.SearchSteps...)
                        }
                        mu.Unlock()
                        return
                    }
                }
            }(pe)
        }

        wg.Wait()
        active = nextBatch
        if len(active) > maxPaths*3 {
            active = active[:maxPaths*3]
        }
    }

    return completePaths, allSearchSteps, totalNodes
}

// Helper function to process a single exploration
func processExploration(
    exploration PathExploration,
    reverseGraph map[int][]struct{ InputA, InputB int },
    graph IndexedGraph,
    resultsChannel chan<- struct {
        path        map[string]RecipeStep
        searchSteps []SearchStep
        nodeCount   int
    },
    explorationsQueue chan<- PathExploration,
    done <-chan struct{},
    ctx context.Context,
) {
    const maxDepth = 50
    if exploration.NodeCount > maxDepth {
        return // Too deep, skip this path
    }
    
    // Continue processing until queue is empty
    for exploration.Queue.Len() > 0 {
        // Check for cancellation
        select {
        case <-done:
            return
        case <-ctx.Done():
            return
        default:
            // Continue processing
        }
        
        // Get next element
        front := exploration.Queue.Front()
        if front == nil {
            break // Guard against nil
        }
        
        curID := exploration.Queue.Remove(front).(int)
        curName := graph.IDToName[curID]
        
        exploration.NodeCount++
        
        // Record search step
        if len(exploration.SearchSteps) < 5000 {
            exploration.SearchSteps = append(exploration.SearchSteps, SearchStep{
                CurrentID:   curID,
                CurrentName: curName,
                QueueIDs:    queueToSlice(exploration.Queue),
                QueueNames:  queueToNameSlice(exploration.Queue, graph),
                StepNumber:  exploration.NodeCount,
                FoundTarget: false, // Not needed for reverse search
            })
        }
        
        // Skip if base element
        if isBaseElement(curName) {
            delete(exploration.IncompletePath, curName)
            continue
        }
        
        // Get recipes
        recipes := reverseGraph[curID]
        if len(recipes) == 0 {
            continue // No recipes
        }
        
        // Process first recipe in this exploration
        firstRecipe := recipes[0]
        exploration.Path[curName] = RecipeStep{
            Combo: IngredientCombo{
                A: graph.IDToName[firstRecipe.InputA],
                B: graph.IDToName[firstRecipe.InputB],
            },
        }
        
        // Process ingredients
        delete(exploration.IncompletePath, curName)
        addIngredientToExploration(firstRecipe.InputA, graph, &exploration)
        addIngredientToExploration(firstRecipe.InputB, graph, &exploration)
        
        // Process additional recipes (branching)
        for i := 1; i < len(recipes) && i < 10; i++ {
            select {
            case <-done:
                return
            case <-ctx.Done():
                return
            default:
                // Continue
            }
            
            // Clone for this branch
            branch := cloneExploration(exploration)
            recipe := recipes[i]
            
            // Update branch
            branch.Path[curName] = RecipeStep{
                Combo: IngredientCombo{
                    A: graph.IDToName[recipe.InputA],
                    B: graph.IDToName[recipe.InputB],
                },
            }
            
            // Process ingredients
            delete(branch.IncompletePath, curName)
            addIngredientToExploration(recipe.InputA, graph, &branch)
            addIngredientToExploration(recipe.InputB, graph, &branch)
            
            // Queue branch
            select {
            case explorationsQueue <- branch:
                // Successfully queued
            case <-done:
                return
            case <-ctx.Done():
                return
            default:
                // Queue full, drop this branch
            }
        }
        
        // Check if path is complete
        if len(exploration.IncompletePath) == 0 {
            // Complete path - send result, but only if channels aren't closed
            select {
            case resultsChannel <- struct {
                path        map[string]RecipeStep
                searchSteps []SearchStep
                nodeCount   int
            }{
                path:        exploration.Path,
                searchSteps: exploration.SearchSteps,
                nodeCount:   exploration.NodeCount,
            }:
                // Successfully sent
            case <-done:
                return
            case <-ctx.Done():
                return
            }
            return
        }
    }
}

// Helper function to add ingredient to exploration
func addIngredientToExploration(ingredientID int, graph IndexedGraph, exp *PathExploration) {
    ingredientName := graph.IDToName[ingredientID]
    if !isBaseElement(ingredientName) {
        exp.IncompletePath[ingredientName] = true
        exp.Queue.PushBack(ingredientID)
    }
}

// Helper function to process individual recipes
func processRecipe(
	exploration PathExploration,
	recipe struct{ InputA, InputB int },
	graph IndexedGraph,
	curName string,
) PathExploration {
	// Update path
	exploration.Path[curName] = RecipeStep{
		Combo: IngredientCombo{
			A: graph.IDToName[recipe.InputA],
			B: graph.IDToName[recipe.InputB],
		},
	}

	// Update incomplete path
	delete(exploration.IncompletePath, curName)
	ingredientA := graph.IDToName[recipe.InputA]
	ingredientB := graph.IDToName[recipe.InputB]

	if !isBaseElement(ingredientA) {
		exploration.IncompletePath[ingredientA] = true
		exploration.Queue.PushBack(recipe.InputA)
	}

	if !isBaseElement(ingredientB) {
		exploration.IncompletePath[ingredientB] = true
		exploration.Queue.PushBack(recipe.InputB)
	}

	return exploration
}

// Helper to create initial exploration state
func createInitialExploration(targetName string, targetID int) PathExploration {
	exploration := PathExploration{
		Path:           make(map[string]RecipeStep),
		IncompletePath: map[string]bool{targetName: true},
		Queue:          list.New(),
		SearchSteps:    []SearchStep{},
		NodeCount:      0,
	}
	
	// Add target to queue
	exploration.Queue.PushBack(targetID)
	
	// Initialize first search step
	exploration.SearchSteps = append(exploration.SearchSteps, SearchStep{
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
	
	return exploration
}
