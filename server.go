// Package sawsij provides a small, opinionated web framework.
package sawsij

import (
	"code.google.com/p/gorilla/sessions"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/bmizerany/pq"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const (
	R_GUEST = 0
)

type IUser interface {
	SetPassword(password string, salt string)
	TestPassword(password string, salt string) (valid bool)
	GetRole() int64
	ClearPasswordHash()
}

// A AppScope is passed along to a request handler and stores application configuration, the handle to the database and any derived information, like the base path.
// This will probably be supplanted soon by something better.
type AppScope struct {
	Config   *yaml.File
	Db       *sql.DB
	BasePath string
	Setup    *AppSetup
}

// A RequestScope is sent to handler functions and contains session and derived URL information.
type RequestScope struct {
	Session   *sessions.Session
	UrlParams map[string]string
}

// AppSetup is used by Configure() to set app configuration variables.
type AppSetup struct {
	User IUser
}

var store *sessions.CookieStore
var appScope *AppScope
var parsedTemplate *template.Template

func parseTemplates() {
	viewPath := appScope.BasePath + "/templates"
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

	pt, err := template.New("dummy").Delims("<%", "%>").Funcs(GetFuncMap()).ParseFiles(templateFiles...)
	parsedTemplate = pt
	if err != nil {
		log.Print(err)
	}
}

// HandlerResponse is a struct that your handler functions return. It contains all the data needed to generate the response. If Redirect is set,
// the contents of View is ignored.
type HandlerResponse struct {
	View     map[string](interface{})
	Redirect string
}

// Init sets up an empty map for the handler response.
func (h *HandlerResponse) Init() {
	h.View = make(map[string]interface{})
}

// RouteConfig is what is supplied to the Route() function to set up a route. More about how this is used in the documentation for the Route function.
type RouteConfig struct {
	Pattern string
	Handler func(*http.Request, *AppScope, *RequestScope) (HandlerResponse, error)
	Roles   []int
}

// Route takes route config and sets up a handler. This is the primary means by which applications interact with the framework.
// Handler functions must accept a pointer to an http.Request, a pointer to a AppScope and a map of strings with a string key, which will contain the URL
// params.
// The RequestScope struct contains a map of url params and a session struct.
// URL params are defined as anything after the pattern that can be split into pairs. So, for example, if your pattern was "/admin/" and the actual URL
// was "/admin/id/14/display/1", the URL param map your handler function gets would be:
// "id" = "14"
// "display" = "1"
//
// Note that these are strings, so you'll need to convert them to whatever types you need. If you just need an Int id, there's a useful utility function,
// sawsij.GetIntId()
//
// If you start a pattern with "/json", whatever you return will be marshalled into JSON instead of being passed through to a template. Same goes for "/xml" though
// this isn't implemented yet.
//
// The template filename to be used is based on the pattern, with slashes being converted to dashes. So "/admin" looks for "[app_root_dir]/templates/admin.html"
// and "/posts/list" will look for "[app_root_dir]/templates/posts-list.html". The pattern "/" will look for "[app_root_dir]/index.html".
//
// You generally call Route() once per pattern after you've called Configure() and before you call Run().
func Route(rcfg RouteConfig) {

	templateId := GetTemplateName(rcfg.Pattern)

	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request method from handler: %q", r.Method)

		cacheTemplates, err := appScope.Config.Get("server.cacheTemplates")
		if err != nil {
			log.Print(err)
		} else {
			if cacheTemplates != "true" {
				parseTemplates()
			}
		}

		log.Printf("URL path: %v", r.URL.Path)
		returnType, restOfUrl := GetReturnType(r.URL.Path)

		urlParams := GetUrlParams(rcfg.Pattern, restOfUrl)
		log.Printf("URL vars: %v", urlParams)
		global := make(map[string]string)
		session, _ := store.Get(r, "session")

		reqScope := RequestScope{UrlParams: urlParams, Session: session}

		handlerResults, err := rcfg.Handler(r, appScope, &reqScope)
		reqScope.Session.Save(r, w)

		if handlerResults.Redirect != "" {
			http.Redirect(w, r, handlerResults.Redirect, http.StatusFound)
		} else {

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
					if len(handlerResults.View) == 1 {

						var keystring string

						for key, value := range handlerResults.View {
							if _, ok := value.(interface{}); ok {
								keystring = key
							}
						}
						log.Printf("handler returned single value array. returning value of %q", keystring)

						iToRender = handlerResults.View[keystring]
					} else {
						iToRender = handlerResults.View
					}

					b, err := json.Marshal(iToRender)
					if err != nil {
						log.Print(err)
					} else {
						fmt.Fprintf(w, "%s", b)
					}
				default:
					templateFilename := templateId + ".html"
					// Add "global" template variables
					handlerResults.View["global"] = global
					err = parsedTemplate.ExecuteTemplate(w, templateFilename, handlerResults.View)
					if err != nil {
						log.Print(err)
					}
				}
			}

		}
	}

	http.HandleFunc(rcfg.Pattern, fn)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving static resource %q - method: %q", r.URL.Path, r.Method)
	http.ServeFile(w, r, appScope.BasePath+r.URL.Path)
}

// Configure gets the application base path from a command line argument.
// It then reads the config file at [app_root_dir]/etc/config.json (This will probably be changed to YAML at some point.)
// It then attempts to grab a handle to the database, which it sticks into the appScope.
// Configure is the first thing your application will call in its "main" method.
func Configure(as *AppSetup) {

	a := AppScope{Setup: as}
	appScope = &a

	if len(os.Args) == 1 {
		log.Fatal("No basepath file specified.")
	}

	appScope.BasePath = string(os.Args[1])
	configFilename := appScope.BasePath + "/etc/config.yaml"

	log.Print("Using config file [" + configFilename + "]")

	c, err := yaml.ReadFile(configFilename)
	if err != nil {
		log.Fatal(err)
	}
	appScope.Config = c

	driver, err := c.Get("database.driver")
	if err != nil {
		log.Fatal(err)
	}
	connect, err := c.Get("database.connect")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open(driver, connect)
	if err != nil {
		log.Fatal(err)
	}

	key, err := c.Get("encryption.key")
	if err != nil {
		log.Fatal(err)
	}

	store = sessions.NewCookieStore([]byte(key))

	appScope.Db = db
	log.Print("Static dir is [" + appScope.BasePath + "/static" + "]")
	http.HandleFunc("/static/", staticHandler)

	parseTemplates()
}

// Run will start a web server on the port specified in the config file, using the configuration in the config file and the routes specified by any Route() calls
// that have been previously made. This is generally the last line of your application's "main" method.
func Run() {

	port, err := appScope.Config.Get("server.port")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port [" + port + "]")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

