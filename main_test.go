package main

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

func TestSomething(t *testing.T) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v", err)
	}
}
