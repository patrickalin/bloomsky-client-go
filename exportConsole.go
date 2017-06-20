package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/jroimartin/gocui"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/patrickalin/bloomsky-client-go/assembly"
	log "github.com/sirupsen/logrus"
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
	t, err := template.New("bloomsky.txt").Funcs(map[string]interface{}{
		"T": config.translateFunc}).Parse(string(assetBloomsky[:]))
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
}
