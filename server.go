package sawsij

import (
	"log"
	"net/http"
	"os"
    "fmt"
    "strings"

	"github.com/stathat/jconfig"
	"launchpad.net/mgo"
    "github.com/hoisie/mustache"
)

var context     *Context

func Route(pattern string,fn func(*http.Request, *Context) interface{}) {
    
    patternParts := strings.Split(pattern,"/")
    var templateId string

    for i := 0; i < len(patternParts); i++ {
		if i > 0{
		    if patternParts[i] != "" {
                templateId += patternParts[i] 
		    } else {
		        templateId += "index"
		    }
		    if(i < len(patternParts) - 1){
		        templateId += "-"		   
		    } 
		}
	}  
    http.HandleFunc(pattern, makeHandler(fn,templateId))
}

func makeHandler(fn func(*http.Request, *Context) interface{},templateId string ) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromHandler := fn(r, context)
		
		layoutId    := "layout"			
		viewPath            := context.BasePath + "/views"						
	    templateFilename    := viewPath + "/templates/" + templateId + ".html.ms"	    
		layoutFilename      := viewPath + "/layouts/" + layoutId + ".html.ms"
		rendered            := mustache.RenderFileInLayout(templateFilename, layoutFilename, fromHandler)
		
		fmt.Fprint(w,rendered)  
	}
}

func Configure() {

	var ctx Context
	context = &ctx

	if len(os.Args) == 1 {
		log.Fatal("No basepath file specified.")
	}
    
	context.BasePath    = string(os.Args[1]) 
	configFilename      := context.BasePath + "/etc/config.json"
	
	log.Print("Using config file [" + configFilename + "]")

	c := jconfig.LoadConfig(configFilename)
	context.Config = c

	log.Print("Using db host [" + context.Config.GetString("mongo_host") + "]")
	
	dbSession, err := mgo.Dial(context.Config.GetString("mongo_host"))
	if err != nil {
		panic(err)
	}
			
	dbSession.SetMode(mgo.Monotonic, true)
	context.DbSession = dbSession
	

}

func Run() {
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(":"+context.Config.GetString("port"), nil))
}
