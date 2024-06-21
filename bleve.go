package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"
	"unicode"

	_ "embed"

	"github.com/blevesearch/bleve"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonLetters = regexp.MustCompile(`[^a-zA-Z0-9:<>= ]+`)

func clean(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	result = nonLetters.ReplaceAllString(result, "")
	return result
}

// go:embed items.json
// var itemsjson []byte
func search(q string) ([]MLBCard, error) {
	q = clean(q)

	mapping := bleve.NewIndexMapping()
	index, _ := bleve.NewUsing("", mapping, bleve.Config.DefaultIndexType,
		bleve.Config.DefaultMemKVStore, nil)
	defer index.Close()

	{
		b := index.NewBatch()
		func() {
			// items.mutex.Lock()
			// defer items.mutex.Unlock()

			log.Println(len(items.items), "items")
			for _, v := range items.items {
				// _ = b.Index(v.UUID, v)
				descr := fmt.Sprintf("%s (%v) | %s %v", clean(v.Name), v.Ovr, v.TeamShortName, v.Series)
				_ = fmt.Sprint(descr)
				_ = b.Index(v.UUID, struct {
					Description string `json:"description"`
					Name        string `json:"name"`
					Ovr         int    `json:"ovr"`
					Team        string `json:"team"`
				}{
					descr,
					clean(v.Name),
					v.Ovr,
					v.Team,
				})
				// descr = v.Name
				// _ = b.Index(v.UUID, descr)
			}
		}()
		log.Println("indexed", index.Batch(b))
	}

	log.Println("query", clean(q))
	results := []MLBCard{}

	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)
	req.SortBy([]string{"-_score", "-ovr"})

	searchResult, err := index.Search(req)
	if err != nil {
		slog.Error("search result error", "err", err)

		return nil, err
	}
	log.Println("-------------------------------------")
	// doc, _ := index.Document("1d7f7d5faea7d8528d45aeaf191868c1")
	// log.Printf("%#v\n", doc)
	log.Printf("results: %#v\n", searchResult.Hits.Len())
	log.Println("-------------------------------------")

	minScore := 1.5
	for _, v := range searchResult.Hits {
		pre := ""
		if v.Score < minScore {
			pre = "\t\t"
		}
		if v.Score >= minScore {
			results = append(results, items.items[v.ID])
		}
		log.Println(
			pre,
			// v.Index, "items.bleve"
			// v.ID, // mlb uuid
			v.Score,
			items.items[v.ID].Name,
			items.items[v.ID].Ovr,
			items.items[v.ID].Team,
			items.items[v.ID].SeriesYear,
			items.items[v.ID].Series,
		)
	}

	return results, nil
}

func bleveit() {
	func() {
		fil, _ := os.OpenFile("items.json", os.O_RDONLY, os.ModeTemporary)
		defer fil.Close()
		{
			_ = json.NewDecoder(fil).Decode(&items.items)
			// itemsjson = nil
			// log.Println(len(itemsjson), len(items.items))
			log.Println(len(items.items))
		}
	}()
	// search("ovr:>90")
	search("jd martinez")
	search("jose ramirez")
	search("cc sabathia")
	search("ronald acuna")
	search("braves ovr:>=90")
	search("travis darnaud")
}
