package export

import (
	"fmt"

	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
)

//print major informations from a bloomsky JSON to console
func displayToConsole(onebloomsky bloomskyStructure.BloomskyStructure) {

	fmt.Printf("\nTimestamp : \t \t%s\n", onebloomsky.GetTimeStamp())
	fmt.Printf("City : \t \t \t%s\n", onebloomsky.GetCity())
	fmt.Printf("Device Id : \t \t%s\n", onebloomsky.GetDeviceID())
	fmt.Printf("Num Of Followers : \t%d\n", onebloomsky.GetNumOfFollowers())
	fmt.Printf("Index UV : \t \t%s\n", onebloomsky.GetIndexUV())
	fmt.Printf("Night : \t \t%t\n", onebloomsky.IsNight())
	fmt.Printf("Wind Direction : \t%s\n", onebloomsky.GetWindDirection())
	fmt.Printf("Wind Gust : \t \t%.2f mPh\n", onebloomsky.GetWindGustMph())
	fmt.Printf("Sustained Wind Speed : \t%.2f mPh\n", onebloomsky.GetSustainedWindSpeedMph())
	fmt.Printf("Wind Gust : \t \t%.2f km/h\n", onebloomsky.GetWindGustMs())
	fmt.Printf("Sustained Wind Speed : \t \t%.2f m/s\n", onebloomsky.GetSustainedWindSpeedMs())
	fmt.Printf("Rain : \t \t \t%t\n", onebloomsky.IsRain())
	fmt.Printf("Rain Daily : \t \t%.2f in\n", onebloomsky.GetRainDailyIn())
	fmt.Printf("24h Rain : \t \t%.2f in\n", onebloomsky.GetRainIn())
	fmt.Printf("Rain Rate : \t \t%.2f in\n", onebloomsky.GetRainRateIn())
	fmt.Printf("Rain Daily : \t \t%.2f mm\n", onebloomsky.GetRainDailyMm())
	fmt.Printf("24h Rain : \t \t%.2f mm\n", onebloomsky.GetRainMm())
	fmt.Printf("Rain Rate : \t \t%.2f mm\n", onebloomsky.GetRainRateMm())
	fmt.Printf("Temperature F : \t%.1f °F\n", onebloomsky.GetTemperatureFahrenheit())
	fmt.Printf("Temperature C : \t%.1f °C\n", onebloomsky.GetTemperatureCelsius())
	fmt.Printf("Humidity : \t \t%.f %%\n", onebloomsky.GetHumidity())
	fmt.Printf("Pressure InHg : \t%.1f inHg\n", onebloomsky.GetPressureInHg())
	fmt.Printf("Pressure HPa : \t \t%.1f hPa\n", onebloomsky.GetPressureHPa())
}

//InitConsole listen on the chanel
func InitConsole(messages chan bloomskyStructure.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to console")

		for {
			mylog.Trace.Println("Receive message to export to console")
			msg := <-messages
			displayToConsole(msg)
		}
	}()
}
