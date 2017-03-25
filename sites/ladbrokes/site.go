package ladbrokes

import (
	"stavkova/database"
	"stavkova/sites"
	"fmt"
	"log"
	"golang.org/x/net/websocket"
	"errors"
	"strings"
)

const (
	host                   = "https://sports.ladbrokes.com"
	wsHost                 = "wss://sports.ladbrokes.com"
	parseSportList    int8 = iota
	parseCompetitions
	parseEvents
	parseEvent
)

type Site struct {
	*sites.TaskStack
	dbSite *database.DbSite
	db     *database.Database
	ok     bool
}

func NewSite(sbSite *database.DbSite, db *database.Database) *sites.SiteInt {
	i := &Site{dbSite: sbSite, db: db, ok: true}
	v := sites.SiteInt(i)
	return &v
}

func (s *Site) Setup(routinesCount, tasksPerTime, waitSeconds int) {
	s.TaskStack = sites.NewTaskStack(routinesCount, tasksPerTime, waitSeconds)
	s.AddTask(parseSportList, host+"/en-gb/az_list", nil)
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
		err = s.parseEvent(task.Params[0].(int), task.Params[1].(int), task.Params[2].(int), task.Params[3].(string))
	default:
		err = errors.New("some is wrong")
	}
	if err != nil {
		s.ok=false
		s.CloseTasks()
		panic(err)
	}
}

func (s *Site) HasNext() bool {
	return s.HasTask() == true && s.ok == true
}

func (s *Site) GetArgs() *database.DbSite {
	return s.dbSite
}

func (s *Site) parseSportList(url string) (err error) {
	defer s.EndTask()
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
	return nil
}

func (s *Site) parseCompetition(url string, sportId int) (err error) {
	defer s.EndTask()
	d, err := sites.NewDownloader("GET", url)
	if err != nil {
		return
	}
	competitions := &allCompetitions{}
	err = d.DownloadJson(competitions)
	if err != nil {
		return
	}

	for _, competitionGroup := range competitions.LinkGroupsForClasses {
		for _, competition := range competitionGroup.List {
			s.AddTask(parseEvents, host+"/en-gb/events/type/0/0/"+competition.Id, []interface{}{sportId})
		}
	}

	return nil
}

func (s *Site) parseEvents(url string, sportId int) (err error) {
	defer s.EndTask()
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
		if len(event.Event.Participants)==0{
			continue
		}
		if len(event.Event.Participants)<2{
			return errors.New("cosi je zle")
		}
		aTeamId, okA := s.db.GetTeamId(sportId, event.Event.Participants[0].Name)
		bTeamId, okB := s.db.GetTeamId(sportId, event.Event.Participants[1].Name)
		if okA == false || okB == false {
			continue
		}
		s.AddTask(parseEvent, "", []interface{}{sportId, aTeamId, bTeamId, event.Id})
	}

	return nil
}

func getJsonMatch(eventId string) (string, error) {

	origin := "https://sports.ladbrokes.com"
	url := "wss://sports.ladbrokes.com/api/055/lwefiu0x/websocket"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return "",err
	}
	defer ws.Close()
	var msg = make([]byte, 1024)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs := string(msg[:n])
	if msgs!="o"{
		panic(errors.New("bad response 1 '"+msgs+"'"))
	}
	if _, err := ws.Write([]byte("[\"CONNECT\\nprotocol-version:1.3\\naccept-version:1.1,1.0\\nheart-beat:10000,10000\\n\\n\\u0000\"]\\n")); err != nil {
		return "",err
	}
	msg = make([]byte, 1024)
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs = string(msg[:n])
	if msgs!="a[\"CONNECTED\\nversion:1.1\\nheart-beat:10000,10000\\n\\n\\u0000\"]"{
		return "",errors.New("bad response 2 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/user/request-response\\ndestination:/user/request-response\\n\\n\\u0000\"]")); err != nil {
		log.Fatal(err)
	}
	msg = make([]byte, 1024)
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs = string(msg[:n])
	if len(msgs)<22 || msgs[:22]!="a[\"MESSAGE\\ntype:READY"{
		return "",errors.New("bad response 3 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/events/"+eventId+"\\ndestination:/api/en-GB/events/"+eventId+"\\n\\n\\u0000\"]")); err != nil {
		//if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\ndestination:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\n\\n\\u0000\"]")); err != nil {
		log.Fatal(err)
	}
	msg=nil
	result := ""
	msg = make([]byte,1024*10)
	for {
		if n, err = ws.Read(msg); err != nil {
			return "",err
		}
		msgs = string(msg[:n])
		result+=msgs
		if (len(result)>=7 && result[len(result)-7:]=="\"null\"]") || (len(result)>=7 && result[len(result)-7:]=="a[\"\\n\"]") || len(msgs)==0 {
			msg := result
			if len(msg)>20 {
				msg=msg[len(msg)-20:]
			}
			return "",errors.New("bad response 4 '"+msg+"'")
		}
		if len(result)>=8 && result[len(result)-8:]=="\\u0000\"]"{
			break
		}
	}
	arrs := strings.Split(result,`\n`)
	result = arrs[len(arrs)-1]
	result = result[0:len(result)-8]
	result = strings.Replace(result,`\"`,`"`,-1)
	return result, nil
}

func (s *Site) parseEvent(sportId, aTeamId, bTeamId int, realSportId string) (err error) {
	defer s.EndTask()
	return nil
}

func init() {
	sites.RegisterSite("ladbrokes", NewSite)
}
