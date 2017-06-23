package utils

import (
	"log"
	text "text/template"

	"html/template"

	"github.com/patrickalin/bloomsky-client-go/assembly"
)

//GetTemplate retrieve a template
func GetTemplate(templateName string, templateLocation string, funcs map[string]interface{}, dev bool) *text.Template {
	if dev {
		t, err := text.New(templateName).Funcs(funcs).ParseFiles(templateLocation)

		if err != nil {
			log.Fatalf("Load template console : %v", err)
		}
		return t
	}

	assetBloomsky, err := assembly.Asset(templateLocation)
	t, err := text.New(templateName).Funcs(funcs).Parse(string(assetBloomsky[:]))
	if err != nil {
		log.Fatalf("Load template console : %v", err)
	}
	return t
}

// "bloomsky_header.html","tmpl/bloomsky_header.html",map[string]interface{}{"T": config.translateFunc,}
func GetHtmlTemplate(templateName string, templatesLocation []string, funcs map[string]interface{}, dev bool) *template.Template {
	if dev {
		t := template.New(templateName)
		t.Funcs(funcs)
		t, err := t.ParseFiles(templatesLocation...)

		if err != nil {
			log.Fatalf("Template part 1 : %v", err)
		}

		return t
	}

	asset, err := assembly.Asset(templatesLocation[0])

	if err != nil {
		log.Fatalf("Template part 1 assembly: %v", err)
	}

	t, err := template.New(templateName).Funcs(funcs).Parse(string(asset[:]))

	if err != nil {
		log.Fatalf("Template part 1 : %v", err)
	}
	return t

}