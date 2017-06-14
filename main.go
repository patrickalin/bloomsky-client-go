// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/nicksnyder/go-i18n/i18n"
	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	"github.com/spf13/viper"
)

//configName name of the config file
const configName = "config"

//VERSION of the code
var Version = "No Version Provided"

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type configuration struct {
	consoleActivated    string
	hTTPActivated       string
	hTTPPort            string
	influxDBActivated   string
	influxDBDatabase    string
	influxDBPassword    string
	influxDBServer      string
	influxDBServerPort  string
	influxDBUsername    string
	logLevel            string
	bloomskyAccessToken string
	bloomskyURL         string
	refreshTimer        string
	mock                bool
	language            string
	translateFunc       i18n.TranslateFunc
	dev                 bool
}

var config configuration

var (
	bloomskyMessageToConsole  = make(chan bloomsky.BloomskyStructure)
	bloomskyMessageToInfluxDB = make(chan bloomsky.BloomskyStructure)
	bloomskyMessageToHTTP     = make(chan bloomsky.BloomskyStructure)

	myTime time.Duration
	debug  = flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
)

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
	fmt.Printf("The config file loaded is :> %s/%s \n \n", dir, configName)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	config.bloomskyURL = viper.GetString("BloomskyURL")
	config.bloomskyAccessToken = viper.GetString("BloomskyAccessToken")
	config.influxDBDatabase = viper.GetString("InfluxDBDatabase")
	config.influxDBPassword = viper.GetString("InfluxDBPassword")
	config.influxDBServer = viper.GetString("InfluxDBServer")
	config.influxDBServerPort = viper.GetString("InfluxDBServerPort")
	config.influxDBUsername = viper.GetString("InfluxDBUsername")
	config.consoleActivated = viper.GetString("ConsoleActivated")
	config.influxDBActivated = viper.GetString("InfluxDBActivated")
	config.refreshTimer = viper.GetString("RefreshTimer")
	config.hTTPActivated = viper.GetString("HTTPActivated")
	config.hTTPPort = viper.GetString("HTTPPort")
	config.logLevel = viper.GetString("LogLevel")
	config.mock = viper.GetBool("mock")
	config.language = viper.GetString("language")
	config.dev = viper.GetBool("dev")

	config.translateFunc, err = i18n.Tfunc(config.language)
	if err != nil {
		mylog.Error.Fatal(fmt.Sprintf("%v", err))
	}

	// Check if one value of the structure is empty
	v := reflect.ValueOf(config)
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i)
		//v.Field(i).SetString(viper.GetString(v.Type().Field(i).Name))
		if values[i] == "" {
			return fmt.Errorf("Check if the key " + v.Type().Field(i).Name + " is present in the file " + dir)
		}
	}
	return nil
}

//go:generate ./command/bindata.sh
//go:generate ./command/bindata-assetfs.sh

func main() {

	if config.dev {
		if err := i18n.LoadTranslationFile("lang/en-us.all.json"); err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}
		if err := i18n.LoadTranslationFile("lang/fr.all.json"); err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}
	} else {
		assetEn, err := assembly.Asset("lang/en-us.all.json")
		if err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}

		assetFr, err := assembly.Asset("lang/fr.all.json")
		if err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}

		if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", assetEn); err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}
		if err := i18n.ParseTranslationFileBytes("lang/fr.all.json", assetFr); err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}
	}

	flag.Parse()

	fmt.Printf("\n%s :> Bloomsky API version %s in Go\n", time.Now().Format(time.RFC850), Version)

	mylog.Init(mylog.ERROR)

	// getConfig from the file config.json
	if err := readConfig(configName); err != nil {
		mylog.Error.Fatal(fmt.Sprintf("%v", err))
	}

	if *debug != "" {
		config.logLevel = *debug
	}

	level, _ := strconv.Atoi(config.logLevel)
	mylog.Init(mylog.Level(level))

	i, _ := strconv.Atoi(config.refreshTimer)
	myTime = time.Duration(i) * time.Second

	//init listeners
	go func() {
		schedule()
	}()
	if config.consoleActivated == "true" {
		initConsole(bloomskyMessageToConsole)
	}
	if config.influxDBActivated == "true" {
		initInfluxDB(bloomskyMessageToInfluxDB, config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
	}
	if config.hTTPActivated == "true" {
		createWebServer(config.hTTPPort)
	}
}

// The scheduler
func schedule() {
	ticker := time.NewTicker(myTime)
	quit := make(chan struct{})
	repeat()
	for {
		select {
		case <-ticker.C:
			repeat()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

//Principal function which one loops each Time Variable
func repeat() {

	mylog.Trace.Printf("Repeat actions each Time Variable : %s secondes", config.refreshTimer)

	// get bloomsky JSON and parse information in bloomsky Go Structure
	var mybloomsky bloomsky.BloomskyStructure
	if config.mock {
		//TODO put in one file
		mylog.Trace.Println("Warning : mock activated !!!")
		body := []byte("[{\"UTC\":2,\"CityName\":\"Thuin\",\"Storm\":{\"UVIndex\":\"1\",\"WindDirection\":\"E\",\"RainDaily\":0,\"WindGust\":0,\"SustainedWindSpeed\":0,\"RainRate\":0,\"24hRain\":0},\"Searchable\":true,\"DeviceName\":\"skyThuin\",\"RegisterTime\":1486905295,\"DST\":1,\"BoundedPoint\":\"\",\"LON\":4.3101,\"Point\":{},\"VideoList\":[\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-27.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-28.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-29.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-30.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-31.mp4\"],\"VideoList_C\":[\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-27_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-28_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-29_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-30_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-31_C.mp4\"],\"DeviceID\":\"442C05954A59\",\"NumOfFollowers\":2,\"LAT\":50.3394,\"ALT\":195,\"Data\":{\"Luminance\":9999,\"Temperature\":70.79,\"ImageURL\":\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5uqmZammJw=.jpg\",\"TS\":1496345207,\"Rain\":false,\"Humidity\":64,\"Pressure\":29.41,\"DeviceType\":\"SKY2\",\"Voltage\":2611,\"Night\":false,\"UVIndex\":9999,\"ImageTS\":1496345207},\"FullAddress\":\"Drève des Alliés, Thuin, Wallonie, BE\",\"StreetName\":\"Drève des Alliés\",\"PreviewImageList\":[\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5qwlZOmn5c=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5qwnZmqmZw=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5unnJakmZg=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5uom5Kkm50=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5upmZiqnps=.jpg\"]}]")
		mybloomsky = bloomsky.NewBloomskyFromBody(body)
	}
	if !config.mock {
		mybloomsky = bloomsky.NewBloomsky(config.bloomskyURL, config.bloomskyAccessToken, true)
	}

	go func() {
		// display major informations to console
		if config.consoleActivated == "true" {
			bloomskyMessageToConsole <- mybloomsky
		}
	}()

	go func() {
		// display major informations to console to influx DB
		if config.influxDBActivated == "true" {
			bloomskyMessageToInfluxDB <- mybloomsky
		}
	}()

	go func() {
		// display major informations to http
		if config.hTTPActivated == "true" {
			bloomskyMessageToHTTP <- mybloomsky
		}
	}()
}
