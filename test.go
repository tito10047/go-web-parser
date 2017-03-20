package main

import (
	"io"
	"golang.org/x/net/websocket"
	"log"
	"errors"
	"strings"
	"time"
	"encoding/json"
	"fmt"
)

type allMarkets struct {
	Status string `json:"status"`
	Markets []market `json:"markets"`
	CorrectScore1Selections []market `json:"correctScore1Selections"`
	CorrectScoreXSelections []market `json:"correctScoreXSelections"`
	CorrectScore2Selections []market `json:"correctScore2Selections"`
	StartTime time.Time `json:"startTime"`
}

type market struct {
	Id string `json:"id"`
	NameTranslations translation `json:"nameTranslations"`
	Selections []selection `json:"selections"`
}

type translation struct {
	Value string `json:"value"`
}

type selection struct {
	NameTranslations translation `json:"nameTranslations"`
	PrimaryPrice price `json:"primaryPrice"`
	Sort string `json:"sort"`
}

type price struct {
	DecimalOdds float64 `json:"decimalOdds"`
}


const end = '\u0000'

func getJsonMatch(eventId string) (string, error) {

	origin := "https://sports.ladbrokes.com"
	url := "wss://sports.ladbrokes.com/api/055/lwefiu0x/websocket"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return "",err
	}
	defer ws.Close()
	var msg = make([]byte, 1024)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs := string(msg[:n])
	if msgs!="o"{
		panic(errors.New("bad response 1 '"+msgs+"'"))
	}
	if _, err := ws.Write([]byte("[\"CONNECT\\nprotocol-version:1.3\\naccept-version:1.1,1.0\\nheart-beat:10000,10000\\n\\n\\u0000\"]\\n")); err != nil {
		return "",err
	}
	msg = make([]byte, 1024)
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs = string(msg[:n])
	if msgs!="a[\"CONNECTED\\nversion:1.1\\nheart-beat:10000,10000\\n\\n\\u0000\"]"{
		return "",errors.New("bad response 2 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/user/request-response\\ndestination:/user/request-response\\n\\n\\u0000\"]")); err != nil {
		log.Fatal(err)
	}
	msg = make([]byte, 1024)
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	msgs = string(msg[:n])
	if len(msgs)<22 || msgs[:22]!="a[\"MESSAGE\\ntype:READY"{
		return "",errors.New("bad response 3 '"+msgs+"'")
	}
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/events/"+eventId+"\\ndestination:/api/en-GB/events/"+eventId+"\\n\\n\\u0000\"]")); err != nil {
	//if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/event-groups/EventGroup-LIVE-110000006-F\\ndestination:/api/en-GB/event-groups/EventGroup-LIVE-110000006-F\\n\\n\\u0000\"]")); err != nil {
		log.Fatal(err)
	}
	msg=nil
	result := ""
	msg = make([]byte,1024*10)
	for {
		if n, err = ws.Read(msg); err != nil {
			return "",err
		}
		msgs = string(msg[:n])
		result+=msgs
		if (len(msgs)>=7 && msgs[len(msgs)-7:]=="\"null\"]") || len(msgs)<1{
			return "",errors.New("bad response 4 '"+msgs+"' '"+result[len(result)-50:]+"'")
		}
		if len(msgs)>=8 && msgs[len(msgs)-8:]=="\\u0000\"]"{
			break
		}
	}
	arrs := strings.Split(result,`\n`)
	result = arrs[len(arrs)-1]
	result = result[0:len(result)-8]
	result = strings.Replace(result,`\"`,`"`,-1)
	return result, nil
}

func main() {
	test, err := getJsonMatch("223404620")
	if err!=nil{
		panic(err)
	}
	//err = ioutil.WriteFile("test2.json", []byte(test), 0644)
	object := &allMarkets{}
	err = json.Unmarshal([]byte(test),object)
	if err!=nil {
		panic(err)
	}
	t:=object.StartTime
	fmt.Printf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	for _,m := range object.Markets {
		for _,s := range m.Selections {
			fmt.Println(s.PrimaryPrice)
		}
	}
	i:=1
	i++
}



func check(e error) {
	if e != nil && e!=io.EOF{
		panic(e)
	}
}