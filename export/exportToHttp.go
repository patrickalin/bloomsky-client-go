package export

import (
	"fmt"
	"net/http"

	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
	config "github.com/patrickalin/bloomsky-client-go/config"
)

var mybloomsky bloomskyStructure.BloomskyStructure

//displayToHTTP print major informations from a bloomsky JSON to console
func displayToHTTP(onebloomsky bloomskyStructure.BloomskyStructure) {
	mybloomsky = onebloomsky
}

func handler(w http.ResponseWriter, r *http.Request) {

	mylog.Trace.Println("Handle")
	fmt.Fprintln(w, "Bloomsky")
	fmt.Fprintln(w, "")
	//fmt.Fprintln(w, mybloomsky.GetTimeStamp())
	fmt.Fprintf(w, "\nTimestamp : \t \t%s\n", mybloomsky.GetTimeStamp())
	fmt.Fprintf(w, "City : \t \t \t%s\n", mybloomsky.GetCity())
	fmt.Fprintf(w, "Device Id : \t \t%s\n", mybloomsky.GetDeviceID())
	fmt.Fprintf(w, "Num Of Followers : \t%d\n", mybloomsky.GetNumOfFollowers())
	fmt.Fprintf(w, "Index UV : \t \t%s\n", mybloomsky.GetIndexUV())
	fmt.Fprintf(w, "Night : \t \t%t\n", mybloomsky.IsNight())
	fmt.Fprintf(w, "Wind Direction : \t%s\n", mybloomsky.GetWindDirection())
	fmt.Fprintf(w, "Wind Gust : \t \t%.2f mPh\n", mybloomsky.GetWindGustMph())
	fmt.Fprintf(w, "Sustained Wind Speed : \t%.2f mPh\n", mybloomsky.GetSustainedWindSpeedMph())
	fmt.Fprintf(w, "Wind Gust : \t \t%.2f km/h\n", mybloomsky.GetWindGustMs())
	fmt.Fprintf(w, "Sustained Wind Speed : \t%.2f m/s\n", mybloomsky.GetSustainedWindSpeedMs())
	fmt.Fprintf(w, "Rain : \t \t \t%t\n", mybloomsky.IsRain())
	fmt.Fprintf(w, "Rain Daily : \t \t%.2f in\n", mybloomsky.GetRainDailyIn())
	fmt.Fprintf(w, "24h Rain : \t \t%.2f in\n", mybloomsky.GetRainIn())
	fmt.Fprintf(w, "Rain Rate : \t \t%.2f in\n", mybloomsky.GetRainRateIn())
	fmt.Fprintf(w, "Rain Daily : \t \t%.2f mm\n", mybloomsky.GetRainDailyMm())
	fmt.Fprintf(w, "24h Rain : \t \t%.2f mm\n", mybloomsky.GetRainMm())
	fmt.Fprintf(w, "Rain Rate : \t \t%.2f mm\n", mybloomsky.GetRainRateMm())
	fmt.Fprintf(w, "Temperature F : \t%.1f °F\n", mybloomsky.GetTemperatureFahrenheit())
	fmt.Fprintf(w, "Temperature C : \t%.1f °C\n", mybloomsky.GetTemperatureCelsius())
	fmt.Fprintf(w, "Humidity : \t \t%.f %%\n", mybloomsky.GetHumidity())
	fmt.Fprintf(w, "Pressure InHg : \t%.1f inHg\n", mybloomsky.GetPressureInHg())
	fmt.Fprintf(w, "Pressure HPa : \t \t%.1f hPa\n", mybloomsky.GetPressureHPa())
}

//NewServer create web server
func NewServer(oneConfig config.Configuration) {
	mylog.Trace.Printf("Init server http %s", oneConfig.HTTPPort)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(oneConfig.HTTPPort, nil)
	if err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error when I create the server : %v", err))
	}
	mylog.Trace.Printf("Server ok on http %s", oneConfig.HTTPPort)
}

//InitHTTP listen on the chanel
func InitHTTP(messages chan bloomskyStructure.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to http")

		for {
			mylog.Trace.Println("Receive message to export to http")
			msg := <-messages
			displayToHTTP(msg)
		}
	}()
}
