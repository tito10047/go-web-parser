package sites

import (
	"stavkova/database"
	"sync"
	"reflect"
)

type Site interface {
	ParseNext(wg sync.WaitGroup,waiter chan struct{})
	HasNext() bool
}
type NewSite func(id int, db *database.Database) *Site

var sites = map[string]NewSite{}

func RegisterSite(siteName string, constructor NewSite)  {
	sites[siteName]=constructor
}

func NextSite(getId func(siteName string) (int,bool), db *database.Database) <-chan *Site {
	ch := make(chan *Site,1)
	go func() {
		for name := range sites {
			siteId, ok := getId(name)
			if !ok{
				continue
			}
			site:=createSite(name, siteId, db)
			ch <- site
		}
		close(ch)
	}()
	return ch
}

func createSite(name string, id int, db *database.Database) *Site {
	f := reflect.ValueOf(sites[name])

	in := []reflect.Value{
		reflect.ValueOf(id),
		reflect.ValueOf(db),
	}

	result := f.Call(in)
	//site := result[0].Convert(reflect.TypeOf((*Site)(nil)))
	site := result[0].Interface().(*Site)
	return site
}