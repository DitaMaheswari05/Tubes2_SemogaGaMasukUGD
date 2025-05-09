package recipeFinder

// BuildIndexedGraph membuat representasi graf yang dioptimalkan dari katalog.
// Fungsi ini mengubah struktur Catalog menjadi IndexedGraph yang lebih efisien
// dengan menggunakan ID integer untuk mempercepat pencarian dan mengurangi penggunaan memori.
//
// Proses konversi dilakukan dalam dua tahap:
// 1. Tahap pertama: Menetapkan ID untuk setiap nama elemen, diutamakan elemen dasar
// 2. Tahap kedua: Membangun tepi graf yang menghubungkan elemen-elemen berdasarkan resep
//
// Parameter:
//   - cat: Struktur Catalog yang berisi seluruh data elemen dan resep
//
// Return:
//   - IndexedGraph: Representasi graf teroptimasi dengan ID integer
func BuildIndexedGraph(cat Catalog) IndexedGraph {
    // Tahap pertama: menetapkan ID untuk semua nama elemen
    nameToID := make(map[string]int)   // Memetakan nama elemen ke ID integer
    idToName := make(map[int]string)   // Memetakan ID integer kembali ke nama elemen
    
    // Mulai menetapkan ID secara berurutan
    nextID := 0
    
    // Pastikan elemen dasar (BaseElements) mendapatkan ID terlebih dahulu
    // Ini penting karena algoritma BFS dan DFS akan memulai dari elemen dasar
    baseIDs := make([]int, len(BaseElements))
    for i, name := range BaseElements {
        nameToID[name] = nextID
        idToName[nextID] = name
        baseIDs[i] = nextID
        nextID++
    }
    
    // Kemudian tetapkan ID untuk semua elemen lainnya dalam katalog
    // Kita menelusuri setiap tier dan elemen di dalamnya
    for _, tier := range cat.Tiers {
        for _, el := range tier.Elements {
            // Periksa apakah elemen sudah memiliki ID
            if _, exists := nameToID[el.Name]; !exists {
                nameToID[el.Name] = nextID
                idToName[nextID] = el.Name
                nextID++
            }
            
            // Juga tetapkan ID untuk semua nama bahan dalam resep
            // Ini memastikan semua elemen yang muncul dalam resep memiliki ID
            for _, rec := range el.Recipes {
                for _, ingredient := range rec {
                    if _, exists := nameToID[ingredient]; !exists {
                        nameToID[ingredient] = nextID
                        idToName[nextID] = ingredient
                        nextID++
                    }
                }
            }
        }
    }
    
    // Tahap kedua: membangun tepi graf berdasarkan resep
    // Setiap resep (A+B→C) akan ditambahkan sebagai tepi dari A ke B dengan hasil C
    // dan juga dari B ke A dengan hasil yang sama
    edges := make(map[int][]IndexedNeighbor)
    
    for _, tier := range cat.Tiers {
        for _, el := range tier.Elements {
            productID := nameToID[el.Name]  // ID dari produk (hasil kombinasi)
            
            for _, rec := range el.Recipes {
                // Pastikan resep terdiri dari 2 bahan
                if len(rec) != 2 {
                    continue
                }
                
                aID := nameToID[rec[0]]  // ID bahan pertama
                bID := nameToID[rec[1]]  // ID bahan kedua
                
                // Tambahkan tepi ke dalam graf untuk kedua arah
                // Karena A+B=C dan B+A=C adalah sama
                edges[aID] = append(edges[aID], IndexedNeighbor{
                    PartnerID: bID,        // Bahan kedua
                    ProductID: productID,  // Produk hasil
                })
                
                edges[bID] = append(edges[bID], IndexedNeighbor{
                    PartnerID: aID,        // Bahan pertama
                    ProductID: productID,  // Produk hasil
                })
            }
        }
    }
    
    // Mengembalikan struktur IndexedGraph yang sudah lengkap
    return IndexedGraph{
        NameToID:  nameToID,  // Pemetaan nama ke ID
        IDToName:  idToName,  // Pemetaan ID ke nama
        Edges:     edges,     // Tepi graf dengan ID
    }
}

// GetBaseElementIDs mengembalikan daftar ID integer untuk semua elemen dasar.
// Fungsi ini berguna untuk mengakses elemen dasar (Air, Earth, Fire, Water)
// dalam bentuk ID mereka, yang diperlukan untuk algoritma pencarian.
//
// Return:
//   - []int: Slice berisi ID integer untuk semua elemen dasar
func (g *IndexedGraph) GetBaseElementIDs() []int {
    ids := make([]int, len(BaseElements))
    for i, name := range BaseElements {
        ids[i] = g.NameToID[name]
    }
    return ids
}