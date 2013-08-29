// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Provides a command line tool for creating a new sawsij application. Sets up app, copies files, creates tables.

Check out http://sawsij.com for more information and documentation.

*/
package main

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"
	"bitbucket.org/jaybill/sawsij/framework/model/mysql"
	"bitbucket.org/jaybill/sawsij/framework/model/postgres"
	"bitbucket.org/jaybill/sawsij/sawsijcmd/resources"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var env map[string]string
var goroot string = "/usr/local/go"
var gopath string = ""
var confdir string = ""
var envfile string = ""
var gobinpath string = ""
var sawsijhome string = ""

func main() {
	var err error

	envStrings := os.Environ()

	env = make(map[string]string)

	for _, envString := range envStrings {
		keyval := strings.Split(envString, "=")
		env[keyval[0]] = keyval[1]
	}

	confdir = env["HOME"] + "/.sawsij"
	envfile = confdir + "/sawsijenv"

	if env["SAWSIJ_SETUP"] != "1" {
		fmt.Println("Environment variables not set, running intial configuration...")
		envSetup()
	}

	if env["SAWSIJ_HOME"] == "" {
		sawsijhome = confdir
	} else {
		sawsijhome = env["SAWSIJ_HOME"]
	}

	if env["GOPATH"] != "" {
		gopath = env["GOPATH"]
	}

	fmt.Printf("GOPATH is %v \n", gopath)

	gobinpath, err = exec.LookPath("go")
	if err != nil {
		fmt.Println("Can't find go. Is it installed and in your path?")
	} else {

		version, err := exec.Command(gobinpath, "version").Output()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Printf("Using %s\n", gobinpath)
			fmt.Printf("%s\n", strings.TrimSpace(string(version)))
		}

	}

	if len(os.Args) == 1 {
		fmt.Println("No command specified.")
		os.Exit(1)
	}
	command := strings.TrimSpace(os.Args[1])

	switch command {
	case "new":
		new()
	case "factory":
		factory()
	default:
		fmt.Printf("Command %q not recognized.\n", command)
		os.Exit(1)
	}

}

func envSetup() {

	envt := `# This file contains sawsij environment variables for development
# Be sure to add the line "source $HOME/.sawsij/sawsijenv" to the end of ` + env["HOME"] + `/.profile
# so this will be read.

SAWSIJ_SETUP=1
export SAWSIJ_SETUP
`

	_, err := os.Open(confdir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(confdir, os.FileMode(0744))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Printf("Created new sawsij config dir %q\n", confdir)
			}
		}
	} else {
		fmt.Printf("Found config dir %q\n", confdir)
	}

	_, err = os.Open(envfile)
	if err != nil {
		if os.IsNotExist(err) {
			err = framework.WriteStringToFile(envt, envfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {

				fmt.Println("****************************\n** NOTICE! ACTION NEEDED! **\n****************************")
				fmt.Println("A new file containing sawsij environment variables was created.")
				fmt.Printf("You must add the line \"source $HOME/.sawsij/sawsijenv\" to the end of %v\n\n", env["HOME"]+"/.profile")

				os.Exit(0)
			}
		}
	} else {
		fmt.Printf("Found env file %q\n", envfile)
	}

}

func new() {
	var err error
	var itWorked bool = true
	var queries model.Queries

	config := make(map[string]string)
	name := ""
	fmt.Println("****************************\n** CREATE NEW APPLICATION **\n****************************")
	for name == "" {
		name, _ = framework.GetUserInput("Application name", name)
	}

	config["name"] = name
	appserver := name + "server"
	path, _ := os.Getwd()
	path = strings.TrimSpace(path) + "/" + name
	path, _ = framework.GetUserInput("Application path", path)

	config["port"], _ = framework.GetUserInput("Server Port", "8078")

	config["salt"] = framework.MakeRandomId()
	config["key"] = framework.MakeRandomId()

	doDb, _ := framework.GetUserInput("Configure database?", "Y")

	var dbes string = ""
	var dbed string = ""

	if doDb == "Y" {

		fmt.Println("****************************\n** DATABASE CONFIGURATION **\n****************************")

		config["driver"], _ = framework.GetUserInput("Database driver", "postgres")
		dbhost, _ := framework.GetUserInput("Database host", "localhost")
		var dp string
		if config["driver"] == "postgres" {
			dp = "5432"
		} else {
			dp = "3306"
		}
		dbport, _ := framework.GetUserInput("Database port", dp)
		dbname, _ := framework.GetUserInput("Database name", name)
		config["schema"], _ = framework.GetUserInput("Database schema", name)
		dbuser, _ := framework.GetUserInput("Database user", name)
		dbpass, _ := framework.GetUserInput("Database password", "")

		switch config["driver"] {
		case "postgres":
			queries = postgres.GetQueries()
			dbes = config["schema"]
			dbed = ""
		case "mysql":
			queries = mysql.GetQueries()
			dbed = dbname
			dbes = ""
		default:
			bomb(&framework.SawsijError{"Driver not supported"})
		}
		//user string, password string, host string, dbname string, port string
		strcon := queries.ConnString(dbuser, dbpass, dbhost, dbname, dbport)
		fmt.Printf("Connection string: %v\n", strcon)
		config["connect"] = strcon
		fmt.Println("****************************\n**   ADMIN ACCOUNT SETUP  **\n****************************")
		config["admin_email"], _ = framework.GetUserInput("Admin Email", name+"@"+name+".com")
		password := ""
		for password == "" {
			password, _ = framework.GetUserInput("Admin Password", password)
		}
		config["password_hash"] = framework.PasswordHash(password, config["salt"])
	} else {
		config["driver"] = "none"
		config["schema"] = ""
		config["connect"] = ""
	}

	fmt.Printf("Creating new sawsij app %q in location %v\n", name, path)

	tplDir := path + "/templates"
	binDir := path + "/bin"
	srcDir := path + "/src"
	etcDir := path + "/etc"
	pkgDir := path + "/pkg"
	sqlChgDir := path + "/sql/changes"
	sqlObjDir := path + "/sql/objects"
	staticDir := path + "/static"

	appdirs := []string{
		tplDir,
		tplDir + "/crud",
		srcDir,
		srcDir + "/" + appserver,
		srcDir + "/" + name,
		etcDir,
		staticDir + "/js",
		staticDir + "/img",
		staticDir + "/css",
		pkgDir,
		sqlChgDir,
		sqlObjDir}

	_, err = os.Stat(path)

	if !os.IsNotExist(err) {
		fmt.Printf("Target directory %v already exists.\n", path)
		os.Exit(1)
	}

	for _, adir := range appdirs {
		err = os.MkdirAll(adir, os.FileMode(0744))
		if err != nil {
			fmt.Println(err)

		} else {
			err = os.MkdirAll(binDir, os.FileMode(0755))
			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

		}
	}

	type TplDef struct {
		Name string
		Dest string
	}

	templateContent := resources.GetTemplateResources()
	staticContent := resources.GetStaticResources()
	var tpls []TplDef
	tpls = append(tpls, TplDef{"admin.html.tpl", path + "/templates/admin.html"})
	tpls = append(tpls, TplDef{"admin-users.html.tpl", path + "/templates/admin-users.html"})
	tpls = append(tpls, TplDef{"admin-users-delete.html.tpl", path + "/templates/admin-users-delete.html"})
	tpls = append(tpls, TplDef{"admin-users-edit.html.tpl", path + "/templates/admin-users-edit.html"})
	tpls = append(tpls, TplDef{"admin-footer.html.tpl", path + "/templates/admin-footer.html"})
	tpls = append(tpls, TplDef{"admin-header.html.tpl", path + "/templates/admin-header.html"})
	tpls = append(tpls, TplDef{"appserver.go.tpl", path + "/src/" + appserver + "/" + appserver + ".go"})
	tpls = append(tpls, TplDef{"config.yaml.tpl", path + "/etc/config.yaml"})
	tpls = append(tpls, TplDef{"dbversions.yaml.tpl", path + "/etc/dbversions.yaml"})
	tpls = append(tpls, TplDef{"constants.go.tpl", path + "/src/" + name + "/constants.go"})
	tpls = append(tpls, TplDef{"footer.html.tpl", path + "/templates/footer.html"})
	tpls = append(tpls, TplDef{"header.html.tpl", path + "/templates/header.html"})
	tpls = append(tpls, TplDef{"index.html.tpl", path + "/templates/index.html"})
	tpls = append(tpls, TplDef{"login.html.tpl", path + "/templates/login.html"})
	tpls = append(tpls, TplDef{"denied.html.tpl", path + "/templates/denied.html"})
	tpls = append(tpls, TplDef{"error.html.tpl", path + "/templates/error.html"})
	tpls = append(tpls, TplDef{"messages.html.tpl", path + "/templates/messages.html"})
	tpls = append(tpls, TplDef{"license.tpl", path + "/LICENSE"})
	tpls = append(tpls, TplDef{"user.go.tpl", path + "/src/" + name + "/user.go"})

	if doDb == "Y" {
		tpls = append(tpls, TplDef{config["driver"] + "_0001.sql.tpl", path + "/sql/changes/" + config["driver"] + "_" + config["schema"] + "_0001.sql"})
		tpls = append(tpls, TplDef{config["driver"] + "_views.sql.tpl", path + "/sql/objects/" + config["driver"] + "_" + config["schema"] + "_views.sql"})

	}
	var spls []TplDef
	spls = append(spls, TplDef{"admin-dashboard.js", path + "/static/js/admin-dashboard.js"})
	spls = append(spls, TplDef{"admin-delete.html.tpl", path + "/templates/crud/admin-delete.html.tpl"})
	spls = append(spls, TplDef{"admin-edit.html.tpl", path + "/templates/crud/admin-edit.html.tpl"})
	spls = append(spls, TplDef{"admin.css", path + "/static/css/admin.css"})
	spls = append(spls, TplDef{"admin.html.tpl", path + "/templates/crud/admin.html.tpl"})
	spls = append(spls, TplDef{"bootstrap-datepicker.min.js", path + "/static/js/bootstrap-datepicker.min.js"})
	spls = append(spls, TplDef{"datepicker.css", path + "/static/css/datepicker.css"})
	spls = append(spls, TplDef{"handler.go.tpl", path + "/templates/crud/handler.go.tpl"})
	spls = append(spls, TplDef{"sawsij.js", path + "/static/js/sawsij.js"})
	spls = append(spls, TplDef{"site.css", path + "/static/css/site.css"})

	if itWorked {
		for _, t := range tpls {

			tpla, err := base64.StdEncoding.DecodeString(templateContent[t.Name])
			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

			fmt.Printf("Parsing template %v\n", t.Name)
			err = framework.ParseTemplate(string(tpla), config, t.Dest)

			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

		}

		for _, s := range spls {
			spla, err := base64.StdEncoding.DecodeString(staticContent[s.Name])
			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

			err = framework.WriteStringToFile(string(spla), s.Dest)
			if err != nil {
				fmt.Println(err)
				itWorked = false
			}
		}

	}

	if itWorked && doDb == "Y" {

		db, err := sql.Open(config["driver"], config["connect"])
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

		// TODO Remove hardcoded sql string, replace with driver based lookup (issue #11)
		tcq := queries.DbEmpty(dbes, dbed)
		row := db.QueryRow(tcq)
		tcount := 0

		err = row.Scan(&tcount)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		} else {
			if tcount > 0 {
				fmt.Println("Specified database/schema is not empty.")
				itWorked = false
			}
		}

		if itWorked {
			dbscript := path + "/sql/changes/" + config["driver"] + "_" + config["schema"] + "_0001.sql"
			fmt.Printf("Running db script: %v\n", dbscript)
			bQuery, err := ioutil.ReadFile(dbscript)
			if err != nil {
				fmt.Println(err)
			} else {
				sQuery := string(bQuery)
				queries := strings.Split(sQuery, ";")

				for _, query := range queries {
					query = strings.TrimSpace(query)
					if query != "" {
						_, err = db.Exec(query)
						if err != nil {
							fmt.Println(err)
							itWorked = false
						}
					}
				}
			}
		}
	}

	if itWorked {

		ef := `
# start %v environment
if [ "$%v" == '' ]; then
	export %v=%v
	GOPATH=$GOPATH:$%v
	PATH=$PATH:$%v/bin
	export GOPATH
	export PATH
fi
# end %v environment
`
		ucname := strings.ToUpper(name)
		envname := "SJ_" + ucname
		ef = fmt.Sprintf(ef, name, envname, envname, path, envname, envname, name)

		err = framework.AppendStringToFile(ef, envfile)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

		// set environment variables so we can build the application.
		err = os.Setenv(envname, path)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
		err = os.Setenv("GOPATH", gopath+":"+path)
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}
		err = os.Setenv("PATH", env["PATH"]+":"+path+"/bin")
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

	}

	// "go install" new app

	cm := `
*************************
**  APPLICATION READY  **
*************************

Your application has been built and lives in %v

New environment variables have been added, so be 
sure to issue to following command before continuing:

source ~/.sawsij/sawsijenv

To start your app:

cd %v
%v .		

(Use CRTL + C to stop it.)

You can then point a browser at http://localhost:%v


`
	cm2 := `The admin panel is at http://localhost:%v/admin
Your username is "admin" and your password is the one you chose above.

`
	if itWorked {
		fmt.Printf("%v %v %v\n", gobinpath, "install", appserver)
		compileMessage, err := exec.Command(gobinpath, "install", appserver).CombinedOutput()

		if err != nil {
			fmt.Println(string(compileMessage))
			fmt.Println(err)
			itWorked = false
		} else {

			fmt.Printf(cm, path, path, appserver, config["port"])
			if doDb == "Y" {
				fmt.Printf(cm2, config["port"])
			}
		}
	}

	if !itWorked {
		err = os.RemoveAll(path)
		if err != nil {
			fmt.Println(err)
		}
		failMsg := `
************************
**    BUILD FAILED    **
************************

Your application could not be built. Please see above for detailed error messages.

`
		fmt.Println(failMsg)
		os.Exit(1)

	}

}

func factory() {
	var basePath string
	var tName string
	var sName string

	fmt.Println("****************************\n**     SAWSIJ FACTORY     **\n****************************")

	// get command line args
	if len(os.Args) == 5 {

		basePath = string(os.Args[2])
		sName = string(os.Args[3])
		tName = string(os.Args[4])

		fmt.Printf("Basedir: %v\n", basePath)
		fmt.Printf("Table: %v\n", tName)

	} else {
		fmt.Println("Usage: sawsijcmd factory [basedir] [schema] [table]")
		os.Exit(1)
	}

	// read config file

	configFilename := basePath + "/etc/config.yaml"

	fmt.Printf("Using config file [%v]\n", configFilename)

	c, err := yaml.ReadFile(configFilename)
	if err != nil {
		bomb(err)
	}

	// determine package name

	pName, err := c.Get("app.pkg")

	if err != nil {
		bomb(err)
	} else {
		fmt.Printf("Package name is [%v]\n", pName)
	}

	// set up database connection

	driver, err := c.Get("database.driver")

	if err != nil {
		bomb(err)
	}

	connect, err := c.Get("database.connect")
	if err != nil {
		bomb(err)
	}

	db, err := sql.Open(driver, connect)
	if err != nil {
		bomb(err)
	}

	var queries model.Queries
	dbName := ""

	switch driver {

	case "postgres":
		queries = postgres.GetQueries()

	case "mysql":
		queries = mysql.GetQueries()
		cm := queries.ParseConnect(connect)
		dbName = cm["dbname"]
		fmt.Printf("Dbname is %v\n", dbName)
	default:
		bomb(&framework.SawsijError{"Driver not supported"})
	}

	query := queries.DescribeTable(tName, sName, dbName)
	fmt.Printf("Table description query is %v\n", query)

	rows, err := db.Query(query)
	tV := make(map[string]interface{})

	type fieldDef struct {
		FName       string
		FType       string
		CanBeNull   bool
		IsPk        bool
		DisplayType string
	}
	var sA []fieldDef
	if err == nil {

		for rows.Next() {
			var colName string
			var dType string
			var canBeNull string

			err = rows.Scan(&colName, &dType, &canBeNull)
			if err != nil {
				bomb(err)
			}

			var sType string
			var sDType string
			switch dType {

			case "bigint":
				sType = "int64"
				sDType = "number"
			case "int":
				sType = "int64"
				sDType = "number"
				tV["importStrconv"] = true
			case "integer":
				sType = "int64"
				sDType = "number"
				tV["importStrconv"] = true
			case "timestamp without time zone":
				sDType = "timestamp"
				sType = "time.Time"
				tV["importTime"] = true
			case "timestamp with time zone":
				sDType = "timestamp"
				sType = "time.Time"
				tV["importTime"] = true
			case "datetime":
				sDType = "timestamp"
				sType = "time.Time"
				tV["importTime"] = true
			case "timestamp":
				sDType = "timestamp"
				sType = "time.Time"
				tV["importTime"] = true
			case "date":
				sDType = "date"
				sType = "time.Time"
				tV["importTime"] = true
			case "text":
				sDType = "text"
				sType = "string"
			case "character varying":
				sDType = "text"
				sType = "string"
			case "varchar":
				sDType = "text"
				sType = "string"
			default:
				sType = "string"
			}

			var sCbn bool = false
			if canBeNull == "YES" {
				sCbn = true
			}

			var isPk bool = false
			if model.MakeFieldName(colName) == "Id" {
				isPk = true
			}

			sA = append(sA, fieldDef{model.MakeFieldName(colName), sType, sCbn, isPk, sDType})

		}

		tV["typeName"] = model.MakeFieldName(tName)
		tV["typeVar"] = strings.ToLower(model.MakeFieldName(tName))
		tV["pName"] = pName
		tV["struct"] = sA

		type sTplDef struct {
			Source string
			Dest   string
		}

		var tpls []sTplDef

		fName := strings.ToLower(model.MakeFieldName(tName))

		tpls = append(tpls, sTplDef{"admin-delete.html.tpl", fmt.Sprintf("%v/templates/admin-%v-delete.html", basePath, fName)})
		tpls = append(tpls, sTplDef{"admin-edit.html.tpl", fmt.Sprintf("%v/templates/admin-%v-edit.html", basePath, fName)})

		tpls = append(tpls, sTplDef{"admin.html.tpl", fmt.Sprintf("%v/templates/admin-%v.html", basePath, fName)})
		hsf := fmt.Sprintf("%v/src/%v/%v.go", basePath, pName, fName)
		tpls = append(tpls, sTplDef{"handler.go.tpl", hsf})

		fcmd := exec.Command("gofmt", "-w", hsf)
		err := fcmd.Start()
		if err != nil {
			bomb(err)
		} else {
			fmt.Printf("Formatted %v\n", hsf)
		}

		for _, tpl := range tpls {
			t, err := framework.ReadFileIntoString(basePath + "/templates/crud/" + tpl.Source)

			if err != nil {
				bomb(err)
			}

			err = framework.ParseTemplate(t, tV, tpl.Dest)
			fmt.Printf("%v\n", tpl.Dest)

			if err != nil {
				bomb(err)
			}
		}

		fmt.Printf("Generated handler and templates for %v.%v\n", sName, tName)

	} else {
		bomb(err)
	}

}

func bomb(err error) {
	fmt.Println(err)
	os.Exit(1)
}
