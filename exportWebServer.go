package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
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

	var err error
	var templateHeader *template.Template
	var templateBody *template.Template

	if config.dev {
		templateHeader, err = template.New("bloomsky_header.html").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).ParseFiles("tmpl/bloomsky_header.html")
	} else {
		assetHeader, err := assembly.Asset("tmpl/bloomsky_header.html")
		if err != nil {
			panic(err)
		}

		templateHeader, err = template.New("bloomsky_header.html").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).Parse(string(assetHeader[:]))
	}

	if err != nil {
		log.Fatal(fmt.Errorf("template part 1 : %v", err))
	}

	//fmt.Println(T("program_greeting"))

	err = templateHeader.Execute(w, "ws://"+r.Host+"/refreshdata")
	if err != nil {
		log.Fatal(fmt.Errorf("write part 1 : %v", err))
	}

	fmt.Println(config.dev)

	if config.dev {
		templateBody, err = template.New("bloomsky_body.html").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).ParseFiles("tmpl/bloomsky_body.html")

	} else {
		assetBody, err := assembly.Asset("tmpl/bloomsky_body.html")
		if err != nil {
			panic(err)
		}

		templateBody, err = template.New("bloomsky_body.html").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).Parse(string(assetBody[:]))
	}

	if err != nil {
		log.Fatal(fmt.Errorf("template part 2 : %v", err))
	}

	//fmt.Println(T("program_greeting"))

	err = templateBody.Execute(w, mybloomsky)
	if err != nil {
		log.Fatal(fmt.Errorf("write part 2 : %v", err))
	}
}

//createWebServer create web server
func createWebServer(HTTPPort string) {
	flag.Parse()
	log.SetFlags(0)

	fs := http.FileServer(&assetfs.AssetFS{Asset: assemblyAssetfs.Asset, AssetDir: assemblyAssetfs.AssetDir, AssetInfo: assemblyAssetfs.AssetInfo, Prefix: "static"})

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/refreshdata", refreshdata)
	http.HandleFunc("/", home)

	fmt.Printf("Init server http port %s \n", HTTPPort)

	if err := http.ListenAndServe(HTTPPort, nil); err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error when I create the server : %v", err))
	}
	mylog.Trace.Printf("Server listen on %s", HTTPPort)
}
