package main

import (
	"context"
	"time"

	"github.com/patrickalin/bloomsky-client-go/pkg/ring"

	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/sirupsen/logrus"
)

type store struct {
	in     chan bloomsky.Bloomsky
	stores map[string]*ring.Ring
}

type measure struct {
	Timestamp time.Time
	value     float64
}

func (m measure) TimeStamp() time.Time {
	return m.Timestamp
}

/**
* Measure  represents a measure that has a GetValue
 */

func (m measure) Value() float64 {
	return m.value
}

//InitConsole listen on the chanel
func createStore(messages chan bloomsky.Bloomsky) (store, error) {
	stores := make(map[string]*ring.Ring)
	stores["temp"] = &ring.Ring{}
	stores["wind"] = &ring.Ring{}
	return store{in: messages, stores: stores}, nil

}

func (c *store) listen(context context.Context) {

	go func() {

		log.WithFields(logrus.Fields{
			"fct": "exportStore.listen",
		}).Info("init")

		for {
			msg := <-c.in
			log.WithFields(logrus.Fields{
				"fct": "exportStore.listen",
			}).Debug("Receive message")
			c.stores["temp"].Enqueue(measure{time.Now(), msg.GetTemperatureCelsius()})
			c.stores["wind"].Enqueue(measure{time.Now(), msg.GetWindGustkmh()})
		}
	}()

}

func (c *store) GetValues(name string) []ring.TimeMeasure {
	return c.stores[name].Values()
}

func (c *store) String(name string) string {
	/*[
	    [new Date(1416013200000), 22],
	    [new Date(2014, 10, 15, 0, 30), 23],
	    [new Date(2014, 10, 15, 0, 00), 22],
	    [new Date(2014, 10, 14, 23, 30), 21],
	    [new Date(2014, 10, 14, 23, 00), 22],
	    [new Date(2014, 10, 14, 22, 30), 18],
	]*/

	/*

		var ret = "["
		for k, v :	= range c.stores[name].Values() {
			ret = ret + "[" + strconv.FormatFloat(v.Value(), 'f', 6, 64) + "," + strconv.Itoa(k) + "],"
		}
		ret += "]"
		return ret*/
	s, _ := c.stores[name].DumpLine()
	return s
}
