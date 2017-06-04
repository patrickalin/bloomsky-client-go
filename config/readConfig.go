//Package config read config bloomsky to structure config
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	mylog "github.com/patrickalin/GoMyLog"
	viper "github.com/spf13/viper"
)

// Configuration is the structure of the config YAML file
//use http://mervine.net/json2struct
type Configuration struct {
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

// Config GetURL return the URL from the config file
type Config interface {
	GetURL() string
}

// ReadConfig read config from config.json
// with the package viper
func ReadConfig(configName string) (conf Configuration, err error) {
	var configInfo Configuration
	viper.SetConfigName(configName)
	viper.AddConfigPath(".")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return configInfo, fmt.Errorf("%v", err)
	}
	dir = dir + "/" + configName
	fmt.Printf("The config file loaded is :> %s/%s \n \n", dir, configName)

	if err := viper.ReadInConfig(); err != nil {
		return configInfo, err
	}

	configInfo.BloomskyURL = viper.GetString("BloomskyURL")
	configInfo.BloomskyAccessToken = viper.GetString("BloomskyAccessToken")
	configInfo.InfluxDBDatabase = viper.GetString("InfluxDBDatabase")
	configInfo.InfluxDBPassword = viper.GetString("InfluxDBPassword")
	configInfo.InfluxDBServer = viper.GetString("InfluxDBServer")
	configInfo.InfluxDBServerPort = viper.GetString("InfluxDBServerPort")
	configInfo.InfluxDBUsername = viper.GetString("InfluxDBUsername")
	configInfo.ConsoleActivated = viper.GetString("ConsoleActivated")
	configInfo.InfluxDBActivated = viper.GetString("InfluxDBActivated")
	configInfo.RefreshTimer = viper.GetString("RefreshTimer")
	configInfo.HTTPActivated = viper.GetString("HTTPActivated")
	configInfo.HTTPPort = viper.GetString("HTTPPort")
	configInfo.LogLevel = viper.GetString("LogLevel")

	// Check if one value of the structure is empty
	v := reflect.ValueOf(configInfo)
	values := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		//v.Field(i).SetString(viper.GetString(v.Type().Field(i).Name))
		if values[i] == "" {
			return configInfo, fmt.Errorf("Check if the key " + v.Type().Field(i).Name + " is present in the file " + dir)
		}
	}

	return configInfo, nil
}

// New create the configStructure and fill in
func New(configName string) Configuration {
	configInfo, err := ReadConfig(configName)
	if err != nil {
		mylog.Error.Fatal(fmt.Sprintf("%v", err))
	}
	return configInfo
}

// GetURL return bloomskyURL
func (configInfo Configuration) GetURL() string {
	return configInfo.BloomskyURL
}
