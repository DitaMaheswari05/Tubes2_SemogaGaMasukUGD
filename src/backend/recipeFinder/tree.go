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
func BuildTree(name string, prev map[string]Info) *RecipeNode {
    // Create a node for the current element.
    node := &RecipeNode{Name: name}
    
    // If this is a base element, we’re done.
    if isBaseElement(name) {
        return node
    }
    
    // If we have recorded recipe info for this element in prev, use it.
    if info, ok := prev[name]; ok {
        node.Children = []*RecipeNode{
            BuildTree(info.Parent, prev),
            BuildTree(info.Partner, prev),
        }
        return node
    }
    
    // Otherwise, we have no recipe info for this element from the current BFS run.
    // As a fallback, try to run a simple BFS (or lookup) to fill this branch.
    fallback := IndexedBFSBuild(name, GlobalIndexedGraph) // indexedGraph should be globally available or passed as parameter.
    if len(fallback) > 0 {
        // Use the fallback result to continue building the tree.
        fbInfo := fallback[name]
        node.Children = []*RecipeNode{
            BuildTree(fbInfo.Parent, fallback),
            BuildTree(fbInfo.Partner, fallback),
        }
    }
    // If even the fallback doesn’t yield a recipe, we simply return the leaf.
    return node
}

func BuildTrees(target string, pathPrev map[string][]Info) []*RecipeNode {
	var trees []*RecipeNode

	for _, info := range pathPrev[target] {
		// Use only the steps from this specific path to build prev
		prev := make(map[string]Info)
		for _, step := range info.Path {
			if len(step) == 3 {
				parent, partner, product := step[0], step[1], step[2]
				prev[product] = Info{Parent: parent, Partner: partner}
			}
		}

		tree := BuildTree(target, prev)
		trees = append(trees, tree)
	}

	return trees
}

func isBaseElement(name string) bool {
	for _, b := range baseElements {
		if b == name {
			return true
		}
	}
	return false
}
