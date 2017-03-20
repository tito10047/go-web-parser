package ladbrokes

import "time"

// all sports
type allSports struct {
	MostPopSportsGroup groups `json:"mostPopSportsGroup"`
	AllSportsGroup     groups `json:"allSportsGroup"`
}

type groups struct {
	List         []group `json:"list"`
	Title        string `json:"title"`
	Href         string `json:"href"`
	DeepLinkHref string `json:"deepLinkHref"`
	SortOrder    int `json:"sortOrder"`
	Id           string `json:"id"`
}

type group struct {
	Meta         string `json:"meta"`
	Text         string `json:"text"`
	Href         string `json:"href"`
	DeepLinkHref string `json:"deepLinkHref"`
	Icon         string `json:"icon"`
	Id           string `json:"id"`
	SortOrder    int `json:"sortOrder"`
	Event        bool `json:"eventDetail"`
}

//sport_competition_links  https://sports.ladbrokes.com/en-gb/sport_competition_links/110000006
type allCompetitions struct {
	LinkGroupsForClasses []groups
}

//sport events https://sports.ladbrokes.com/en-gb/events/type/0/0/84

type allEvents struct {
	AllEventsGroup eventGroup `json:"allEventsGroup"`
}

type eventGroup struct {
	List []eventDetail `json:"list"`
}

type eventDetail struct {
	Id           string `json:"id"`
	Event event `json:"event"`
}

type event struct {
	Participants []participant `json:"participants"`
}

type participant struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

//stavky wss://sports.ladbrokes.com/api/055/lwefiu0x/websocket

type allMarkets struct {
	Status string `json:"status"`
	Markets []market `json:"markets"`
	CorrectScore1Selections []market `json:"correctScore1Selections"`
	CorrectScoreXSelections []market `json:"correctScoreXSelections"`
	CorrectScore2Selections []market `json:"correctScore2Selections"`
	StartTime time.Time `json:"startTime"`
}

type market struct {
	Id string `json:"id"`
	NameTranslations translation `json:"nameTranslations"`
	Selections []selection `json:"selections"`
}

type translation struct {
	Value string `json:"value"`
}

type selection struct {
	NameTranslations translation `json:"nameTranslations"`
	PrimaryPrice price `json:"primaryPrice"`
	Sort string `json:"sort"`
}

type price struct {
	DecimalOdds float64 `json:"decimalOdds"`
}