package ladbrokes

import (
	"stavkova/database"
	"sync"
	"stavkova/sites"
	"fmt"
)

type Site struct {
	id int
	db *database.Database
}

func NewSite(id int, db *database.Database) *sites.Site {
	i := Site{id, db}
	var site sites.Site = &i
	return &site
}

func (s *Site) ParseNext(wg sync.WaitGroup, waiter chan struct{}) {
	panic("implement me")
}

func (s *Site) HasNext() bool {
	fmt.Println("haha")
	return true
}

func init() {
	fmt.Println("registering ladbrokes")
	sites.RegisterSite("ladbrokes",NewSite)
}