package main

import (
	"io"
	"golang.org/x/net/websocket"
	"log"
	"errors"
	"io/ioutil"
	"strings"
)




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
		if len(msgs)<7 || msgs[len(msgs)-7:]=="\"null\"]"{
			return "",errors.New("bad response 4 '"+msgs+"' '"+result[len(result)-50:]+"'")
		}
		if msgs[len(msgs)-8:]=="\\u0000\"]"{
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
	err = ioutil.WriteFile("test.json", []byte(test), 0644)
}



func check(e error) {
	if e != nil && e!=io.EOF{
		panic(e)
	}
}