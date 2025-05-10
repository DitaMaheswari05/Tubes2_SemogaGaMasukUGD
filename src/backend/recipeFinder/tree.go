package recipeFinder

// RecipeNode itu bentuk data JSON yang nanti kita kirim ke frontend.
// - Name: nama elemen (misal "Brick")
// - Children: daftar subtree resep bahan-bahannya.
//   kalau leaf (elemen dasar), Children bisa kosong atau nil.
type RecipeNode struct {
	Name     string        `json:"name"`
	Children []*RecipeNode `json:"children,omitempty"`
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
	if fb, _ := IndexedBFSBuild(name, GlobalIndexedGraph); len(fb) > 0 {
		if step, ok := fb[name]; ok {
			node.Children = []*RecipeNode{
				buildTreeRec(step.Combo.A, fb, visited, depth+1),
				buildTreeRec(step.Combo.B, fb, visited, depth+1),
			}
		}
	}
	return node
}