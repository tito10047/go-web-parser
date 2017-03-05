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

	db, err := database.NewDatabase(dbSource)
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
	for site := range sites.NextSite(getId,nil){
		sitesArr = append(sitesArr, *site)
	}

	var wg sync.WaitGroup
	var MAX_ROUTINES int = settings.System.RoutineCount/len(sitesArr)
	fmt.Println("MAX_ROUTINES",MAX_ROUTINES)
	if MAX_ROUTINES<10{
		fmt.Println("minimum coroutines is 10")
		return
	}
	for _,site := range sitesArr {
		wg.Add(1)
		fmt.Println("starting main routine for site ",site.GetId())
		go func(site sites.Site) {
			defer wg.Done()
			guard := make(chan struct{}, MAX_ROUTINES)

			for site.HasNext() {
				wg.Add(1)
				guard <- struct{}{}
				go func(site sites.Site) {
					defer wg.Done()

					site.ParseNext()

					<-guard
				}(site)
			}
			close(guard)
		}(site)
	}
	wg.Wait()
}

func getDbSource(dbSett src.DatabaseInfo) string {
	timezone := url.QueryEscape(dbSett.Timezone)
	sourceName := dbSett.User + ":" + dbSett.Password + "@tcp(" + dbSett.Host + ":" + strconv.Itoa(dbSett.Port) + ")/" + dbSett.Name + "?charset=" + dbSett.Charset + "&parseTime=true&loc=" + timezone
	return sourceName
}
