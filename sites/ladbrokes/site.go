package ladbrokes

import (
	"stavkova/database"
	"stavkova/sites"
	"fmt"
)

const (
	host                   = "https://sports.ladbrokes.com"
	parseSportList    int8 = iota
	parseCompetitions
	parseEvents
	parseEvent
)

type Site struct {
	*sites.TaskStack
	id  int
	db  *database.Database
	cnt int
	ok  bool
}

func NewSite(id int, db *database.Database) *sites.Site {
	i := Site{id: id, db: db, ok: true}
	var site sites.Site = &i
	return &site
}

func (s *Site) Setup(routinesCount int) {
	s.AddTask(parseSportList, host+"/en-gb/az_list", nil)
}

func (s *Site) GetId() int {
	return s.id
}

func (s *Site) ParseNext() {
	task, ok := s.NextTask()
	if ok == false {
		return
	}
	var err error
	switch task.TaskNum {
	case parseSportList:
		err = s.parseSportList(task.Url)
	case parseCompetitions:
		err = s.parseCompetition(task.Url, task.Params[0].(int))
	case parseEvents:
		err = s.parseEvents(task.Url, task.Params[0].(int))
	case parseEvent:
		err = s.parseEvent(task.Params[0].(int),task.Params[1].(int),task.Params[2].(int),task.Params[3].(string))
	}
	if err != nil {
		s.CloseTasks()
		panic(err)
	}
}

func (s *Site) HasNext() bool {
	return s.HasTask() == false && s.ok == true
}

func (s *Site) parseSportList(url string) (err error) {
	d, err := sites.NewDownloader("GET", url)
	if err != nil {
		return
	}
	alSp := &allSports{}
	err = d.DownloadJson(alSp)
	if err != nil {
		return
	}
	for _, sport := range alSp.AllSportsGroup.List {
		id, ok := s.db.GetSportId(sport.Text)
		if ok {
			s.AddTask(parseCompetitions, host+"/en-gb/sport_competition_links/"+sport.Icon, []interface{}{id})
		}
	}
	s.EndTask()
	return nil
}

func (s *Site) parseCompetition(url string, sportId int) (err error) {
	d, err := sites.NewDownloader("GET", url)
	if err != nil {
		return
	}
	competitions := &allCompetitions{}
	err = d.DownloadJson(competitions)
	if err != nil {
		return
	}

	for _, competition := range competitions.LinkGroupsForClasses {
		s.AddTask(parseEvents, host+"/en-gb/events/type/0/0/"+competition.Id, []interface{}{sportId})
	}

	s.EndTask()
	return nil
}

func (s *Site) parseEvents(url string, sportId int) (err error) {
	d, err := sites.NewDownloader("GET", url)
	if err != nil {
		return
	}
	events := &allEvents{}
	err = d.DownloadJson(events)
	if err != nil {
		return
	}

	for _, event := range events.AllEventsGroup.List {
		aTeamId, okA := s.db.GetTeamId(sportId,event.Stats.NameA)
		bTeamId, okB := s.db.GetTeamId(sportId,event.Stats.NameB)
		if okA==false || okB==false {
			continue
		}
		s.AddTask(parseEvent, "", []interface{}{sportId,aTeamId,bTeamId, event.Id})
	}

	s.EndTask()
	return nil
}

func (s *Site) parseEvent(sportId,aTeamId,bTeamId int, realSportId string) (err error) {
	
}

func init() {
	fmt.Println("registering ladbrokes")
	sites.RegisterSite("ladbrokes", NewSite)
}
