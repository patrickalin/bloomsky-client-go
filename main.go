// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

//go:generate echo Go Generate!
//go:generate ./command/bindata.sh
//go:generate ./command/bindata-assetfs.sh

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"time"

	_ "net/http/pprof"

	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//configName name of the config file
const configNameFile = "config"
const logFile = "bloomsky.log"

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type configuration struct {
	consoleActivated    bool
	hTTPActivated       bool
	historyActivated    bool
	hTTPPort            string
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
	translateFunc       i18n.TranslateFunc
	dev                 bool
}

var (
	//Version of the code, fill in in compile.sh -ldflags "-X main.Version=`cat VERSION`"
	Version = "No Version Provided"
	//logger
	log = logrus.New()
)

func init() {
	log.Formatter = new(logrus.JSONFormatter)

	err := os.Remove(logFile)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("Failed to log to file, using default stderr")
		return
	}
	log.Out = file
}

func main() {

	//Create context
	logDebug(funcName(), "Create context", "")
	myContext, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
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
		"time":    time.Now().Format(time.RFC850),
		"version": Version,
		"config":  configNameFile,
		"fct":     funcName(),
	}).Info("Bloomsky API")

	//Read configuration from config file
	config, err := readConfig(configNameFile)
	checkErr(err, funcName(), "Problem reading config file", "")

	//Read flags
	logDebug(funcName(), "Get flag from command line", "")
	debug := flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
	flag.Parse()
	if *debug != "" {
		config.logLevel = *debug
	}

	// Set Level log
	level, err := logrus.ParseLevel(config.logLevel)
	checkErr(err, funcName(), "Error parse level", "")
	log.Level = level
	logInfo(funcName(), "Level log", config.logLevel)

	// Context
	ctxsch := context.Context(myContext)

	channels := make(map[string]chan bloomsky.Bloomsky)

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readFile("lang/en-us.all.json", config.dev)); err != nil {
		logFatal(err, funcName(), "Error read language file check in config.yaml if dev=false", "")
	}
	if err := i18n.ParseTranslationFileBytes("lang/fr.all.json", readFile("lang/fr.all.json", config.dev)); err != nil {
		logFatal(err, funcName(), "Error read language file check in config.yaml if dev=false", "")
	}

	translateFunc, err := i18n.Tfunc(config.language)
	checkErr(err, funcName(), "Problem with loading translate file", "")

	if config.historyActivated {
		channels["store"] = make(chan bloomsky.Bloomsky)
		s, err := createStore(channels["store"])
		checkErr(err, funcName(), "Error with history create store", "")
		s.listen(context.Background())
	}

	// Console initialisation
	if config.consoleActivated {
		channels["console"] = make(chan bloomsky.Bloomsky)
		c, err := createConsole(channels["console"], translateFunc, dev)
		checkErr(err, funcName(), "Error with initConsol", "")
		c.listen(context.Background())
	}

	// InfluxDB initialisation
	if config.influxDBActivated {
		channels["influxdb"] = make(chan bloomsky.Bloomsky)
		c, err := initClient(channels["influxdb"], config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
		checkErr(err, funcName(), "Error with initClientInfluxDB", "")
		c.listen(context.Background())
	}

	// WebServer initialisation
	var httpServ *httpServer
	if config.hTTPActivated {
		var err error
		channels["web"] = make(chan bloomsky.Bloomsky)
		httpServ, err = createWebServer(channels["web"], config.hTTPPort, translateFunc, config.dev)
		checkErr(err, funcName(), "Error with initWebServer", "")
		httpServ.listen(context.Background())

	}

	// get bloomsky JSON and parse information in bloomsky Go Structure
	mybloomsky := bloomsky.New(config.bloomskyURL, config.bloomskyAccessToken, config.mock, log)
	//Call scheduler
	schedule(ctxsch, mybloomsky, channels, config.refreshTimer)

	//If signal to close the program
	<-myContext.Done()
	if httpServ.httpServ != nil {
		logDebug(funcName(), "Shutting down webserver", "")
		err := httpServ.httpServ.Shutdown(myContext)
		checkErr(err, funcName(), "Impossible to shutdown context", "")
	}

	logrus.WithFields(logrus.Fields{
		"fct": "main.main",
	}).Debug("Terminated see bloomsky.log")
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
			logDebug(funcName(), "Stoping ticker", "")
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
	logDebug(funcName(), "Parse informations from API bloomsky", "")

	mybloomsky.Refresh()

	//send message on each channels
	for _, v := range channels {
		v <- mybloomsky
	}
}

// ReadConfig read config from config.json with the package viper
func readConfig(configName string) (configuration, error) {

	var conf configuration
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	checkErr(err, funcName(), "Fielpaths", "")
	dir = dir + "/" + configName

	if err := viper.ReadInConfig(); err != nil {
		logFatal(err, funcName(), "The config file loaded", dir)
		return conf, err
	}
	logInfo(funcName(), "The config file loaded", dir)

	//TODO#16 find to simplify this section
	conf.bloomskyURL = viper.GetString("BloomskyURL")
	conf.bloomskyAccessToken = viper.GetString("BloomskyAccessToken")
	conf.influxDBDatabase = viper.GetString("InfluxDBDatabase")
	conf.influxDBPassword = viper.GetString("InfluxDBPassword")
	conf.influxDBServer = viper.GetString("InfluxDBServer")
	conf.influxDBServerPort = viper.GetString("InfluxDBServerPort")
	conf.influxDBUsername = viper.GetString("InfluxDBUsername")
	conf.consoleActivated = viper.GetBool("ConsoleActivated")
	conf.influxDBActivated = viper.GetBool("InfluxDBActivated")
	conf.historyActivated = viper.GetBool("historyActivated")
	conf.refreshTimer = time.Duration(viper.GetInt("RefreshTimer")) * time.Second
	conf.hTTPActivated = viper.GetBool("HTTPActivated")
	conf.hTTPPort = viper.GetString("HTTPPort")
	conf.logLevel = viper.GetString("LogLevel")
	conf.mock = viper.GetBool("mock")
	conf.language = viper.GetString("language")
	conf.dev = viper.GetBool("dev")

	// Check if one value of the structure is empty
	v := reflect.ValueOf(conf)
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i)
		//TODO#16
		//v.Field(i).SetString(viper.GetString(v.Type().Field(i).Name))
		if values[i] == "" {
			return conf, fmt.Errorf("Check if the key " + v.Type().Field(i).Name + " is present in the file " + dir)
		}
	}
	if token := os.Getenv("bloomskyAccessToken"); token != "" {
		conf.bloomskyAccessToken = token
	}
	return conf, nil
}

//Read file and return []byte
func readFile(fileName string, dev bool) []byte {
	if dev {
		fileByte, err := ioutil.ReadFile(fileName)
		checkErr(err, funcName(), "Error reading the file", fileName)
		return fileByte
	}

	fileByte, err := assembly.Asset(fileName)
	checkErr(err, funcName(), "Error reading the file", fileName)
	return fileByte
}
