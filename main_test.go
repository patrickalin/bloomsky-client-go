package main

import (
	"fmt"
	"os"
	"testing"

	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
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

func Test_displayToConsole(t *testing.T) {
	type args struct {
		onebloomsky bloomskyStructure.BloomskyStructure
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayToConsole(tt.args.onebloomsky)
		})
	}
}
