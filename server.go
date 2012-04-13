package sawsij

import (
	"log"
	"net/http"
	"os"
    //"fmt"
    "strings"
    "text/template"
    "bytes"

	"github.com/stathat/jconfig"
	"launchpad.net/mgo"

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
		handlerResults := fn(r, context)
		
		layoutId            := "layout"			
		viewPath            := context.BasePath + "/views"							    
		templateFilename    := templateId + ".html"
		layoutFilename      := layoutId + ".html"
	    templatePath        := viewPath + "/templates/" + templateFilename
		layoutPath          := viewPath + "/layouts/" + layoutFilename
				
		tmpl, err := template.ParseFiles(templatePath,layoutPath)

		alltemplates := tmpl.Templates()

		for i :=0; i < len(alltemplates); i++ {
    		log.Printf("template: %q",alltemplates[i].Name());
		}
		
        if err != nil { log.Print(err) }

        // parse the content template first
        content := new(bytes.Buffer)        
        err = tmpl.ExecuteTemplate(content, templateFilename,handlerResults)
        type ParsedContent struct{
        	Content string		        
        }
        
        var parsedContent ParsedContent
        parsedContent.Content = content.String()
        
        if err != nil { 
            log.Print(err) 
        } else {
            // take the results of parsing the content template and cram it into the layout template, then write that out
            err = tmpl.ExecuteTemplate(w, layoutFilename,parsedContent)
        }
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
			
	dbSession.SetMode(mgo.Monotonic, true)
	context.DbSession = dbSession
    log.Print("Static dir is [" + context.BasePath + "/static" + "]")	
    http.HandleFunc("/static/",staticHandler )
}

func Run() {
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(":"+context.Config.GetString("port"), nil))
}
