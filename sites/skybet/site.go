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
	host = "https://www.skybet.com/"
	parseSportList    int8 = iota
	parseSport
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
		s.CloseTasks()
		panic(err)
	}
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
					s.AddTask(parseSport, host+attr.Val,[]interface{}{sportId})
				}
			}
		}
	}

	return nil
}


func init() {
	fmt.Println("registering skybet")
	sites.RegisterSite("skybet", NewSite)
}