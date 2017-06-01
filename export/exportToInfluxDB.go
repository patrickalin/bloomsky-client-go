package export

import (
	"fmt"
	"time"

	bloomskyStructure "github.com/patrickalin/bloomsky-client-go-source/bloomskyStructure"
	config "github.com/patrickalin/bloomsky-client-go-source/config"

	clientinfluxdb "github.com/influxdata/influxdb/client/v2"

	mylog "github.com/patrickalin/GoMyLog"
)

type influxDBStruct struct {
	Columns []string        `json:"columns"`
	Serie   string          `json:"name"`
	Points  [][]interface{} `json:"points"`
}

type influxDBError struct {
	message error
	advice  string
}

var clientInflux clientinfluxdb.Client

func (e *influxDBError) Error() string {
	return fmt.Sprintf("\n \t InfluxDBError :> %s \n\t InfluxDB Advice:> %s", e.message, e.advice)
}

func sendbloomskyToInfluxDB(onebloomsky bloomskyStructure.BloomskyStructure, oneConfig config.ConfigStructure) {

	fmt.Printf("\n%s :> Send bloomsky Data to InfluxDB\n", time.Now().Format(time.RFC850))

	// Create a point and add to batch
	tags := map[string]string{"bloomsky": "living"}
	fields := map[string]interface{}{
		"NumOfFollowers": onebloomsky.GetNumOfFollowers(),
	}

	// Create a new point batch
	bp, err := clientinfluxdb.NewBatchPoints(clientinfluxdb.BatchPointsConfig{
		Database:  oneConfig.InfluxDBDatabase,
		Precision: "s",
	})

	if err != nil {
		mylog.Error.Fatal(&influxDBError{err, "Error sent Data to Influx DB"})
	}

	pt, err := clientinfluxdb.NewPoint("bloomskyData", tags, fields, time.Now())
	bp.AddPoint(pt)

	// Write the batch
	err = clientInflux.Write(bp)

	if err != nil {
		err2 := createDB(oneConfig)
		if err2 != nil {
			mylog.Error.Fatal(&influxDBError{err, "Error with Post : Check if InfluxData is running or if the database bloomsky exists"})
		}
	}

}

func createDB(oneConfig config.ConfigStructure) error {
	fmt.Println("Create Database bloomsky in InfluxData")

	query := fmt.Sprint("CREATE DATABASE ", oneConfig.InfluxDBDatabase)
	q := clientinfluxdb.NewQuery(query, "", "")

	fmt.Println("Query: ", query)

	_, err := clientInflux.Query(q)
	if err != nil {
		return &influxDBError{err, "Error with : Create database bloomsky, check if InfluxDB is running"}
	}
	fmt.Println("Database bloomsky created in InfluxDB")
	return nil
}

func makeClient(oneConfig config.ConfigStructure) (client clientinfluxdb.Client, err error) {
	client, err = clientinfluxdb.NewHTTPClient(
		clientinfluxdb.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", oneConfig.InfluxDBServer, oneConfig.InfluxDBServerPort),
			Username: oneConfig.InfluxDBUsername,
			Password: oneConfig.InfluxDBPassword,
		})

	if err != nil || client == nil {
		return nil, &influxDBError{err, "Error with creating InfluxDB Client : , check if InfluxDB is running"}
	}
	return client, nil
}

// InitInfluxDB initiate the client influxDB
// Arguments bloomsky informations, configuration from config file
// Wait events to send to influxDB
func InitInfluxDB(messagesbloomsky chan bloomskyStructure.BloomskyStructure, oneConfig config.ConfigStructure) {

	clientInflux, _ = makeClient(oneConfig)

	go func() {
		mylog.Trace.Println("Receive messagesbloomsky to export InfluxDB")
		for {
			msg := <-messagesbloomsky
			sendbloomskyToInfluxDB(msg, oneConfig)
		}
	}()

}
