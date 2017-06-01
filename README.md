# BloomSky Client in Go

[![Build Status](https://travis-ci.org/patrickalin/bloomsky-client-go-source.svg?branch=master)](https://travis-ci.org/patrickalin/bloomsky-client-go-source)

![Build size](https://reposs.herokuapp.com/?path=patrickalin/bloomsky-client-go-source)

[![Go Report Card](https://goreportcard.com/badge/github.com/patrickalin/bloomsky-client-go-source)](https://goreportcard.com/report/github.com/patrickalin/bloomsky-client-go-source)

A simple Go client for the BloomSky API.
I's possible to export informations in the console or in in Time Series Database InfluxData.

## Prerequisites

* BloomSky API key (get it here: https://dashboard.bloomsky.com/)

## Getting Started

### Installation

Download the [binary](https://github.com/patrickalin/bloomsky-client-go/releases) for your OS and the [config.yaml](https://github.com/patrickalin/bloomsky-client-go/blob/master/config.yaml) in the same folder.

### Usage

You have to change the API Key in the config.yaml.

###Example : result in the standard console.

Example : result in the standard console.

    Tuesday, 30-May-17 16:14:43 CEST :> Bloomsky API 0.1 in Go
   
    Tuesday, 30-May-17 16:14:47 CEST :> Send bloomsky Data to InfluxDB
    
    Timestamp : 	 	2017-05-30 16:08:49 +0200 CEST
    City : 	 	Thuin
    Device Id : 	 	442C05954A59
    Num Of Followers : 	 	2
    Index UV : 	 	1
    Night : 	 	false
    Wind Direction : 	 	SW
    Wind Gust : 	 	7.40 mPh
    Sustained Wind Speed : 	 	4.47 mPh
    Wind Gust : 	 	11.91 km/h
    Sustained Wind Speed : 	 	7.20 m/s
    Rain : 	 	false
    Rain Daily : 	 	0.00 in
    24h Rain : 	 	0.00 in
    Rain Rate : 	 	0.00 in
    Rain Daily : 	 	0.00 mm
    24h Rain : 	 	0.00 mm
    Rain Rate : 	 	0.00 mm
    Temperature F : 	 	67.9 °F
    Temperature C : 	 	20.0 °C
    Humidity : 	 	62 %
    Pressure InHg : 	 	29.4 inHg
    Pressure HPa : 	 	993.9 hPa

### Example : result in a influxData.

![InfluxData Image ](https://github.com/patrickalin/bloomsky-client-go-source/blob/master/img/InfluxDB.png)

You can display the result with Chronograph

![Chronograph Image ](https://github.com/patrickalin/bloomsky-client-go-source/blob/master/img/Chronograph.png)

You can display the result with Grafana

![Grafana Image ](https://github.com/patrickalin/GobloomskyThermostatAPIRest/blob/master/img/Grafana.png)

If you want I have a similar code for openweather to save the temperature of you location.
[GoOpenWeatherToInfluxDB](https://github.com/patrickalin/GoOpenWeatherToInfluxDB)

## Compilation

### Pre installation

install git

install go from http://golang.org/

If you want use influxData, version > 0.13

#Installation

    git clone https://github.com/patrickalin/GobloomskyThermostatAPIRest.git
    cd GobloomskyThermostatAPIRest
    export GOPATH=$PWD
    go get -v .
    go build

#Extra installation influxDB

[InfluxData download](https://influxdata.com/downloads/#influxdb)

Execution

    influxd@

#Configuration

1 You must copy the config.json.example to config.json

    cp config.json.example config.json

2 In the config file modify the secret key receive on https://developer.bloomsky.com/

Don't forget to receive a secretCode, it's a POST to developer-api.bloomsky.com not a GET. I lose a lot of time with this error.

To test, execute one time :

    curl -L -X GET -H "Accept: application/json" "https://developer-api.bloomsky.com/?auth=c.557ToBeCompleted"
    with you key

4 Modify all paramameters in config.json

- For InfluxData, isntall the software https://influxdata.com/, the Go program create himself the database "bloomsky"

#Execution

    ./GobloomskyThermostatAPIRest

#Debug

In the config file, you can change the log level.

#Thanks

https://github.com/tixu for testing and review

http://mervine.net/json2struct "transform JSON to Go struct library"

http://github.com/spf13/viper "read config library"

### License

The code is licensed under the permissive Apache v2.0 licence. This means you can do what you like with the software, as long as you include the required notices. [Read this](https://tldrlegal.com/license/apache-license-2.0-(apache-2.0)) for a summary.
