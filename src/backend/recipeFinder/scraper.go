// backend/recipeFinder/scraper.go
// Web scraper for Little Alchemy 2 wiki that extracts element recipes and images

package recipeFinder

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery" // HTML parsing library
)

// Wiki URL containing all Little Alchemy 2 elements and their recipes
const baseURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

// Element represents a single item in Little Alchemy 2
// Each element has a name, SVG image paths, and recipes to create it
type Element struct {
	Name           string     `json:"name"`           // Element name (e.g., "Water", "Fire")
	LocalSVGPath   string     `json:"local_svg_path"` // Relative path to locally saved SVG
	OriginalSVGURL string     `json:"original_svg_url"` // Original URL of the element's image
	Recipes        [][]string `json:"recipes"`        // List of ingredient pairs that make this element
}

// Tier represents a group of elements of similar complexity
// In Little Alchemy 2, elements are organized in tiers (1-13 + special tiers)
type Tier struct {
	Name     string    `json:"name"`     // Tier name (e.g., "1", "2", "Starting")
	Elements []Element `json:"elements"` // Elements belonging to this tier
}

// Catalog is the root data structure containing all tiers and elements
// This will be serialized to JSON as the game's recipe database
type Catalog struct {
	Tiers []Tier `json:"tiers"` // List of all tiers in the game
}

// ScrapeAll retrieves and parses the entire Little Alchemy 2 wiki
// It extracts all elements, their recipes, and image references
// Returns a complete Catalog and any errors encountered during scraping
func ScrapeAll() (Catalog, error) {
	// First make an HTTP GET request to the wiki page
	resp, err := http.Get(baseURL)
	if err != nil {
		return Catalog{}, err
	}
	defer resp.Body.Close()

	// Parse the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Catalog{}, err
	}

	var catalog Catalog

	// Find all h3 headers which divide elements into tiers
	doc.Find("h3").Each(func(_ int, hdr *goquery.Selection) {
    	// Extract tier title from headline span
		rawTitle := hdr.Find("span.mw-headline").Text()
		if rawTitle == "" {
			return // Skip headers without titles
		}
    
		// Find the first table after this header
		// This table contains all elements in this tier
		tbl := hdr.Next()
		for tbl.Length() > 0 && !tbl.Is("table.list-table") {
			tbl = tbl.Next()
		}
		if tbl.Length() == 0 {
			return // No table found for this tier
		}

		// Process the tier name and prepare directory for SVG files
		tierName := cleanTierName(rawTitle)

		// Skip the "Special" tier entirely
		if tierName == "Special" {
			return // Skip this tier and continue to the next h3
		}

		tierDir := filepath.Join("svgs", strings.ReplaceAll(tierName, " ", "_"))
		os.MkdirAll(tierDir, 0755) // Create directory if it doesn't exist

		// Extract each element from table rows
		var elems []Element
		tbl.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // Skip table header row
			}
		
			// Get all columns in this row
			cols := row.Find("td")
			if cols.Length() < 2 {
				return // Skip rows without enough columns
			}
		
			// Extract element name from the first column
			name := cols.Eq(0).Find("a[title]").First().Text()
			if name == "" {
				return // Skip unnamed elements
			}

			// Extract SVG image link
			fileA := cols.Eq(0).Find("a.mw-file-description")
			href, _ := fileA.Attr("href")
			local := ""
			if href != "" {
				// Create local path for SVG (not actually downloading in this code)
				fname := strings.ReplaceAll(name, " ", "_") + ".svg"
				local = filepath.Join(strings.ReplaceAll(tierName, " ", "_"), fname)
				// downloadSVG(href, filepath.Join(tierDir, fname)) // Commented out
			}

			// Extract all recipes from the second column
			recipes := make([][]string, 0)
			cols.Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
				// Each recipe is a pair of element links
				parts := li.Find("a[title]").Map(func(_ int, a *goquery.Selection) string {
					return a.Text()
				})
				if len(parts) == 2 {
					recipes = append(recipes, []string{parts[0], parts[1]})
				}
			})

			// Create Element struct with all collected data
			elems = append(elems, Element{
				Name:           name,
				LocalSVGPath:   local,
				OriginalSVGURL: href,
				Recipes:        recipes,
			})
		})

		// Add this tier to the catalog if it has elements
		if len(elems) > 0 {
			catalog.Tiers = append(catalog.Tiers, Tier{
				Name:     tierName,
				Elements: elems,
			})
		}
	})

  return catalog, nil
}

// cleanTierName normalizes tier names from the wiki format
// Removes "Tier " prefix and " elements"/" element" suffix
// For example: "Tier 1 elements" becomes simply "1"
func cleanTierName(raw string) string {
	s := raw
	s = strings.TrimPrefix(s, "Tier ") // Remove "Tier " prefix
	s = strings.TrimSuffix(s, " elements") // Remove plural suffix
	s = strings.TrimSuffix(s, " element")  // Remove singular suffix
	return s
}

// downloadSVG downloads an SVG image from a URL and saves it to a local path
// This function is used to create a local cache of all element images
// Note: This function handles errors internally and logs them but doesn't return them
func downloadSVG(url, dest string) {
	// Download the SVG file
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("svg GET %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
  
	// Create the destination file
	f, err := os.Create(dest)
	if err != nil {
		log.Printf("create %s: %v", dest, err)
		return
	}
	defer f.Close()
  
	// Copy the downloaded content to the file
	io.Copy(f, resp.Body)
}