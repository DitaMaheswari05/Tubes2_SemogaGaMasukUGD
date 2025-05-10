package recipeFinder

// RecipeNode itu bentuk data JSON yang nanti kita kirim ke frontend.
// - Name: nama elemen (misal "Brick")
// - Children: daftar subtree resep bahan-bahannya.
//   kalau leaf (elemen dasar), Children bisa kosong atau nil.
type RecipeNode struct {
	Name     string        `json:"name"`
	Children []*RecipeNode `json:"children,omitempty"`
}

// BuildTree bangun pohon resep mulai dari `name` (target akhir),
// pakai peta prev hasil BFSBuild yang isinya:
//   product → Info{Parent, Partner}
// artinya product itu pertama kali dibuat dari Parent+Partner.
// Ini rekursif:
// 1) bikin node buat elemen `name`.
// 2) kalau ada prev[name], berarti dia bukan base,
//    turun (recursively) buat Parent dan Partner.
// 3) gabung hasilnya jadi Children dua buah subtree.
// BuildTree builds a recipe tree for the given element using the recipe info from prev.
// If no recipe info for an element is found (and it isn’t a base element), 
// use a fallback BFS to try to fill in that branch until a base element is reached.
// BuildTree builds a recipe tree for the given element using the recipe info from prev.
// func BuildTree(name string, prev ProductToIngredients) *RecipeNode {
//     // Create a node for the current element.
//     node := &RecipeNode{Name: name}
    
//     // If this is a base element, we're done.
//     if isBaseElement(name) {
//         return node
//     }
    
//     // If we have recorded recipe info for this element in prev, use it.
//     if recipeStep, ok := prev[name]; ok {
//         node.Children = []*RecipeNode{
//             BuildTree(recipeStep.Combo.A, prev),
//             BuildTree(recipeStep.Combo.B, prev),
//         }
//         return node
//     }
    
//     // Otherwise, we have no recipe info for this element from the current BFS run.
//     // As a fallback, try to run a simple BFS (or lookup) to fill this branch.
//     fallback, _ := IndexedBFSBuild(name, GlobalIndexedGraph)
//     if len(fallback) > 0 {
//         // Use the fallback result to continue building the tree.
//         fbRecipe := fallback[name]
//         node.Children = []*RecipeNode{
//             BuildTree(fbRecipe.Combo.A, fallback),
//             BuildTree(fbRecipe.Combo.B, fallback),
//         }
//     }
    
//     // If even the fallback doesn't yield a recipe, we simply return the leaf.
//     return node
// }

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