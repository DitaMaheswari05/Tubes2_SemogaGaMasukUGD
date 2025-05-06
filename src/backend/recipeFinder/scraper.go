// backend/recipeFinder/scraper.go

package recipeFinder

import (
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

type Tier struct {
  Name     string    `json:"name"`
  Elements []Element `json:"elements"`
}

type Catalog struct {
  Tiers []Tier `json:"tiers"`
}

func ScrapeAll() (Catalog, error) {
  resp, err := http.Get(baseURL)
  if err != nil {
    return Catalog{}, err
  }
  defer resp.Body.Close()

  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil {
    return Catalog{}, err
  }

  var catalog Catalog

  doc.Find("h3").Each(func(_ int, hdr *goquery.Selection) {
    rawTitle := hdr.Find("span.mw-headline").Text()
    if rawTitle == "" {
      return
    }
    // find the next table.list-table for this section
    tbl := hdr.Next()
    for tbl.Length() > 0 && !tbl.Is("table.list-table") {
      tbl = tbl.Next()
    }
    if tbl.Length() == 0 {
      return
    }

    // clean the tier name: strip "Tier " prefix and " elements"/" element" suffix
    tierName := cleanTierName(rawTitle)
    tierDir := filepath.Join("svgs", strings.ReplaceAll(tierName, " ", "_"))
    os.MkdirAll(tierDir, 0755)

    var elems []Element
    tbl.Find("tr").Each(func(i int, row *goquery.Selection) {
      if i == 0 {
        return // skip header
      }
      cols := row.Find("td")
      if cols.Length() < 2 {
        return
      }
      name := cols.Eq(0).Find("a[title]").First().Text()
      if name == "" {
        return
      }

      // SVG
      fileA := cols.Eq(0).Find("a.mw-file-description")
      href, _ := fileA.Attr("href")
      local := ""
      if href != "" {
        fname := strings.ReplaceAll(name, " ", "_") + ".svg"
        local = filepath.Join(strings.ReplaceAll(tierName, " ", "_"), fname)
        // downloadSVG(href, filepath.Join(tierDir, fname))
      }

      // recipes
      recipes := make([][]string, 0)
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
        LocalSVGPath:   local,
        OriginalSVGURL: href,
        Recipes:        recipes,
      })
    })

    if len(elems) > 0 {
      catalog.Tiers = append(catalog.Tiers, Tier{
        Name:     tierName,
        Elements: elems,
      })
    }
  })

  return catalog, nil
}

func cleanTierName(raw string) string {
  s := raw
  // strip "Tier " prefix
  s = strings.TrimPrefix(s, "Tier ")
  // strip " elements" or " element" suffix
  s = strings.TrimSuffix(s, " elements")
  s = strings.TrimSuffix(s, " element")
  return s
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
  io.Copy(f, resp.Body)
}