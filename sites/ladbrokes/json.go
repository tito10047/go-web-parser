package ladbrokes

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
	Id string `json:"id"`
}

type group struct {
	Meta      string `json:"meta"`
	Text      string `json:"text"`
	Href      string `json:"href"`
	DeepLinkHref string `json:"deepLinkHref"`
	Icon      string `json:"icon"`
	Id string `json:"id"`
	SortOrder int `json:"sortOrder"`
	Event     bool `json:"event"`
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
	List []event `json:"list"`
}

type event struct {
	Id string `json:"id"`
	Stats eventStat `json:"stats"`
}

type eventStat struct {
	NameA string `json:"nameA"`
	NameB string `json:"nameB"`
}