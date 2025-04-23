// main.go
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

type Element struct {
	Name           string     `json:"name"`
	LocalSVGPath   string     `json:"local_svg_path"`
	OriginalSVGURL string     `json:"original_svg_url"`
	Recipes        [][]string `json:"recipes"`
}

func main() {
	// 1) Scrape & write JSON
	data, err := scrapeAll()
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	// ensure data/dirs  
	if err := os.MkdirAll("json", 0755); err != nil {
		log.Fatal(err)
	}
	raw, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile("json/recipe.json", raw, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote json/recipe.json (%d sections)", len(data))

	// 2) HTTP handlers
	http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
	})

	// serve SVG files under /svgs/
	fs := http.FileServer(http.Dir("svgs"))
	http.Handle("/svgs/", http.StripPrefix("/svgs/", fs))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func scrapeAll() (map[string][]Element, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]Element)
	doc.Find("h3").Each(func(_ int, hdr *goquery.Selection) {
		title := hdr.Find("span.mw-headline").Text()
		if title == "" {
			return
		}

		// find next table.list-table sibling
		tbl := hdr.Next()
		for tbl.Length() > 0 && !tbl.Is("table.list-table") {
			tbl = tbl.Next()
		}
		if tbl.Length() == 0 {
			return
		}

		sectionKey := title
		dir := filepath.Join("svgs", strings.ReplaceAll(title, " ", "_"))
		os.MkdirAll(dir, 0755)

		var elems []Element
		// skip header row, iterate rows
		tbl.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return
			}
			cols := row.Find("td")
			if cols.Length() < 2 {
				return
			}
			name := cols.Eq(0).Find("a[title]").First().Text()
			if name == "" {
				return
			}

			// SVG link
			fileA := cols.Eq(0).Find("a.mw-file-description")
			href, _ := fileA.Attr("href")
			localPath := ""
			if href != "" {
				fname := strings.ReplaceAll(name, " ", "_") + ".svg"
				localPath = filepath.Join(strings.ReplaceAll(title, " ", "_"), fname)
				downloadSVG(href, filepath.Join(dir, fname))
			}

			// recipes
			recipes := [][]string{} // <-- fix here
			cols.Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
				parts := li.Find("a[title]").Map(func(_ int, a *goquery.Selection) string {
					return a.Text()
				})
				if len(parts) == 2 {
					recipes = append(recipes, []string{parts[0], parts[1]})
				}
			})

			elems = append(elems, Element{
				Name:           name,
				LocalSVGPath:   localPath,
				OriginalSVGURL: href,
				Recipes:        recipes,
			})
		})

		if len(elems) > 0 {
			out[sectionKey] = elems
		}
	})

	return out, nil
}

func downloadSVG(url, dest string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("svg GET %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		log.Printf("create %s: %v", dest, err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Printf("write %s: %v", dest, err)
	}
}
