package sawsij

import(
	"os"	
	"log"	
	"net/http"
	
	"launchpad.net/mgo"	
	"github.com/stathat/jconfig"
)

var context *Context;

func makeHandler(fn func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r,context)
	}
}

func configure(){	

	context =: new(Context)
	
	if len(os.Args) == 1 {
		log.Fatal("No configuration file specified.")
	}
		
	configFilename := string(os.Args[1])	
	log.Print("Using config file [" + configFilename + "]")
	
	c := jconfig.LoadConfig(configFilename)		
	context.Config = c
	
	dbSession, err := mgo.Dial(context.Config.GetString("mongo_host"))
	if err != nil {
			panic(err)
	}
	//defer dbSession.Close()		
	dbSession.SetMode(mgo.Monotonic, true)	
	context.DbSession = dbSession	
}


func Run() {
	
	configure()	
	http.HandleFunc("/", makeHandler(Handler))
	
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(":" + context.Config.GetString("port"), nil))
}
