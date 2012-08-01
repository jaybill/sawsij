// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sawsij provides a small, opinionated web framework.
package framework

import (
	"bitbucket.org/jaybill/sawsij/framework/model"
	"code.google.com/p/gorilla/sessions"

	"database/sql"
	"encoding/base64"
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
type AppSetup struct {
	GetUser func(username string, a *AppScope) User
	Roles   *map[string]int
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
	if len(templateFiles) > 0 {
		pt, err := template.New("dummy").Delims("<%", "%>").Funcs(GetFuncMap()).ParseFiles(templateFiles...)
		parsedTemplate = pt
		if err != nil {
			log.Print(err)
		}
	}
}

// HandlerResponse is a struct that your handler functions return. It contains all the data needed to generate the response. If Redirect is set,
// the contents of View is ignored. 
// Note: If you only supply one entry in your View map, the *contents* of the map will be passed to the view rather than the whole map. This is done 
// to simplify templates and JSON responses with only one entry.
type HandlerResponse struct {
	View     map[string](interface{})
	Redirect string
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
	// Setting this to framework.RT_JSON or framework.RT_HTML will force the return type and ignore any URL hints.
	ReturnType int
	// How parameters will be specified on the URL. Will default to PARAMS_MAP, a key value map. Can be set to PARAMS_ARRAY to return
	// an ordered array of values
	ParamsAs int
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
	templateId := GetTemplateName(rcfg.Pattern)

	var slashRoute string = ""
	if p := strings.LastIndex(rcfg.Pattern, "/"); p != len(rcfg.Pattern)-1 {
		slashRoute = rcfg.Pattern + "/"
		log.Printf("Specified %q, implying %q", rcfg.Pattern, slashRoute)
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
				dest := base64.URLEncoding.EncodeToString([]byte(rcfg.Pattern))
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
				switch returnType {
				case RT_XML:
					//TODO Return actual XML here (issue #6)
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
					log.Printf("Using template file %v", templateFilename)
					// Add "global" template variables
					if len(global) > 0 && handlerResults.View != nil {

						global["roles"] = *appScope.Setup.Roles
						global["url"] = rcfg.Pattern
						handlerResults.View["global"] = global

					}
					err = parsedTemplate.ExecuteTemplate(w, templateFilename, handlerResults.View)
					if err != nil {
						log.Print(err)
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
func Configure(as *AppSetup, basePath string) (err error) {
	migrateAndExit := false
	a := AppScope{Setup: as}
	appScope = &a
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

	if err == nil {

		for _, schema := range allSchemas {
			// TODO Remove hardcoded sql string, replace with driver based lookup (issue #11)
			query := fmt.Sprintf("SELECT version_id from %v.sawsij_db_version ORDER BY ran_on DESC LIMIT 1;", schema.Name)
			row := db.QueryRow(query)
			var dbversion int64 = 0

			err = row.Scan(&dbversion)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Schema: %v App: %v Db: %v", schema.Name, schema.Version, dbversion)
				if schema.Version != dbversion {

					if migrateAndExit {
						dbs := &model.DbSetup{Db: db}
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

						}
						viewfile := fmt.Sprintf("%v/sql/objects/%v_%v_views.sql", appScope.BasePath, driver, schema.Name)
						log.Printf("Running script %v", viewfile)
						err = model.RunScript(db, viewfile)
						if err != nil {
							log.Fatal(err)
						}

					} else {
						log.Fatal("Schema/App version mismatch. Please run migrate to update the database.")
					}

				}
			}

		}
		appScope.Db = &model.DbSetup{Db: db, DefaultSchema: defaultSchema, Schemas: allSchemas}
		if migrateAndExit {
			log.Print("All schemas updated. Exiting.")
			os.Exit(0)
		}

	} else {
		log.Fatal(err)
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

	port, err := appScope.Config.Get("server.port")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port [" + port + "]")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
