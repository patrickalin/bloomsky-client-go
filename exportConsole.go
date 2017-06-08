package main

import (
	"fmt"
	"html/template"
	"os"

	mylog "github.com/patrickalin/GoMyLog"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

// displayToConsole print major informations from a bloomsky JSON to console
func displayToConsole(bloomsky bloomsky.BloomskyStructure) {
	t, err := template.ParseFiles("tmpl/bloomsky.txt")
	if err != nil {
		fmt.Printf("%v", err)
	}
	if err = t.Execute(os.Stdout, bloomsky); err != nil {
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
