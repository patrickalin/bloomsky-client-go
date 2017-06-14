package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/nicksnyder/go-i18n/i18n"
	mylog "github.com/patrickalin/GoMyLog"
	"github.com/spf13/viper"
)

func TestSomething(t *testing.T) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v", err)
	}
}
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
}

/*func TestReadConfigNotFound(t *testing.T) {
	if err := readConfig("configError"); err != nil {
		fmt.Printf("%v", err)
	}
}*/
