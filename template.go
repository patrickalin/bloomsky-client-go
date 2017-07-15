package main

import (
	text "text/template"

	"html/template"

	"github.com/patrickalin/bloomsky-client-go/assembly"
)

//GetTemplate retrieve a template
func GetTemplate(templateName string, templateLocation string, funcs map[string]interface{}, dev bool) *text.Template {
	if dev {
		t, err := text.New(templateName).Funcs(funcs).ParseFiles(templateLocation)
		checkErr(err, funcName(), "Load template console", templateLocation)
		return t
	}

	assetBloomsky, err := assembly.Asset(templateLocation)
	t, err := text.New(templateName).Funcs(funcs).Parse(string(assetBloomsky[:]))
	checkErr(err, funcName(), "Load template console", templateLocation)
	return t
}

// GetHTMLTemplate "bloomsky_header.html","tmpl/bloomsky_header.html",map[string]interface{}{"T": config.translateFunc,}
func GetHTMLTemplate(templateName string, templatesLocation []string, funcs map[string]interface{}, dev bool) *template.Template {
	t := template.New(templateName)
	t.Funcs(funcs)
	if dev {
		t, err := t.ParseFiles(templatesLocation...)
		checkErr(err, funcName(), "ParseFiles")
		return t
	}

	for _, l := range templatesLocation {
		asset, err := assembly.Asset(l)
		checkErr(err, funcName(), "Assembly template", l)

		t, err = t.Parse(string(asset[:]))
		checkErr(err, funcName(), "Parse file")
	}

	return t
}
