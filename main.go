// Bloomsky application to export Data bloomsky to console or to influxdb.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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

//Version of the code
var Version = "No Version Provided"

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
	refreshTimer        string
	mock                bool
	language            string
	translateFunc       i18n.TranslateFunc
	dev                 bool
}

var config configuration

var (
	channels = make(map[string]chan bloomsky.BloomskyStructure)

	myTime time.Duration
	debug  = flag.String("debug", "", "Error=1, Warning=2, Info=3, Trace=4")
	c      *httpServer
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
	config.consoleActivated = viper.GetBool("ConsoleActivated")
	config.influxDBActivated = viper.GetBool("InfluxDBActivated")
	config.refreshTimer = viper.GetString("RefreshTimer")
	config.hTTPActivated = viper.GetBool("HTTPActivated")
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
	if token := os.Getenv("bloomskyAccessToken"); token != "" {
		config.bloomskyAccessToken = token
	}
	return nil
}

//go:generate ./command/bindata.sh
//go:generate ./command/bindata-assetfs.sh

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh)
	go func() {
		select {
		case i := <-signalCh:
			fmt.Printf("receive interrupt  %v", i)
			cancel()
			return
		}
	}()

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readTranslationResource("lang/en-us.all.json")); err != nil {
		log.Fatal(fmt.Errorf("error read language file : %v", err))
	}
	if err := i18n.ParseTranslationFileBytes("lang/fr.all.json", readTranslationResource("lang/fr.all.json")); err != nil {
		log.Fatal(fmt.Errorf("error read language file : %v", err))
	}

	flag.Parse()

	fmt.Printf("\n%s :> Bloomsky API %s in Go\n", time.Now().Format(time.RFC850), Version)

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
	ctxsch, cancelsch := context.WithCancel(ctx)
	//init listeners

	if config.consoleActivated {
		channels["console"] = make(chan bloomsky.BloomskyStructure)
		c, err := initConsole(channels["console"])
		if err != nil {
			mylog.Error.Fatal(fmt.Sprintf("%v", err))
		}
		c.listen(context.Background())
	}
	if config.influxDBActivated {
		channels["influxdb"] = make(chan bloomsky.BloomskyStructure)
		c, err := initClient(channels["influxdb"], config.influxDBServer, config.influxDBServerPort, config.influxDBUsername, config.influxDBPassword, config.influxDBDatabase)
		if err != nil {
			mylog.Error.Fatal(fmt.Sprintf("%v", err))
		}
		c.listen(context.Background())

	}
	if config.hTTPActivated {
		var err error
		channels["web"] = make(chan bloomsky.BloomskyStructure)
		c, err = createWebServer(channels["web"], config.hTTPPort)
		if err != nil {
			mylog.Error.Fatal(fmt.Sprintf("%v", err))
		}
		c.listen(context.Background())

	}

	schedule(ctxsch)

	<-ctx.Done()
	cancelsch()
	if c.h != nil {
		fmt.Println("shutting down ws")
		c.h.Shutdown(ctx)
	}

	fmt.Println("terminated")
}

// The scheduler
func schedule(ctx context.Context) {
	ticker := time.NewTicker(myTime)

	collect(ctx)
	for {
		select {
		case <-ticker.C:
			collect(ctx)
		case <-ctx.Done():
			fmt.Println("stoping ticker")
			ticker.Stop()
			for _, v := range channels {
				close(v)
			}
			return
		}
	}
}

//Principal function which one loops each Time Variable
func collect(ctx context.Context) {

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

	for _, v := range channels {
		v <- mybloomsky
	}

}

func readTranslationResource(name string) []byte {

	if config.dev {
		b, err := ioutil.ReadFile(name)
		if err != nil {
			log.Fatal(fmt.Errorf("error read language file : %v", err))
		}
		return b
	}

	b, err := assembly.Asset(name)
	if err != nil {
		log.Fatal(fmt.Errorf("error read language file : %v", err))
	}

	return b
}
