package ladbrokes

import (
	"go-web-parser/database"
	"go-web-parser/sites"
	"golang.org/x/net/websocket"
	"errors"
	"strings"
	"encoding/json"
	"strconv"
	"fmt"
	"math/rand"
	"regexp"
)

const (
	host                   = "https://sports.ladbrokes.com"
	wsHost                 = "wss://sports.ladbrokes.com"
	parseSportList    int8 = iota
	parseCompetitions
	parseEvents
	parseEvent
	parseHorseRace
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
		err = s.parseEvent(task.Params[0].(int), task.Params[1].(string))
	case parseHorseRace:
		err = s.parseHorseRace(task.Params[0].(int),task.Params[1].(string))
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
			if competition.Id==""{
				fmt.Println("BAD competetion id is empty '"+url+"' - '"+competition.Text)
			}
			if competition.Event==false {
				s.AddTask(parseEvents, host+"/en-gb/events/type/0/0/"+competition.Id, []interface{}{sportId})
			}else{
				var re = regexp.MustCompile(`(?U)\/racing\/.*\/.*\/\d*\/(\d*)\/?$`)
				if len(re.FindStringIndex(competition.Href)) > 0 {
					var idStr = re.FindStringSubmatch(competition.Href)[0]
					raceId, err := strconv.ParseInt(idStr, 10, 64)
					if err != nil {
						continue
					}
					s.AddTask(parseHorseRace,"",[]interface{}{sportId, raceId})
				}
			}
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
		s.AddTask(parseEvent, "", []interface{}{sportId, event.Id})
	}

	return nil
}
const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
type matchType string
const (
	eventsType matchType = "events"
	racesType matchType = "races"
)
func getJsonMatch(eventId string, mType matchType) (string, error) {

	uid := RandStringBytes(8)
	uid2 := strconv.Itoa(rand.Intn(3))

	origin := "https://sports.ladbrokes.com"
	//url := "wss://sports.ladbrokes.com/api/055/lwefiu0x/websocket"
	url := "wss://sports.ladbrokes.com/api/"+uid2+"/"+uid+"/websocket"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return "",err
	}
	defer ws.Close()
	var msg = make([]byte, 1024)
	var n int
	if n, err = ws.Read(msg); err != nil {
		panic(err)
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
		panic(err)
	}
	msgs = string(msg[:n])
	if msgs!="a[\"CONNECTED\\nversion:1.1\\nheart-beat:10000,10000\\n\\n\\u0000\"]"{
		return "",errors.New("bad response 2 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/user/request-response\\ndestination:/user/request-response\\n\\n\\u0000\"]")); err != nil {
		panic(err)
	}
	msg = make([]byte, 1024)
	if n, err = ws.Read(msg); err != nil {
		panic(err)
	}
	msgs = string(msg[:n])
	if len(msgs)<22 || msgs[:22]!="a[\"MESSAGE\\ntype:READY"{
		return "",errors.New("bad response 3 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/"+string(mType)+"/"+eventId+"\\ndestination:/api/en-GB/"+string(mType)+"/"+eventId+"\\n\\n\\u0000\"]")); err != nil {
		//if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\ndestination:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\n\\n\\u0000\"]")); err != nil {
		panic(err)
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

func (s *Site) parseEvent(sportId int, realSportId string) (err error) {
	defer s.EndTask()
	jsonStream, err := getJsonMatch(realSportId, eventsType)
	if err!=nil{
		fmt.Println(err)
		return err
	}

	markets := &allMarkets{}
	err = json.Unmarshal([]byte(jsonStream),markets)
	if err!=nil {
		fmt.Println(err)
		return err
	}

	if markets.Status!="ACTIVE"{
		return nil
	}

	var teamA, teamB = -1,-1
	if markets.Stats!=nil{
		maybeTeamA, okA := s.db.GetTeamId(sportId, markets.Stats.NameA)
		maybeTeamB, okB := s.db.GetTeamId(sportId, markets.Stats.NameB)
		if okA {
			teamA = maybeTeamA
		}
		if okB {
			teamB = maybeTeamB
		}
	}
	id, err := strconv.ParseInt(realSportId, 10, 64)
	if err != nil {
		return err
	}

	matchId, err := s.db.InsertMatch(s.dbSite.Id,sportId,markets.Name,teamA,teamB,markets.StartTime,int(id))
	if err!=nil{
		return err
	}

	var marketsArr = []*[]market{
		markets.Markets,
		markets.CorrectScore1Selections,
		markets.CorrectScoreXSelections,
		markets.CorrectScore2Selections,
	}
	for _, market := range marketsArr {
		if market == nil {
			continue
		}
		for _, m := range *market {
			if m.Selections == nil {
				continue
			}
			typeId, ok := s.db.GetTypeId(m.Name)
			if !ok {
				continue
			}
			for _, selections := range *m.Selections {

				selectionId, err := strconv.ParseInt(selections.Id, 10, 64)
				if err != nil {
					return err
				}
				s.db.InsertMatchSelection(matchId, typeId, selections.Name, selections.PrimaryPrice.DecimalOdds, int(selectionId))
			}
		}
	}

	return nil
}

func (s *Site ) parseHorseRace(sportId int, raceId string) error {
	defer s.EndTask()
	jsonStream, err := getJsonMatch(raceId, racesType)
	if err!=nil{
		fmt.Println(err)
		return err
	}

	race := &allHorseRaces{}
	err = json.Unmarshal([]byte(jsonStream),race)
	if err!=nil {
		fmt.Println(err)
		return err
	}

	raceName := race.Race.SportsBookClass.Name

	id, err := strconv.ParseInt(raceId, 10, 64)
	if err != nil {
		return err
	}

	matchId, err := s.db.InsertMatch(s.dbSite.Id,sportId,raceName,-1,-1,race.Race.StartTime,int(id))
	if err!=nil{
		return err
	}

	for _, group := range race.MarketGroups {
		for _, m := range group.Markets {
			if m.Selections == nil {
				continue
			}
			typeId, ok := s.db.GetTypeId(m.Name)
			if !ok {
				continue
			}
			for _, selections := range *m.Selections {

				selectionId, err := strconv.ParseInt(selections.Id, 10, 64)
				if err != nil {
					return err
				}
				s.db.InsertMatchSelection(matchId, typeId, selections.Name, selections.PrimaryPrice.DecimalOdds, int(selectionId))
			}
		}
	}

	return nil
}

func init() {
	sites.RegisterSite("ladbrokes", NewSite)
}
