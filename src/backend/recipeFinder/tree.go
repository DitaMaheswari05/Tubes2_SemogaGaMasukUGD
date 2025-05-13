package recipeFinder

import "log"

// RecipeNode itu bentuk data JSON yang nanti kita kirim ke frontend.
// - Name: nama elemen (misal "Brick")
// - Children: daftar subtree resep bahan-bahannya.
//   kalau leaf (elemen dasar), Children bisa kosong atau nil.
type RecipeNode struct {
	Name     string			`json:"name"`
	Children []*RecipeNode	`json:"children,omitempty"`
}

func BuildTrees(target string, pathPrev map[string][]RecipeStep) []*RecipeNode {
	var trees []*RecipeNode

	for _, recipeStep := range pathPrev[target] {
		// Use only the steps from this specific path to build prev
		prev := make(ProductToIngredients)
		for _, step := range recipeStep.Path {
			if len(step) == 3 {
				product := step[2]
				prev[product] = RecipeStep{
					Combo: IngredientCombo{
						A: step[0],
						B: step[1],
					},
				}
			}
		}

		tree := BuildTree(target, prev)
		trees = append(trees, tree)
	}

	return trees
}

func isBaseElement(name string) bool {
	for _, b := range BaseElements {
		if b == name {
			return true
		}
	}
	return false
}

const treeDepthLimit = 150

func BuildTree(name string, prev ProductToIngredients) *RecipeNode {
	return buildTreeRec(name, prev, make(map[string]bool), 0)
}

func buildTreeRec(
	name string,
	prev ProductToIngredients,
	visited map[string]bool,
	depth int,
) *RecipeNode {
	node := &RecipeNode{Name: name}

	// 1. stop-conditions ----------------------------------------------------
	if isBaseElement(name) || depth >= treeDepthLimit {
		return node
	}
	if visited[name] { // siklus terdeteksi
		return node
	}
	visited[name] = true

	// 2. gunakan info resep dari prev jika ada -----------------------------
	if step, ok := prev[name]; ok {
		node.Children = []*RecipeNode{
			buildTreeRec(step.Combo.A, prev, visited, depth+1),
			buildTreeRec(step.Combo.B, prev, visited, depth+1),
		}
		return node
	}

	// 3. fallback sekali saja ----------------------------------------------
	if _, ok := prev[name]; !ok {
		// Log whenever fallback is attempted
		log.Printf("FALLBACK TRIGGERED for element %q at depth %d (visited elements: %v)", 
			name, depth, getVisitedKeys(visited))
		
		if fb, _, nodesVisited := IndexedBFSBuild(name, GlobalIndexedGraph); len(fb) > 0 {
			// Log fallback success details
			log.Printf("FALLBACK SUCCESS for %q: found recipe via BFS (%d nodes visited)", 
			name, nodesVisited)
			
			if step, ok := fb[name]; ok {
				// Log the exact recipe found
				log.Printf("FALLBACK RECIPE for %q: %s + %s", 
					name, step.Combo.A, step.Combo.B)
					
				node.Children = []*RecipeNode{
					buildTreeRec(step.Combo.A, fb, visited, depth+1),
					buildTreeRec(step.Combo.B, fb, visited, depth+1),
				}
			} else {
				log.Printf("FALLBACK ERROR: BFS returned recipes but none for %q?!", name)
			}
		} else {
			log.Printf("FALLBACK FAILED for %q: BFS found no recipes", name)
		}
	}
	return node
}

// Helper function to get keys from map for logging
func getVisitedKeys(visited map[string]bool) []string {
    keys := make([]string, 0, len(visited))
    for k := range visited {
        keys = append(keys, k)
    }
    return keys
}
