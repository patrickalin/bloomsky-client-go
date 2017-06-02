package bloomskyStructure

import (
	"os"
	"testing"

	mylog "github.com/patrickalin/GoMyLog"
)

var mybloomsky BloomskyStructure

func TestMain(m *testing.M) {
	mylog.Init(mylog.ERROR)
	body := []byte("[{\"UTC\":2,\"CityName\":\"Thuin\",\"Storm\":{\"UVIndex\":\"1\",\"WindDirection\":\"E\",\"RainDaily\":0,\"WindGust\":0,\"SustainedWindSpeed\":0,\"RainRate\":0,\"24hRain\":0},\"Searchable\":true,\"DeviceName\":\"skyThuin\",\"RegisterTime\":1486905295,\"DST\":1,\"BoundedPoint\":\"\",\"LON\":4.3101,\"Point\":{},\"VideoList\":[\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-27.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-28.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-29.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-30.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-31.mp4\"],\"VideoList_C\":[\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-27_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-28_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-29_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-30_C.mp4\",\"http://s3.amazonaws.com/bskytimelapses/faBiuZWsnpaoqZqr_2_2017-05-31_C.mp4\"],\"DeviceID\":\"442C05954A59\",\"NumOfFollowers\":2,\"LAT\":50.3394,\"ALT\":195,\"Data\":{\"Luminance\":9999,\"Temperature\":70.79,\"ImageURL\":\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5uqmZammJw=.jpg\",\"TS\":1496345207,\"Rain\":false,\"Humidity\":64,\"Pressure\":29.41,\"DeviceType\":\"SKY2\",\"Voltage\":2611,\"Night\":false,\"UVIndex\":9999,\"ImageTS\":1496345207},\"FullAddress\":\"Drève des Alliés, Thuin, Wallonie, BE\",\"StreetName\":\"Drève des Alliés\",\"PreviewImageList\":[\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5qwlZOmn5c=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5qwnZmqmZw=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5unnJakmZg=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5uom5Kkm50=.jpg\",\"http://s3-us-west-1.amazonaws.com/bskyimgs/faBiuZWsnpaoqZqrqJ1kr5upmZiqnps=.jpg\"]}]")

	mybloomsky = NewBloomskyFromBody(body)

	os.Exit(m.Run())
}

func TestGetCity(t *testing.T) {

	t.Log("City Thuin")
	if city := mybloomsky.GetCity(); city != "Thuin" {
		t.Errorf("Expected Thuin, but it was %s instead.", city)
	}
}

func TestDeviceId(t *testing.T) {

	t.Log("DevideID 442C05954A59 ")
	if city := mybloomsky.GetDeviceID(); city != "442C05954A59" {
		t.Errorf("Expected 442C05954A59, but it was %s instead.", city)
	}
}
