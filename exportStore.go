package main

import (
	"context"

	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/sirupsen/logrus"
)

type store struct {
	in     chan bloomsky.Bloomsky
	stores map[string]*Ring
}

type measure struct {
	value float64
}

func (m measure) Value() float64 {
	return m.value
}

//InitConsole listen on the chanel
func createStore(messages chan bloomsky.Bloomsky) (store, error) {
	stores := make(map[string]*Ring)
	stores["temp"] = &Ring{}
	stores["wind"] = &Ring{}
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
			c.stores["temp"].Enqueue(measure{msg.GetTemperatureCelsius()})
			c.stores["wind"].Enqueue(measure{msg.GetWindGustkmh()})
		}
	}()

}

func (c *store) GetValues(name string) []Measure {
	return c.stores[name].Values()

}
