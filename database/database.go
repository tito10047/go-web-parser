package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
	"stavkova/misc"
)

type Database struct {
	db           *sql.DB
	similarity   float64
	sports       map[string]int
	types        map[string]int
	teams        map[int]map[string]int
	muxSports    sync.Mutex
	muxTypes     sync.Mutex
	muxTeams     sync.Mutex
	muxDB        sync.Mutex
}

func NewDatabase(dbSource *sql.DB, similarity float64) (*Database, error) {

	db := &Database{
		db:           dbSource,
		similarity:   similarity,
		sports:       make(map[string]int),
		types:        make(map[string]int),
		teams:        make(map[int]map[string]int),
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

func (d *Database) GetSites() ([]DbSite, error) {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	rows, err := d.db.Query("SELECT `id`,`name`,`routines_count`,`tasks_per_time`,`wait_sec_per_tasks`,`Enabled` FROM `bet_company`;")
	if err != nil {
		fmt.Println("cant select from bet_company")
		return nil, err
	}
	defer rows.Close()

	var sites []DbSite

	for rows.Next() {
		var id, routinesCount, tasksPerTime, waitSeconds int
		var name string
		var enabled bool
		if err := rows.Scan(&id, &name, &routinesCount, &tasksPerTime, &waitSeconds, &enabled); err != nil {
			fmt.Println(err)
			continue
		}
		sites = append(sites, DbSite{
			id,
			routinesCount,
			tasksPerTime,
			waitSeconds, name,
			enabled,
		})
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

	sportId := d.findSimilarId(d.sports, name)

	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT bet_sport_name SET name=?, id_bet_sport=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, sportId)
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

	typeId := d.findSimilarId(d.types, name)

	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT bet_match_type_name SET name=?, id_bet_type=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, typeId)
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

	teamId  := d.findSimilarId(d.teams[sportId], name)

	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT `bet_team_name` SET `id_bet_sport`=?, id_bet_team=?, `name`=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer stmt.Close()
	_, err = stmt.Exec(sportId, teamId, name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	d.teams[sportId][name] = -1
	return 0, false
}

type EntryTeam struct {
	TeamId sql.NullInt64
	Odd	float64
}
/**
 * @
 */
func (d *Database) InsertEntry(siteId, sportId, typeId int, teams[]EntryTeam, date time.Time, orgId int) {
	//TODO implement new database structure
	return
	/*d.muxDB.Lock()
	defer d.muxDB.Unlock()

	stmt, err := d.db.Prepare("SELECT * FROM `bet_entry` WHERE `id_bet_company`=? AND `id_bet_sport`=? AND `id_bet_type`=? AND `org_id`=?")
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	entryRow := stmt.QueryRow(sportId)

	var entryId int64
	err = entryRow.Scan(&entryId)
	if err!=nil{
		if err!=sql.ErrNoRows{
			fmt.Println(err)
			return
		}

		stmt, err = d.db.Prepare("INSERT `bet_entry` SET `id_bet_company`=?, `id_bet_sport`=?, `id_bet_type`=?, `max_bet`=?, `date`=?, `org_id`=?, ")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer stmt.Close()
		//TODO: max bet
		res, err := stmt.Exec(siteId, sportId, typeId,0,date,orgId)
		if err != nil {
			fmt.Println(err)
			return
		}
		entryId, err = res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for _,entry := range teams{
		stmt, err = d.db.Prepare("INSERT INTO `bet_entry_team` (`id_entry`, `id_team`, `odd`) VALUES (?,?,?) on duplicate key update `odd`=?")
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err := stmt.Exec(entryId, typeId, entry.TeamId, entry.Odd)
		if err != nil {
			fmt.Println(err)
		}
		stmt.Close()
	}

	return*/
}


func (d *Database) findSimilarId(m map[string]int, name string) sql.NullInt64 {
	id := sql.NullInt64{Valid: false}
	var max float64 = 0
	for maybeName, maybeId := range m {
		percent := misc.SimilarTextPercent(name, maybeName)
		if maybeId != -1 && percent > d.similarity && percent > max {
			max = percent
			id.Valid = true
			id.Int64 = int64(maybeId)
		}
	}
	return id
}