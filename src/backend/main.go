// backend/main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	recipeFinder.GlobalIndexedGraph = indexedGraph

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
    // In main.go, modify your /api/find handler:

	http.HandleFunc("/api/find", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing ?target=", http.StatusBadRequest)
			return
		}
	
		// Get maxPaths from query parameter with default fallback
		maxPaths := int64(5)
		if maxPathsParam := r.URL.Query().Get("maxPaths"); maxPathsParam != "" {
			if val, err := strconv.ParseInt(maxPathsParam, 10, 64); err == nil && val > 0 {
				maxPaths = val
			}
		}
		
		// Get search mode (multi or single path)
		multi := true
		if multiParam := r.URL.Query().Get("multi"); multiParam != "" {
			multi = multiParam == "true"
		}
		
		// Get algorithm type (bfs, dfs, bidirectional)
		algorithm := "bfs"  // Default to BFS
		if algoParam := r.URL.Query().Get("algorithm"); algoParam != "" {
			algorithm = algoParam
		}
	
		// Create response structure with timing
		type FindResponse struct {
			Tree      interface{} `json:"tree"`
			DurationMs float64    `json:"duration_ms"`
			Algorithm string      `json:"algorithm"`
		}
	
		var response FindResponse
		startTime := time.Now()
		
		// Handle different algorithms
		switch algorithm {
		case "dfs":
			// DFS algorithm placeholder
			if multi {
				response.Algorithm = "dfs"
				// Placeholder for multi-path DFS
				response.Tree = []*recipeFinder.RecipeNode{} // Empty result for now
			} else {
				response.Algorithm = "dfs"
				// Placeholder for single-path DFS
				prev := recipeFinder.IndexedBFSBuild(target, indexedGraph) // Using BFS as placeholder
				response.Tree = recipeFinder.BuildTree(target, prev)
			}
			
		case "bidirectional":
			// Bidirectional search placeholder
			if multi {
				response.Algorithm = "bidirectional"
				// Placeholder for multi-path bidirectional
				response.Tree = []*recipeFinder.RecipeNode{} // Empty result for now
			} else {
				response.Algorithm = "bidirectional"
				// Placeholder for single-path bidirectional
				prev := recipeFinder.IndexedBFSBuild(target, indexedGraph) // Using BFS as placeholder
				response.Tree = recipeFinder.BuildTree(target, prev)
			}
			
		default: // "bfs" or any other value defaults to BFS
			if multi {
				// Multi-path BFS
				response.Algorithm = "bfs"
				
				desired := int(maxPaths)   // user-requested unique recipe count
				batch := 20                // pull 20 raw paths per round
				
				var trees []*recipeFinder.RecipeNode
				printed := map[string]bool{} // tree-level de-dup
				skip := 0                    // global path index
				
			outer:
				for len(trees) < desired {
					infos := recipeFinder.RangePathsIndexed(
						indexedGraph.NameToID[target], skip, batch, indexedGraph)
					if len(infos) == 0 {     // search exhausted
						break
					}
					skip += len(infos)
					
					for _, info := range infos {
						// convert Info -> map -> tree
						single := make(recipeFinder.ProductToIngredients)
						
						for _, step := range info.Path {
							if len(step) == 3 {
								single[step[2]] = recipeFinder.RecipeStep{
									Combo: recipeFinder.IngredientCombo{
										A: step[0],
										B: step[1],
									},
								}
							}
						}
						
						tree := recipeFinder.BuildTree(target, single)
						
						keyBytes, _ := json.Marshal(tree)   // stable string key
						if printed[string(keyBytes)] {
							continue // duplicate visual recipe; skip
						}
						printed[string(keyBytes)] = true
						trees = append(trees, tree)
						if len(trees) == desired {
							break outer
						}
					}
				}
				
				response.Tree = trees
			} else {
				// Single-path BFS
				response.Algorithm = "bfs"
				
				prev := recipeFinder.IndexedBFSBuild(target, indexedGraph)
				response.Tree = recipeFinder.BuildTree(target, prev)
			}
		}
		
		// Calculate duration
		response.DurationMs = float64(time.Since(startTime).Microseconds()) / 1000
		
		// Dump raw JSON to file
		rawJSON, err := json.MarshalIndent(response, "", "  ")
		if err == nil {
			// Ensure directory exists
			os.MkdirAll(filepath.Join("json"), 0755)
			// Write to file
			err = os.WriteFile(filepath.Join("json", "queryResult.json"), rawJSON, 0644)
			if err != nil {
				log.Printf("Failed to write query result: %v", err)
			}
		}
		
		// Send response to client
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

    log.Printf("listening on %s…", *addr)
    log.Fatal(http.ListenAndServe(*addr, nil))
}