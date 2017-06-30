package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/fatih/color"
	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

type console struct {
	in           chan bloomsky.Bloomsky
	textTemplate *template.Template
}

//InitConsole listen on the chanel
func createConsole(messages chan bloomsky.Bloomsky, translateFunc i18n.TranslateFunc, dev bool) (console, error) {
	f := map[string]interface{}{"T": translateFunc}
	//Get template
	c := console{in: messages, textTemplate: GetTemplate("bloomsky.txt", "tmpl/bloomsky.txt", f, dev)}
	logInfo(funcName(), "Console listen", "")
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

		logDebug(funcName(), "Init the queue console to display message", "")

		for {
			select {
			case msg := <-c.in:

				color.Set(color.FgBlack)
				color.Set(color.BgWhite)

				if err := c.textTemplate.Execute(os.Stdout, msg); err != nil {
					fmt.Printf("%v", err)
				}

				color.Unset()

			case <-context.Done():
				fmt.Println("console done")
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
