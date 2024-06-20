package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/blevesearch/bleve"
)

const APIPathItems = `https://mlb24.theshow.com/apis/items.json?type=mlb_card`

type ItemsResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
	Items      []MLBCard `json:"items"`
}

type ItemsCache struct {
	mutex *sync.Mutex
	items map[string]MLBCard
}

var items = ItemsCache{
	&sync.Mutex{},
	map[string]MLBCard{},
}

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

func loadMLBCards() {
	defer log.Println("i exited")
	// apiURL := os.Getenv("MLB_LISTINGS_URL")
	apiURL := APIPathItems
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

			listings.mutex.Lock()
			defer listings.mutex.Unlock()
			for _, l := range result.Items {
				items.items[l.UUID] = l
			}
		}()

		if page > result.TotalPages {
			page = 1
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetIndent("", "  ")
			_ = enc.Encode(items.items)

			log.Println("writing items.json")
			_ = os.WriteFile("items.json", buf.Bytes(), os.ModePerm)
			time.Sleep(loadItemsInterval)
		}
	}
}
