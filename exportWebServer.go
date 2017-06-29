package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
	"github.com/sirupsen/logrus"
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

type page struct {
	Websockerurl string
}

func (httpServ *httpServer) listen(context context.Context) {
	go func() {
		for {
			var err error

			mybloomsky := <-httpServ.bloomskyMessageToHTTP
			msgJSON, err = json.Marshal(mybloomsky.GetBloomskyStruct())
			if err != nil {
				log.WithFields(logrus.Fields{
					"error": err,
					"fct":   "exportWebServer.listen",
				}).Fatal("Marshal json Error")
			}

			if msgJSON == nil {
				logrus.WithFields(logrus.Fields{
					"fct": "exportWebServer.listen",
				}).Fatal("JSON Empty")
			}

			if conn != nil {
				err = conn.WriteMessage(websocket.TextMessage, msgJSON)
				if err != nil {
					log.WithFields(logrus.Fields{
						"error": err,
						"fct":   "exportWebServer.listen",
					}).Fatal("Impossible to write to websocket")
				}
			}

			log.WithFields(logrus.Fields{
				"fct": "exportWebServer.listen",
			}).Debug("Message send to browser")
		}
	}()
}

// Websocket handler to send data
func (httpServ *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"fct": "exportWebServer.refreshdata",
	}).Debug("Refresh data Websocket handle")

	upgrader := websocket.Upgrader{}
	var err error

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"fct":   "exportWebServer.refreshdata",
		}).Fatal("Upgrade upgrader")
		return
	}

		if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"fct":   "exportWebServer.refreshdata",
		}).Fatal("Impossible to write to websocket")
	}
}

func (httpServ *httpServer) home(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"JSON": string(msgJSON),
		"fct":  "exportWebServer.home",
	}).Debug("Home Http handle")

	t := GetHTMLTemplate("bloomsky", []string{"tmpl/bloomsky.html", "tmpl/bloomsky_header.html", "tmpl/bloomsky_body.html"}, map[string]interface{}{"T": config.translateFunc}, config.dev)

	//p := page{Websockerurl: "wss://" + r.Host + "/refreshdata"}
	p := page{Websockerurl: "ws://" + r.Host + "/refreshdata"}
	if err := t.Execute(w, p); err != nil {
		log.Fatalf("Write field ws : %v", err)
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
			log.WithFields(logrus.Fields{
				"error": err,
				"fct":   "exportWebServer.createWebServer",
			}).Fatal("Error when I create the server")
		}
	}()

	logrus.WithFields(logrus.Fields{
		"port": HTTPPort,
		"fct":  "exportWebServer.createWebServer",
	}).Info("Server listen")

	server.httpServ = h
	return server, nil
}
