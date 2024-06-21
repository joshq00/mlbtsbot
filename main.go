package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/gempir/go-twitch-irc/v4"
)

const (
	loadListingsInterval = time.Minute * 10
	loadItemsInterval    = time.Minute * 30
)

func findListing(q string) {
	matches, _ := items.Search(q)
	ll := []Listing{}
	for _, v := range matches {
		if l, ok := listings.Get(v.UUID); ok {
			ll = append(ll, l)
			log.Println(l.BestSellPrice, l.BestBuyPrice)
		}
	}
}

func lineReader() {
	// results, err := items.Search("jose ramirez live guardians")
	//
	// log.Println(err)
	// for _, item := range results {
	// 	log.Println(
	// 		item.Name,
	// 		item.Ovr,
	// 		item.Team,
	// 		item.SeriesYear,
	// 		item.Series,
	// 	)
	// }
	// return

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text (press Ctrl+D or Ctrl+Z to exit):")
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("You entered:", line)
		{
			results, err := items.Search(line)
			log.Println(err)
			for _, item := range results {
				log.Println(
					item.Name,
					item.Ovr,
					item.Team,
					item.SeriesYear,
					item.Series,
				)
			}
		}
		{
			results, err := listings.Search(line)
			log.Println(err)
			for _, item := range results {
				log.Println(
					item.BestBuyPrice,
					item.BestSellPrice,
					item.Item.Name,
					item.Item.Ovr,
					item.Item.Team,
					item.Item.SeriesYear,
					item.Item.Series,
				)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func main() {
	captains.loadFromFile()
	listings.loadFromFile()
	items.loadFromFile()

	// findcaptains()
	// return
	// go loadCaptainsAsync()
	// go loadMLBCardsAsync()
	// go loadListingsasync()
	// // loadMLBCardsFromFile()
	// select {}
	// lineReader()
	// return

	// go loadMLBCards()
	// go loadMLBCardsAsync()
	// go loadListingsasync()

	// bleveit()
	// notmain()

	log.Println(os.Getenv("TWITCH_CHANNELS"))

	channels := strings.Split(os.Getenv("TWITCH_CHANNELS"), ",")

	clientID := os.Getenv("CLIENT_ID")
	log.Println("Get an access token\n",
		"https://id.twitch.tv/oauth2/authorize?client_id="+
			clientID+
			"&redirect_uri=http://localhost&response_type=token&scope=chat:read%20chat:edit")

	client := twitch.NewClient("joshq00", os.Getenv("TWITCH_OAUTH_TOKEN"))

	client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		fmt.Println(msg.Message)
	})
	client.OnConnect(func() {
		slog.Info("connected")
	})
	// client.SetRateLimiter(twitch.CreateVerifiedRateLimiter())
	client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		// log.Println(msg.Channel, msg.Message, msg.User.Name)
		// log.Printf("#%15s @%15s : %s\n", msg.Channel, msg.User.Name, msg.Message)
		slog.Debug("pm", "channel", msg.Channel, "user", msg.User.Name, "msg", msg.Message)
		// log.Printf("%#v\n", msg)
	})

	client.OnPrivateMessage(pmhandler(client))

	client.OnWhisperMessage(func(msg twitch.WhisperMessage) {})
	client.OnClearChatMessage(func(msg twitch.ClearChatMessage) {})
	client.OnClearMessage(func(msg twitch.ClearMessage) {
		log.Printf("[DELETED] #%s @%s : %s\n", msg.Channel, msg.Login, msg.Message)
	})
	client.OnRoomStateMessage(func(msg twitch.RoomStateMessage) {})
	client.OnUserNoticeMessage(func(msg twitch.UserNoticeMessage) {})
	client.OnUserStateMessage(func(msg twitch.UserStateMessage) {
		// log.Printf("%#v\n", msg)
	})
	client.OnGlobalUserStateMessage(func(msg twitch.GlobalUserStateMessage) {
		gusm = msg
	})
	client.OnNoticeMessage(func(msg twitch.NoticeMessage) {})
	client.OnUserJoinMessage(func(msg twitch.UserJoinMessage) {
		// fmt.Printf("@%s %sED %s\n", msg.User, "JOIN", msg.Channel)
	})
	client.OnUserPartMessage(func(msg twitch.UserPartMessage) {
		// fmt.Printf("@%s %sED %s\n", msg.User, "PART", msg.Channel)
	})

	client.Join(channels...)

	go func() {
		err := client.Connect()
		if err != nil {
			panic(err)
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	for _, v := range channels {
		client.Depart(v)
	}
	log.Println(client.Disconnect())
	time.Sleep(time.Second)
}

func notmain() {

	// go loadall()
	go loadMLBCards()

	/*
		go func() {
			_ = os.RemoveAll("items.bleve")
			mapping := bleve.NewIndexMapping()
			index, err := bleve.New("items.bleve", mapping)

			if err != nil {
				panic(err)
			}

			for {
				time.Sleep(time.Second * 5)

				func() {
					return
					items.mutex.Lock()
					defer items.mutex.Unlock()
					for k, v := range items.items {
						// log.Println(k, ":", v.Name, v.Ovr, v.SeriesYear, v.Series)
						// log.Println(k, ":", v.Name, v.Ovr, v.SeriesYear, v.Series)
						_ = fmt.Sprint(k, v)
					}
				}()

				func() {
					items.mutex.Lock()
					defer items.mutex.Unlock()

					log.Println(len(items.items), "items")
					for k, v := range items.items {
						if strings.HasPrefix(v.UUID, "27992e4bc24123be43724a326667f0dd") {
							log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
							log.Println("there's acuna", v)
							log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
						}
						err = index.Index(k, v)
						if err != nil {
							log.Println(err)
						}
					}
				}()
			}
		}()
		//*/

	go func() {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Println("recovering", err)
					}
				}()

				idx, _ := bleve.Open("items.bleve")
				defer idx.Close()
				query := bleve.NewQueryStringQuery("olson")
				req := bleve.NewSearchRequest(query)
				searchResult, _ := idx.Search(req)
				log.Println("-------------------------------------")
				// doc, _ := idx.Document("1d7f7d5faea7d8528d45aeaf191868c1")
				// log.Printf("%#v\n", doc)
				log.Printf("results %#v\n", searchResult)
				log.Printf("status %#v\n", searchResult.Status.Errors)
				log.Println("-------------------------------------")
				for _, v := range searchResult.Hits {
					log.Println(v.Index, v.ID, v.Score, v.String())
				}
			}()
			time.Sleep(time.Second * 10)
		}
	}()

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
	ll := []Listing{}
	for _, v := range listings.All() {
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
