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

const MIN_ROUTINES = 1

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

	dbSites, err := db.GetSites()
	if err!=nil {
		panic(err)
	}

	getDbSite := func(siteName string) (*database.DbSite,bool){
		for _,site := range dbSites {
			if site.Name==siteName{
				return &site, true
			}
		}
		fmt.Println("cant find id for site "+siteName)
		return nil,false
	}

	var sitesArr []sites.SiteInt
	for site := range sites.NextSite(getDbSite,db){
		sitesArr = append(sitesArr, *site)
	}

	var wg sync.WaitGroup

	guard := make([]chan struct{},len(sitesArr))
	for index,site := range sitesArr {
		wg.Add(1)
		fmt.Println("starting main routine for site ",site.GetArgs().Id)
		go func(site sites.SiteInt, index int) {
			defer wg.Done()
			guard[index] = make(chan struct{}, site.GetArgs().RoutinesCount)
			site.Setup(site.GetArgs().RoutinesCount,site.GetArgs().TasksPerTime,site.GetArgs().WaitSeconds)
			for site.HasNext() {
				guard[index] <- struct{}{}
				wg.Add(1)
				go func(site sites.SiteInt, index int) {
					defer wg.Done()
					defer func() {
						if r:=recover(); r!=nil{
							fmt.Println("Recovered in f", r)
						}
					}()

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
