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

	serv, err = createWebServer(channels["web"], ":1110", ":2220", translateFunc, true, store{})
	checkErr(err, funcName(), "Impossible to create server", "")

	os.Exit(m.Run())
}

func TestConfig(t *testing.T) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v", err)
	}
}

func TestHanlderHome(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:1110/",
		nil,
	)
	if err != nil {
		logFatal(err, funcName(), "Could not request: %v")
	}

	rec := httptest.NewRecorder()
	serv.home(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200; got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "est") {
		t.Errorf("unexpected body; %q", rec.Body.String())
	}
}

func TestHanlderLog(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:1110/log",
		nil,
	)
	if err != nil {
		logFatal(err, funcName(), "Could not request: %v")
	}

	rec := httptest.NewRecorder()
	serv.log(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200; got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Server HTTPS listen on port") {
		t.Errorf("unexpected body; %q", rec.Body.String())
	}
}

func TestHanlderHistory(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:1110/history",
		nil,
	)
	if err != nil {
		logFatal(err, funcName(), "Could not request: %v")
	}

	rec := httptest.NewRecorder()
	serv.history(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200; got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "corechart") {
		t.Errorf("unexpected body; %q", rec.Body.String())
	}
}

func BenchmarkHanlder(b *testing.B) {

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(
			http.MethodGet,
			"http://localhost:1110",
			nil,
		)
		if err != nil {
			logFatal(err, funcName(), "Could not request: %v")
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

/*
func TestRing(t *testing.T) {

	channels := make(map[string]chan bloomsky.Bloomsky)
	channels["store"] = make(chan bloomsky.Bloomsky)

	store, err := createStore(channels["store"])
	checkErr(err, funcName(), "Error with history create store")

	store.listen(context.Background())

	mybloomsky := bloomsky.New("", "", true, nil)
	collect(mybloomsky, channels)
	mybloomsky = bloomsky.New("", "", true, nil)
	collect(mybloomsky, channels)
	mybloomsky = bloomsky.New("", "", true, nil)
	collect(mybloomsky, channels)

	if store.String("temp") == "[ [new Date(\"Tue Jul  4 22:16:25 2017\"),21.55] ,[new Date(\"Tue Jul  4 22:16:25 2017\"),21.55] ,[new Date(\"Tue Jul  4 22:16:25 2017\"),21.55]]" {
		t.Errorf("unexpected string : |%s|", store.String("temp"))
	}
}*/
