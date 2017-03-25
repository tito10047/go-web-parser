package paddy

import (
	"stavkova/sites"
	"stavkova/database"
	"strings"
	"github.com/andybalholm/cascadia"
	xhtml "golang.org/x/net/html"
	"errors"
)

const(
	host = "http://www.paddypower.com"
	parseSportList int8 = iota
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

func init() {
	sites.RegisterSite("paddy", NewSite)
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
	selector, err := cascadia.Compile("#nav>#sitemap li>a[href^=\"http\"]")
	if err!=nil {
		return err
	}
	aNodes := selector.MatchAll(doc)

	for _,aNode := range aNodes {
		sportId, ok := s.db.GetSportId(aNode.FirstChild.Data)
		if ok {
			for _,attr := range aNode.Attr {
				if attr.Key=="href" {
					s.AddTask(parseSport, attr.Val,[]interface{}{"","today",1,sportId})
					break
				}
			}
		}
	}

	return nil
}