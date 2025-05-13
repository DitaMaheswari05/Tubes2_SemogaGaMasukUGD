package recipeFinder


// UnifiedRecipeTree builds a complete tree showing all ways to make an element
func UnifiedRecipeTree(targetName string, graph IndexedGraph) *RecipeNode {
    // Create reverse mapping: product → all recipes that make it
    reverseGraph := buildReverseGraph(graph)
    
    // Visited map to avoid duplication in visualization
    visited := make(map[string]bool)
    
    // Build the complete tree recursively
    return buildUnifiedTree(targetName, reverseGraph, graph, visited, 0)
}

// Recursive helper to build the tree
func buildUnifiedTree(elementName string, reverseGraph map[int][]struct{InputA, InputB int}, 
                       graph IndexedGraph, visited map[string]bool, depth int) *RecipeNode {
    // Create node for this element
    node := &RecipeNode{Name: elementName}
    
    // Base case: stop at base elements or max depth
    if isBaseElement(elementName) || depth > 30 {
        return node
    }
    
    // Avoid cycles (though the tier system should prevent them)
    if visited[elementName] {
        return node
    }
    visited[elementName] = true
    defer delete(visited, elementName) // Remove when done with this branch
    
    // Get element ID
    elementID := graph.NameToID[elementName]
    
    // Find all recipes that make this element
    recipes := reverseGraph[elementID]
    
    // If no recipes, return just the node
    if len(recipes) == 0 {
        return node
    }
    
    // For each recipe that makes this element
    var children []*RecipeNode
    for _, recipe := range recipes {
        // Get ingredient names
        ingredientA := graph.IDToName[recipe.InputA]
        ingredientB := graph.IDToName[recipe.InputB]
        
        // Recursively build trees for both ingredients
        childA := buildUnifiedTree(ingredientA, reverseGraph, graph, visited, depth+1)
        childB := buildUnifiedTree(ingredientB, reverseGraph, graph, visited, depth+1)
        
        // Create a combiner node to represent this specific recipe
        combiner := &RecipeNode{
            Name: elementName + " Recipe", 
            Children: []*RecipeNode{childA, childB},
        }
        
        children = append(children, combiner)
    }
    
    // Attach all recipe variants to this element
    node.Children = children
    
    return node
}