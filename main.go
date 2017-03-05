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
)

type tomlSettings struct {
	DB databaseInfo `toml:"database"`
}

type databaseInfo struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	Charset  string
	Timezone string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.NumCPU())

	var settings tomlSettings

	if _, err := toml.DecodeFile("settings.toml", &settings); err != nil {
		fmt.Println("cant read settings.toml file");
		return
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
		return
	}
	defer db.FlushEntities()

	sites, err := db.GetSites()
	for _, site := range sites {
		fmt.Println(site)
	}

}

func getDbSource(dbSett databaseInfo) string {
	timezone := url.QueryEscape(dbSett.Timezone)
	sourceName := dbSett.User + ":" + dbSett.Password + "@tcp(" + dbSett.Host + ":" + strconv.Itoa(dbSett.Port) + ")/" + dbSett.Name + "?charset=" + dbSett.Charset + "&parseTime=true&loc=" + timezone
	return sourceName
}
