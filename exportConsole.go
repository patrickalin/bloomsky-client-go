package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/template"

	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
)

type console struct {
	in           chan bloomsky.BloomskyStructure
	testTemplate *template.Template
}

func initTemplate() *template.Template {
	if config.dev {
		t, err := template.New("bloomsky.txt").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).ParseFiles("tmpl/bloomsky.txt")

		if err != nil {
			log.Fatal(fmt.Errorf("template console : %v", err))
		}
		return t
	}

	assetBloomsky, err := assembly.Asset("tmpl/bloomsky.txt")
	if err != nil {
		log.Fatal(fmt.Errorf("template console : %v", err))
	}

	t, err := template.New("bloomsky.txt").Funcs(map[string]interface{}{
		"T": config.translateFunc,
	}).Parse(string(assetBloomsky[:]))
	if err != nil {
		log.Fatal(fmt.Errorf("template console : %v", err))
	}
	return t
}

//InitConsole listen on the chanel
func initConsole(messages chan bloomsky.BloomskyStructure) (console, error) {
	c := console{in: messages, testTemplate: initTemplate()}
	return c, nil

}

func (c *console) listen(context context.Context) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to console")

		for {
			mylog.Trace.Println("Receive message to export to console")
			msg := <-c.in
			if err := c.testTemplate.Execute(os.Stdout, msg); err != nil {
				fmt.Printf("%v", err)
			}
		}

	}()

}
