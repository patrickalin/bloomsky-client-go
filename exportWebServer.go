package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
	"github.com/patrickalin/bloomsky-client-go/utils"
	"github.com/sirupsen/logrus"
)

var (
	conn       *websocket.Conn
	mybloomsky bloomsky.BloomskyStructure
	msgJSON    []byte
)

type httpServer struct {
	bloomskyMessageToHTTP chan bloomsky.BloomskyStructure
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
			msgJSON, err = json.Marshal(mybloomsky)
			log.Debugf("JSON : %s", msgJSON)

			if err != nil {
				log.Infof("Marshal json Error: %v", err)
				return
			}

			if conn != nil {
				err = conn.WriteMessage(websocket.TextMessage, msgJSON)
				if err != nil {
					log.Infof("Impossible to write to websocket : %v", err)
				}
			}
			log.Debug("Message send to browser")
		}
	}()
}

// Websocket handler to send data
func (httpServ *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Refresdata WS handle Send JSON : %s", msgJSON)

	upgrader := websocket.Upgrader{}
	var err error

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade upgrader : %v", err)
		return
	}

	if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
		log.Errorf("Impossible to write to websocket : %v", err)
	}
}

func (h *httpServer) home(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Home Http handle Send JSON : %s", msgJSON)

	t := utils.GetHtmlTemplate("bloomsky", []string{"tmpl/bloomsky.html", "tmpl/bloomsky_header.html", "tmpl/bloomsky_body.html"}, map[string]interface{}{"T": config.translateFunc}, config.dev)

	p := page{Websockerurl: "wss://" + r.Host + "/refreshdata"}
	if err := t.Execute(w, p); err != nil {
		log.Fatalf("Write part 1 : %v", err)
	}
}

//createWebServer create web server
func createWebServer(in chan bloomsky.BloomskyStructure, HTTPPort string) (*httpServer, error) {
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
	s.Handle("/favicon.ico", fs)

	h := &http.Server{Addr: HTTPPort, Handler: s}

	/*i := &http.Server{Addr: ":2222", Handler: s}

	go func() {
		if err := i.ListenAndServeTLS("server.crt", "server.key"); err != nil {
			logrus.Errorf("Error when I create the server HTTPS : %v", err)
		}
	}()*/

	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Errorf("Error when I create the server : %v", err)
		}
	}()
	logrus.Infof("Server listen on port %s", HTTPPort)
	server.httpServ = h
	return server, nil
}
