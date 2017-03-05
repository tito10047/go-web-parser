package ladbrokes

import (
	"stavkova/database"
	"stavkova/sites"
	"fmt"
	"net/http"
	"crypto/tls"
	"bytes"
	"io"
)

type Site struct {
	id  int
	db  *database.Database
	cnt int
	ok bool
}

func (s *Site) GetId() int {
	return s.id
}

func NewSite(id int, db *database.Database) *sites.Site {
	i := Site{id:id, db:db, ok:true}
	var site sites.Site = &i
	return &site
}

func (s *Site) ParseNext() {
	//fmt.Println("stating read in step",s.cnt)
	req, err := http.NewRequest("GET", "https://warofbastards.com", nil)
	if err != nil {
		fmt.Println(err)
		s.ok=false
		return
	}
	req.Header.Set("Accept", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		s.ok=false
		return
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)

	if buf.Len()==0{
		fmt.Println("bad:",response.StatusCode)
		s.ok=false
	}

	//body := string(b)
	fmt.Println("reader",buf.Len(),"bytes in step",s.cnt)
}

func (s *Site) HasNext() bool {
	s.cnt++
	return s.cnt<10000 && s.ok
}

func init() {
	fmt.Println("registering ladbrokes")
	sites.RegisterSite("ladbrokes", NewSite)
}
