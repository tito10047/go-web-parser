package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type database struct {
	db     *sql.DB
	sports map[string]int
	mux    sync.Mutex
}

func NewDatabase(dbSource *sql.DB) (*database, error) {

	db := &database{db: dbSource, sports: make(map[string]int)}

	names, err := db.db.Query("SELECT `id_bet_sport`, `name` FROM `bet_sport_name`;")
	if err != nil {
		fmt.Println("cant select from bet_sport")
		return nil, err
	}
	for names.Next() {
		var idSport sql.NullInt64
		var name string
		if err := names.Scan(&idSport, &name); err != nil {
			fmt.Println(err)
			continue
		}
		if idSport.Valid {
			db.sports[name] = int(idSport.Int64)
		}else{
			db.sports[name] = -1
		}
	}

	return db, nil
}

func (d *database) GetSites() ([]dbSite, error) {
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

func (d *database) GetSportId(name string) (int, bool) {
	if id, ok := d.sports[name]; ok {
		if id==-1{
			return 0, false
		}
		return id, true
	}
	stmt, err := d.db.Prepare("INSERT bet_sport_name SET name=?")
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	_ , err = stmt.Exec(name)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}

	return 0, false
}
