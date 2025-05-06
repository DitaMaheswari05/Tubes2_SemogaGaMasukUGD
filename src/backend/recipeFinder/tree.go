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
func BuildTree(name string, prev map[string]Info) *RecipeNode {
	// bikin node untuk elemen sekarang
	node := &RecipeNode{Name: name}

	if isBaseElement(name) {
		return node
	}

	// cek: apakah elemen ini pernah dicatat di prev?
	// kalau iya, berarti dia dibuat dari kombinasi dua bahan
	if info, ok := prev[name]; ok {
		// info.Parent + info.Partner → name
		// jadi anak-anaknya adalah subtree Parent dan Partner
		node.Children = []*RecipeNode{
			BuildTree(info.Parent, prev),
			BuildTree(info.Partner, prev),
		}
	}
	// return subtree ini
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
