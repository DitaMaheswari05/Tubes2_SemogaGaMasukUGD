// backend/main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD/backend/recipeFinder"
)

const (
    jsonDir  = "json"
    jsonFile = jsonDir + "/recipe.json"
    svgDir   = "svgs"
)

var (
	recipeGraph recipeFinder.Graph
    doScrape = flag.Bool("scrape", false, "rebuild recipe.json by scraping") // kalau -scrape=true, jalankan scraping dulu
    addr     = flag.String("addr", ":8080", "listen address")                // alamat HTTP server
)

func main() {
    flag.Parse() // parse flags: doScrape sama addr

    // 1) Kalau dipanggil dengan flag -scrape, langsung scrape dan tulis ulang recipe.json
    if *doScrape {
        catalog, err := recipeFinder.ScrapeAll() // ambil data dari Fandom
        if err != nil {
            log.Fatalf("scrape failed: %v", err)
        }
        // pastikan folder json/ ada
        os.MkdirAll(jsonDir, 0755)
        // jadikan objek Go -> JSON ter-format
        raw, _ := json.MarshalIndent(catalog, "", "  ")
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

    // 3) "Unmarshal" JSON → struct Go kita
    //
    //    rawJSON adalah byte slice berisi konten file recipe.json,
    //    misalnya '[{"Tier 1 elements":[{"name":"Mud","recipes":[...]}], ...}]'
    //
    //    json.Unmarshal artinya: ambil data JSON mentah (rawJSON)
    //    lalu konversi ke tipe Go yang kita tentukan (di sini map[string][]Element).
    //    Hasilnya disimpan di variabel "sections".
    var catalog recipeFinder.Catalog
    if err := json.Unmarshal(rawJSON, &catalog); err != nil {
        // Kalau error, kita fatal: artinya data JSON-nya tidak valid / berbeda dengan definisi struct Element
        log.Fatalf("invalid JSON: %v", err)
    }

    // 4) Bangun index byPair: untuk tiap pasangan (A,B) daftar produk yang bisa dibuat
    // combinationMap := recipeFinder.BuildCombinationMap(catalog)
	// recipeGraph = recipeFinder.BuildGraphFromCatalog(catalog)
	indexedGraph := recipeFinder.BuildIndexedGraph(catalog)

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

    // Get full working directory path
    wd, err := os.Getwd()
    if err != nil {
        log.Fatalf("failed to get working directory: %v", err)
    }
    log.Printf("Current working directory: %s", wd)

    // Determine SVG path based on working directory
    var svgPath string
    if filepath.Base(wd) == "backend" {
        // Already in backend dir
        svgPath = svgDir
    } else {
        // Might be in repo root
        svgPath = filepath.Join("src", "backend", svgDir)
    }

    // Check if the path exists
    if _, err := os.Stat(svgPath); os.IsNotExist(err) {
        log.Fatalf("SVG directory not found at: %s", svgPath)
    }
    log.Printf("Serving SVGs from: %s", svgPath)

    // Serve SVGs with FileServer
    http.Handle("/svgs/", http.StripPrefix("/svgs/", http.FileServer(http.Dir(svgPath))))

    // 6) Endpoint /api/find?target=NamaElemen
    //    => jalanin BFSBuild (mencari jalur terpendek dari starting elements ke target)
    //    => terus build tree-nya (rekonstruksi recipe)
    http.HandleFunc("/api/find", func(w http.ResponseWriter, r *http.Request) {
        target := r.URL.Query().Get("target")
        if target == "" {
            http.Error(w, "missing ?target=", http.StatusBadRequest)
            return
        }

        var maxPaths int64 = 20
        multi := false

        var trees interface{} // Adjust type based on actual return type of BuildTrees

        if multi {
            pathPrev := recipeFinder.BFSBuildMulti(target, recipeGraph, maxPaths)
            trees = recipeFinder.BuildTrees(target, pathPrev)
        } else {
            prev := recipeFinder.IndexedBFSBuild(target, indexedGraph)
            trees = recipeFinder.BuildTree(target, prev)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(trees)
    })

    log.Printf("listening on %s…", *addr)
    log.Fatal(http.ListenAndServe(*addr, nil))
}