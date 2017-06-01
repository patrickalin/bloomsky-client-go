package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	bloomskyStructure "github.com/patrickalin/bloomsky-client-go-source/bloomskyStructure"
	config "github.com/patrickalin/bloomsky-client-go-source/config"
	export "github.com/patrickalin/bloomsky-client-go-source/export"

	mylog "github.com/patrickalin/GoMyLog"
)

/*
Get bloomsky Thermostat Information
*/

//configName name of the config file
const configName = "config"

//VERSION version of the code
const VERSION = "0.1"

var (
	bloomskyMessageToConsole  = make(chan bloomskyStructure.BloomskyStructure)
	bloomskyMessageToInfluxDB = make(chan bloomskyStructure.BloomskyStructure)

	myTime time.Duration

	myConfig config.ConfigStructure

	debug = flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
)

func main() {

	flag.Parse()

	fmt.Printf("\n %s :> Bloomsky API %s in Go\n", time.Now().Format(time.RFC850), VERSION)

	mylog.Init(mylog.ERROR)

	// getConfig from the file config.json
	myConfig = config.New(configName)

	if *debug != "" {
		myConfig.LogLevel = *debug
	}

	level, _ := strconv.Atoi(myConfig.LogLevel)
	mylog.Init(mylog.Level(level))

	i, _ := strconv.Atoi(myConfig.RefreshTimer)
	myTime = time.Duration(i) * time.Second

	//init listeners
	if myConfig.ConsoleActivated == "true" {
		export.InitConsole(bloomskyMessageToConsole)
	}
	if myConfig.InfluxDBActivated == "true" {
		export.InitInfluxDB(bloomskyMessageToInfluxDB, myConfig)
	}

	schedule()
}

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

//principal function which loops
func repeat() {
	mylog.Trace.Println("Repeat actions each Time Variable")

	// get bloomsky JSON and parse information in bloomsky Go Structure
	mybloomsky := bloomskyStructure.MakeNew(myConfig)

	go func() {
		// display major informations to console

		if myConfig.ConsoleActivated == "true" {
			bloomskyMessageToConsole <- mybloomsky
		}
	}()

	go func() {
		// display major informations to console to influx DB
		if myConfig.InfluxDBActivated == "true" {
			bloomskyMessageToInfluxDB <- mybloomsky
		}
	}()

}
