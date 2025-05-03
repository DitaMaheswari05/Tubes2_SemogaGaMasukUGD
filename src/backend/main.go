// backend/main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD/backend/recipeFinder"
)

const (
	jsonDir  = "json"                   // folder tempat nyimpen recipe.json
	jsonFile = jsonDir + "/recipe.json" // path lengkap ke file JSON
)

var (
	doScrape = flag.Bool("scrape", false, "rebuild recipe.json by scraping") // kalau -scrape=true, jalankan scraping dulu
	addr     = flag.String("addr", ":8080", "listen address")                // alamat HTTP server
)

func main() {
	flag.Parse() // parse flags: doScrape sama addr

	// 1) Kalau dipanggil dengan flag -scrape, langsung scrape dan tulis ulang recipe.json
	if *doScrape {
		sections, err := recipeFinder.ScrapeAll() // ambil data dari Fandom
		if err != nil {
			log.Fatalf("scrape failed: %v", err)
		}
		// pastikan folder json/ ada
		os.MkdirAll(jsonDir, 0755)
		// jadikan objek Go -> JSON ter-format
		raw, _ := json.MarshalIndent(sections, "", "  ")
		// simpan di disk
		if err := os.WriteFile(jsonFile, raw, 0644); err != nil {
			log.Fatal(err)
		}
		log.Printf("wrote %s", jsonFile)
		return
	}

	// 2) Baca file JSON yang udah ada
	rawJSON, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("cannot read %s: %v\nRun with -scrape first.", jsonFile, err)
	}

	// 3) “Unmarshal” JSON → struct Go kita
	//
	//    rawJSON adalah byte slice berisi konten file recipe.json,
	//    misalnya '[{"Tier 1 elements":[{"name":"Mud","recipes":[...]}], ...}]'
	//
	//    json.Unmarshal artinya: ambil data JSON mentah (rawJSON)
	//    lalu konversi ke tipe Go yang kita tentukan (di sini map[string][]Element).
	//    Hasilnya disimpan di variabel “sections”.
	var sections map[string][]recipeFinder.Element
	if err := json.Unmarshal(rawJSON, &sections); err != nil {
		// Kalau error, kita fatal: artinya data JSON-nya tidak valid / berbeda dengan definisi struct Element
		log.Fatalf("invalid JSON: %v", err)
	}

	// 4) Bangun index byPair: untuk tiap pasangan (A,B) daftar produk yang bisa dibuat
	byPair := recipeFinder.MakeByPair(sections)

	// 5) Endpoint /api/recipes: langsung dump rawJSON-nya (UNTUK SEKARANG)
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		// CORS bebas
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(rawJSON)
	})

	// 6) Endpoint /api/find?target=NamaElemen
	//    => jalanin BFSBuild (mencari jalur terpendek dari starting elements ke target)
	//    => terus build tree-nya (rekonstruksi recipe)
	http.HandleFunc("/api/find", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing ?target=", http.StatusBadRequest)
			return
		}
		prev := recipeFinder.DFSBuild(target, byPair) // prev map: node ← parent-nya
		tree := recipeFinder.BuildTree(target, prev)  // bangun struktur pohon recipe
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tree) // kirim JSON ke client
	})

	log.Printf("listening on %s…", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
