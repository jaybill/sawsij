package main

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"bitbucket.org/jaybill/sawsij/framework/model"
	"database/sql"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
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

				// Download seed file.
				seedUrl := "https://bitbucket.org/jaybill/sawsij/downloads/seed.zip"

				fmt.Println("Attempting to download seed...")

				zipfile := confdir + "/seed.zip"
				err = framework.CopyUrlToFile(seedUrl, zipfile)
				if err != nil {
					fmt.Println(err)
					err = os.RemoveAll(confdir)
					if err != nil {
						fmt.Println(err)
					}
					os.Exit(1)
				}

				err = framework.UnzipFileToPath(zipfile, confdir+"/seed")
				if err != nil {
					fmt.Printf("Can't unzip file: %q\n", err.Error())
					/*
						err = os.RemoveAll(confdir)
						if err != nil {
							fmt.Println(err)
						}*/
					os.Exit(1)
				}

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

	doDb, _ := framework.GetUserInput("Configure database?", "Y")

	if doDb == "Y" {

		fmt.Println("****************************\n** DATABASE CONFIGURATION **\n****************************")

		config["driver"], _ = framework.GetUserInput("Database driver", "postgres")
		dbname, _ := framework.GetUserInput("Database name", name)
		config["schema"], _ = framework.GetUserInput("Database schema", name)
		dbuser, _ := framework.GetUserInput("Database user", name)
		dbpass, _ := framework.GetUserInput("Database password", "")
		dbssl, _ := framework.GetUserInput("Database SSL Mode", "disable")
		config["connect"] = fmt.Sprintf("user=%v password=%v dbname=%v sslmode=%v", dbuser, dbpass, dbname, dbssl)
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

	config["salt"] = framework.MakeRandomId()
	config["key"] = framework.MakeRandomId()

	fmt.Printf("Creating new sawsij app %q in location %v\n", name, path)

	seeddir := sawsijhome + "/seed"

	fmt.Printf("seed dir: %v\n", seeddir)

	tplDir := path + "/templates"
	binDir := path + "/bin"
	srcDir := path + "/src"
	etcDir := path + "/etc"
	pkgDir := path + "/pkg"
	sqlChgDir := path + "/sql/changes"
	sqlObjDir := path + "/sql/objects"

	appdirs := []string{
		tplDir,
		srcDir,
		srcDir + "/" + appserver,
		srcDir + "/" + name,
		etcDir,
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

	if itWorked {
		for _, t := range tpls {
			fmt.Printf("Parsing template %v\n", t.Name)
			pt, err := template.New(t.Name).ParseFiles(seeddir + "/templates/" + t.Name)

			if err != nil {
				fmt.Println(err)
				itWorked = false
			}

			f, err := os.Create(t.Dest)
			if err != nil {
				fmt.Println(err)
				itWorked = false
			} else {
				defer f.Close()

				err = pt.ExecuteTemplate(f, t.Name, config)
				if err != nil {
					fmt.Println(err)
					itWorked = false
				}
			}

		}

		err = framework.CopyDir(seeddir+"/static", path+"/static")
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

		err = framework.CopyDir(seeddir+"/crud", path+"/templates/crud")
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

	}

	if itWorked && doDb == "Y" {

		db, err := sql.Open(config["driver"], config["connect"])
		if err != nil {
			fmt.Println(err)
			itWorked = false
		}

		// TODO Remove hardcoded sql string, replace with driver based lookup (issue #11)
		tcq := "SELECT count(*) as tables FROM information_schema.tables WHERE table_schema = $1;"
		row := db.QueryRow(tcq, config["schema"])
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

	query := "select column_name,data_type from information_schema.columns where table_name = $1 and table_schema = $2 order by ordinal_position desc;"

	rows, err := db.Query(query, tName, sName)
	tV := make(map[string]interface{})

	type fieldDef struct {
		FName string
		FType string
	}
	var sA []fieldDef
	if err == nil {

		for rows.Next() {
			var colName string
			var dType string

			err = rows.Scan(&colName, &dType)
			if err != nil {
				bomb(err)
			}

			var sType string

			switch dType {

			case "bigint":
				sType = "int64"
			case "integer":
				sType = "int64"
			case "timestamp without time zone":
				sType = "time.Time"
				tV["importTime"] = true
			case "text":
				sType = "string"
			case "character varying":
				sType = "string"
			default:
				sType = "string"
			}

			sA = append(sA, fieldDef{model.MakeFieldName(colName), sType})

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
		tpls = append(tpls, sTplDef{"handler.go.tpl", fmt.Sprintf("%v/src/%v/%v.go", basePath, pName, fName)})

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
