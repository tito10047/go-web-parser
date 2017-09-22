package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
	"go-web-parser/misc"
)

type Database struct {
	db         *sql.DB
	similarity float64
	muxDB      sync.Mutex
}

func NewDatabase(dbSource *sql.DB, similarity float64) (*Database, error) {

	db := &Database{
		db:         dbSource,
		similarity: similarity,
	}


	return db, nil
}

func (d *Database) GetSites() ([]DbSite, error) {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	rows, err := d.db.Query("SELECT `id`,`name`,`routines_count`,`tasks_per_time`,`wait_sec_per_tasks`,`enabled`,`timezone` FROM `parse_site`;")
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

func (d *Database) InsertMatch(siteId, sportId int, name string, teamA, teamB int, date time.Time, orgId int) (int, error) {
	d.muxDB.Lock()
	defer d.muxDB.Unlock()

	stmt, err := d.db.Prepare("SELECT `id`, `date` FROM `bet_match` WHERE `id_bet_company` = ? AND `id_bet_sport` = ? AND `org_id` = ?")
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	matchRow := stmt.QueryRow(siteId, sportId, orgId)

	var matchId int64
	var prevDate time.Time
	err = matchRow.Scan(&matchId, &date)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			return 0, err
		}
		stmt, err = d.db.Prepare("INSERT INTO `bet_match` SET `id_bet_company`=?, `id_bet_sport`=?, `name`=?, `team_a`=?, `team_b`=?, `date`=?, `org_id`=?")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		ta := sql.NullInt64{Int64: int64(teamA), Valid:teamA != -1}
		tb := sql.NullInt64{Int64: int64(teamB), Valid:teamB != -1}
		fmt.Println(siteId, sportId, name, ta, tb, date, orgId)
		res, err := stmt.Exec(siteId, sportId, name, ta, tb, date, orgId)
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

func (d *Database) InsertMatchSelection(matchId, typeId int, name string, odds float64, orgId int) error {
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
		stmt, err = d.db.Prepare("INSERT INTO `bet_mach_selection` SET `id_bet_match`=?, id_bet_match_selection_type=?, `name`=?, `odds`=?, `org_id`=?")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		_, err := stmt.Exec(matchId, typeId, name, odds, orgId)
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
