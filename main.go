package main

import (
	"github.com/BurntSushi/toml"
	"fmt"
	"stavkova/database"
	"strconv"
	"database/sql"
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
}

func main() {
	var settings tomlSettings

	if _, err := toml.DecodeFile("settings.toml", &settings); err != nil {
		fmt.Println("cant read settings.toml file");
		return
	}
	dbSett := settings.DB

	sourceName := dbSett.User + ":" + dbSett.Password + "@tcp(" + dbSett.Host + ":" + strconv.Itoa(dbSett.Port) + ")/" + dbSett.Name + "?charset=utf8"
	fmt.Println(sourceName)
	dbSource, err := sql.Open("mysql", sourceName)
	if err != nil {
		fmt.Println("cant connect do database")
		return
	}
	defer dbSource.Close()

	db, err := database.NewDatabase(dbSource)
	if err != nil {
		fmt.Println("cant create db")
		return
	}

	sites, err := db.GetSites()
	for _,site := range sites {
		fmt.Println(site)
	}
	db.GetSportId("test")
}
