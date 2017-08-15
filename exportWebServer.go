package main

import (
	"bufio"
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly-assetfs"
)

type httpServer struct {
	bloomskyMessageToHTTP chan bloomsky.Bloomsky
	httpServs             [2]*http.Server
	conn                  *websocket.Conn
	msgJSON               []byte
	templates             map[string]*template.Template
	store                 store
	wss                   bool
	config                configuration
}
type pageHome struct {
	Websockerurl string
}

type pageParam struct {
	ConsoleActivated    string
	HTTPActivated       string
	HTTPPort            string
	HTTPSPort           string
	InfluxDBActivated   string
	InfluxDBDatabase    string
	InfluxDBPassword    string
	InfluxDBServer      string
	InfluxDBServerPort  string
	InfluxDBUsername    string
	LogLevel            string
	BloomskyAccessToken string
	BloomskyURL         string
	RefreshTimer        string
	Mock                string
	Language            string
	Dev                 string
	Wss                 string
}

type logStru struct {
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	Level string `json:"level"`
	Param string `json:"param"`
	Fct   string `json:"fct"`
}

func (h *httpServer) shutdown(mycontext context.Context) error {
	log.Debug("shutting down web server")

	for _, h := range h.httpServs {
		if h == nil {
			return nil
		}
		if err := h.Shutdown(mycontext); err != nil {
			logWarn("shutdown", "error while shutting down an http server")
		}
	}
	return nil
}

const logfile = "bloomsky.log"

//listen
func (h *httpServer) listen(context context.Context) {
	go func() {
		for {
			select {
			case mybloomsky := <-h.bloomskyMessageToHTTP:
				var err error

				h.msgJSON, err = json.Marshal(mybloomsky.GetBloomskyStruct())
				checkErr(err, funcName(), "Marshal json Error")

				if h.msgJSON == nil {
					logFatal(err, funcName(), "JSON Empty")
				}

				if h.conn != nil {
					h.refreshWebsocket()
				}

				logDebug(funcName(), "Listen", string(h.msgJSON))
			case <-context.Done():
				return
			}

		}
	}()
}

func (h *httpServer) refreshWebsocket() {
	t := append(h.msgJSON, []byte("SEPARATOR"+h.store.String("temperatureCelsius"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("pressureHPa"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("windGustkmh"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("windSustainedSpeedkmh"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("humidity"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("rainDailyMm"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("rainMm"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("rainRate"))...)
	t = append(t, []byte("SEPARATOR"+h.store.String("indexUV"))...)
	err := h.conn.WriteMessage(websocket.TextMessage, t)
	checkErr(err, funcName(), "Impossible to write to websocket")
}

// Websocket handler to send data
func (h *httpServer) refreshdata(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Refresh data Websocket handle")

	upgrader := websocket.Upgrader{}

	var err error

	h.conn, err = upgrader.Upgrade(w, r, nil)
	checkErr(err, funcName(), "Upgrade upgrader")

	if err = h.conn.WriteMessage(websocket.TextMessage, h.msgJSON); err != nil {
		logFatal(err, funcName(), "Impossible to write to websocket")
	}
}

// Websocket handler to send data
func (h *httpServer) refreshHistory(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Refresh history Websocket handle")

	upgrader := websocket.Upgrader{}

	var err error

	h.conn, err = upgrader.Upgrade(w, r, nil)
	checkErr(err, funcName(), "Upgrade upgrader")

	h.refreshWebsocket()
}

func getWs(r *http.Request, wss bool) string {
	if r.TLS != nil || wss {
		return "wss://"
	}
	return "ws://"
}

// Home bloomsky handler
func (h *httpServer) home(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Home Http handle")

	p := pageHome{Websockerurl: getWs(r, h.wss) + r.Host + "/refreshdata"}
	if err := h.templates["home"].Execute(w, p); err != nil {
		logFatal(err, funcName(), "Execute template home")
	}
}

// History bloomsky handler
func (h *httpServer) history(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "History Http handle")

	p := pageHome{Websockerurl: getWs(r, h.wss) + r.Host + "/refreshhistory"}
	if err := h.templates["history"].Execute(w, p); err != nil {
		logFatal(err, funcName(), "Execute template history")
	}
}

// Parameter bloomsky handler
func (h *httpServer) parameter(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Parameter Http handle")

	p := pageParam{
		ConsoleActivated:    strconv.FormatBool(h.config.consoleActivated),
		HTTPActivated:       strconv.FormatBool(h.config.hTTPActivated),
		HTTPPort:            h.config.hTTPPort,
		HTTPSPort:           h.config.hTTPSPort,
		InfluxDBActivated:   strconv.FormatBool(h.config.influxDBActivated),
		InfluxDBDatabase:    h.config.influxDBDatabase,
		InfluxDBPassword:    h.config.influxDBPassword,
		InfluxDBServer:      h.config.influxDBServer,
		InfluxDBServerPort:  h.config.influxDBServerPort,
		InfluxDBUsername:    h.config.influxDBUsername,
		LogLevel:            h.config.logLevel,
		BloomskyAccessToken: h.config.bloomskyAccessToken,
		BloomskyURL:         h.config.bloomskyURL,
		RefreshTimer:        h.config.refreshTimer.String(),
		Mock:                strconv.FormatBool(h.config.mock),
		Language:            h.config.language,
		Dev:                 strconv.FormatBool(h.config.dev),
		Wss:                 strconv.FormatBool(h.config.wss)}

	if err := h.templates["parameter"].Execute(w, p); err != nil {
		logFatal(err, funcName(), "Execute template parameter")
	}
}

// Log handler
func (h *httpServer) log(w http.ResponseWriter, r *http.Request) {
	logDebug(funcName(), "Log Http handle")

	p := map[string]interface{}{"logRange": createArrayLog(logfile)}

	err := h.templates["log"].Execute(w, p)
	checkErr(err, funcName(), "Compile template log")
}

func getFileServer(dev bool) http.FileSystem {
	if dev {
		return http.Dir("static")
	}
	return &assetfs.AssetFS{Asset: assemblyAssetfs.Asset, AssetDir: assemblyAssetfs.AssetDir, AssetInfo: assemblyAssetfs.AssetInfo, Prefix: "static"}
}

//createWebServer create web server
func createWebServer(in chan bloomsky.Bloomsky, HTTPPort string, HTTPSPort string, translate i18n.TranslateFunc, devel bool, store store, wss bool) (*httpServer, error) {

	t := make(map[string]*template.Template)
	t["home"] = GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/bloomsky/script.html", "tmpl/bloomsky/body.html", "tmpl/bloomsky/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": translate}, devel)
	t["history"] = GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/history/script.html", "tmpl/history/body.html", "tmpl/history/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": translate}, devel)
	t["log"] = GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/log/script.html", "tmpl/log/body.html", "tmpl/log/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": translate}, devel)
	t["parameter"] = GetHTMLTemplate("bloomsky", []string{"tmpl/index.html", "tmpl/parameter/script.html", "tmpl/parameter/body.html", "tmpl/parameter/menu.html", "tmpl/header.html", "tmpl/endScript.html"}, map[string]interface{}{"T": translate}, devel)

	server := &httpServer{bloomskyMessageToHTTP: in,
		templates: t,
		store:     store}

	fs := http.FileServer(getFileServer(devel))

	s := http.NewServeMux()

	s.Handle("/static/", http.StripPrefix("/static/", fs))
	s.Handle("/favicon.ico", fs)
	s.HandleFunc("/", server.home)
	s.HandleFunc("/refreshdata", server.refreshdata)
	s.HandleFunc("/refreshhistory", server.refreshHistory)
	s.HandleFunc("/log", server.log)
	s.HandleFunc("/history", server.history)
	s.HandleFunc("/parameter", server.parameter)
	s.HandleFunc("/debug/pprof/", pprof.Index)
	s.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.HandleFunc("/debug/pprof/trace", pprof.Trace)

	h := &http.Server{Addr: HTTPPort, Handler: s}
	go func() {
		err := h.ListenAndServe()
		checkErr(err, funcName(), "Error when I create the server HTTP (don't forget ':')")
	}()

	hs := &http.Server{Addr: HTTPSPort, Handler: s}
	go func() {
		err := hs.ListenAndServeTLS("server.crt", "server.key")
		checkErr(err, funcName(), "Error when I create the server HTTPS (don't forget ':')")
	}()

	logInfo(funcName(), "Server HTTP listen on port", HTTPPort)
	logInfo(funcName(), "Server HTTPS listen on port", HTTPSPort)

	server.httpServs[0] = h
	server.httpServs[1] = hs
	server.wss = wss

	return server, nil
}

func createArrayLog(logFile string) (logRange []logStru) {
	file, err := os.Open(logFile)
	checkErr(err, funcName(), "Imposible to open file", logFile)

	defer func() {
		err = file.Close()
		checkErr(err, funcName(), "Imposible to close file", logFile)
	}()
	scanner := bufio.NewScanner(file)

	var tt logStru
	for scanner.Scan() {
		err = json.Unmarshal([]byte(scanner.Text()), &tt)
		checkErr(err, funcName(), "Impossible to unmarshall log", scanner.Text())

		logRange = append(logRange, tt)
	}

	err = scanner.Err()
	checkErr(err, funcName(), "Scanner Err")

	return logRange
}
