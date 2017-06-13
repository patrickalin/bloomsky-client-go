package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/nicksnyder/go-i18n/i18n"
)

var testTemplate2 *template.Template

var funcMap2 = map[string]interface{}{
	"T": i18n.IdentityTfunc,
}

func initTranslation() {
	i18n.MustLoadTranslationFile("lang/en-US.all.json")
	T, _ := i18n.Tfunc("en-US")

	var err error
	//testTemplate, err = template.New("hello.gohtml").Funcs(funcMap).ParseFiles("hello.gohtml")

	testTemplate2, err = template.New("hello.gohtml").Funcs(map[string]interface{}{
		"T": T,
	}).ParseFiles("hello.gohtml")

	if err != nil {
		panic(err)
	}

	fmt.Println(T("program_greeting"))

	err = testTemplate2.Execute(os.Stdout, map[string]interface{}{
		"Person": "Bob",
	})
	if err != nil {
		fmt.Printf("%v", err)
	}
}
