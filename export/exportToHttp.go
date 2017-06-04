package export

import (
	"fmt"
	"html/template"
	"net/http"

	mylog "github.com/patrickalin/GoMyLog"
	bloomskyStructure "github.com/patrickalin/bloomsky-client-go/bloomskyStructure"
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
func NewServer(HTTPPort string) {
	mylog.Trace.Printf("Init server http port %s", HTTPPort)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(HTTPPort, nil)
	if err != nil {
		mylog.Error.Fatal(fmt.Errorf("Error when I create the server : %v", err))
	}
	mylog.Trace.Printf("Server ok on http %s", HTTPPort)
}

//InitHTTP listen on the chanel
func InitHTTP(messages chan bloomskyStructure.BloomskyStructure) {
	go func() {

		mylog.Trace.Println("Init the queue to receive message to export to http")

		for {
			msg := <-messages
			mylog.Trace.Println("Receive message to export to http")
			displayToHTTP(msg)
		}
	}()
}
