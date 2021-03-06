package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/i18n"
	bloomsky "github.com/patrickalin/bloomsky-api-go"
)

var serv *httpServer

func TestMain(m *testing.M) {

	if err := i18n.ParseTranslationFileBytes("lang/en-us.all.json", readFile("lang/en-us.all.json", true)); err != nil {
		logFatal(err, funcName(), "Error read language file check in config.yaml if dev=false", "")
	}

	translateFunc, err := i18n.Tfunc("en-us")
	checkErr(err, funcName(), "Problem with loading translate file", "")

	channels := make(map[string]chan bloomsky.Bloomsky)

	serv, err = createWebServer(channels["web"], ":1110", ":2220", translateFunc, true, store{}, false)
	checkErr(err, funcName(), "Impossible to create server", "")

	os.Exit(m.Run())
}

func TestLoadCorrectConfig(t *testing.T) {
	conf := readConfig("configForTest")
	//conf2 := initServerConfiguration("wrongConfigForTest")
	tests := []struct {
		name   string
		fields bool
		want   bool
	}{
		{"Test Mock good conf", conf.mock, true},
		{"Test Devel good conf", conf.dev, true},
		//{"Test Mock wrong conf", conf2.mock, true},
		//{"Test Devel wrong conf", conf2.dev, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields; got != tt.want {
				t.Errorf("mock = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestHanlderHome(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost:1110/",
		nil,
	)
	if err != nil {
		logFatal(err, funcName(), "Could not create request: %v")
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
