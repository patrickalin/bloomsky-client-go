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
const mockFile = "test-mock/mock.json"
const logFile = "bloomsky.log"

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type configuration struct {
	consoleActivated    bool
	hTTPActivated       bool
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
	//Version of the code
	Version = "No Version Provided"
	//record the configuration parameter
	config configuration

	channels = make(map[string]chan bloomsky.Bloomsky)

	debug = flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
	//logger
	log              = logrus.New()
	responseBloomsky []byte
)

func init() {
	//log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("Failed to log to file, using default stderr")
		return
	}
	log.Out = file
}

func main() {
	log.Debug("Create context")
	myContext, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		select {
		case i := <-signalCh:
			log.Debugf("Receive interrupt  %v", i)
			cancel()
			return
		}
	}()

	logrus.WithFields(logrus.Fields{
		"time":    time.Now().Format(time.RFC850),
		"version": Version,
		"config":  configNameFile,
		"fct":     "main.main",
	}).Info("Bloomsky API")

	if err := readConfig(configNameFile); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"fct":   "main.main",
		}).Info("Problem reading config file")
	}

	log.WithFields(logrus.Fields{
		"fct": "main.main",
	}).Debug("Get flag from command line")

	flag.Parse()
	if *debug != "" {
		config.logLevel = *debug
	}

	// Level Debug
	level, err := logrus.ParseLevel(config.logLevel)
	if err != nil {
		log.WithFields(logrus.Fields{
			"fct":   "main.main",
			"error": err,
		}).Fatal("Error parse level")
	}

	log.Level = level
	log.WithFields(logrus.Fields{
		"fct":   "main.main",
		"level": level,
	}).Info("Level trace")

	// Context
	ctxsch := context.Context(myContext)

	// Mock
	if config.mock {
		responseBloomsky = readMockFile(mockFile)
	}

	// Console
	if config.consoleActivated {
		channels["console"] = make(chan bloomsky.Bloomsky)
		c, err := createConsole(channels["console"])
		if err != nil {
			log.WithFields(logrus.Fields{
				"fct":   "main.main",
				"error": err,
			}).Fatal("Error with initConsol")
		}
		c.listen(context.Background())
	}

	// InfluxDB
	if config.influxDBActivated {
		channels["influxdb"] = make(chan bloomsky.Bloomsky)
		c, err := initClient(channels["influxdb"], config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
		if err != nil {
			log.WithFields(logrus.Fields{
				"fct":   "main.main",
				"error": err,
			}).Fatal("Error with initClientInfluxDB")
		}
		c.listen(context.Background())

	}

	// WebServer
	var httpServ *httpServer
	if config.hTTPActivated {
		var err error
		channels["web"] = make(chan bloomsky.Bloomsky)
		httpServ, err = createWebServer(channels["web"], config.hTTPPort)
		if err != nil {
			log.WithFields(logrus.Fields{
				"fct":   "main.main",
				"error": err,
			}).Fatal("Error with initWebServer")
		}
		httpServ.listen(context.Background())

	}

	schedule(ctxsch)

	<-myContext.Done()
	if httpServ.httpServ != nil {
		log.WithFields(logrus.Fields{
			"fct": "main.main",
		}).Debug("Shutting down ws")
		httpServ.httpServ.Shutdown(myContext)
	}

	logrus.WithFields(logrus.Fields{
		"fct": "main.main",
	}).Debug("Terminated see bloomsky.log")
}

// The scheduler execute each time collect
func schedule(myContext context.Context) {
	ticker := time.NewTicker(config.refreshTimer)
	log.WithFields(logrus.Fields{
		"fct": "main.schedule",
	}).Debug("Create scheduler")

	collect()
	for {
		select {
		case <-ticker.C:
			collect()
		case <-myContext.Done():
			log.WithFields(logrus.Fields{
				"fct": "main.schedule",
			}).Debug("Stoping ticker")
			ticker.Stop()
			for _, v := range channels {
				close(v)
			}
			return
		}
	}
}

//Principal function which one loops each Time Variable
func collect() {
	log.WithFields(logrus.Fields{
		"fct":          "main.collect",
		"Refresh Time": config.refreshTimer,
	}).Debug("Parse informations from API bloomsky")

	// get bloomsky JSON and parse information in bloomsky Go Structure
	var mybloomsky = bloomsky.New(config.bloomskyURL, config.bloomskyAccessToken, log)
	if config.mock {
		mybloomsky.RefreshFromBody(responseBloomsky)
	} else {
		log.Debug("Mock desactivated")
		mybloomsky.RefreshFromRest()
	}

	//send message on each channels
	for _, v := range channels {
		v <- mybloomsky
	}
}

// ReadConfig read config from config.json
// with the package viper
func readConfig(configName string) (err error) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	dir = dir + "/" + configName
	log.WithFields(logrus.Fields{
		"config": dir + configName,
		"fct":    "main.readConfig",
	}).Info("The config file loaded")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	//TODO#16 find to simplify this section
	config.bloomskyURL = viper.GetString("BloomskyURL")
	config.bloomskyAccessToken = viper.GetString("BloomskyAccessToken")
	config.influxDBDatabase = viper.GetString("InfluxDBDatabase")
	config.influxDBPassword = viper.GetString("InfluxDBPassword")
	config.influxDBServer = viper.GetString("InfluxDBServer")
	config.influxDBServerPort = viper.GetString("InfluxDBServerPort")
	config.influxDBUsername = viper.GetString("InfluxDBUsername")
	config.consoleActivated = viper.GetBool("ConsoleActivated")
	config.influxDBActivated = viper.GetBool("InfluxDBActivated")
	config.refreshTimer = time.Duration(viper.GetInt("RefreshTimer")) * time.Second
	config.hTTPActivated = viper.GetBool("HTTPActivated")
	config.hTTPPort = viper.GetString("HTTPPort")
	config.logLevel = viper.GetString("LogLevel")
	config.mock = viper.GetBool("mock")
	config.language = viper.GetString("language")
	config.dev = viper.GetBool("dev")

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readTranslationResource("lang/en-us.all.json")); err != nil {
		log.Fatalf("Error read language file : %v check in config.yaml if dev=false", err)
	}
	if err := i18n.ParseTranslationFileBytes("lang/fr.all.json", readTranslationResource("lang/fr.all.json")); err != nil {
		log.Fatalf("Error read language file : %v check in config.yaml if dev=false", err)
	}

	config.translateFunc, err = i18n.Tfunc(config.language)
	if err != nil {
		log.Errorf("Problem with loading translate file, %v", err)
	}

	// Check if one value of the structure is empty
	v := reflect.ValueOf(config)
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i)
		//TODO#16
		//v.Field(i).SetString(viper.GetString(v.Type().Field(i).Name))
		if values[i] == "" {
			return fmt.Errorf("Check if the key " + v.Type().Field(i).Name + " is present in the file " + dir)
		}
	}
	if token := os.Getenv("bloomskyAccessToken"); token != "" {
		config.bloomskyAccessToken = token
	}
	return nil
}

//Read translation ressources from /lang or the assembly
func readTranslationResource(name string) []byte {
	if config.dev {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			log.Errorf("Error read language file from folder /lang : %v check in config.yaml if dev=false", err)
		}
		return b
	}

	b, err := assembly.Asset(name)
	if err != nil {
		log.Errorf("Error read language file from assembly : %v", err)
	}
	return b
}

//If mock activated load the file mock and place it in the responseBloomsky
func readMockFile(fMock string) []byte {
	logrus.Warn("Mock activated !!!")

	if config.dev {
		mockFile, err := ioutil.ReadFile(fMock)
		if err != nil {
			log.Fatalf("Error in reading the file %s Err:  %v", fMock, err)
		}
		return mockFile
	}

	mockFile, err := assembly.Asset(fMock)
	if err != nil {
		log.Fatalf("Error in reading the file %s Err:  %v", fMock, err)
	}
	return mockFile
}
