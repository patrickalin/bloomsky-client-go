package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/utils"
	"github.com/sirupsen/logrus"
)

type console struct {
	in           chan bloomsky.Bloomsky
	testTemplate *template.Template
}

//InitConsole listen on the chanel
func createConsole(messages chan bloomsky.Bloomsky) (console, error) {
	f := map[string]interface{}{"T": config.translateFunc}
	c := console{in: messages, testTemplate: utils.GetTemplate("bloomsky.txt", "tmpl/bloomsky.txt", f, config.dev)}
	logrus.WithFields(logrus.Fields{
		"fct": "exportConsole.initConsole",
	}).Info("Init console")
	return c, nil
}

func (c *console) listen(context context.Context) {
	go func() {

		/*
			TODO

			g, err := gocui.NewGui(gocui.OutputNormal)
			if err != nil {
				log.Panicln(err)
			}
			defer g.Close()

			g.SetManagerFunc(layout)

			if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
				log.Panicln(err)
			}

			if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
				log.Panicln(err)
			}*/

		log.Debug("Init the queue to receive message to export to console")

		for {
			msg := <-c.in

			if err := c.testTemplate.Execute(os.Stdout, msg); err != nil {
				fmt.Printf("%v", err)
			}
		}

	}()

}

/*
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "Hello world!")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}*/
