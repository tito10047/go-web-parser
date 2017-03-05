package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
	"strings"
)

const entriesBufferLength = 50
const entriesColumnCount = 8

type Database struct {
	db           *sql.DB
	sports       map[string]int
	types        map[string]int
	teams        map[int]map[string]int
	entries      [entriesBufferLength * entriesColumnCount]interface{}
	entriesIndex int
	muxSports    sync.Mutex
	muxTypes     sync.Mutex
	muxTeams     sync.Mutex
	muxDB        sync.Mutex
	muxEntries   sync.Mutex
}

func NewDatabase(dbSource *sql.DB) (*Database, error) {

	db := &Database{
		db:           dbSource,
		sports:       make(map[string]int),
		types:        make(map[string]int),
		teams:        make(map[int]map[string]int),
		entriesIndex: 0,
	}

	if err := db.loadSports(); err != nil {
		return db, err
	}

	if err := db.loadTypes(); err != nil {
		return db, err
	}

	if err := db.loadTeams(); err != nil {
		return db, err
	}

	return db, nil
}

func (d *Database) GetSites() ([]dbSite, error) {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	rows, err := d.db.Query("SELECT `id`,`name` FROM `bet_company`;")
	if err != nil {
		fmt.Println("cant select from bet_company")
		return nil, err
	}
	defer rows.Close()

	var sites []dbSite

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			fmt.Println(err)
			continue
		}
		sites = append(sites, dbSite{id, name})
	}
	return sites, nil
}

func (d *Database) GetSportId(name string) (int, bool) {
	d.muxSports.Lock()
	defer d.muxSports.Unlock()
	if id, ok := d.sports[name]; ok {
		if id == -1 {
			return 0, false
		}
		return id, true
	}
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT bet_sport_name SET name=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	d.sports[name] = -1
	return 0, false
}

func (d *Database) GetTypeId(name string) (int, bool) {
	d.muxTypes.Lock()
	defer d.muxTypes.Unlock()
	if id, ok := d.types[name]; ok {
		if id == -1 {
			return 0, false
		}
		return id, true
	}
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT bet_type_name SET name=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()
	_, err = stmt.Exec(name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	d.types[name] = -1
	return 0, false
}

func (d *Database) GetTeamId(sportId int, name string) (int, bool) {
	d.muxTeams.Lock()
	defer d.muxTeams.Unlock()
	if teams, ok := d.teams[sportId]; ok {
		if id, ok := teams[name]; ok {
			return id, true
		}
	} else {
		d.teams[sportId] = make(map[string]int)
	}
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT `bet_team_name` SET `id_bet_sport`=?, `name`=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()
	_, err = stmt.Exec(sportId, name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	d.teams[sportId][name] = -1
	return 0, false
}

/**
 * @
 */
func (d *Database) InsertEntry(siteId, sportId, typeId, teamId int, rate, maxBet float32, date time.Time, orgId int) {
	d.muxEntries.Lock()
	defer d.muxEntries.Unlock()

	d.entries[d.entriesIndex+0] = siteId
	d.entries[d.entriesIndex+1] = sportId
	d.entries[d.entriesIndex+2] = typeId
	d.entries[d.entriesIndex+3] = teamId
	d.entries[d.entriesIndex+4] = rate
	d.entries[d.entriesIndex+5] = maxBet
	d.entries[d.entriesIndex+6] = date
	d.entries[d.entriesIndex+7] = orgId
	d.entriesIndex+=entriesColumnCount

	if d.entriesIndex<entriesBufferLength+entriesColumnCount-1{
		return
	}

	d.FlushEntities()
}

func (d *Database) FlushEntities() {
	if d.entriesIndex==0{
		return
	}
	sqlStr := "INSERT INTO bet_entry(id_bet_company, id_bet_sport, id_bet_type, id_bet_team, rate, max_bet, date, org_id) VALUES "
	sqlSubStr := strings.Repeat("?,", entriesColumnCount)
	sqlSubStr = sqlSubStr[0:len(sqlSubStr)-1]
	sqlStr += strings.Repeat("("+sqlSubStr+"),", d.entriesIndex/entriesColumnCount)
	sqlStr = sqlStr[0:len(sqlStr)-1]
	stmt, err := d.db.Prepare(sqlStr)
	if err != nil {
		fmt.Println(sqlStr)
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(d.entries[0:d.entriesIndex]...)
	if err != nil {
		fmt.Println(d.entries[0:d.entriesIndex])
		panic(err)
	}
	d.entriesIndex=0
}
