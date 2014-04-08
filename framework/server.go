// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Package sawsij provides a small, opinionated web framework.

Sawsij is a framework for building web applications. Generally, new sawsij applications are created with the sawsijcmd tool. It will talk you
through an intitial application configuration and generate all the required files and code. You can learn more about sawsijcmd over here:
https://bitbucket.org/jaybill/sawsij/wiki/Installation

Check out http://sawsij.com for more information and documentation.

*/
package framework

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
	"bitbucket.org/jaybill/sawsij/framework/model/mysql"
	"bitbucket.org/jaybill/sawsij/framework/model/postgres"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/bmizerany/pq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/kylelemons/go-gypsy/yaml"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	R_GUEST = 0
)

// An AppScope is passed along to a request handler and stores application configuration, the handle to the database and any derived information,
// like the base path.
type AppScope struct {
	// A reference to the config file
	Config   *yaml.File
	Db       *model.DbSetup
	BasePath string
	Setup    *AppSetup
	// Can be used to store arbitrary data in the application scope.
	Custom *map[string]interface{}
}

// A RequestScope is sent to handler functions and contains session and derived URL information.
// A note about sessions: If you use the AddFlash() function, the value will be placed in the .global map as "flash".
// The value in "flash" will *only* be present on the next request made against that session and will be gone after that.
// This is useful for "success" messages that need to show up on another page after you've redirected there.
type RequestScope struct {
	// The user session handle
	Session *sessions.Session
	// An array of URL parameters, in the order they came through. Will be populated if ParamAs field of the RouteConfig is set to PARAMS_ARRAY
	UrlParamArray []string
	// A map of URL parameters as a key value map. Will be populated if ParamAs field of the RouteConfig is set to PARAMS_MAP
	UrlParamMap map[string]string
}

// The User interface describes the methods that the framework needs to interact with a user for the purposes of auth and session management.
// Sawsij does not describe its own user struct, that's up to the application.
type User interface {
	// How the framework determines if the user has supplied the correct password
	TestPassword(password string, a *AppScope) bool
	// How the framework determines what role the user has. Currently only has one role.
	GetRole() int64
	// If you're storing a password hash in your user object, implement ClearPasswordHash() so that it blanks that.
	// Otherwise the hash will get stored in the session cookie, which is no good.
	ClearPasswordHash()
}

// AppSetup is used by Configure() to set up callback functions that your application implements to extend the framework
// functionality. It serves as the basis of the "plugin" system. The only exception is GetUser(), which your app must implement
// for the framework to function. The GetUser function supplies a type conforming to the User specification. It's used for auth and
// session mangement.
// Roles is a map of ints with string keys that allow you to make role identifiers available by name from within templates. This isn't
// checked in any way and is solely for ease of use.
// TemplateFuncs is a map of functions that can be called from your templates. If you make the keys the same as any of the built in functions,
// you'll effectively override it.

type AppSetup struct {
	GetUser func(username string, a *AppScope) User

	Roles         *map[string]int
	TemplateFuncs template.FuncMap
}

var store *sessions.CookieStore
var appScope *AppScope
var parsedTemplate *template.Template

// SetCustom is used to add custom data that will be placed in the AppScope. This can later be retrieved in handler functions.
// You supply it with a function that returns a map, and it will set the AppScope that gets passed to handlers to that.
// You should always call SetCustom *after* you call Configure, that way the AppScope your function recieves will
// have all the configuration stuff, like database connections and the config file.
func SetCustom(f func(a *AppScope) *map[string]interface{}) {
	log.Print("Calling custom function")
	appScope.Custom = f(appScope)
	log.Print("Custom is now %+v", *appScope.Custom)
	return
}

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

	if len(templateFiles) > 0 {
		fnm := GetFuncMap()
		if len(appScope.Setup.TemplateFuncs) > 0 {
			for name, fn := range appScope.Setup.TemplateFuncs {
				fnm[name] = fn
			}
		}
		pt, err := template.New("dummy").Delims("<%", "%>").Funcs(fnm).ParseFiles(templateFiles...)
		parsedTemplate = pt
		if err != nil {
			log.Printf("** TEMPLATE PARSE ERROR: %v", err)
		}
	}
}

// HandlerResponse is a struct that your handler functions return. It contains all the data needed to generate the response. If Redirect is set,
// the contents of View is ignored.
// Note: If you only supply one entry in your View map, the *contents* of the map will be passed to the view rather than the whole map. This is done
// to simplify templates and JSON responses with only one entry.
// Headers is an array of standard http headers that will be set on the response.
// Modtime is the last modified time, which is only used when the RouteConfig's ReturnType is RT_RAW
type HandlerResponse struct {
	View     map[string](interface{})
	Redirect string
	Header   http.Header
	Content  io.ReadSeeker
	Modtime  time.Time
}

// Init sets up an empty map for the handler response. Generally the first thing you'll call in your handler function.
func (h *HandlerResponse) Init() {
	h.View = make(map[string]interface{})
}

// RouteConfig is what is supplied to the Route() function to set up a route. More about how this is used in the documentation for the Route function.
type RouteConfig struct {
	// The URL pattern to be matched for this route, i.e. "/admin/users"
	Pattern string
	// A function that will handle this route.
	Handler func(*http.Request, *AppScope, *RequestScope) (HandlerResponse, error)
	// An array of role (ints) that are allowed to access this route.
	Roles []int
	// Setting this to framework.RT_JSON or framework.RT_HTML will force the return type and ignore any URL hints. Setting this to framework.RT_RAW
	// will use http.ServeContent to pass whatever is returned in HandlerResponse.Content (useful for sending binary data like images)
	ReturnType int
	// How parameters will be specified on the URL. Will default to PARAMS_MAP, a key value map. Can be set to PARAMS_ARRAY to return
	// an ordered array of values
	ParamsAs int
	// If TemplateFilename is set, it will be used instead of template name derived from then pattern-based naming convention.
	// The specified template must exist in the [app_root]/templates folder.
	TemplateFilename string
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
// The template filename to be used is based on the pattern, with slashes being converted to dashes. So "/admin" looks for "[app_root_dir]/templates/admin.html"
// and "/posts/list" will look for "[app_root_dir]/templates/posts-list.html". The pattern "/" will look for "[app_root_dir]/index.html".
//
// You generally call Route() once per pattern after you've called Configure() and before you call Run().
func Route(rcfg RouteConfig) {

	var slashRoute string = ""
	if p := strings.LastIndex(rcfg.Pattern, "/"); p != len(rcfg.Pattern)-1 {
		slashRoute = rcfg.Pattern + "/"
	}

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
		var returnType int

		if rcfg.ReturnType == 0 {
			returnType = RT_HTML
		} else {
			returnType = rcfg.ReturnType
		}

		global := make(map[string]interface{})
		session, _ := store.Get(r, "session")
		role := R_GUEST // Set to guest by default
		su := session.Values["user"]

		log.Printf("User: %+v", su)
		log.Printf("Session vals: %+v", session.Values)
		if su != nil {
			u := su.(User)
			role = int(u.GetRole())
		}

		log.Printf("pattern: %v roles that can see this: %v user role: %v", rcfg.Pattern, rcfg.Roles, role)

		var handlerResults HandlerResponse

		if !InArray(role, rcfg.Roles) {
			// This user does not have the right role
			if su == nil {
				// User isn't logged in, send to login page, passing along desired destination
				log.Printf("Request URI for redirect: %v", r.URL.RequestURI())
				dest := base64.URLEncoding.EncodeToString([]byte(r.URL.RequestURI()))
				handlerResults.Redirect = fmt.Sprintf("/login/dest/%v", dest)
			} else {
				// The user IS logged in, they're just not permitted to go here
				handlerResults.Redirect = "/denied"
				handlerResults.Init()
			}
		} else {
			// Everything is ok. Proceed normally.
			reqScope := RequestScope{Session: session}

			switch rcfg.ParamsAs {
			case PARAMS_ARRAY:
				reqScope.UrlParamArray = GetUrlParamsArray(rcfg.Pattern, r.URL.Path)
			default:
				reqScope.UrlParamMap = GetUrlParamsMap(rcfg.Pattern, r.URL.Path)

			}

			global["user"] = session.Values["user"]

			if flashes := session.Flashes(); len(flashes) > 0 {
				log.Printf("Setting error flashes to %+v", flashes[0])
				global["flash"] = flashes[0]
			}

			// Call the supplied handler function and get the results back.
			handlerResults, err = rcfg.Handler(r, appScope, &reqScope)
			reqScope.Session.Save(r, w)
		}

		if handlerResults.Redirect != "" {
			http.Redirect(w, r, handlerResults.Redirect, http.StatusFound)
		} else {

			if err != nil {
				log.Print(err)
				http.Error(w, "An error occured. See log for details.", http.StatusInternalServerError)
			} else {

				for key, values := range handlerResults.Header {

					for _, value := range values {
						w.Header().Add(key, value)
					}

				}

				switch returnType {
				case RT_XML:
					//TODO Return actual XML here (issue #6)
					w.Header().Add("Content-Type", "text/xml")
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
					w.Header().Add("Content-Type", "application/json")
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

				case RT_RAW:

					http.ServeContent(w, r, "", handlerResults.Modtime, handlerResults.Content)
				default:
					var templateFilename string
					if rcfg.TemplateFilename == "" {
						templateFilename = GetTemplateName(rcfg.Pattern) + ".html"
					} else {
						templateFilename = rcfg.TemplateFilename
					}
					log.Printf("Using template file %v", templateFilename)
					// Add "global" template variables
					global["roles"] = *appScope.Setup.Roles
					global["url"] = rcfg.Pattern
					log.Printf("URL sent to template: %v", global["url"])
					if len(global) > 0 {
						if handlerResults.View == nil {
							handlerResults.Init()
						}
						handlerResults.View["global"] = global
					}

					defer func() {
						if err := recover(); err != nil {
							log.Print(err)
							parseTemplates() // parse templates again so we can throw any errors
							return
						}
					}()
					err = parsedTemplate.ExecuteTemplate(w, templateFilename, handlerResults.View)
					if err != nil {
						log.Printf("** TEMPLATE EXECUTION ERROR: %v", err)
					}

				}
			}

		}
	}

	http.HandleFunc(rcfg.Pattern, fn)

	if slashRoute != "" {
		http.HandleFunc(slashRoute, fn)
	}

	return
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving static resource %q - method: %q", r.URL.Path, r.Method)
	http.ServeFile(w, r, appScope.BasePath+r.URL.Path)
}

// Configure gets the application base path from a command line argument unless you specify it.  It then reads the config file at [app_root_dir]/etc/config.yaml.
// It then attempts to grab a handle to the database, which it sticks into the appScope.
// It will also set up a static handler for any files in [app_root_dir]/static, which can be used to serve up images, CSS and JavaScript.
// Configure is the first thing your application will call in its "main" method.
func Configure(as *AppSetup, basePath string) (a *AppScope, err error) {
	migrateAndExit := false
	a = &AppScope{Setup: as}
	appScope = a
	log.Printf("Basepath is currently %q", basePath)
	if basePath == "" {

		if len(os.Args) == 1 {
			log.Fatal("No basepath file specified.")
		}

		appScope.BasePath = string(os.Args[1])
	} else {
		appScope.BasePath = basePath
	}

	if len(os.Args) == 3 {
		switch os.Args[2] {
		case "migrate":
			migrateAndExit = true
		default:
			log.Printf("Command line option %q not valid.", os.Args[2])
		}
	}

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

	if driver != "none" {

		connect, err := c.Get("database.connect")
		if err != nil {
			log.Fatal(err)
		}

		db, err := sql.Open(driver, connect)
		if err != nil {
			log.Fatal(err)
		}

		dBconfigFilename := appScope.BasePath + "/etc/dbversions.yaml"
		defaultSchema, allSchemas, err := model.ParseDbVersionsFile(dBconfigFilename)
		appScope.Db = &model.DbSetup{Db: db, DefaultSchema: defaultSchema, Schemas: allSchemas}
		switch driver {
		case "postgres":
			appScope.Db.GetQueries = postgres.GetQueries
		case "mysql":
			appScope.Db.GetQueries = mysql.GetQueries
		default:
			log.Fatal("Database driver not supported.")
		}

		if err == nil {

			for _, schema := range allSchemas {

				// Count the tables in the schema. If it's empty, make the dbversion 0.
				q := fmt.Sprintf(appScope.Db.GetQueries().TableCount(), schema.Name)
				r := db.QueryRow(q)
				var tc int64 = 0
				err = r.Scan(&tc)
				if err != nil {
					log.Fatal(err)
				}
				var dbversion int64 = 0

				if tc != 0 {
					query := fmt.Sprintf(appScope.Db.GetQueries().DbVersion(), schema.Name)
					row := db.QueryRow(query)
					err = row.Scan(&dbversion)
					if err != nil {
						log.Fatal(err)
					}
				}

				log.Printf("Schema: %v App: %v Db: %v", schema.Name, schema.Version, dbversion)
				if schema.Version != dbversion {

					if migrateAndExit {
						dbs := &model.DbSetup{Db: db}
						dbs.GetQueries = appScope.Db.GetQueries
						t := &model.Table{Db: dbs, Schema: schema.Name}
						log.Printf("Running database migration on %q", schema.Name)
						for i := dbversion + 1; i <= schema.Version; i++ {
							scriptfile := fmt.Sprintf("%v/sql/changes/%v_%v_%04d.sql", appScope.BasePath, driver, schema.Name, i)
							log.Printf("Running script %v", scriptfile)

							err = model.RunScript(db, scriptfile)
							if err != nil {
								log.Fatal(err)
							}
							dbv := &model.SawsijDbVersion{VersionId: i, RanOn: time.Now()}
							t.Insert(dbv)
							log.Printf("Inserted record: %+v", dbv)

						}

					} else {
						log.Fatal("Schema/App version mismatch. Please run migrate to update the database.")
					}

				}

				if migrateAndExit {
					viewfile := fmt.Sprintf("%v/sql/objects/%v_%v_views.sql", appScope.BasePath, driver, schema.Name)
					log.Printf("Running script %v", viewfile)
					err = model.RunScript(db, viewfile)
					if err != nil {
						log.Fatal(err)
					}

				}

			}

			if migrateAndExit {

				log.Print("All schemas updated. Exiting.")
				os.Exit(0)
			}
		}
	}

	key, err := c.Get("encryption.key")
	if err != nil {
		log.Fatal(err)
	}

	store = sessions.NewCookieStore([]byte(key))

	log.Print("Static dir is [" + appScope.BasePath + "/static" + "]")
	http.HandleFunc("/static/", staticHandler)

	parseTemplates()

	return
}

// Run will start a web server on the port specified in the config file, using the configuration in the config file and the routes specified by any Route() calls
// that have been previously made. This is generally the last line of your application's "main" method.
func Run() {

	log.Printf("Number of processors: %d", runtime.NumCPU())

	runtime.GOMAXPROCS(runtime.NumCPU())
	listen := ""
	listen, err := appScope.Config.Get("server.listen")
	if err != nil {
		port, err := appScope.Config.Get("server.port")
		if err != nil {
			log.Print(err)
			log.Fatal("Config file must specify 'listen' or 'port'.")
		} else {
			listen = fmt.Sprintf(":%v", port)
		}
	} else {

	}

	log.Printf("Listening on %v", listen)
	log.Fatal(http.ListenAndServe(listen, context.ClearHandler(http.DefaultServeMux)))
}
