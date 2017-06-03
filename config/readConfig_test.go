package config

import (
	"os"
	"testing"

	mylog "github.com/patrickalin/GoMyLog"
)

func TestMain(m *testing.M) {
	mylog.Init(mylog.ERROR)

	os.Exit(m.Run())
}

func TestReadConfigFound(t *testing.T) {
	New("configForTest")
}

func TestReadConfigNotFound(t *testing.T) {
	//New("configError")
}

func TestReadURL(t *testing.T) {
	if url := New("configForTest").GetURL(); url != "https://api.bloomsky.com/api/skydata/" {
		t.Errorf("Expected https://api.bloomsky.com/api/skydata/, but it was %s instead.", url)
	}

}
