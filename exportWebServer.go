package main

import (
	"bufio"
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/pprof"
	"os"

	"fmt"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
)

var (
	conn       *websocket.Conn
	mybloomsky bloomsky.Bloomsky
	msgJSON    []byte
)

type httpServer struct {
	bloomskyMessageToHTTP chan bloomsky.Bloomsky
	httpServ              *http.Server
}

type pageHome struct {
	Websockerurl string
}

type pageLog struct {
	LogTxt string
}

type logStru struct {
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	Level string `json:"level"`
	Param string `json:"param"`
	Fct   string `json:"fct"`
}

//listen
func (httpServ *httpServer) listen(context context.Context) {
	go func() {
		for {
			mybloomsky := <-httpServ.bloomskyMessageToHTTP

			localJSON, err := json.Marshal(mybloomsky.GetBloomskyStruct())
			msgJSON = localJSON
			checkErr(err, funcName(), "Marshal json Error", "")

			if msgJSON == nil {
				logFatal(err, funcName(), "JSON Empty", "")
			}

			if conn != nil {
				err = conn.WriteMessage(websocket.TextMessage, msgJSON)
				checkErr(err, funcName(), "Impossible to write to websocket", "")
			}

			logDebug(funcName(), "Listen", string(msgJSON))
		}
	}()
}

// Websocket handler to send data
func (httpServ *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Refresh data Websocket handle", "")

	fmt.Printf("ici:%v", msgJSON)

	upgrader := websocket.Upgrader{}

	conn, err := upgrader.Upgrade(w, r, nil)
	checkErr(err, funcName(), "Upgrade upgrader", "")

	if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
		logFatal(err, funcName(), "Impossible to write to websocket", "")
	}
}

// Home bloomsky handler
func (httpServ *httpServer) home(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Home Http handle", "")

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/bloomsky.html", "tmpl/bloomsky_header.html", "tmpl/bloomsky_body.html"}, map[string]interface{}{"T": config.translateFunc}, config.dev)

	p := pageHome{Websockerurl: "ws://" + r.Host + "/refreshdata"}
	if err := t.Execute(w, p); err != nil {
		logFatal(err, funcName(), "Execute template home", "")
	}
}

// Log handler
func (httpServ *httpServer) log(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Log Http handle", "")

	file, err := os.Open("bloomsky.log")
	checkErr(err, funcName(), "Imposible to open file", "bloomsky.log")
	defer file.Close()

	scanner := bufio.NewScanner(file)

	str := "<table class=\"table table-striped\"> <tr> <th> Time </th><th> Level </th><th> Fonction </th><th> Message </th> <th> Parameter </th></tr>"
	for scanner.Scan() {
		var tt logStru
		if err := json.Unmarshal([]byte(scanner.Text()), &tt); err != nil {
			logFatal(err, funcName(), "Impossible to unmarshall log", scanner.Text())
		}

		str += "<tr>"
		str += "<th>" + tt.Time + "</th>"
		str += "<th> <img src=\"\\static\\" + tt.Level + ".png\" width=\"40\"> </th>"
		str += "<th>" + tt.Fct + "</th>"
		str += "<th>" + tt.Msg + "</th>"
		str += "<th>" + tt.Param + "</th>"
		str += "</tr>"
	}
	str += "</table>"

	if err := scanner.Err(); err != nil {
		logFatal(err, funcName(), "Scanner Err", "")
	}

	p := map[string]interface{}{"LogTxt": template.HTML(str)}

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/bloomsky.html", "tmpl/log_header.html", "tmpl/log_body.html"}, map[string]interface{}{"T": config.translateFunc}, config.dev)
	if err := t.Execute(w, p); err != nil {
		logFatal(err, funcName(), "Compile template log", "")
	}
}

//createWebServer create web server
func createWebServer(in chan bloomsky.Bloomsky, HTTPPort string) (*httpServer, error) {
	server := &httpServer{bloomskyMessageToHTTP: in}

	var fs http.Handler
	if config.dev {
		fs = http.FileServer(http.Dir("static"))
	} else {
		fs = http.FileServer(&assetfs.AssetFS{Asset: assemblyAssetfs.Asset, AssetDir: assemblyAssetfs.AssetDir, AssetInfo: assemblyAssetfs.AssetInfo, Prefix: "static"})
	}

	s := http.NewServeMux()

	s.Handle("/static/", http.StripPrefix("/static/", fs))
	s.HandleFunc("/refreshdata", server.refreshdata)
	s.HandleFunc("/", server.home)

	s.HandleFunc("/log", server.log)

	s.HandleFunc("/debug/pprof/", pprof.Index)
	s.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.Handle("/favicon.ico", fs)

	h := &http.Server{Addr: HTTPPort, Handler: s}

	/*
		HTTPS
		i := &http.Server{Addr: ":2222", Handler: s}
		go func() {
			if err := i.ListenAndServeTLS("server.crt", "server.key"); err != nil {
				logrus.Errorf("Error when I create the server HTTPS : %v", err)
			}
		}()
	*/

	go func() {
		if err := h.ListenAndServe(); err != nil {
			logFatal(err, funcName(), "Error when I create the server", "")
		}
	}()

	logInfo(funcName(), "Server listen on port", HTTPPort)

	server.httpServ = h
	return server, nil
}
