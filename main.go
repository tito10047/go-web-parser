package main

import (
	"stavkova/database"
	_ "stavkova/sites/ladbrokes"
	"github.com/BurntSushi/toml"
	"fmt"
	"strconv"
	"database/sql"
	"runtime"
	"net/url"
	"stavkova/sites"
	"stavkova/src"
	"sync"
)



func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var settings src.TomlSettings

	if _, err := toml.DecodeFile("settings.toml", &settings); err != nil {
		fmt.Println("cant read settings.toml file")
		panic(err)
	}

	dbSett := settings.DB

	sourceName := getDbSource(dbSett)
	dbSource, err := sql.Open("mysql", sourceName)
	if err != nil {
		fmt.Println("cant connect do database")
		panic(err)
	}
	defer dbSource.Close()

	db, err := database.NewDatabase(dbSource, settings.System.Similar)
	if err != nil {
		fmt.Println("cant create db")
		panic(err)
	}
	defer db.FlushEntities()

	dbSites, err := db.GetSites()
	if err!=nil {
		panic(err)
	}

	getId := func(siteName string) (int,bool){
		for _,site := range dbSites {
			if site.Name==siteName{
				return site.Id, true
			}
		}
		fmt.Println("cant find id for site "+siteName)
		return 0,false
	}

	var sitesArr []sites.Site
	for site := range sites.NextSite(getId,db){
		sitesArr = append(sitesArr, *site)
	}

	var wg sync.WaitGroup
	var MAX_ROUTINES int = settings.System.RoutineCount/len(sitesArr)
	fmt.Println("MAX_ROUTINES",MAX_ROUTINES)
	if MAX_ROUTINES<3{
		fmt.Println("minimum coroutines is 10")
		return
	}
	guard := make([]chan struct{},len(sitesArr))
	for index,site := range sitesArr {
		wg.Add(1)
		fmt.Println("starting main routine for site ",site.GetId())
		go func(site sites.Site, index int) {
			defer wg.Done()
			guard[index] = make(chan struct{}, MAX_ROUTINES)
			site.Setup(MAX_ROUTINES)
			for site.HasNext() {
				guard[index] <- struct{}{}
				wg.Add(1)
				go func(site sites.Site, index int) {
					defer wg.Done()

					site.ParseNext()

					<-guard[index]
				}(site, index)
			}
		}(site, index)
	}
	wg.Wait()
	for _,ch := range guard{
		close(ch)
	}
}

func getDbSource(dbSett src.DatabaseInfo) string {
	timezone := url.QueryEscape(dbSett.Timezone)
	sourceName := dbSett.User + ":" + dbSett.Password + "@tcp(" + dbSett.Host + ":" + strconv.Itoa(dbSett.Port) + ")/" + dbSett.Name + "?charset=" + dbSett.Charset + "&parseTime=true&loc=" + timezone
	return sourceName
}
