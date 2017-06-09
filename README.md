# BloomSky Client in Go

[![Build Status](https://travis-ci.org/patrickalin/bloomsky-client-go.svg?branch=master)](https://travis-ci.org/patrickalin/bloomsky-client-go)

![Build size](https://reposs.herokuapp.com/?path=patrickalin/bloomsky-client-go)

[![Go Report Card](https://goreportcard.com/badge/github.com/patrickalin/bloomsky-client-go)](https://goreportcard.com/report/github.com/patrickalin/bloomsky-client-go)

[![Coverage Status](https://coveralls.io/repos/github/patrickalin/bloomsky-client-go/badge.svg)](https://coveralls.io/github/patrickalin/bloomsky-client-go)

A simple Go client for the BloomSky API.
I's possible to export informations in the console or in in Time Series Database InfluxData.

## Prerequisites

* BloomSky API key (get it here: https://dashboard.bloomsky.com/)

## Getting Started

### Installation

Download the [binary](https://github.com/patrickalin/bloomsky-client-go/releases) for your OS and the [config.yaml](https://github.com/patrickalin/bloomsky-client-go/blob/master/config.yaml) in the same folder.

### Usage

You have to change the API Key in the config.yaml.

### Example : result in the standard console.

    Bloomsky
    --------
    Timestamp :         2017-06-09 22:07:10 &#43;0200 CEST
    City :              Thuin
    Device Id :         442C05954A59
    Num Of Followers :  2
    Index UV :          1
    Night :             true
    Wind Direction :    SW
    Wind Gust :         4.16
    Sustained Wind Speed : 2.17
    Wind Gust :         6.6976
    Sustained Wind Speed 3.4937
    Rain :              false
    Daily :             0.44
    24h Rain :          0.44
    Rain Rate :         0
    Rain Daily :        0.44
    24h Rain :          11.18
    Rain Rate :         0
    Temperature F :     59.13 °F
    Temperature C :     15.07 °C
    Humidity :          65
    Pressure InHg :     29.38
    Pressure HPa :      994.92

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
