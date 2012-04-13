package sawsij

import (
	"log"
	"net/http"
	"os"
    
    "strings"
    "text/template"
    

	"github.com/stathat/jconfig"
	"launchpad.net/mgo"

)

var context             *Context
var parsedTemplate      *template.Template

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

func parseTemplates(){
    viewPath            := context.BasePath + "/templates"
    templateDir,err     := os.Open(viewPath)
    if err != nil { log.Print(err) }
    
    allFiles,err   := templateDir.Readdirnames(0)
    if err != nil { log.Print(err) }
    templateExt   := "html"
    var templateFiles []string
    for i := 0; i < len(allFiles); i++ {
       if si := strings.Index(allFiles[i], templateExt); si != -1{
            if si == len(allFiles[i]) - len(templateExt){
                templateFiles = append(templateFiles,viewPath + "/" + allFiles[i])
            }
        }
    }
    log.Printf("Templates: %v",templateFiles)
    
    pt, err := template.ParseFiles(templateFiles...)
    parsedTemplate = pt
    if err != nil { log.Print(err) }    
}

func makeHandler(fn func(*http.Request, *Context) interface{},templateId string ) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerResults := fn(r, context)
		templateFilename    := templateId + ".html"
		err := parsedTemplate.ExecuteTemplate(w, templateFilename,handlerResults)
        if err != nil { log.Print(err) }
    }
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("Serving static resource %q",r.URL.Path)	
    http.ServeFile(w,r,context.BasePath + r.URL.Path)
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
	parseTemplates()		
	dbSession.SetMode(mgo.Monotonic, true)
	context.DbSession = dbSession
    log.Print("Static dir is [" + context.BasePath + "/static" + "]")	
    http.HandleFunc("/static/",staticHandler )
}

func Run() {
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(":"+context.Config.GetString("port"), nil))
}
