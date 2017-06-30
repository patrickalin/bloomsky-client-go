package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
)

type httpServer struct {
	bloomskyMessageToHTTP chan bloomsky.Bloomsky
	httpServ              *http.Server
	conn                  *websocket.Conn
	msgJSON               []byte
	translateFunc         i18n.TranslateFunc
	dev                   bool
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
			var err error

			httpServ.msgJSON, err = json.Marshal(mybloomsky.GetBloomskyStruct())
			checkErr(err, funcName(), "Marshal json Error", "")

			if httpServ.msgJSON == nil {
				logFatal(err, funcName(), "JSON Empty", "")
			}

			if httpServ.conn != nil {
				err = httpServ.conn.WriteMessage(websocket.TextMessage, httpServ.msgJSON)
				checkErr(err, funcName(), "Impossible to write to websocket", "")
			}

			logDebug(funcName(), "Listen", string(httpServ.msgJSON))
		}
	}()
}

// Websocket handler to send data
func (httpServ *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Refresh data Websocket handle", "")

	upgrader := websocket.Upgrader{}

	var err error

	httpServ.conn, err = upgrader.Upgrade(w, r, nil)
	checkErr(err, funcName(), "Upgrade upgrader", "")

	if err = httpServ.conn.WriteMessage(websocket.TextMessage, httpServ.msgJSON); err != nil {
		logFatal(err, funcName(), "Impossible to write to websocket", "")
	}
}

func getWs(r *http.Request) string {
	if r.TLS == nil {
		return "ws://"
	}
	return "wss://"
}

// Home bloomsky handler
func (httpServ *httpServer) home(w http.ResponseWriter, r *http.Request) {

	logDebug(funcName(), "Home Http handle", "")

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/bloomsky/script.html", "tmpl/bloomsky/body.html", "tmpl/bloomsky/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": httpServ.translateFunc}, httpServ.dev)

	p := pageHome{Websockerurl: getWs(r) + r.Host + "/refreshdata"}
	if err := t.Execute(w, p); err != nil {
		logFatal(err, funcName(), "Execute template home", "")
	}
}

// Home bloomsky handler
func (httpServ *httpServer) history(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Home History handle", "")

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/history/script.html", "tmpl/history/body.html", "tmpl/history/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": httpServ.translateFunc}, httpServ.dev)

	p := pageHome{Websockerurl: getWs(r) + r.Host + "/refreshdata"}
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

	var logRange []logStru

	for scanner.Scan() {
		var tt logStru
		if err := json.Unmarshal([]byte(scanner.Text()), &tt); err != nil {
			logFatal(err, funcName(), "Impossible to unmarshall log", scanner.Text())
		}
		logRange = append(logRange, tt)
	}

	if err := scanner.Err(); err != nil {
		logFatal(err, funcName(), "Scanner Err", "")
	}

	p := map[string]interface{}{"logRange": logRange}

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/log/script.html", "tmpl/log/body.html", "tmpl/log/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": httpServ.translateFunc}, httpServ.dev)
	if err := t.Execute(w, p); err != nil {
		logFatal(err, funcName(), "Compile template log", "")
	}
}

func getFileServer(dev bool) http.FileSystem {
	if dev {
		return http.Dir("static")
	}
	return &assetfs.AssetFS{Asset: assemblyAssetfs.Asset, AssetDir: assemblyAssetfs.AssetDir, AssetInfo: assemblyAssetfs.AssetInfo, Prefix: "static"}
}

//createWebServer create web server
func createWebServer(in chan bloomsky.Bloomsky, HTTPPort string, HTTPSPort string, translate i18n.TranslateFunc, devel bool) (*httpServer, error) {
	server := &httpServer{bloomskyMessageToHTTP: in,
		dev:           devel,
		translateFunc: translate}

	fs := http.FileServer(getFileServer(devel))

	s := http.NewServeMux()

	s.Handle("/static/", http.StripPrefix("/static/", fs))
	s.Handle("/favicon.ico", fs)
	s.HandleFunc("/", server.home)
	s.HandleFunc("/refreshdata", server.refreshdata)
	s.HandleFunc("/log", server.log)
	s.HandleFunc("/history", server.history)
	s.HandleFunc("/debug/pprof/", pprof.Index)
	s.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.HandleFunc("/debug/pprof/trace", pprof.Trace)

	h := &http.Server{Addr: HTTPPort, Handler: s}
	go func() {
		err := h.ListenAndServe()
		checkErr(err, funcName(), "Error when I create the server HTTP (don't forget ':')", "")
	}()

	hs := &http.Server{Addr: HTTPSPort, Handler: s}
	go func() {
		err := hs.ListenAndServeTLS("server.crt", "server.key")
		checkErr(err, funcName(), "Error when I create the server HTTPS (don't forget ':')", "")
	}()

	logInfo(funcName(), "Server HTTP listen on port", HTTPPort)
	logInfo(funcName(), "Server HTTPS listen on port", HTTPSPort)

	server.httpServ = h
	return server, nil
}
