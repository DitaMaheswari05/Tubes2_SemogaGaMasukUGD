package recipeFinder

// Pair: pasangan dua nama bahan.
type Pair struct {
	A, B string
}

// MakeByPair bikin map di mana:
//   key   = pasangan bahan (Pair{A,B})
//   value = list nama produk yang muncul dari kombinasi A+B
//
// Kita masukin juga (B,A) supaya nanti waktu ngecek
// gak usah mikir urutan, tinggal swap aja langsung ketemu.
func MakeByPair(sections map[string][]Element) map[Pair][]string {
	byPair := make(map[Pair][]string)

	// 1) loop tiap kategori (Starting elements, Tier1, dst):
	for _, elems := range sections {
		// 2) loop tiap element di kategori itu:
		for _, el := range elems {
			// 3) loop tiap recipe [bahan1, bahan2] di element tersebut:
			for _, rec := range el.Recipes {
				// jadinya rec bisa panjangnya bukan 2? skip aja
				if len(rec) != 2 {
					continue
				}
				a, b := rec[0], rec[1]         // misal ["Water","Earth"]
				p1 := Pair{A: a, B: b}         // Pair{A:"Water", B:"Earth"}
				p2 := Pair{A: b, B: a}         // Pair{A:"Earth", B:"Water"}

				// tambahkan produk (el.Name) ke daftar byPair[p1]
				byPair[p1] = append(byPair[p1], el.Name)
				// juga ke byPair[p2], biar komut byPair[{Earth,Water}] juga ada
				byPair[p2] = append(byPair[p2], el.Name)
			}
		}
	}

	return byPair
}
