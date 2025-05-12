// backend/main.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD/backend/recipeFinder"
)

// -----------------------------------------------------------------------------
// Constants
// -----------------------------------------------------------------------------
const (
	jsonDir  = "json"                   // directory for recipe.json & query results
	jsonFile = jsonDir + "/recipe.json" // full path to recipe.json
	svgDir   = "svgs"                   // directory for SVG icons for frontend
)

// -----------------------------------------------------------------------------
// Command line flags
// -----------------------------------------------------------------------------
var (
	// If -scrape flag is set, run Fandom web-scraper and write new recipe.json
	doScrape = flag.Bool("scrape", false, "rebuild recipe.json by scraping")
	// HTTP server address & port
	addr = flag.String("addr", ":8080", "listen address")
)

func main() {
	flag.Parse() // parse all flags above

	var rawJSON []byte
	var err error

	// ---------------------------------------------------------------------
	// 1) Run scraper if requested
	// ---------------------------------------------------------------------
	if *doScrape {
		// Scrape and assign directly to global catalog
		recipeFinder.GlobalCatalog, err = recipeFinder.ScrapeAll()
		if err != nil {
			log.Fatalf("scrape failed: %v", err)
		}

		os.MkdirAll(jsonDir, 0o755) // ensure directory exists
		rawJSON, _ = json.MarshalIndent(recipeFinder.GlobalCatalog, "", "  ")
		if err := os.WriteFile(jsonFile, rawJSON, 0o644); err != nil {
			log.Fatal(err)
		}

		log.Printf("wrote %s", jsonFile)
	} else {
		// ---------------------------------------------------------------------
		// 2) Read existing recipe.json
		// ---------------------------------------------------------------------
		rawJSON, err = os.ReadFile(jsonFile)
		// if err != nil {
		// 	log.Fatalf("cannot read %s: %v\nRun with -scrape first.", jsonFile, err)
		// }

		// // ---------------------------------------------------------------------
		// // 3) Parse JSON → Catalog struct (only if we didn't just scrape)
		// // ---------------------------------------------------------------------
		// if err := json.Unmarshal(rawJSON, &recipeFinder.GlobalCatalog); err != nil {
		// 	log.Fatalf("invalid JSON: %v", err)
		// }
	}

	// Sort tiers in catalog - "Starting" first, then numeric tiers in order
	sortCatalogTiers(&recipeFinder.GlobalCatalog)

	recipeFinder.InitElementTiers(recipeFinder.GlobalCatalog)

	// ---------------------------------------------------------------------
	// 4) Build indexed graph for fast searches and save globally
	// ---------------------------------------------------------------------
	indexedGraph := recipeFinder.BuildIndexedGraph(recipeFinder.GlobalCatalog)
	recipeFinder.GlobalIndexedGraph = indexedGraph // can be accessed by other packages

	// ---------------------------------------------------------------------
	// 5) Static endpoint: /api/recipes — send raw catalog to frontend
	// ---------------------------------------------------------------------
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(rawJSON)
	})

	// ---------------------------------------------------------------------
	// 6) Static file server for SVG icons (/svgs/...)
	// ---------------------------------------------------------------------
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)

	var svgPath string
	if filepath.Base(wd) == "backend" {
		svgPath = svgDir
	} else {
		svgPath = filepath.Join("src", "backend", svgDir)
	}
	if _, err := os.Stat(svgPath); os.IsNotExist(err) {
		log.Fatalf("SVG directory not found at: %s", svgPath)
	}
	log.Printf("Serving SVGs from: %s", svgPath)
	http.Handle("/svgs/", http.StripPrefix("/svgs/", http.FileServer(http.Dir(svgPath))))

	// ---------------------------------------------------------------------
	// 7) Recipe search endpoint: /api/find?target=Name&maxPaths=5&multi=true
	// ---------------------------------------------------------------------
	type FindResponse struct {
		Tree         interface{} `json:"tree"`
		DurationMs   float64     `json:"duration_ms"`
		Algorithm    string      `json:"algorithm"`
		NodesVisited int         `json:"nodes_visited"`
		SearchSteps  interface{} `json:"search_steps,omitempty"` // Use interface{} for flexibility
	}

	http.HandleFunc("/api/find", func(w http.ResponseWriter, r *http.Request) {
		// ---------- parameter validation ----------
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing ?target=", http.StatusBadRequest)
			return
		}

		// maxPaths (default 5)
		maxPaths := int64(5)
		if v := r.URL.Query().Get("maxPaths"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
				maxPaths = n
			}
		}
		// multi=true/false
		multi := true
		if v := r.URL.Query().Get("multi"); v != "" {
			multi = v == "true"
		}
		// algorithm=bfs|dfs|bidirectional
		algo := r.URL.Query().Get("algorithm")
		if algo == "" {
			algo = "bfs"
		}

		resp := FindResponse{Algorithm: algo}
		t0 := time.Now()

		// ---------- choose algorithm ----------
		switch algo {
		//-----------------------------------------------------------------
		case "dfs":
			recipeFinder.BuildReverseIndex(recipeFinder.GlobalIndexedGraph)
			if multi {
				// Get N unique paths (multi DFS)
				effectiveMaxPaths := int(maxPaths) * 2
				steps, nodes := recipeFinder.RangeDFSPaths(target, effectiveMaxPaths, recipeFinder.GlobalIndexedGraph)
				resp.NodesVisited = nodes
				trees := stepsToTrees(target, steps)

				// Apply tree-based deduplication (just like in BFS)
				if len(trees) > 0 {
					trees = recipeFinder.DeduplicateRecipeTrees(trees)
				}

				resp.Tree = trees
			} else {
				// Single path (single DFS)
				rec, nodes := recipeFinder.DFSBuildTargetToBase(target, recipeFinder.GlobalIndexedGraph)
				resp.NodesVisited = nodes
				resp.Tree = recipeFinder.BuildTree(target, rec)
			}

		//-----------------------------------------------------------------
		case "bidirectional": // placeholder
			if multi {
				resp.Tree = []*recipeFinder.RecipeNode{}
			} else {
				prev, _, nodes := recipeFinder.IndexedBFSBuild(target, recipeFinder.GlobalIndexedGraph)
				resp.NodesVisited = nodes
				resp.Tree = recipeFinder.BuildTree(target, prev)
			}

		//-----------------------------------------------------------------
		default: // bfs
			if multi {
				desired := int(maxPaths)
				batch := 4 // get 20 paths per iteration
				printed := map[string]bool{}
				var trees []*recipeFinder.RecipeNode
				skip := 0
				for len(trees) < desired*2 && skip < 20 {
					infos, nodes := recipeFinder.RangePathsIndexed(recipeFinder.GlobalIndexedGraph.NameToID[target], skip, batch, recipeFinder.GlobalIndexedGraph)
					resp.NodesVisited += nodes
					if len(infos) == 0 {
						break
					}
					skip += len(infos)
					t := infosToTrees(target, infos, printed)
					trees = append(trees, t...)
				}
				if len(trees) > 0 {
					// Apply tree-based deduplication
					trees = recipeFinder.DeduplicateRecipeTrees(trees)
					resp.Tree = trees
				}
				resp.Tree = trees
			} else {
				prev, searchSteps, nodes := recipeFinder.IndexedBFSBuild(target, recipeFinder.GlobalIndexedGraph)
				resp.NodesVisited = nodes
				resp.Tree = recipeFinder.BuildTree(target, prev)
				resp.SearchSteps = searchSteps

				// Add this before sending response
				log.Printf("SearchSteps length: %d", len(searchSteps))
				if len(searchSteps) > 0 {
					stepJSON, _ := json.Marshal(searchSteps[0])
					log.Printf("First step sample: %s", string(stepJSON))
				}
			}
		}

		// ---------- write response ----------

		resp.DurationMs = float64(time.Since(t0).Microseconds()) / 1000.0
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		// save to file for easy checking/debugging
		raw, _ := json.MarshalIndent(resp, "", "  ")
		os.MkdirAll(jsonDir, 0o755)
		_ = os.WriteFile(filepath.Join(jsonDir, "queryResult.json"), raw, 0o644)
	})

	// ---------------------------------------------------------------------
	// 8) Recipe scrape endpoint: /api/scrape
	// ---------------------------------------------------------------------
	http.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Scrape requested via API")

		// Run the same scraping code as with the -scrape flag
		catalog, err := recipeFinder.ScrapeAll()
		if err != nil {
			log.Printf("API scrape failed: %v", err)
			http.Error(w, "Failed to scrape data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update global catalog
		recipeFinder.GlobalCatalog = catalog

		// Sort catalog tiers
		sortCatalogTiers(&recipeFinder.GlobalCatalog)

		// Save to file
		os.MkdirAll(jsonDir, 0o755)
		rawJSON, _ = json.MarshalIndent(recipeFinder.GlobalCatalog, "", "  ")
		if err := os.WriteFile(jsonFile, rawJSON, 0o644); err != nil {
			log.Printf("Failed to write scraped data: %v", err)
			http.Error(w, "Failed to save scraped data", http.StatusInternalServerError)
			return
		}

		// Re-initialize with new data
		recipeFinder.InitElementTiers(recipeFinder.GlobalCatalog)

		// Rebuild indexed graph
		indexedGraph = recipeFinder.BuildIndexedGraph(recipeFinder.GlobalCatalog)
		recipeFinder.GlobalIndexedGraph = indexedGraph

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "success",
			"message":        "Scraping completed successfully",
			"elements_count": len(catalog.Tiers),
		})
	})

	// ---------------------------------------------------------------------
	// 9) Run server
	// ---------------------------------------------------------------------
	log.Printf("listening on %s…", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
-------------------------------------------------------------------------

	Helper transform: steps -> tree slice (used in multi-DFS/BFS)
*/
func stepsToTrees(target string, steps []recipeFinder.RecipeStep) []*recipeFinder.RecipeNode {
	var trees []*recipeFinder.RecipeNode
	printed := map[string]bool{}
	for _, step := range steps {
		single := make(recipeFinder.ProductToIngredients)
		for _, p := range step.Path {
			if len(p) == 3 {
				single[p[2]] = recipeFinder.RecipeStep{Combo: recipeFinder.IngredientCombo{A: p[0], B: p[1]}}
			}
		}
		tree := recipeFinder.BuildTree(target, single)
		key, _ := json.Marshal(tree)
		if !printed[string(key)] {
			printed[string(key)] = true
			trees = append(trees, tree)
		}
	}
	return trees
}

func infosToTrees(target string, infos []recipeFinder.RecipeStep, printed map[string]bool) []*recipeFinder.RecipeNode {
	var out []*recipeFinder.RecipeNode
	for _, info := range infos {
		single := make(recipeFinder.ProductToIngredients)
		for _, p := range info.Path {
			if len(p) == 3 {
				single[p[2]] = recipeFinder.RecipeStep{Combo: recipeFinder.IngredientCombo{A: p[0], B: p[1]}}
			}
		}
		tree := recipeFinder.BuildTree(target, single)
		key, _ := json.Marshal(tree)
		if !printed[string(key)] {
			printed[string(key)] = true
			out = append(out, tree)
		}
	}
	return out
}

// sortCatalogTiers sorts the tiers in catalog - "Starting" first, then numeric tiers in order
func sortCatalogTiers(catalog *recipeFinder.Catalog) {
	// "Starting" tier always comes first, then numeric tiers in order
	sort.Slice(catalog.Tiers, func(i, j int) bool {
		// "Starting" tier always comes first
		if catalog.Tiers[i].Name == "Starting" {
			return true
		}
		if catalog.Tiers[j].Name == "Starting" {
			return false
		}

		// For numeric tiers, sort by number
		iNum, _ := strconv.Atoi(catalog.Tiers[i].Name)
		jNum, _ := strconv.Atoi(catalog.Tiers[j].Name)
		return iNum < jNum
	})
}
