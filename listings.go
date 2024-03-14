package main

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
