package recipeFinder

// DFSBuild menjalankan DFS mulai dari elemen dasar
// hingga menemukan `target`. Hasilnya adalah map
// dari produk ke Info (Parent & Partner-nya),
// supaya kita bisa membangun kembali urutan resepnya nanti.

func DFSBuild(target string, combinationMap CombinationMap) map[string]Info {
	seen := make(map[string]bool)
	prev := make(map[string]Info)
	found := false

	for _, e := range baseElements {
		seen[e] = true
	}

	for _, start := range baseElements {
		if dfs(start, target, seen, prev, combinationMap, &found) {
			break
		}
	}

	return prev
}

func dfs(
	cur string,
	target string,
	seen map[string]bool,
	prev map[string]Info,
	combinationMap CombinationMap,
	found *bool,
) bool {
	if *found {
		return true
	}

	if cur == target {
		*found = true
		return true
	}

	for partner := range seen {
		combo := IngredientCombo{A: cur, B: partner}
		for _, prod := range combinationMap[combo] {
			if !seen[prod] {
				seen[prod] = true
				prev[prod] = Info{Parent: cur, Partner: partner}
				if dfs(prod, target, seen, prev, combinationMap, found) {
					return true
				}
			}
		}
	}

	return false
}
