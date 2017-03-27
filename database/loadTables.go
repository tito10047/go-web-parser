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
	names, err := d.db.Query("SELECT `id_bet_sport`, `bet_sport_name`.`name`, `enabled` FROM `bet_sport_name` LEFT JOIN bet_sport on bet_sport.id=id_bet_sport;")
	if err != nil {
		panic(err)
	}
	for names.Next() {
		var idSport sql.NullInt64
		var name string
		var enabled sql.NullBool
		if err := names.Scan(&idSport, &name, &enabled); err != nil {
			fmt.Println(err)
			continue
		}
		if idSport.Valid && enabled.Valid && enabled.Bool==true {
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
	types, err := d.db.Query("SELECT `id_bet_match_type`, `bet_match_selection_type_name`.`name`, Enabled FROM `bet_match_selection_type_name` LEFT JOIN bet_match_selection_type ON bet_match_selection_type.id=id_bet_match_type;")
	if err != nil {
		panic(err)
	}
	for types.Next() {
		var idType sql.NullInt64
		var name string
		var enabled sql.NullBool
		if err := types.Scan(&idType, &name, &enabled); err != nil {
			fmt.Println(err)
			continue
		}
		if idType.Valid && enabled.Valid && enabled.Bool==true {
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
		fmt.Println("cant select from bet_sport")
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
			stmt, err := d.db.Prepare("SELECT `id_bet_team`,`bet_team_name`.`name`, Enabled FROM `bet_team_name` left join bet_sport on bet_sport.id=id_bet_sport WHERE id_bet_sport=?")
			defer stmt.Close()
			if err != nil {
				panic(err)
			}
			teams, err := stmt.Query(sportId)
			if err != nil {
				fmt.Println(err)
				return
			}
			for teams.Next() {
				var teamId sql.NullInt64
				var name string
				var enabled sql.NullBool
				if err := teams.Scan(&teamId, &name, &enabled); err != nil {
					fmt.Println(err)
					continue
				}
				if teamId.Valid && enabled.Valid && enabled.Bool==true {
					d.teams[sportId][name] = int(teamId.Int64)
				} else {
					d.teams[sportId][name] = -1
				}
			}
		}()
	}
	return nil
}