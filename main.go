package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"
)

type ListingCache struct {
	mutex    *sync.Mutex
	listings map[string]Listing
}

var listings = ListingCache{
	&sync.Mutex{},
	map[string]Listing{},
}

func loadall() {
	defer log.Println("i exited")
	p := 1
	apiURL := os.Getenv("MLB_LISTINGS_URL")
	for {
		result := ListingsResponse{}
		func() {
			log.Println("starting a load")
			defer func() {
				if err := recover(); err != nil {
					log.Println("recovering", err)
				}
			}()
			req, _ := http.NewRequest("GET", apiURL, nil)
			q := req.URL.Query()
			q.Add("page", strconv.Itoa(p))
			req.URL.RawQuery = q.Encode()

			log.Println("url", req.URL.String())

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("error getting listings", err)
				p = 1
				return
			}
			defer resp.Body.Close()
			log.Println("got page", p)

			_ = json.NewDecoder(resp.Body).Decode(&result)
			log.Println("result", result.Page, result.TotalPages)
			listings.mutex.Lock()
			defer listings.mutex.Unlock()
			for _, l := range result.Listings {
				listings.listings[l.Item.UUID] = l
			}
		}()
		p = result.Page + 1
		if p > result.TotalPages {
			p = 1
			time.Sleep(time.Second * 60)
		}
	}
}

func main() {
	http.DefaultClient.Timeout = time.Second * 15

	godotenv.Load()

	go loadall()

	c := twitch.NewClient("joshq00", os.Getenv("TWITCH_OAUTH_TOKEN"))
	c.Join(strings.Split(os.Getenv("TWITCH_CHANNELS"), ",")...)
	c.OnPrivateMessage(func(m twitch.PrivateMessage) {
		log.Println(m.Channel, m.User.DisplayName, m.Message)
		if strings.HasPrefix(m.Message, "!price ") {
			playerName := strings.TrimPrefix(m.Message, "!price ")
			cards := findCard(playerName)

			for _, item := range cards {
				// fmt.Println(k, item.ListingName, item.Item.Series, item.BestBuyPrice, item.BestSellPrice)
				c.Say(m.Channel,
					fmt.Sprintf("[PRICE] %s (%v) | %s %v | Buy now: %v | Sell now: %v\n", item.ListingName, item.Item.Ovr, item.Item.TeamShortName, item.Item.Series, item.BestSellPrice, item.BestBuyPrice),
				)
			}

			if len(cards) == 0 {
				c.Say(m.Channel,
					"Player Card not found")
			}

		}

	})
	c.Connect()
}

func findCard(playerName string) []Listing {
	log.Println("looking for player", playerName)
	// call api with ?name=playername
	// parse the json into a ListingsResponse{}
	// filter the listings down to playerName
	// return the matching cards
	listings.mutex.Lock()
	defer listings.mutex.Unlock()
	ll := []Listing{}
	for _, v := range listings.listings {
		ll = append(ll, v)
	}

	return filterListings(playerName, ll)
}

func filterListings(needle string, haystack []Listing) []Listing {
	log.Println("filtering for player", needle)
	results := []Listing{}
	for _, item := range haystack {
		lcneedle, lcname := strings.ToLower(needle), strings.ToLower(item.ListingName)
		if strings.Contains(lcname, lcneedle) {
			results = append(results, item)
		}
	}
	return results
}

/*
!price PlayerName -> current buy now and sell now price for PlayerName
!price LS PlayerName -> live series prices
* */
