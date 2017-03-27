package main

import (
	"log"
	"strings"
	"golang.org/x/net/websocket"
	"errors"
	"io/ioutil"
	"strconv"
	"math/rand"
)

type matchType string

const (
	eventsType matchType = "events"
	racesType matchType = "races"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func getJsonMatch(eventId string, mType matchType) (string, error) {

	uid := RandStringBytes(8)
	uid2 := strconv.Itoa(rand.Intn(3))

	origin := "https://sports.ladbrokes.com"
	//url := "wss://sports.ladbrokes.com/api/055/lwefiu0x/websocket"
	url := "wss://sports.ladbrokes.com/api/"+uid2+"/"+uid+"/websocket"

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
	if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/"+string(mType)+"/"+eventId+"\\ndestination:/api/en-GB/"+string(mType)+"/"+eventId+"\\n\\n\\u0000\"]")); err != nil {
		//if _, err := ws.Write([]byte("[\"SUBSCRIBE\\nid:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\ndestination:/api/en-GB/eventDetail-groups/EventGroup-LIVE-110000006-F\\n\\n\\u0000\"]")); err != nil {
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
		if (len(result)>=7 && result[len(result)-7:]=="\"null\"]") || (len(result)>=7 && result[len(result)-7:]=="a[\"\\n\"]") || len(msgs)==0 {
			msg := result
			if len(msg)>20 {
				msg=msg[len(msg)-20:]
			}
			return "",errors.New("bad response 4 '"+msg+"'")
		}
		if len(result)>=8 && result[len(result)-8:]=="\\u0000\"]"{
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
	json, err := getJsonMatch("223503532",racesType)
	if err!=nil {
		panic(err)
	}
	err = ioutil.WriteFile("./testHorse.json", []byte(json), 0644)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
