package database

import (
	"fmt"
	"database/sql"
)

func (d *Database) loadSports() error {
	d.muxSports.Lock()
	defer d.muxSports.Unlock()
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	names, err := d.db.Query("SELECT `id_bet_sport`, `name` FROM `bet_sport_name`;")
	if err != nil {
		fmt.Println("cant select from bet_sport")
		return err
	}
	for names.Next() {
		var idSport sql.NullInt64
		var name string
		if err := names.Scan(&idSport, &name); err != nil {
			fmt.Println(err)
			continue
		}
		if idSport.Valid {
			d.sports[name] = int(idSport.Int64)
		} else {
			d.sports[name] = -1
		}
	}
	return nil
}

func (d *Database) loadTypes() error {
	d.muxTypes.Lock()
	defer d.muxTypes.Unlock()
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	types, err := d.db.Query("SELECT `id_bet_type`, `name` FROM `bet_type_name`;")
	if err != nil {
		fmt.Println("cant select from id_bet_type")
		return err
	}
	for types.Next() {
		var idType sql.NullInt64
		var name string
		if err := types.Scan(&idType, &name); err != nil {
			fmt.Println(err)
			continue
		}
		if idType.Valid {
			d.types[name] = int(idType.Int64)
		} else {
			d.types[name] = -1
		}
	}
	return nil
}

func (d *Database) loadTeams() error {
	d.muxTeams.Lock()
	defer d.muxTeams.Unlock()
	d.muxDB.Lock()
	defer d.muxDB.Unlock()
	sports, err := d.db.Query("SELECT `id` FROM `bet_sport`;")
	if err != nil {
		fmt.Println("cant select from id_bet_type")
		return err
	}
	for sports.Next() {
		var sportId int
		if err := sports.Scan(&sportId); err != nil {
			fmt.Println(err)
			continue
		}
		d.teams[sportId] = make(map[string]int)

		func() {
			stmt, err := d.db.Prepare("SELECT `id_bet_team`,`name` FROM `bet_team_name` WHERE id_bet_sport=?")
			defer stmt.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			teams, err := stmt.Query(sportId)
			if err != nil {
				fmt.Println(err)
				return
			}
			for teams.Next() {
				var teamId sql.NullInt64
				var name string
				if err := teams.Scan(&teamId, &name); err != nil {
					fmt.Println(err)
					continue
				}
				if teamId.Valid {
					d.teams[sportId][name] = int(teamId.Int64)
				} else {
					d.teams[sportId][name] = -1
				}
			}
		}()
	}
	return nil
}