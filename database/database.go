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
	db         *sql.DB
	similarity float64
	sports     map[string]int
	types      map[string]int
	teams      map[int]map[string]int
	muxSports  sync.Mutex
	muxTypes   sync.Mutex
	muxTeams   sync.Mutex
	muxDB      sync.Mutex
}

func NewDatabase(dbSource *sql.DB, similarity float64) (*Database, error) {

	db := &Database{
		db:         dbSource,
		similarity: similarity,
		sports:     make(map[string]int),
		types:      make(map[string]int),
		teams:      make(map[int]map[string]int),
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
	rows, err := d.db.Query("SELECT `id`,`name`,`routines_count`,`tasks_per_time`,`wait_sec_per_tasks`,`enabled`,`timezone` FROM `bet_company`;")
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
		var timezone int
		if err := rows.Scan(&id, &name, &routinesCount, &tasksPerTime, &waitSeconds, &enabled, &timezone); err != nil {
			fmt.Println(err)
			continue
		}
		sites = append(sites, DbSite{
			id,
			routinesCount,
			tasksPerTime,
			waitSeconds, name,
			enabled,
			timezone,
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
		panic(err)
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
	stmt, err := d.db.Prepare("INSERT bet_match_type_name SET name=?, id_bet_match_type=?")
	if err != nil {
		panic(err)
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

	teamId := d.findSimilarId(d.teams[sportId], name)

	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	stmt, err := d.db.Prepare("INSERT `bet_team_name` SET `id_bet_sport`=?, id_bet_team=?, `name`=?")
	if err != nil {
		panic(err)
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

/**
 * @
 */
func (d *Database) InsertMatch(siteId, sportId, typeId int, name string, teamA, teamB int, date time.Time, orgId int) (int, error) {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()

	stmt, err := d.db.Prepare("SELECT `id`, `date` FROM `bet_match` WHERE `id_bet_company` = ? AND `id_bet_sport` = ? AND `id_bet_match_type` = ? AND `org_id` = ?")
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	matchRow := stmt.QueryRow(siteId, sportId, typeId, orgId)

	var matchId int64
	var prevDate time.Time
	err = matchRow.Scan(&matchId, &date)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			return 0, err
		}
		stmt, err = d.db.Prepare("INSERT INTO `bet_match` SET `id_bet_company`=?, `id_bet_sport`=?, `id_bet_match_type`=?, `name`=?, `team_a`=?, `team_b`=?, `date`=?, `org_id`=?")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		ta := sql.NullInt64{Int64: int64(teamA), Valid:teamA != -1}
		tb := sql.NullInt64{Int64: int64(teamA), Valid:teamB != -1}
		res, err := stmt.Exec(siteId, sportId, typeId, name, ta, tb, date, orgId)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		matchId, err = res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
	} else {
		if !prevDate.Equal(date) {
			stmt, err = d.db.Prepare("UPDATE `bet_match` SET `date` = ? WHERE `id` = ?")
			if err != nil {
				panic(err)
			}
			defer stmt.Close()
			_, err := stmt.Exec(date, matchId)
			if err != nil {
				fmt.Println(err)
				return 0, err
			}
		}
	}
	return int(matchId), nil
}

func (d *Database) InsertMatchSelection(matchId int, name string, odds float64, orgId int) error {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()

	stmt, err := d.db.Prepare("SELECT id, odds FROM `bet_mach_selection` WHERE `id_bet_match`=? AND `org_id`=?")
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	selectionRow := stmt.QueryRow(matchId, orgId)
	var selectionId int64
	var orgOdds float64
	err = selectionRow.Scan(&selectionId, &orgOdds)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			return err
		}
		stmt, err = d.db.Prepare("INSERT INTO `bet_mach_selection` SET `id_bet_match`=?, `name`=?, `odds`=?, `org_id`=?")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		_, err := stmt.Exec(matchId, name, odds, orgId)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}else{
		if orgOdds!=odds {
			stmt, err = d.db.Prepare("UPDATE `bet_mach_selection` SET `odds`=?, last_update=NOW() WHERE `id`=?")
			if err != nil {
				panic(err)
			}
			defer stmt.Close()
			_, err := stmt.Exec(odds, matchId)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
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
