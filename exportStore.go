package main

import (
	"context"
	"strconv"
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
	stores["temperatureCelsius"] = &ring.Ring{}
	stores["pressureHPa"] = &ring.Ring{}
	stores["windGustkmh"] = &ring.Ring{}
	stores["windSustainedSpeedkmh"] = &ring.Ring{}
	stores["humidity"] = &ring.Ring{}
	stores["rainDailyMm"] = &ring.Ring{}
	stores["rainMm"] = &ring.Ring{}
	stores["rainRate"] = &ring.Ring{}
	stores["indexUV"] = &ring.Ring{}

	return store{in: messages, stores: stores}, nil

}

func (c *store) listen(context context.Context) {

	go func() {

		log.WithFields(logrus.Fields{
			"fct": "exportStore.listen",
		}).Info("init")

		for {
			select {
			case msg := <-c.in:
				log.WithFields(logrus.Fields{
					"fct": "exportStore.listen",
				}).Debug("Receive message")

				c.stores["temperatureCelsius"].Enqueue(measure{time.Now(), msg.GetTemperatureCelsius()})
				c.stores["pressureHPa"].Enqueue(measure{time.Now(), msg.GetPressureHPa()})
				c.stores["windGustkmh"].Enqueue(measure{time.Now(), msg.GetWindGustkmh()})
				c.stores["windSustainedSpeedkmh"].Enqueue(measure{time.Now(), msg.GetSustainedWindSpeedkmh()})
				c.stores["humidity"].Enqueue(measure{time.Now(), msg.GetHumidity()})
				c.stores["rainDailyMm"].Enqueue(measure{time.Now(), msg.GetRainDailyMm()})
				c.stores["rainMm"].Enqueue(measure{time.Now(), msg.GetRainMm()})
				c.stores["rainRate"].Enqueue(measure{time.Now(), msg.GetRainRateMm()})

				i, err := strconv.ParseFloat(msg.GetIndexUV(), 64)
				checkErr(err, funcName(), "impossible to convert string to int")
				c.stores["indexUV"].Enqueue(measure{time.Now(), i})

			case <-context.Done():
				return
			}
		}
	}()

}

func (c *store) GetValues(name string) []ring.TimeMeasure {
	return c.stores[name].Values()
}

func (c *store) String(name string) string {
	s, _ := c.stores[name].DumpLine()
	return s
}
