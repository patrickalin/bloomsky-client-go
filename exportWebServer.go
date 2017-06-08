package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

var conn *websocket.Conn
var mybloomsky bloomsky.BloomskyStructure
var msgJSON []byte

func write() {
	for {
		msg := <-bloomskyMessageToHTTP
		mybloomsky = msg
		var err error

		mylog.Trace.Println("Receive message to export to http")

		msgJSON, err = json.Marshal(msg)
		if err != nil {
			mylog.Trace.Printf("%v", fmt.Errorf("Error: %s", err))
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, msgJSON)
		if err != nil {
			mylog.Trace.Printf("Impossible to write to websocket : %v", err)
		}
		mylog.Trace.Println("Message send to browser")
	}
}

// websocket handler
func refreshdata(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	var err error
	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		mylog.Trace.Printf("upgrade: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		mylog.Trace.Printf("Impossible to write to websocket : %v", err)
	}

	go write()
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate, err := template.ParseFiles("tmpl/bloomsky.html")
	if err != nil {
		log.Fatal(fmt.Errorf("template part 1 : %v", err))
	}
	err = homeTemplate.Execute(w, "ws://"+r.Host+"/refreshdata")
	if err != nil {
		log.Fatal(fmt.Errorf("write part 2 : %v", err))
	}
	homeTemplate2, err := template.ParseFiles("tmpl/bloomsky2.html")
	if err != nil {
		log.Fatal(fmt.Errorf("template part 2 : %v", err))
	}
	err = homeTemplate2.Execute(w, mybloomsky)
	if err != nil {
		log.Fatal(fmt.Errorf("write part 2 : %v", err))
	}
}

//createWebServer create web server
func createWebServer(HTTPPort string) {
	flag.Parse()
	log.SetFlags(0)

	http.HandleFunc("/refreshdata", refreshdata)
	http.HandleFunc("/", home)

	mylog.Trace.Printf("Init server http port %s \n", HTTPPort)

	if err := http.ListenAndServe(HTTPPort, nil); err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error when I create the server : %v", err))
	}
	mylog.Trace.Printf("Server listen on %s", HTTPPort)
}
