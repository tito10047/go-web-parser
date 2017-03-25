package skybet

import (
	"stavkova/database"
	"stavkova/sites"
	"errors"
	"fmt"
	"strings"
	xhtml "golang.org/x/net/html"
	"github.com/andybalholm/cascadia"
)

const (
	host = "https://www.skybet.com"
	parseSportList    int8 = iota
	parseSport
	parseEvents
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
	case parseSport:
		err = s.parseSport(task.Url, task.Params[0].(string), task.Params[1].(string), task.Params[2].(int), task.Params[3].(int))
	case parseEvents:
		err = s.parseEvents(task.Url,task.Params[0].(int))
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

func (s *Site) parseSportList(url string) error {
	defer s.EndTask()
	d, err := sites.NewDownloader("GET", url)
	if err!=nil {
		return err
	}
	html, err := d.Download()
	if err!=nil {
		return err
	}
	doc, err := xhtml.Parse(strings.NewReader(html))
	if err!=nil {
		return err
	}
	selector, err := cascadia.Compile("#nav .section:nth-child(5n+4)>ul>a")
	if err!=nil {
		return err
	}
	aNodes := selector.MatchAll(doc)

	for _,aNode := range aNodes {
		sportId, ok := s.db.GetSportId(aNode.FirstChild.Data)
		if ok {
			for _,attr := range aNode.Attr {
				if attr.Key=="href" {
					s.AddTask(parseSport, host+attr.Val+"/ev_time/all",[]interface{}{"","today",1,sportId})
					break
				}
			}
		}
	}

	return nil
}

func (s *Site) parseSport(url, params, day string, page, sportId int) error {
	defer s.EndTask()
	d, err := sites.NewDownloader("GET", url+params)
	if err!=nil {
		return err
	}
	html, err := d.Download()
	if err!=nil {
		return err
	}
	doc, err := xhtml.Parse(strings.NewReader(html))
	if err!=nil {
		return err
	}

	// paginate days
	if day=="today" && page==1 {
		selector, err := cascadia.Compile(".pagination.day-pagination>li>a")
		if err!=nil {
			return err
		}
		aNodes := selector.MatchAll(doc)
		for index, aNode := range aNodes {
			if index==0 {
				continue
			}
			for _,attr := range aNode.Attr {
				if attr.Key == "href" {
					s.AddTask(parseSport, url,[]interface{}{attr.Val,"",1,sportId})
					break
				}
			}
		}
	}

	// paginate pages
	if page==1 {
		selector, err := cascadia.Compile("ol.pagination.below.meta-pagination>li>a")
		if err != nil {
			return err
		}
		aNodes := selector.MatchAll(doc)
		for index, aNode := range aNodes {
			if index == 0 {
				continue
			}
			for _, attr := range aNode.Attr {
				if attr.Key == "href" {
					s.AddTask(parseSport, url, []interface{}{attr.Val,"", index+1, sportId})
					break
				}
			}
		}
	}

	selector, err := cascadia.Compile(".market-wdw>table.mkt>tbody>tr>td>.all-bets-link")
	if err!=nil {
		return err
	}
	trNodes := selector.MatchAll(doc)
	for _, trNode := range trNodes {
		for _, attr := range trNode.Attr {
			if attr.Key == "href" {
				s.AddTask(parseEvents, host+attr.Val, []interface{}{sportId})
				break
			}
		}
	}

	return nil
}

func (s *Site) parseEvents(url string, sportId int) error {
	defer s.EndTask()
	fmt.Println(url)
	return nil
}


func init() {
	sites.RegisterSite("skybet", NewSite)
}