package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var listings = Cache[Listing]{}

type ListingsResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
	Listings   []Listing `json:"listings"`
}
type Item struct {
	UUID              string      `json:"uuid"`
	Type              string      `json:"type"`
	Img               string      `json:"img"`
	BakedImg          string      `json:"baked_img"`
	ScBakedImg        interface{} `json:"sc_baked_img"`
	Name              string      `json:"name"`
	Rarity            string      `json:"rarity"`
	Team              string      `json:"team"`
	TeamShortName     string      `json:"team_short_name"`
	Ovr               int         `json:"ovr"`
	Series            string      `json:"series"`
	SeriesTextureName string      `json:"series_texture_name"`
	SeriesYear        int         `json:"series_year"`
	DisplayPosition   string      `json:"display_position"`
	HasAugment        bool        `json:"has_augment"`
	AugmentText       interface{} `json:"augment_text"`
	AugmentEndDate    interface{} `json:"augment_end_date"`
	HasMatchup        bool        `json:"has_matchup"`
	Stars             interface{} `json:"stars"`
	Trend             interface{} `json:"trend"`
	NewRank           int         `json:"new_rank"`
	HasRankChange     bool        `json:"has_rank_change"`
	Event             bool        `json:"event"`
	SetName           string      `json:"set_name"`
	IsLiveSet         bool        `json:"is_live_set"`
	UIAnimIndex       int         `json:"ui_anim_index"`
}

type Listing struct {
	ListingName   string `json:"listing_name"`
	BestSellPrice int    `json:"best_sell_price"`
	BestBuyPrice  int    `json:"best_buy_price"`
	Item          Item   `json:"item"`
}

func (l Listing) ToSearchable() (string, any) {
	v := l.Item
	descr := fmt.Sprintf("%s (%v) | %v %s %v", clean(v.Name), v.Ovr, v.SeriesYear, v.TeamShortName, v.SeriesYear)

	return v.UUID, struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Ovr         int    `json:"ovr"`
		Team        string `json:"team"`
	}{
		descr,
		clean(v.Name),
		v.Ovr,
		v.Team,
	}
}

func loadListingsasync() {
	_ = listings.loadFromFile()
	defer log.Println("i exited")
	// p := 1
	apiURL := os.Getenv("MLB_LISTINGS_URL")
	loadPage := func(page int) (ListingsResponse, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovering", err)
			}
		}()
		result := ListingsResponse{}
		// log.Println("starting a load")
		req, _ := http.NewRequest("GET", apiURL, nil)
		q := req.URL.Query()
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("error getting listings", err)
			return result, err
		}
		defer resp.Body.Close()
		log.Println("got page", page)

		_ = json.NewDecoder(resp.Body).Decode(&result)
		log.Println("result", result.Page, result.TotalPages)
		for _, l := range result.Listings {
			listings.Set(l.Item.UUID, l)
		}
		return result, err
	}

	wg := sync.WaitGroup{}
	for {
		func() {
			p1, err := loadPage(1)
			if err != nil {
				log.Println(err)
				return
			}
			for i := p1.Page + 1; i <= p1.TotalPages; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					loadPage(i)
				}()
			}
		}()

		log.Println("waiting")
		wg.Wait()
		log.Println("done waiting")
		listings.Save()
		time.Sleep(loadListingsInterval)
	}
}

func loadListings() {
	defer log.Println("i exited")
	p := 1
	apiURL := os.Getenv("MLB_LISTINGS_URL")
	for {
		result := ListingsResponse{}
		func() {
			// log.Println("starting a load")
			defer func() {
				if err := recover(); err != nil {
					log.Println("recovering", err)
				}
			}()
			req, _ := http.NewRequest("GET", apiURL, nil)
			q := req.URL.Query()
			q.Add("page", strconv.Itoa(p))
			req.URL.RawQuery = q.Encode()

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("error getting listings", err)
				p = 1
				return
			}
			defer resp.Body.Close()

			_ = json.NewDecoder(resp.Body).Decode(&result)
			log.Println("listings page", result.Page, "/", result.TotalPages)
			for _, l := range result.Listings {
				listings.Set(l.Item.UUID, l)
			}
		}()
		p = result.Page + 1
		if p > result.TotalPages {
			p = 1
			time.Sleep(loadListingsInterval)
		}
	}
}
