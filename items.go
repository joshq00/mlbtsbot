package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/blevesearch/bleve"
)

// const APIPathItems = `https://mlb24.theshow.com/apis/items.json?type=mlb_card`
var items = Cache[MLBCard]{}

type MLBCard struct {
	UUID                      string    `json:"uuid"`
	Type                      string    `json:"type"`
	Img                       string    `json:"img"`
	BakedImg                  string    `json:"baked_img"`
	ScBakedImg                any       `json:"sc_baked_img"`
	Name                      string    `json:"name"`
	Rarity                    string    `json:"rarity"`
	Team                      string    `json:"team"`
	TeamShortName             string    `json:"team_short_name"`
	Ovr                       int       `json:"ovr"`
	Series                    string    `json:"series"`
	SeriesTextureName         string    `json:"series_texture_name"`
	SeriesYear                int       `json:"series_year"`
	DisplayPosition           string    `json:"display_position"`
	DisplaySecondaryPositions string    `json:"display_secondary_positions"`
	JerseyNumber              string    `json:"jersey_number"`
	Age                       int       `json:"age"`
	BatHand                   string    `json:"bat_hand"`
	ThrowHand                 string    `json:"throw_hand"`
	Weight                    string    `json:"weight"`
	Height                    string    `json:"height"`
	Born                      string    `json:"born"`
	IsHitter                  bool      `json:"is_hitter"`
	Stamina                   int       `json:"stamina"`
	PitchingClutch            int       `json:"pitching_clutch"`
	HitsPerBf                 int       `json:"hits_per_bf"`
	KPerBf                    int       `json:"k_per_bf"`
	BbPerBf                   int       `json:"bb_per_bf"`
	HrPerBf                   int       `json:"hr_per_bf"`
	PitchVelocity             int       `json:"pitch_velocity"`
	PitchControl              int       `json:"pitch_control"`
	PitchMovement             int       `json:"pitch_movement"`
	ContactLeft               int       `json:"contact_left"`
	ContactRight              int       `json:"contact_right"`
	PowerLeft                 int       `json:"power_left"`
	PowerRight                int       `json:"power_right"`
	PlateVision               int       `json:"plate_vision"`
	PlateDiscipline           int       `json:"plate_discipline"`
	BattingClutch             int       `json:"batting_clutch"`
	BuntingAbility            int       `json:"bunting_ability"`
	DragBuntingAbility        int       `json:"drag_bunting_ability"`
	HittingDurability         int       `json:"hitting_durability"`
	FieldingDurability        int       `json:"fielding_durability"`
	FieldingAbility           int       `json:"fielding_ability"`
	ArmStrength               int       `json:"arm_strength"`
	ArmAccuracy               int       `json:"arm_accuracy"`
	ReactionTime              int       `json:"reaction_time"`
	Blocking                  int       `json:"blocking"`
	Speed                     int       `json:"speed"`
	BaserunningAbility        int       `json:"baserunning_ability"`
	BaserunningAggression     int       `json:"baserunning_aggression"`
	HitRankImage              string    `json:"hit_rank_image"`
	FieldingRankImage         string    `json:"fielding_rank_image"`
	Pitches                   []Pitches `json:"pitches"`
	Quirks                    []Quirks  `json:"quirks"`
	IsSellable                bool      `json:"is_sellable"`
	HasAugment                bool      `json:"has_augment"`
	AugmentText               any       `json:"augment_text"`
	AugmentEndDate            any       `json:"augment_end_date"`
	HasMatchup                bool      `json:"has_matchup"`
	Stars                     any       `json:"stars"`
	Trend                     any       `json:"trend"`
	NewRank                   int       `json:"new_rank"`
	HasRankChange             bool      `json:"has_rank_change"`
	Event                     bool      `json:"event"`
	SetName                   string    `json:"set_name"`
	IsLiveSet                 bool      `json:"is_live_set"`
	UIAnimIndex               int       `json:"ui_anim_index"`
}
type Pitches struct {
	Name     string `json:"name"`
	Speed    int    `json:"speed"`
	Control  int    `json:"control"`
	Movement int    `json:"movement"`
}
type Quirks struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Img         string `json:"img"`
}

func (v MLBCard) ToSearchable() (string, any) {
	descr := fmt.Sprintf("%s (%v) | %v %s %v", clean(v.Name), v.Ovr, v.SeriesYear, v.TeamShortName, v.Series)

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

type ItemsResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
	Items      []MLBCard `json:"items"`
}

func loadMLBCards() {
	_ = items.loadFromFile()

	defer log.Println("i exited")
	apiURL := os.Getenv("MLB_ITEMS_URL")
	page := 1

	for {
		result := ItemsResponse{}
		func() {
			log.Println("starting a load")
			defer func() {
				if err := recover(); err != nil {
					log.Println("recovering", err)
				}
			}()
			req, _ := http.NewRequest("GET", apiURL, nil)
			q := req.URL.Query()
			q.Add("page", strconv.Itoa(page))
			req.URL.RawQuery = q.Encode()

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("error getting items", err)
				page = 1
				return
			}
			defer resp.Body.Close()
			log.Println("got items page", page)

			_ = json.NewDecoder(resp.Body).Decode(&result)

			page = result.Page + 1

			for _, l := range result.Items {
				items.Set(l.UUID, l)
			}
		}()

		if page > result.TotalPages {
			page = 1
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetIndent("", "  ")

			_ = enc.Encode(items.All())

			log.Println("writing items.json")
			_ = os.WriteFile("items.json", buf.Bytes(), os.ModePerm)
			time.Sleep(loadItemsInterval)
		}
	}
}

func loadMLBCardsAsync() {
	defer log.Println("i exited")
	apiURL := os.Getenv("MLB_ITEMS_URL")
	loadPage := func(page int) (ItemsResponse, error) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovering", err)
			}
		}()
		result := ItemsResponse{}
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
		log.Println("items page", result.Page, "/", result.TotalPages)
		for _, l := range result.Items {
			items.Set(l.UUID, l)
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
		items.Save()
		time.Sleep(loadItemsInterval)
	}
}
