package sawsij

import (
	"log"
	"net/http"
	"os"

	"github.com/stathat/jconfig"
	"launchpad.net/mgo"
)

var context *Context

func MakeHandler(fn func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, context)
	}
}

func Configure() {

	var ctx Context
	context = &ctx

	if len(os.Args) == 1 {
		log.Fatal("No configuration file specified.")
	}

	configFilename := string(os.Args[1])
	log.Print("Using config file [" + configFilename + "]")

	c := jconfig.LoadConfig(configFilename)
	context.Config = c

	log.Print("Using db host [" + context.Config.GetString("mongo_host") + "]")
	dbSession, err := mgo.Dial(context.Config.GetString("mongo_host"))
	if err != nil {
		panic(err)
	}
	//defer dbSession.Close()		
	dbSession.SetMode(mgo.Monotonic, true)
	context.DbSession = dbSession

}

func Run() {
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(":"+context.Config.GetString("port"), nil))
}
