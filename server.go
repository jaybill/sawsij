package sawsij

import (
	"database/sql"
    "encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	_ "github.com/bmizerany/pq"
	"github.com/stathat/jconfig"
)

type Context struct {
	Config   *jconfig.Config
	Db       *sql.DB
	BasePath string
}

var context *Context
var parsedTemplate *template.Template

func parseTemplates() {
	viewPath := context.BasePath + "/templates"
	templateDir, err := os.Open(viewPath)
	if err != nil {
		log.Print(err)
	}

	allFiles, err := templateDir.Readdirnames(0)
	if err != nil {
		log.Print(err)
	}
	templateExt := "html"
	var templateFiles []string
	for i := 0; i < len(allFiles); i++ {
		if si := strings.Index(allFiles[i], templateExt); si != -1 {
			if si == len(allFiles[i])-len(templateExt) {
				templateFiles = append(templateFiles, viewPath+"/"+allFiles[i])
			}
		}
	}
	log.Printf("Templates: %v", templateFiles)

	pt, err := template.New("dummy").Delims("<%", "%>").ParseFiles(templateFiles...)
	parsedTemplate = pt
	if err != nil {
		log.Print(err)
	}
}

func Route(pattern string, fn func(*http.Request, *Context, map[string](string)) (map[string](interface{}), error)) {

	patternParts := strings.Split(pattern, "/")
	maxParts := len(patternParts)
	log.Printf("Pattern length: %d\tLastIndexOf /:%d", len(pattern)-1, strings.LastIndex(pattern, "/"))

	if strings.LastIndex(pattern, "/") == len(pattern)-1 && len(pattern) > 1 {
		maxParts = maxParts - 1
	}

	templateParts := make([]string, 0)
	for i := 0; i < maxParts; i++ {
		if i > 0 {
			if patternParts[i] != "" {
				templateParts = append(templateParts, patternParts[i])
			} else {
				templateParts = append(templateParts, "index")
			}
		}

	}
	templateId := strings.Join(templateParts, "-")
	http.HandleFunc(pattern, makeHandler(fn, templateId, pattern))
}

func makeHandler(fn func(*http.Request, *Context, map[string](string)) (map[string](interface{}), error), templateId string, pattern string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	    log.Printf("Request method from handler: %q",r.Method) 
		
		
		if !context.Config.GetBool("cacheTemplates") {
			parseTemplates()
		}

		log.Printf("URL path: %v", r.URL.Path)
        returnType,restOfUrl := GetReturnType(r.URL.Path)

        urlParams := GetUrlParams(pattern,restOfUrl)
		log.Printf("URL vars: %v", urlParams)
		handlerResults, err := fn(r, context, urlParams)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			switch returnType {
			case RT_XML:
				//TODO Return actual XML here
				w.Header().Set("Content-Type", "text/xml")
				fmt.Fprintf(w, "%s", xml.Header)
				log.Print("returning xml")
				type Response struct {
					Error string
				}
				r := Response{Error: "NOT YET IMPLEMENTED"}
				b, err := xml.Marshal(r)
				if err != nil {
					log.Print(err)
				} else {
					fmt.Fprintf(w, "%s", b)
				}
			case RT_JSON:
				w.Header().Set("Content-Type", "application/json")                
				log.Print("returning json")

                var iToRender interface{}                            
                if len(handlerResults) == 1{
                
      				var keystring string
    
                	for key, value := range handlerResults {
	                    if _, ok := value.(interface{}); ok {
		                    keystring = key
	                    }
                    }
    	            log.Printf("handler returned single value array. returning value of %q", keystring)
                    
                    iToRender = handlerResults[keystring]
                } else {
                    iToRender = handlerResults
                }				
                
				b, err := json.Marshal(iToRender)
				if err != nil {
					log.Print(err)
				} else {
					fmt.Fprintf(w, "%s", b)
				}
			default:
				templateFilename := templateId + ".html"
				err = parsedTemplate.ExecuteTemplate(w, templateFilename, handlerResults)
				if err != nil {
					log.Print(err)
				}
			}

		}
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving static resource %q - method: %q", r.URL.Path,r.Method)
	http.ServeFile(w, r, context.BasePath+r.URL.Path)
}

func Configure() {

	var ctx Context
	context = &ctx

	if len(os.Args) == 1 {
		log.Fatal("No basepath file specified.")
	}

	context.BasePath = string(os.Args[1])
	configFilename := context.BasePath + "/etc/config.json"

	log.Print("Using config file [" + configFilename + "]")

	c := jconfig.LoadConfig(configFilename)
	context.Config = c

	db, err := sql.Open("postgres", context.Config.GetString("dbConnect"))
	if err != nil {
		log.Fatal(err)
	}
	parseTemplates()

	context.Db = db
	log.Print("Static dir is [" + context.BasePath + "/static" + "]")
	http.HandleFunc("/static/", staticHandler)

}


func Run() {  
	log.Print("Listening on port [" + context.Config.GetString("port") + "]")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v",context.Config.GetString("port")), nil))
}

