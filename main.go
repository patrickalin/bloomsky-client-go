// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

//go:generate echo Go Generate!
//go:generate ./scripts/build/bindata.sh
//go:generate ./scripts/build/bindata-assetfs.sh

import (
	"context"
	"fmt"
	"regexp"

	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	_ "net/http/pprof"

	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//configName name of the config file and log file
const (
	configNameFile = "config"
	logFile        = "bloomsky.log"
)

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type configuration struct {
	consoleActivated    bool
	hTTPActivated       bool
	historyActivated    bool
	hTTPPort            string
	hTTPSPort           string
	influxDBActivated   bool
	influxDBDatabase    string
	influxDBPassword    string
	influxDBServer      string
	influxDBServerPort  string
	influxDBUsername    string
	logLevel            string
	bloomskyAccessToken string
	bloomskyURL         string
	refreshTimer        time.Duration
	mock                bool
	language            string
	dev                 bool
	wss                 bool
}

// DO NOT EDIT THIS FILE DIRECTLY. These are build-time constants
// set through ‘buildscripts/gen-ldflags.go’.
var (
	// Go get development tag.
	goGetTag = "DEVELOPMENT.GOGET"
	// Version - version time.RFC3339.
	Version = goGetTag
	// ReleaseTag - release tag in TAG.%Y-%m-%dT%H-%M-%SZ.
	ReleaseTag = goGetTag
	// CommitID - latest commit id.
	CommitID = goGetTag
	// ShortCommitID - first 12 characters from CommitID.
	ShortCommitID = CommitID[:12]
	//logger
	log = logrus.New()
)

func init() {
	log.Formatter = new(logrus.JSONFormatter)

	err := os.Remove(logFile)
	if err != nil {
		log.Info("Failed to remove log file")
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("Failed to log to file, using default stderr")
		return
	}
	log.Out = file
}

type stopServer func()

func startServer(mycontext context.Context, config configuration) stopServer {
	// Set Level log
	level, err := logrus.ParseLevel(config.logLevel)
	checkErr(err, funcName(), "Error parse level")
	log.Level = level
	logInfo(funcName(), "Level log", config.logLevel)

	// Context
	ctxsch := context.Context(mycontext)

	channels := make(map[string]chan bloomsky.Bloomsky)

	// Traduction
	err = i18n.ParseTranslationFileBytes("lang/en-us.all.json", readFile("lang/en-us.all.json", config.dev))
	checkErr(err, funcName(), "Error read language file check in config.yaml if dev=false")
	err = i18n.ParseTranslationFileBytes("lang/fr.all.json", readFile("lang/fr.all.json", config.dev))
	checkErr(err, funcName(), "Error read language file check in config.yaml if dev=false")
	translateFunc, err := i18n.Tfunc(config.language)
	checkErr(err, funcName(), "Problem with loading translate file")

	// Console initialisation
	if config.consoleActivated {
		channels["console"] = make(chan bloomsky.Bloomsky)
		c, err := createConsole(channels["console"], translateFunc, config.dev)
		checkErr(err, funcName(), "Error with initConsol")
		ctxcsl, cancelcsl := context.WithCancel(mycontext)
		defer cancelcsl()
		c.listen(ctxcsl)
	}

	// InfluxDB initialisation
	if config.influxDBActivated {
		channels["influxdb"] = make(chan bloomsky.Bloomsky)
		c, err := initClient(channels["influxdb"], config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
		checkErr(err, funcName(), "Error with initClientInfluxDB")
		c.listen(context.Background())
	}

	// WebServer initialisation
	var httpServ *httpServer

	if config.hTTPActivated {
		channels["store"] = make(chan bloomsky.Bloomsky)

		store, err := createStore(channels["store"])
		checkErr(err, funcName(), "Error with history create store")
		ctxtstroe, cancelstore := context.WithCancel(mycontext)
		defer cancelstore()

		store.listen(ctxtstroe)

		channels["web"] = make(chan bloomsky.Bloomsky)

		httpServ, err = createWebServer(channels["web"], config.hTTPPort, config.hTTPSPort, translateFunc, config.dev, store, config.wss)
		checkErr(err, funcName(), "Error with initWebServer")
		ctxthttp, cancelhttp := context.WithCancel(mycontext)
		defer cancelhttp()
		httpServ.listen(ctxthttp)
	}

	// get bloomsky JSON and parse information in bloomsky Go Structure
	mybloomsky := bloomsky.New(config.bloomskyURL, config.bloomskyAccessToken, config.mock, log)
	//Call scheduler
	schedule(ctxsch, mybloomsky, channels, config.refreshTimer)

	return func() {
		log.Debug(funcName(), "shutting down")
		checkErr(httpServ.shutdown(context.Context(mycontext)), funcName(), "http server issue")
		logrus.WithFields(logrus.Fields{
			"fct": "main.main",
		}).Debug("Terminated see bloomsky.log")
		os.Exit(0)

	}

}
func main() {

	//Create context
	logDebug(funcName(), "Create context")
	myContext, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh)
	go func() {
		select {
		case i := <-signalCh:
			logDebug(funcName(), "Receive interrupt", i.String())
			cancel()
			return
		}
	}()

	logrus.WithFields(logrus.Fields{
		"time":          time.Now().Format(time.RFC850),
		"version":       Version,
		"release-tag":   ReleaseTag,
		"Commit-ID":     CommitID,
		"ShortCommitID": ShortCommitID,
		"config":        configNameFile,
		"fct":           funcName(),
	}).Info("Bloomsky API")
	config := readConfig(configNameFile, validateHTTPPort, validateHTTSPort)
	stop := startServer(myContext, config)
	defer stop()
	//If signal to close the program

	<-myContext.Done()
	log.Debug("going to stop")

}

// The scheduler executes each time "collect"
func schedule(myContext context.Context, mybloomsky bloomsky.Bloomsky, channels map[string]chan bloomsky.Bloomsky, refreshTime time.Duration) {
	ticker := time.NewTicker(refreshTime)
	logDebug(funcName(), "Create scheduler", refreshTime.String())

	collect(mybloomsky, channels)
	for {
		select {
		case <-ticker.C:
			collect(mybloomsky, channels)
		case <-myContext.Done():
			logDebug(funcName(), "Stoping ticker")
			ticker.Stop()
			for _, v := range channels {
				close(v)
			}
			return
		}
	}
}

//Principal function which one loops each Time Variable
func collect(mybloomsky bloomsky.Bloomsky, channels map[string]chan bloomsky.Bloomsky) {
	logDebug(funcName(), "Parse informations from API bloomsky")

	mybloomsky.Refresh()

	//send message on each channels
	for _, v := range channels {
		v <- mybloomsky
	}
}

// ReadConfig read config from config.json with the package viper
func readConfig(configName string, options ...validation) configuration {

	var conf configuration

	pflag.String("main.bloomsky.token", "rrrrr", "yourtoken")
	pflag.Bool("main.dev", false, "developpement mode")
	pflag.Bool("main.mock", false, "use mock  mode")

	fmt.Println("ici")

	pflag.Parse()

	fmt.Println("la")

	//viper.BindFlagValue("main.bloomsky.token")
	viper.SetConfigType("yaml")
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("test")
	err := viper.BindPFlags(pflag.CommandLine)
	checkErr(err, funcName(), "Error with bindPFlags")

	viper.SetDefault("main.language", "en-us")
	viper.SetDefault("main.RefreshTimer", 60)
	viper.SetDefault("main.bloomsky.url", "https://api.bloomsky.com/api/skydata/")
	viper.SetDefault("main.log.level", "panic")
	viper.SetDefault("outputs.influxdb.activated", false)
	viper.SetDefault("outputs.web.activated", true)
	viper.SetDefault("outputs.web.port", ":1111")
	viper.SetDefault("outputs.web.secureport", ":1112")
	viper.SetDefault("outputs.console.activated", true)
	// trying to read config file
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	dir = dir + "/" + configName
	checkErr(err, funcName(), "Fielpaths", dir)
	if err := viper.ReadInConfig(); err != nil {
		logWarn(funcName(), "Config file not loaded error we use flag and default value", os.Args[0])
	}

	conf.mock = viper.GetBool("main.mock")

	conf.dev = viper.GetBool("main.dev")

	//TODO#16 find to simplify this section
	main := viper.Sub("main")
	conf.bloomskyURL = main.GetString("bloomsky.url")

	conf.bloomskyAccessToken = viper.GetString("main.bloomsky.token")

	conf.language = main.GetString("language")
	conf.logLevel = main.GetString("log.level")
	conf.wss = main.GetBool("wss")
	conf.historyActivated = viper.GetBool("historyActivated")
	conf.refreshTimer = time.Duration(main.GetInt("refreshTimer")) * time.Second

	web := viper.Sub("outputs.web")
	conf.hTTPActivated = web.GetBool("activated")
	conf.hTTPPort = web.GetString("port")
	conf.hTTPSPort = web.GetString("secureport")

	console := viper.Sub("outputs.console")
	conf.consoleActivated = console.GetBool("activated")

	influxdb := viper.Sub("outputs.influxdb")
	conf.influxDBDatabase = influxdb.GetString("database")
	conf.influxDBPassword = influxdb.GetString("password")
	conf.influxDBServer = influxdb.GetString("server")
	conf.influxDBServerPort = influxdb.GetString("port")
	conf.influxDBUsername = influxdb.GetString("username")
	conf.influxDBActivated = viper.GetBool("activated")

	// configuration validation
	for _, option := range options {
		err := option(&conf)
		if err != nil {
			if IsTemporary(err) {
				logDebug("readconfig", "temporary error reading config file", err.Error())
			}
			panic(err)
		}

	}
	return conf
}

//Read file and return []byte
func readFile(fileName string, dev bool) []byte {
	if dev {
		fileByte, err := ioutil.ReadFile(fileName)
		checkErr(err, funcName(), "Error reading the file", fileName)
		return fileByte
	}

	fileByte, err := assembly.Asset(fileName)
	checkErr(err, funcName(), "Error reading the file as an asset", fileName)
	return fileByte
}

type validation func(conf *configuration) error

type validationError struct {
	fields    string
	msg       string
	cause     error
	temporary bool
}

func (err validationError) Temporary() bool {
	return err.temporary
}

func (err validationError) Error() string {
	return fmt.Sprintf("fields %s  give msg %s with err %s", err.fields, err.msg, err.cause)
}

type temporary interface {
	Temporary() bool
}

// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

/* Validation Funciton  */
func validateHTTPPort(conf *configuration) error {
	if err := validatePort(conf.hTTPPort); err != nil {
		return validationError{
			fields:    "HTTPport",
			cause:     err,
			temporary: !conf.hTTPActivated,
		}
	}
	return nil
}
func validateHTTSPort(conf *configuration) error {
	if err := validatePort(conf.hTTPSPort); err != nil {
		return validationError{
			fields:    "HTTPport",
			cause:     err,
			temporary: !conf.hTTPActivated,
		}
	}
	return nil
}

func validatePort(port string) error {
	matched, err := regexp.MatchString("^:[0-9]*$", port)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid string for a port %s", port)
	}
	return nil
}
