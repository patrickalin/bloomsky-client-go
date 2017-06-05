package main

import (
	"fmt"
	"os"
	"testing"

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
