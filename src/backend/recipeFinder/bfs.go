package recipeFinder

// Info merekam dua bahan (“Parent” dan “Partner”) pertama kali
// yang kita gabungkan untuk membuat suatu produk.
// Parent = bahan yang sedang kita “pop” dari queue (yang lagi kita proses)
// Partner = bahan lain (dari set seen) yang kita gabungkan bersama Parent untuk bikin produk itu.
type Info struct {
    Parent, Partner string
}

// Starting elements
var baseElements = []string{"Air", "Earth", "Fire", "Water"}

// BFSBuild menjalankan BFS mulai dari elemen dasar
// hingga menemukan `target` (atau habis opsi). Hasilnya adalah map
// dari produk ke Info (siapa Parent & Partner-nya), supaya kita bisa
// membangun kembali urutan resepnya nanti.

func BFSBuild(target string, byPair map[Pair][]string) map[string]Info {
    // 1) Siapkan antrean (queue) awal berisi elemen dasar
    queue := make([]string, len(baseElements))
    copy(queue, baseElements)

    // 2) seen = set untuk menandai elemen yang sudah kita kunjungi
    seen := make(map[string]bool, len(baseElements))
    for _, b := range baseElements {
        seen[b] = true
    }

    // 3) prev akan menyimpan, untuk tiap produk, Info{Parent, Partner}
    //    yang pertama kali menghasilkannya
    prev := make(map[string]Info)

    // 4) Loop BFS: selama masih ada di queue (i < len(queue))
    for i := 0; i < len(queue); i++ {
        cur := queue[i]

        // Jika cur sudah sama dengan target, kita bisa stop cari.
        if cur == target {
            break
        }

        // 5) Untuk setiap bahan “partner” yang sudah pernah dilihat (seen),
        //    coba gabungkan cur + partner, lihat byPair[pair] → daftar produk.
        for partner := range seen {
            // Pair{A: cur, B: partner} lookup produk apa saja
            for _, prod := range byPair[Pair{A: cur, B: partner}] {
                // 6) Kalau produk baru (belum seen), tandai dan enqueue
                if !seen[prod] {
                    seen[prod] = true
                    // catat siapa parent & partner-nya
                    prev[prod] = Info{Parent: cur, Partner: partner}
                    // tambahkan ke antrean untuk nanti diproses
                    queue = append(queue, prod)
                }
            }
        }
    }

    // 7) Kembalikan map prev. Dari sini kita bisa rekonstruksi jalur
    //    resep: mulai dari target, lihat prev[target], lalu prev[parent], dst.
    return prev
}
