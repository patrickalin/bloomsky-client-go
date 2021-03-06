package main

import (
	"context"
	"fmt"
	"time"

	clientinfluxdb "github.com/influxdata/influxdb1-client/v2"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

type client struct {
	in       chan bloomsky.Bloomsky
	c        clientinfluxdb.Client
	database string
}

func (c *client) sendbloomskyToInfluxDB(onebloomsky bloomsky.Bloomsky) {

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
		"WindDirectionDeg":      onebloomsky.GetWindDirectionDeg(),
		"WindGustkmh":           onebloomsky.GetWindGustkmh(),
		"WindGustMph":           onebloomsky.GetWindGustMph(),
		"WindGustMs":            onebloomsky.GetWindGustMs(),
		"TemperatureCelsius":    onebloomsky.GetTemperatureCelsius(),
		"TemperatureFahrenheit": onebloomsky.GetTemperatureFahrenheit(),
		"TimeStamp":             onebloomsky.GetTimeStamp(),
	}

	// Create a new point batch
	bp, err := clientinfluxdb.NewBatchPoints(clientinfluxdb.BatchPointsConfig{
		Database:  c.database,
		Precision: "s",
	})
	checkErr(err, funcName(), "Error sent Data to Influx DB")

	pt, err := clientinfluxdb.NewPoint("bloomskyData", tags, fields, time.Now())
	checkErr(err, funcName(), "New point")
	bp.AddPoint(pt)

	// Write the batch
	err = c.c.Write(bp)

	if err != nil {
		err := c.createDB(c.database)
		checkErr(err, funcName(), "Check if InfluxData is running or if the database bloomsky exists")
	}
}

func (c *client) createDB(InfluxDBDatabase string) error {
	fmt.Println("Create Database bloomsky in InfluxData")

	query := fmt.Sprint("CREATE DATABASE ", InfluxDBDatabase)
	q := clientinfluxdb.NewQuery(query, "", "")

	fmt.Println("Query: ", query)

	_, err := c.c.Query(q)
	checkErr(err, funcName(), "Error with : Create database bloomsky, check if InfluxDB is running")
	fmt.Println("Database bloomsky created in InfluxDB")
	return nil
}

func initClient(messagesbloomsky chan bloomsky.Bloomsky, InfluxDBServer, InfluxDBServerPort, InfluxDBUsername, InfluxDBPassword, InfluxDatabase string) (*client, error) {
	c, err := clientinfluxdb.NewHTTPClient(
		clientinfluxdb.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", InfluxDBServer, InfluxDBServerPort),
			Username: InfluxDBUsername,
			Password: InfluxDBPassword,
		})

	if err != nil || c == nil {
		return nil, fmt.Errorf("Error creating database bloomsky, check if InfluxDB is running : %v", err)
	}
	cl := &client{c: c, in: messagesbloomsky, database: InfluxDatabase}
	//need to check how to verify that the db is running
	err = cl.createDB(InfluxDatabase)
	checkErr(err, funcName(), "impossible to create DB", InfluxDatabase)
	return cl, nil
}

// InitInfluxDB initiate the client influxDB
// Arguments bloomsky informations, configuration from config file
// Wait events to send to influxDB
func (c *client) listen(context context.Context) {

	go func() {
		log.Info("Receive messagesbloomsky to export InfluxDB")
		for {
			msg := <-c.in
			c.sendbloomskyToInfluxDB(msg)
		}
	}()
}
