package main

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
	"github.com/patrickalin/bloomsky-client-go/utils"
)

var conn *websocket.Conn
var mybloomsky bloomsky.BloomskyStructure
var msgJSON []byte

type httpServer struct {
	bloomskyMessageToHTTP chan bloomsky.BloomskyStructure
	h                     *http.Server
}

func (h *httpServer) listen(context context.Context) {
	go func() {
		for {
			mybloomsky := <-h.bloomskyMessageToHTTP
			var err error
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

// Websocket handler
func (h *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Refresdata WS handle Send JSON : %s", msgJSON)

	upgrader := websocket.Upgrader{}
	var err error

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Upgrade upgrader : %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		log.Errorf("Impossible to write to websocket : %v", err)
	}

}

func (h *httpServer) home(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Home Http handle Send JSON : %s", msgJSON)

	var err error
	var templateHeader *template.Template
	var templateBody *template.Template

	templateHeader = utils.GetHtmlTemplate("bloomsky_header.html", "tmpl/bloomsky_header.html", map[string]interface{}{"T": config.translateFunc}, config.dev)

	err = templateHeader.Execute(w, "ws://"+r.Host+"/refreshdata")
	if err != nil {
		log.Fatalf("Write part 1 : %v", err)
	}
	templateBody = utils.GetHtmlTemplate("bloomsky_body.html", "tmpl/bloomsky_body.html", map[string]interface{}{"T": config.translateFunc}, config.dev)

	err = templateBody.Execute(w, mybloomsky)
	if err != nil {
		log.Fatalf("Write part 2 : %v", err)
	}
}

//createWebServer create web server
func createWebServer(in chan bloomsky.BloomskyStructure, HTTPPort string) (*httpServer, error) {
	server := &httpServer{bloomskyMessageToHTTP: in}

	fs := http.FileServer(&assetfs.AssetFS{Asset: assemblyAssetfs.Asset, AssetDir: assemblyAssetfs.AssetDir, AssetInfo: assemblyAssetfs.AssetInfo, Prefix: "static"})

	s := http.NewServeMux()

	s.Handle("/static/", http.StripPrefix("/static/", fs))
	s.HandleFunc("/refreshdata", server.refreshdata)
	s.HandleFunc("/", server.home)

	h := &http.Server{Addr: HTTPPort, Handler: s}
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Errorf("Error when I create the server : %v", err)
		}
	}()
	log.Infof("Server listen on port %s", HTTPPort)
	server.h = h
	return server, nil
}
