package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/nicksnyder/go-i18n/i18n"
	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

var funcMap = map[string]interface{}{
	"T": i18n.IdentityTfunc,
}

var testTemplate *template.Template

// displayToConsole print major informations from a bloomsky JSON to console
func displayToConsole(bloomsky bloomsky.BloomskyStructure) {

	var err error
	//testTemplate, err = template.New("hello.gohtml").Funcs(funcMap).ParseFiles("hello.gohtml")

	testTemplate, err = template.New("bloomsky.txt").Funcs(map[string]interface{}{
		"T": config.translateFunc,
	}).ParseFiles("tmpl/bloomsky.txt")

	if err != nil {
		panic(err)
	}

	//fmt.Println(T("program_greeting"))

	if testTemplate.Execute(os.Stdout, bloomsky) != nil {
		fmt.Printf("%v", err)
	}
}

//InitConsole listen on the chanel
func initConsole(messages chan bloomsky.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to console")

		for {
			mylog.Trace.Println("Receive message to export to console")
			msg := <-messages
			displayToConsole(msg)
		}
	}()
}
