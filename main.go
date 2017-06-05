// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	clientinfluxdb "github.com/influxdata/influxdb/client/v2"
	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
	"github.com/spf13/viper"
)

//configName name of the config file
const configName = "config"

//VERSION of the code
const VERSION = "0.2"

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

// displayToConsole print major informations from a bloomsky JSON to console
func displayToConsole(bloomsky bloomskyStructure.BloomskyStructure) {
	t, err := template.ParseFiles("tmpl/bloomsky.txt")
	if err != nil {
		fmt.Printf("%v", err)
	}
	if err = t.Execute(os.Stdout, bloomsky); err != nil {
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

var myBloomskyHTTP bloomskyStructure.BloomskyStructure

//displayToHTTP TODO normally push information with websocket to page
func displayToHTTP(bloomsky bloomskyStructure.BloomskyStructure) {
	fmt.Println("normally push websocket")
	myBloomskyHTTP = bloomsky
}

func handler(w http.ResponseWriter, r *http.Request) {
	mylog.Trace.Println("Handle")

	//t := template.New("bloomsky") // Create a template.
	//t = t.Funcs(template.FuncMap{"GetTimeStamp": mybloomsky.GetTimeStamp()})

	t, err := template.ParseFiles("tmpl/bloomsky.html") // Parse template file.
	if err != nil {
		fmt.Printf("%v", err)
	}
	err = t.Execute(w, myBloomskyHTTP) // merge.
	if err != nil {
		fmt.Printf("%v", err)
	}

}

//newWebServer create web server
func newWebServer(HTTPPort string) {
	mylog.Trace.Printf("Init server http port %s", HTTPPort)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(HTTPPort, nil)
	if err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error when I create the server : %v", err))
	}
	mylog.Trace.Printf("Server ok on http %s", HTTPPort)
}

//initWebServer listen on the chanel
func initWebServer(messages chan bloomskyStructure.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to http")

		for {
			msg := <-messages
			mylog.Trace.Println("Receive message to export to http")
			displayToHTTP(msg)
		}
	}()
}

func sendbloomskyToInfluxDB(onebloomsky bloomskyStructure.BloomskyStructure, clientInflux clientinfluxdb.Client, InfluxDBDatabase string) {

	fmt.Printf("\n%s :> Send bloomsky Data to InfluxDB\n", time.Now().Format(time.RFC850))

	// Create a point and add to batch
	tags := map[string]string{"bloomsky": "living"}
	fields := map[string]interface{}{
		"NumOfFollowers": onebloomsky.GetNumOfFollowers(),
	}

	// Create a new point batch
	bp, err := clientinfluxdb.NewBatchPoints(clientinfluxdb.BatchPointsConfig{
		Database:  InfluxDBDatabase,
		Precision: "s",
	})

	if err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error sent Data to Influx DB : %v", err))
	}

	pt, err := clientinfluxdb.NewPoint("bloomskyData", tags, fields, time.Now())
	bp.AddPoint(pt)

	// Write the batch
	err = clientInflux.Write(bp)

	if err != nil {
		err2 := createDB(clientInflux, InfluxDBDatabase)
		if err2 != nil {
			mylog.Error.Fatal(fmt.Errorf("Check if InfluxData is running or if the database bloomsky exists : %v", err))
		}
	}
}

func createDB(clientInflux clientinfluxdb.Client, InfluxDBDatabase string) error {
	fmt.Println("Create Database bloomsky in InfluxData")

	query := fmt.Sprint("CREATE DATABASE ", InfluxDBDatabase)
	q := clientinfluxdb.NewQuery(query, "", "")

	fmt.Println("Query: ", query)

	_, err := clientInflux.Query(q)
	if err != nil {
		return fmt.Errorf("Error with : Create database bloomsky, check if InfluxDB is running : %v", err)
	}
	fmt.Println("Database bloomsky created in InfluxDB")
	return nil
}

func makeClientInfluxDB(InfluxDBServer, InfluxDBServerPort, InfluxDBUsername, InfluxDBPassword string) (client clientinfluxdb.Client, err error) {
	client, err = clientinfluxdb.NewHTTPClient(
		clientinfluxdb.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", InfluxDBServer, InfluxDBServerPort),
			Username: InfluxDBUsername,
			Password: InfluxDBPassword,
		})

	if err != nil || client == nil {
		return nil, fmt.Errorf("Error creating database bloomsky, check if InfluxDB is running : %v", err)
	}
	return client, nil
}

// InitInfluxDB initiate the client influxDB
// Arguments bloomsky informations, configuration from config file
// Wait events to send to influxDB
func initInfluxDB(messagesbloomsky chan bloomskyStructure.BloomskyStructure, influxDBServer, influxDBServerPort, influxDBUsername, influxDBPassword, influxDBDatabase string) {

	clientInflux, _ := makeClientInfluxDB(influxDBServer, influxDBServerPort, influxDBUsername, influxDBPassword)

	go func() {
		mylog.Trace.Println("Receive messagesbloomsky to export InfluxDB")
		for {
			msg := <-messagesbloomsky
			sendbloomskyToInfluxDB(msg, clientInflux, influxDBDatabase)
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
		config.logLevel = *debug
	}

	level, _ := strconv.Atoi(config.logLevel)
	mylog.Init(mylog.Level(level))

	i, _ := strconv.Atoi(config.refreshTimer)
	myTime = time.Duration(i) * time.Second

	//init listeners
	if config.consoleActivated == "true" {
		initConsole(bloomskyMessageToConsole)
	}
	if config.influxDBActivated == "true" {
		initInfluxDB(bloomskyMessageToInfluxDB, config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
	}
	if config.hTTPActivated == "true" {
		fmt.Printf("cici %s", config.hTTPActivated)
		initWebServer(bloomskyMessageToHTTP)
	}
	go func() {
		schedule()
	}()
	if config.hTTPActivated == "true" {
		newWebServer(config.hTTPPort)
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
	mylog.Trace.Printf("Repeat actions each Time Variable %s secondes", config.refreshTimer)

	// get bloomsky JSON and parse information in bloomsky Go Structure
	mybloomsky := bloomskyStructure.NewBloomsky(config.bloomskyURL, config.bloomskyAccessToken)

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
