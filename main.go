// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
	export "github.com/patrickalin/bloomsky-client-go/export"
	"github.com/spf13/viper"

	mylog "github.com/patrickalin/GoMyLog"
)

//configName name of the config file
const configName = "config"

//VERSION of the code
const VERSION = "0.2"

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type configuration struct {
	ConsoleActivated    string
	HTTPActivated       string
	HTTPPort            string
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
}

var config configuration

var (
	bloomskyMessageToConsole  = make(chan bloomskyStructure.BloomskyStructure)
	bloomskyMessageToInfluxDB = make(chan bloomskyStructure.BloomskyStructure)
	bloomskyMessageToHTTP     = make(chan bloomskyStructure.BloomskyStructure)

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

	config.BloomskyURL = viper.GetString("BloomskyURL")
	config.BloomskyAccessToken = viper.GetString("BloomskyAccessToken")
	config.InfluxDBDatabase = viper.GetString("InfluxDBDatabase")
	config.InfluxDBPassword = viper.GetString("InfluxDBPassword")
	config.InfluxDBServer = viper.GetString("InfluxDBServer")
	config.InfluxDBServerPort = viper.GetString("InfluxDBServerPort")
	config.InfluxDBUsername = viper.GetString("InfluxDBUsername")
	config.ConsoleActivated = viper.GetString("ConsoleActivated")
	config.InfluxDBActivated = viper.GetString("InfluxDBActivated")
	config.RefreshTimer = viper.GetString("RefreshTimer")
	config.HTTPActivated = viper.GetString("HTTPActivated")
	config.HTTPPort = viper.GetString("HTTPPort")
	config.LogLevel = viper.GetString("LogLevel")

	// Check if one value of the structure is empty
	v := reflect.ValueOf(config)
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		//v.Field(i).SetString(viper.GetString(v.Type().Field(i).Name))
		if values[i] == "" {
			return fmt.Errorf("Check if the key " + v.Type().Field(i).Name + " is present in the file " + dir)
		}
	}

	return nil
}

// displayToConsole print major informations from a bloomsky JSON to console
func displayToConsole(onebloomsky bloomskyStructure.BloomskyStructure) {
	t, err := template.ParseFiles("tmpl/bloomsky.txt")
	if err != nil {
		fmt.Printf("%v", err)
	}
	if err = t.Execute(os.Stdout, onebloomsky); err != nil {
		fmt.Printf("%v", err)
	}
}

//InitConsole listen on the chanel
func initConsole(messages chan bloomskyStructure.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to console")

		for {
			mylog.Trace.Println("Receive message to export to console")
			msg := <-messages
			displayToConsole(msg)
		}
	}()
}

func main() {

	flag.Parse()

	fmt.Printf("\n %s :> Bloomsky API %s in Go\n", time.Now().Format(time.RFC850), VERSION)

	mylog.Init(mylog.ERROR)

	// getConfig from the file config.json
	if err := readConfig(configName); err != nil {
		mylog.Error.Fatal(fmt.Sprintf("%v", err))
	}

	if *debug != "" {
		config.LogLevel = *debug
	}

	level, _ := strconv.Atoi(config.LogLevel)
	mylog.Init(mylog.Level(level))

	i, _ := strconv.Atoi(config.RefreshTimer)
	myTime = time.Duration(i) * time.Second

	//init listeners
	if config.ConsoleActivated == "true" {
		initConsole(bloomskyMessageToConsole)
	}
	if config.InfluxDBActivated == "true" {
		export.InitInfluxDB(bloomskyMessageToInfluxDB, config.InfluxDBServer, config.InfluxDBServerPort, config.InfluxDBUsername, config.InfluxDBPassword, config.InfluxDBDatabase)
	}
	if config.HTTPActivated == "true" {
		export.InitHTTP(bloomskyMessageToHTTP)
	}
	go func() {
		schedule()
	}()
	export.NewServer(config.HTTPPort)
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
	mylog.Trace.Printf("Repeat actions each Time Variable %s secondes", config.RefreshTimer)

	// get bloomsky JSON and parse information in bloomsky Go Structure
	mybloomsky := bloomskyStructure.NewBloomsky(config.BloomskyURL, config.BloomskyAccessToken)

	go func() {
		// display major informations to console
		if config.ConsoleActivated == "true" {
			bloomskyMessageToConsole <- mybloomsky
		}
	}()

	go func() {
		// display major informations to console to influx DB
		if config.InfluxDBActivated == "true" {
			bloomskyMessageToInfluxDB <- mybloomsky
		}
	}()

	go func() {
		// display major informations to http
		if config.HTTPActivated == "true" {
			bloomskyMessageToHTTP <- mybloomsky
		}
	}()
}
