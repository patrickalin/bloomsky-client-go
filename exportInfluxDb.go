package main

import (
	"fmt"
	"time"

	clientinfluxdb "github.com/influxdata/influxdb/client/v2"
	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-api-go"
)

func sendbloomskyToInfluxDB(onebloomsky bloomskyStructure.BloomskyStructure, clientInflux clientinfluxdb.Client, InfluxDBDatabase string) {

	fmt.Printf("\n%s :> Send bloomsky Data to InfluxDB\n", time.Now().Format(time.RFC850))

	// Create a point and add to batch
	tags := map[string]string{"bloomsky": onebloomsky.GetCity()}
	fields := map[string]interface{}{
		"NumOfFollowers":        onebloomsky.GetNumOfFollowers(),
		"Humidity":              onebloomsky.GetHumidity(),
		"Uv":                    onebloomsky.GetIndexUV(),
		"PressureHpa":           onebloomsky.GetPressureHPa(),
		"PressureInHg":          onebloomsky.GetPressureInHg(),
		"Night":                 onebloomsky.IsNight(),
		"Rain":                  onebloomsky.IsRain(),
		"RainDailyIn":           onebloomsky.GetRainDailyIn(),
		"RainDailyMm":           onebloomsky.GetRainDailyMm(),
		"RainIn":                onebloomsky.GetRainIn(),
		"RainMm":                onebloomsky.GetRainMm(),
		"RainRateIn":            onebloomsky.GetRainRateIn(),
		"RainRateMm":            onebloomsky.GetRainRateMm(),
		"ustainedWindSpeedkmh":  onebloomsky.GetSustainedWindSpeedkmh(),
		"SustainedWindSpeedMph": onebloomsky.GetSustainedWindSpeedMph(),
		"SustainedWindSpeedMs":  onebloomsky.GetSustainedWindSpeedMs(),
		"WindDirection":         onebloomsky.GetWindDirection(),
		"WindGustkmh":           onebloomsky.GetWindGustkmh(),
		"WindGustMph":           onebloomsky.GetWindGustMph(),
		"WindGustMs":            onebloomsky.GetWindGustMs(),
		"TemperatureCelsius":    onebloomsky.GetTemperatureCelsius(),
		"TemperatureFahrenheit": onebloomsky.GetTemperatureFahrenheit(),
		"TimeStamp":             onebloomsky.GetTimeStamp(),
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
