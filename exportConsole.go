package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	log "github.com/sirupsen/logrus"
)

var funcMap = map[string]interface{}{
	"T": i18n.IdentityTfunc,
}

var testTemplate *template.Template

// displayToConsole print major informations from a bloomsky JSON to console
func displayToConsole(bloomsky bloomsky.BloomskyStructure) {

	var err error
	if config.dev {
		testTemplate, err = template.New("bloomsky.txt").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).ParseFiles("tmpl/bloomsky.txt")

		if err != nil {
			log.Fatal(fmt.Errorf("template console : %v", err))
		}
	} else {
		assetBloomsky, err := assembly.Asset("tmpl/bloomsky.txt")
		if err != nil {
			log.Fatalf("template console : %v", err)
		}

		testTemplate, err = template.New("bloomsky.txt").Funcs(map[string]interface{}{
			"T": config.translateFunc,
		}).Parse(string(assetBloomsky[:]))
		if err != nil {
			log.Fatalf("template console : %v", err)
		}
	}

	if testTemplate.Execute(os.Stdout, bloomsky) != nil {
		fmt.Printf("%v", err)
	}
}

//InitConsole listen on the chanel
func initConsole(messages chan bloomsky.BloomskyStructure) {
	go func() {

		log.Info("Init the queue to receive message to export to console")

		for {
			log.Info("Receive message to export to console")
			msg := <-messages
			displayToConsole(msg)
		}
	}()
}
