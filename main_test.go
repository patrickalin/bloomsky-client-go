package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
	"github.com/spf13/viper"
)

var serv *httpServer

func TestMain(m *testing.M) {

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readFile("lang/en-us.all.json", true)); err != nil {
		logFatal(err, funcName(), "Error read language file check in config.yaml if dev=false", "")
	}

	translateFunc, err := i18n.Tfunc("en-us")
	checkErr(err, funcName(), "Problem with loading translate file", "")

	channels := make(map[string]chan bloomsky.Bloomsky)
	serv, err = createWebServer(channels["web"], ":1111", ":2222", translateFunc, true)
	os.Exit(m.Run())
}

func TestSomething(t *testing.T) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v", err)
	}
}

/*
func TestMain(m *testing.M) {
	i18n.MustLoadTranslationFile("lang/en-US.all.json")
	i18n.MustLoadTranslationFile("lang/fr.all.json")
	mylog.Init(mylog.ERROR)

	os.Exit(m.Run())
}

func TestReadConfigFound(t *testing.T) {
	if err := readConfig("configForTest"); err != nil {
		fmt.Printf("%v", err)
	}
}*/

/*func TestReadConfigNotFound(t *testing.T) {
	if err := readConfig("configError"); err != nil {
		fmt.Printf("%v", err)
	}
}*/

func TestHanlder(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:1111",
		nil,
	)
	if err != nil {
		t.Fatalf("Could not request: %v", err)
	}

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readFile("lang/en-us.all.json", true)); err != nil {
		logFatal(err, funcName(), "Error read language file check in config.yaml if dev=false", "")
	}

	translateFunc, err := i18n.Tfunc("en-us")
	checkErr(err, funcName(), "Problem with loading translate file", "")

	channels := make(map[string]chan bloomsky.Bloomsky)

	rec := httptest.NewRecorder()

	httpServ, err := createWebServer(channels["web"], ":1112", ":2223", translateFunc, true)
	httpServ.home(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200; got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Est") {
		t.Errorf("unexpected body; %q", rec.Body.String())
	}
}

func BenchmarkHanlder(b *testing.B) {

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			"http://localhost:1111",
			nil,
		)
		if err != nil {
			logFatal(err, funcName(), "Could not request: %v", "")
		}

		rec := httptest.NewRecorder()
		serv.home(rec, req)

		if rec.Code != http.StatusOK {
			b.Errorf("expected 200; got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Est") {
			b.Errorf("unexpected body; %q", rec.Body.String())
		}
	}
}
