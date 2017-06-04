package export

import (
	"fmt"
	"html/template"
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

	//t := template.New("bloomsky") // Create a template.
	//t = t.Funcs(template.FuncMap{"GetTimeStamp": mybloomsky.GetTimeStamp()})

	t, err := template.ParseFiles("tmpl/bloomsky.html") // Parse template file.
	if err != nil {
		fmt.Printf("%v", err)
	}
	err = t.Execute(w, mybloomsky) // merge.
	if err != nil {
		fmt.Printf("%v", err)
	}

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
