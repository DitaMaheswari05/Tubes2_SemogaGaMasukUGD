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

// ScrapeAll crawls the Fandom page, downloads SVGs and builds the same
// map[string][]Element that you were saving as JSON before.
func ScrapeAll() (map[string][]Element, error) {
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
        tbl := hdr.Next()
        for tbl.Length() > 0 && !tbl.Is("table.list-table") {
            tbl = tbl.Next()
        }
        if tbl.Length() == 0 {
            return
        }

        dir := filepath.Join("svgs", strings.ReplaceAll(title, " ", "_"))
        os.MkdirAll(dir, 0755)

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

            href, _ := cols.Eq(0).Find("a.mw-file-description").Attr("href")
            localPath := ""
            if href != "" {
                fname := strings.ReplaceAll(name, " ", "_") + ".svg"
                localPath = filepath.Join(strings.ReplaceAll(title, " ", "_"), fname)
                // downloadSVG(href, filepath.Join(dir, fname))
            }

            recipes := [][]string{}
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
            out[title] = elems
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

    if _, err := io.Copy(f, resp.Body); err != nil {
        log.Printf("write %s: %v", dest, err)
    }
}
