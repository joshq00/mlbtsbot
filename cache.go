package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/blevesearch/bleve"
)

type Cache[T searchable] struct {
	mutex *sync.Mutex
	items map[string]T
}

func (c *Cache[T]) Get(id string) (T, bool) {
	if c.items == nil {
		var t T
		return t, false
	}
	i, ok := c.items[id]
	return i, ok
}
func (c *Cache[T]) Set(id string, v T) {
	if c.mutex == nil {
		c.mutex = &sync.Mutex{}
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.items == nil {
		c.items = map[string]T{}
	}
	c.items[id] = v
}
func (c *Cache[T]) All() map[string]T {
	return c.items
}

type searchable interface {
	ToSearchable() (string, any)
}

func (c *Cache[T]) Search(q string) ([]T, error) {
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

			log.Println(len(c.items), "items")
			for _, v := range c.All() {
				id, val := v.ToSearchable()
				_ = b.Index(id, val)
			}
		}()
		log.Println("indexed", index.Batch(b))
	}

	log.Println("query", clean(q))
	results := []T{}

	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)
	req.SortBy([]string{"-_score", "-ovr"})

	searchResult, err := index.Search(req)
	if err != nil {
		slog.Error("search result error", "err", err)

		return nil, err
	}

	minScore := 1.5

	log.Println("-------------------------------------")
	// doc, _ := index.Document("1d7f7d5faea7d8528d45aeaf191868c1")
	// log.Printf("%#v\n", doc)
	log.Printf("results: %#v\n", searchResult.Hits.Len())
	log.Println("-------------------------------------")

	for _, v := range searchResult.Hits {
		pre := ""
		if v.Score < minScore {
			pre = "\t\t"
		}

		item, _ := c.Get(v.ID)
		if v.Score >= minScore {
			results = append(results, item)
		}

		searched, _ := index.Document(v.ID)
		log.Printf("%s %.03f %s\n", pre, v.Score, string(searched.Fields[0].Value()))
	}

	return results, nil
}

func (c *Cache[T]) Save() {
	filnm := c.filname()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")

	_ = enc.Encode(c.All())

	log.Println("writing", filnm)
	_ = os.WriteFile(filnm, buf.Bytes(), os.ModePerm)
}

func (c *Cache[T]) loadFromFile() error {
	dat, err := os.ReadFile(c.filname())
	if err != nil {
		log.Println("cant read items from file")
		return err
	}

	return json.Unmarshal(dat, &c.items)
}

func (c *Cache[T]) filname() string {
	return strings.ToLower(
		fmt.Sprintf("%s.json", c.typ()))
}
func (c *Cache[T]) typ() string {
	var t T
	typeStr := fmt.Sprintf("%T", t)
	parts := strings.Split(typeStr, ".")
	structType := parts[len(parts)-1]
	log.Printf("%v", structType)
	return structType
}
