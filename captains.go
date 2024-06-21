package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var captains = Cache[Captain]{}

type CaptainsResponse struct {
	Page          int       `json:"page"`
	PerPage       int       `json:"per_page"`
	TotalPages    int       `json:"total_pages"`
	TotalCaptains int       `json:"total_captains"`
	Captains      []Captain `json:"captains"`
}
type Attributes struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Boosts struct {
	Tier        string       `json:"tier"`
	Description string       `json:"description"`
	Attributes  []Attributes `json:"attributes"`
}
type Captain struct {
	UUID            string    `json:"uuid"`
	Img             string    `json:"img"`
	BakedImg        string    `json:"baked_img"`
	ScBakedImg      any       `json:"sc_baked_img"`
	Name            string    `json:"name"`
	DisplayPosition string    `json:"display_position"`
	Team            string    `json:"team"`
	Ovr             int       `json:"ovr"`
	AbilityName     string    `json:"ability_name"`
	AbilityDesc     string    `json:"ability_desc"`
	UpdateDate      time.Time `json:"update_date"`
	Boosts          []Boosts  `json:"boosts"`
}

func (v Captain) ToSearchable() (string, any) {
	descr := fmt.Sprintf("%s (%v) | %v %s", clean(v.Name), v.Ovr, v.AbilityName, v.AbilityDesc)

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

func loadCaptainsAsync() {
	captains.loadFromFile()
	defer log.Println("i exited")
	apiURL := os.Getenv("MLB_CAPTAINS_URL")
	loadPage := func(page int) (CaptainsResponse, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovering", err)
			}
		}()
		result := CaptainsResponse{}
		// log.Println("starting a load")
		req, _ := http.NewRequest("GET", apiURL, nil)
		q := req.URL.Query()
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("error getting items", err)
			return result, err
		}
		defer resp.Body.Close()

		_ = json.NewDecoder(resp.Body).Decode(&result)
		log.Println("captains page", result.Page, "/", result.TotalPages)
		for _, l := range result.Captains {
			captains.Set(l.UUID, l)
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
		captains.Save()
		time.Sleep(loadItemsInterval)
	}
}

func findcaptains(plyr MLBCard) []Captain {
	capts := []Captain{}
	for _, capt := range captains.All() {
		mtch := capt.Matches(plyr)
		if mtch {
			capts = append(capts, capt)
		}
	}
	return capts
}

func removes(orig string, drops ...string) string {
	orig = strings.ToLower(orig)
	for _, d := range drops {
		orig = strings.ReplaceAll(orig, d, "")
	}
	return strings.TrimSpace(orig)
}

func (c Captain) Matches(p MLBCard) bool {
	switch true {
	// Players from the X team
	case strings.Contains(c.AbilityDesc, "Players from the"):
		parts := strings.Split(c.AbilityDesc, " ")
		tmnm := parts[len(parts)-1]
		return strings.EqualFold(p.Team, tmnm)

	// Players in the X Series
	// X Series Players
	case strings.Contains(c.AbilityDesc, "Series"):
		want := removes(c.AbilityDesc, "players", "in", "the", "leagues", "series")
		got := removes(p.Series, "players", "in", "the", "leagues", "series")
		return strings.EqualFold(got, want)

	// hitters from the N decade
	// TODO: not series year
	case strings.Contains(c.AbilityDesc, "'s decade"):
		// Hitters from the N's decade
		want := removes(c.AbilityDesc, "hitters from the", "'s decade")
		decade, _ := strconv.Atoi(want)
		got := p.SeriesYear
		return got >= decade && got-decade < 10

	// switch hitters
	case strings.Contains(c.AbilityDesc, "Switch Hitters"):
		return strings.EqualFold(p.BatHand, "S") && p.IsHitter

		// X-handed
		// TODO: unclear
	case strings.Contains(c.AbilityDesc, "-handed hitters and pitchers"):
		// Hitters from the N's decade
		want := removes(c.AbilityDesc, "-handed hitters and pitchers", "eft", "ight")
		return (strings.EqualFold(p.BatHand, want) && p.IsHitter == false) ||
			(strings.EqualFold(p.ThrowHand, want) && p.IsHitter == true)

		// TODO
	case c.AbilityDesc == "Pitchers that have reached 45 Saves in a Season":
		return false

		// TODO
	case c.AbilityDesc == "Players born in Asia or who have represented an Asian national team ^c59^(boost includes Lars Nootbaar and Tommy Edman)^c50^":
		return false

	case c.AbilityDesc == "Hitters with over 84 Speed":
		return p.Speed > 84 && p.IsHitter
	case c.AbilityDesc == "Hitters with under 45 Speed":
		return p.Speed < 45 && p.IsHitter
	case c.AbilityDesc == "Hitters with under 60 Vision":
		return p.PlateVision < 60 && p.IsHitter
	case c.AbilityDesc == "Hitters with under 70 Power and Pitchers with under 75 K/9":
		return (p.PowerLeft < 70 && p.PowerRight < 70 && p.IsHitter) ||
			(p.KPerBf < 75 && p.IsHitter == false)
	case c.AbilityDesc == "Hitters with under 75 Fielding":
		return p.FieldingAbility < 75 && p.IsHitter
	case c.AbilityDesc == "Pitchers with under 65 BB/9":
		return (p.BbPerBf < 65 && p.IsHitter == false)
	default:
		return false
	}

	return false
}
